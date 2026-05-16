package mqtt

import (
	"encoding/json"
	"fmt"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type Publisher interface {
	Publish(topic string, payload interface{}) error
}

type publisher struct {
	client mqtt.Client
}

func NewPublisher(broker string, clientID string) (Publisher, error) {
	opts := mqtt.NewClientOptions().AddBroker(broker)
	opts.SetClientID(clientID)
	opts.SetCleanSession(true)
	opts.SetAutoReconnect(true)

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return nil, token.Error()
	}

	return &publisher{client: client}, nil
}

func (p *publisher) Publish(topic string, payload interface{}) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	token := p.client.Publish(topic, 1, false, data)
	// เราจะไม่รอแบบ Blocking นานเกินไป (ใช้ Timeout 5 วินาที)
	go func() {
		if token.WaitTimeout(5 * time.Second) && token.Error() != nil {
			fmt.Printf("MQTT Publish Error to %s: %v\n", topic, token.Error())
		}
	}()

	return nil
}
