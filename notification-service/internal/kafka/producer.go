package kafka

import (
	"encoding/json"
	"fmt"
	"github.com/IBM/sarama"
	"log"
)

type Producer struct {
	producer sarama.SyncProducer
	topic    string
}

type NotificationMessage struct {
	UserId  int64  `json:"user_id"`
	Message string `json:"message"`
}

// NewKafkaProducer создает Kafka-продюсер
func NewKafkaProducer(brokerURL, topic string) (*Producer, error) {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true

	producer, err := sarama.NewSyncProducer([]string{brokerURL}, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create producer: %v", err)
	}

	return &Producer{
		producer: producer,
		topic:    topic,
	}, nil
}

func (p *Producer) SendNotification(userId int64, message string) error {
	notification := NotificationMessage{
		UserId:  userId,
		Message: message,
	}

	messageBytes, err := json.Marshal(notification)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %v", err)
	}

	msg := &sarama.ProducerMessage{
		Topic: p.topic,
		Value: sarama.ByteEncoder(messageBytes),
	}

	_, _, err = p.producer.SendMessage(msg)
	if err != nil {
		return fmt.Errorf("failed to produce message: %v", err)
	}

	log.Printf("Message delivered to topic %v\n", p.topic)
	return nil
}

func (p *Producer) Close() {
	if err := p.producer.Close(); err != nil {
		log.Printf("Error closing producer: %v", err)
	}
}
