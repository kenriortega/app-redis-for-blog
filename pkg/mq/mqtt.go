package mq

import (
	"app/examples/minibroker/domain"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var (
	mqttTopic      = os.Getenv("MQTT_TOPIC")  // "/hfp/v2/journey/+/+/#"
	mqttBrokerHost = os.Getenv("MQTT_BROKER") // "mqtt.hsl.fi"
	mqttPort       = os.Getenv("MQTT_PORT")   // "8883"
)

// MsgBroker ...
// https://medium.com/swlh/golang-tips-why-pointers-to-slices-are-useful-and-how-ignoring-them-can-lead-to-tricky-bugs-cac90f72e77b
type MsgBroker struct {
	StagingC chan []byte
}

// NewMsgBroker ...
func NewMsgBroker(n int) *MsgBroker {
	return &MsgBroker{
		StagingC: make(chan []byte, n),
	}
}

// messageHandler implements mqtt.PublishHandler/mqtt.MessageHandler,function passes
// all messages along to a single staging channel
//
// NOTE: mqtt.messageHandler must be safe for concurrent use by multiple goroutines;
// and should not be blocking (or call blocking code).
//
// WARNING: For this application specifically, chose to sacrifice the delivered at least
// once property for expediency, set very short 10ms timeout  so don't launch new goroutine
// or block for each message...
func (mb *MsgBroker) messageHandler(client mqtt.Client, msg mqtt.Message) {

	select {
	case mb.StagingC <- msg.Payload(): // Push to staging Channel...
		log.Println("Msg Recv")

	default: // Channel blocked && drop message..
		log.Println("Msg Recv Timeout")
	}

}

// connectHandler implements mqtt.OnConnectHandler, handler logs new connections
// to MQTT
func connectHandler(client mqtt.Client) {

	// NOTE: For each topic, begin listening on a separate goroutine; set the
	// subscriptions onn connection s.t if the client is disconnected, resumes
	// previous subscriptions on reconnect...
	go func(topic string) {
		token := client.Subscribe(topic, 1, nil)
		token.Wait()
	}(mqttTopic)

	log.Println("Subscribed to New Topic")
}

// connectLostHandler implements mqtt.ConnectionLostHandler
func connectionLostHandler(client mqtt.Client, err error) {
	log.Printf("Connect lost: %v", err)
}

// InitMQTTClient - Initializes the MQTT Client w. a fixed set of behavior
// for onConnect, onRecv, and onDisconnect
func InitMQTTClient(StgC *MsgBroker) *mqtt.Client {

	// Initialize default options; instantiates a new *mqtt.ClientOptions
	opts := mqtt.NewClientOptions()

	// Add options to `mqtt.ClientOptions`, per suggestion in lib docs, use
	// setters rather than setting values in opts directly
	opts.AddBroker(
		fmt.Sprintf("mqtts://%s:%s", mqttBrokerHost, mqttPort),
	)

	opts.SetOrderMatters(false)
	opts.SetDefaultPublishHandler(StgC.messageHandler)
	opts.SetOnConnectHandler(connectHandler)
	opts.SetConnectionLostHandler(connectionLostHandler)

	// Create Client
	client := mqtt.NewClient(opts)

	log.Println("New MQTT Client Connection")

	// Open new connection w. Client
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		// Critical Error - Cannot Connect to MQTT, most likely the mqtt server is down....
		log.Panic(token.Error())
	}

	return &client
}

// DeserializeMQTTBody ...
func DeserializeMQTTBody(msgb []byte, hold *domain.EventHolder) error {

	if err := json.Unmarshal(msgb, &hold); err != nil {
		return err
	}

	if lat, lng := hold.VP.Lat, hold.VP.Lng; lat == 0.0 || lng == 0.0 {
		return errors.New("custom error; Missing coords")
	}

	return nil
}
