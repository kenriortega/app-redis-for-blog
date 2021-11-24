package cmd

import (
	"app/examples/minibroker/domain"
	"app/pkg/mq"
	"context"
	"log"

	"github.com/go-redis/redis/v8"
)

// SendData method receive data from mqtt like
// mqtt subscribe -h mqtt.hsl.fi -p 8883 -l mqtts -v -t "/hfp/v2/journey/+/vp/bus/#"
// and then send data to redis stream
func SendData(ctx context.Context, payload <-chan []byte, client *redis.Client) {

	for msg := range payload {

		// Receive the content of the MQTT message and de-serialize bytes into
		// struct
		e := &domain.EventHolder{}
		err := mq.DeserializeMQTTBody(msg, e)

		if err != nil {
			log.Println(err)
			continue
		}

		pipe := client.TxPipeline()
		// // 2. XADD the full event body to a stream of events, these
		value, err := e.VP.ToMAP()
		if err != nil {
			log.Fatal(err)
		}
		pipe.XAdd(
			ctx, &redis.XAddArgs{
				Stream: "events:vp:bus",
				Values: value,
			},
		)
		// Execute Pipe!
		_, err = pipe.Exec(ctx)
		// Failed to Write an Event
		if err != nil {
			log.Fatal(err)
		}
	}
}
