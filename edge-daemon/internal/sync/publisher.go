package sync

import (
	"fmt"
	"log"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"

	"rigtelemetry/shared/topic"
)

// Publisher publishes JSON telemetry payloads to MQTT.
type Publisher struct {
	client mqtt.Client
	qos    byte
}

// NewPublisher connects to the MQTT broker.
func NewPublisher(broker, clientID string, qos byte) (*Publisher, error) {
	opts := mqtt.NewClientOptions().
		AddBroker(broker).
		SetClientID(clientID).
		SetAutoReconnect(true).
		SetConnectRetry(true).
		SetConnectRetryInterval(2 * time.Second).
		SetKeepAlive(30 * time.Second).
		SetOrderMatters(false)

	client := mqtt.NewClient(opts)
	token := client.Connect()
	if !token.WaitTimeout(10 * time.Second) {
		return nil, fmt.Errorf("mqtt connect timeout to %s", broker)
	}
	if err := token.Error(); err != nil {
		return nil, fmt.Errorf("mqtt connect: %w", err)
	}
	log.Printf("sync: connected to MQTT broker %s", broker)
	return &Publisher{client: client, qos: qos}, nil
}

// PublishPoint publishes payload to telemetry/{wellID}/points.
func (p *Publisher) PublishPoint(wellID string, payload []byte) error {
	t := topic.Points(wellID)
	token := p.client.Publish(t, p.qos, false, payload)
	if !token.WaitTimeout(5 * time.Second) {
		return fmt.Errorf("mqtt publish timeout topic=%s", t)
	}
	return token.Error()
}

// Close disconnects the MQTT client.
func (p *Publisher) Close() {
	if p.client != nil && p.client.IsConnected() {
		p.client.Disconnect(250)
	}
}
