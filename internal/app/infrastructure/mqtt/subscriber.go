package mqtt

import (
	"github.com/Bangnus/Bidkan-backend/internal/app/domain/entity"
	"github.com/Bangnus/Bidkan-backend/internal/app/usecase/bike/tracking"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type Subscriber struct {
	client  mqtt.Client
	service tracking.Service
}

func NewSubscriber(broker string, clientID string, service tracking.Service) *Subscriber {
	opts := mqtt.NewClientOptions().AddBroker(broker)
	opts.SetClientID(clientID)

	client := mqtt.NewClient(opts)
	return &Subscriber{
		client:  client,
		service: service,
	}
}

func (s *Subscriber) Start() error {
	if token := s.client.Connect(); token.Wait() && token.Error() != nil {
		return token.Error()
	}

	// ฟังก์ชันจัดการเมื่อได้รับข้อความ
	handler := func(client mqtt.Client, msg mqtt.Message) {
		// ดึง BikeID จาก Topic: bidkan/bikes/{id}/gps
		parts := strings.Split(msg.Topic(), "/")
		if len(parts) < 3 {
			return
		}
		bikeID := parts[2]

		var data entity.BikeData
		if err := json.Unmarshal(msg.Payload(), &data); err != nil {
			fmt.Printf("Error parsing MQTT JSON: %v\n", err)
			return
		}
		data.BikeID = bikeID

		// ส่งให้ Service ประมวลผล (ใช้ Goroutine เพื่อไม่ให้บล็อกการรับข้อความถัดไป)
		go s.service.ProcessTracking(context.Background(), data)
	}

	// Subscribe
	topic := "bidkan/bikes/+/gps"
	if token := s.client.Subscribe(topic, 1, handler); token.Wait() && token.Error() != nil {
		return token.Error()
	}

	fmt.Printf("🚀 MQTT Subscriber started on topic: %s\n", topic)
	return nil
}
