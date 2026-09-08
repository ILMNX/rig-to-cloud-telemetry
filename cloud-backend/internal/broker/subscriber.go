package broker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"

	"rigtelemetry/cloud/internal/hub"
	"rigtelemetry/cloud/internal/store"
	"rigtelemetry/shared/model"
	"rigtelemetry/shared/topic"
)

// Subscriber consumes MQTT telemetry and persists + fans out points.
type Subscriber struct {
	client mqtt.Client
	store  *store.Store
	hub    *hub.Hub
	qos    byte
}

// NewSubscriber connects to MQTT and prepares a subscriber.
func NewSubscriber(brokerURL, clientID string, qos byte, st *store.Store, h *hub.Hub) (*Subscriber, error) {
	s := &Subscriber{store: st, hub: h, qos: qos}

	opts := mqtt.NewClientOptions().
		AddBroker(brokerURL).
		SetClientID(clientID).
		SetAutoReconnect(true).
		SetConnectRetry(true).
		SetConnectRetryInterval(2 * time.Second).
		SetKeepAlive(30 * time.Second).
		SetOrderMatters(false)

	opts.OnConnect = func(c mqtt.Client) {
		filter := topic.PointsWildcard()
		token := c.Subscribe(filter, qos, s.handleMessage)
		token.Wait()
		if err := token.Error(); err != nil {
			log.Printf("broker: subscribe %s failed: %v", filter, err)
			return
		}
		log.Printf("broker: subscribed to %s", filter)
	}

	client := mqtt.NewClient(opts)
	token := client.Connect()
	if !token.WaitTimeout(10 * time.Second) {
		return nil, fmt.Errorf("mqtt connect timeout to %s", brokerURL)
	}
	if err := token.Error(); err != nil {
		return nil, fmt.Errorf("mqtt connect: %w", err)
	}
	s.client = client
	log.Printf("broker: connected to %s", brokerURL)
	return s, nil
}

func (s *Subscriber) handleMessage(_ mqtt.Client, msg mqtt.Message) {
	var point model.TelemetryPoint
	if err := json.Unmarshal(msg.Payload(), &point); err != nil {
		log.Printf("broker: bad json on %s: %v", msg.Topic(), err)
		return
	}
	if wellID, ok := topic.ParseWellID(msg.Topic()); ok && point.WellID == "" {
		point.WellID = wellID
	}
	if err := point.Validate(); err != nil {
		log.Printf("broker: invalid point: %v", err)
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := s.store.Insert(ctx, point); err != nil {
		log.Printf("broker: insert failed: %v", err)
		return
	}
	s.hub.Publish(point)
}

// Connected reports MQTT connection state.
func (s *Subscriber) Connected() bool {
	return s.client != nil && s.client.IsConnected()
}

// Close disconnects the client.
func (s *Subscriber) Close() {
	if s.client != nil && s.client.IsConnected() {
		s.client.Disconnect(250)
	}
}
