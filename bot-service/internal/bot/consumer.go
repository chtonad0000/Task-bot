package bot

import (
	"encoding/json"
	"fmt"
	"github.com/IBM/sarama"
	tgbot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
)

type Consumer struct {
	consumer sarama.Consumer
	bot      *tgbot.BotAPI
	topic    string
}

type NotificationMessage struct {
	UserID  int64  `json:"user_id"`
	Message string `json:"message"`
}

func NewKafkaConsumer(brokerURL, topic string, bot *tgbot.BotAPI) (*Consumer, error) {
	consumer, err := sarama.NewConsumer([]string{brokerURL}, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create consumer: %v", err)
	}

	return &Consumer{
		consumer: consumer,
		bot:      bot,
		topic:    topic,
	}, nil
}

func (c *Consumer) Start() {
	partitions, err := c.consumer.Partitions(c.topic)
	if err != nil {
		log.Fatalf("Failed to get partitions: %v", err)
	}

	for _, partition := range partitions {
		pc, err := c.consumer.ConsumePartition(c.topic, partition, sarama.OffsetNewest)
		if err != nil {
			log.Fatalf("Failed to start consumer for partition %d: %v", partition, err)
		}
		defer pc.Close()

		for msg := range pc.Messages() {
			var notification NotificationMessage
			if err := json.Unmarshal(msg.Value, &notification); err != nil {
				log.Printf("Failed to unmarshal message: %v", err)
				continue
			}

			if err != nil {
				log.Printf("Failed to decrypt Telegram ID: %v", err)
				continue
			}

			err = SendTelegramNotification(c.bot, notification.UserID, notification.Message)
			if err != nil {
				log.Printf("Failed to send notification: %v", err)
			}
		}
	}
}

func (c *Consumer) Close() {
	if err := c.consumer.Close(); err != nil {
		log.Printf("Error closing consumer: %v", err)
	}
}
