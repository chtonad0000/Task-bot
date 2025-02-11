package worker

import (
	"context"
	"log"
	"time"

	"github.com/Task-bot/notification-service/internal/kafka"
	"github.com/Task-bot/notification-service/internal/service"
)

type NotificationWorker struct {
	storage       service.NotificationStorage
	kafkaProducer *kafka.Producer
}

func NewNotificationWorker(storage service.NotificationStorage, kafkaProducer *kafka.Producer) *NotificationWorker {
	return &NotificationWorker{
		storage:       storage,
		kafkaProducer: kafkaProducer,
	}
}

func (w *NotificationWorker) Start(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			w.checkAndSendNotifications(ctx)
		case <-ctx.Done():
			log.Println("Notification worker stopped")
			return
		}
	}
}

func (w *NotificationWorker) checkAndSendNotifications(ctx context.Context) {
	now := time.Now().UTC()
	notifications, err := w.storage.GetDueNotifications(ctx, now)
	if err != nil {
		log.Printf("Error getting due notifications: %v", err)
		return
	}
	if len(notifications) == 0 {
		return
	}
	for _, notification := range notifications {
		err = w.kafkaProducer.SendNotification(notification.UserID, notification.Message)
		if err != nil {
			log.Printf("Error sending notification to Kafka: %v", err)
		}
		err = w.storage.DeleteNotification(ctx, notification.ID)
		if err != nil {
			log.Printf("Error deleting notification from DB: %v", err)
		} else {
			log.Printf("Notification sent to Kafka and deleted from DB: %v", notification.Message)
		}
	}
}
