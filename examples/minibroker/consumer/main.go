package main

import (
	"app/pkg/db"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"runtime"
	"time"

	"github.com/go-redis/redis/v8"
)

var (
	groupName = "vp"
	// seededRand random number
	// #nosec
	seededRand *rand.Rand = rand.New(
		rand.NewSource(time.Now().UnixNano()))
)

const (
	// CHARSET of characters
	CHARSET = "abcdefghijklmnopqrstuvwxyz" +
		"ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789$*-/+"
	// SALT for define the size of the buffer
	SALT = 10
)

// StringWithCharset generate randoms words
func StringWithCharset() string {
	b := make([]byte, SALT)
	for i := range b {
		b[i] = CHARSET[seededRand.Intn(len(CHARSET))]
	}
	return string(b)
}
func init() {
	numcpu := runtime.NumCPU()
	runtime.GOMAXPROCS(numcpu) // Try to use all available CPUs.
	flag.StringVar(&groupName, "group", groupName, "name for consume group")
	flag.Parse()
}

func main() {
	ctx := context.Background()
	rdb := db.GetRedisDbClient(context.Background())
	streams := []string{"events:vp:bus"}
	var ids []string
	if groupName == "" {
		groupName = "consumer-" + StringWithCharset()
	}
	for _, v := range streams {
		ids = append(ids, ">")
		err := rdb.XGroupCreate(ctx, v, groupName, "0").Err()
		if err != nil {
			log.Println(err)
		}

	}

	streams = append(streams, ids...) // for each stream it requires an '>' :{"events:vp:bus", ">"}
	fmt.Printf("Consumer gruop with name: [%s]\n", groupName)
	for {
		entries, err := rdb.XReadGroup(ctx, &redis.XReadGroupArgs{
			Group:    groupName,
			Consumer: fmt.Sprintf("%d", time.Now().UnixNano()),
			Streams:  streams,
			Count:    2,
			Block:    0,
			NoAck:    false,
		}).Result()
		if err != nil {
			log.Fatal(err)
		}

		for _, stream := range entries {
			ReceiveMSG(ctx, stream, rdb, groupName)
		}

	}
}

func ReceiveMSG(ctx context.Context, stream redis.XStream, rdb *redis.Client, groupName string) {
	for i := 0; i < len(stream.Messages); i++ {
		messageID := stream.Messages[i].ID
		values := stream.Messages[i].Values
		bytes, err := json.Marshal(values)
		if err != nil {
			log.Fatal(err)
		}

		rdb.XAck(
			ctx,
			stream.Stream,
			groupName,
			messageID,
		)

		fmt.Printf("ConsumerGroup: [%s] data : %s\n", groupName, string(bytes))
	}
}
