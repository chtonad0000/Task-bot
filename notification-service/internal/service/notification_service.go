package service

import (
	"context"
	pb "github.com/Task-bot/notification-service/internal/generated/notification"
	"time"
)

type Notification struct {
	ID         int64
	UserID     int64
	Message    string
	NotifyTime time.Time
}

type NotificationStorage interface {
	CreateNotification(ctx context.Context, userID int64, message string, notifyTime time.Time) (*Notification, error)
	GetDueNotifications(ctx context.Context, now time.Time) ([]Notification, error)
	DeleteNotification(ctx context.Context, notificationID int64) error
}

type NotificationService struct {
	storage NotificationStorage
	pb.UnimplementedNotificationServiceServer
}

func NewNotificationService(storage NotificationStorage) *NotificationService {
	return &NotificationService{storage: storage}
}

func (n *NotificationService) CreateNotification(ctx context.Context, req *pb.Notification) (*pb.CreateNotificationResponse, error) {
	_, err := n.storage.CreateNotification(ctx, req.UserId, req.Message, req.NotifyTime.AsTime())
	if err != nil {
		return nil, err
	}
	return &pb.CreateNotificationResponse{
		Success: true,
	}, nil
}

func (n *NotificationService) GetDueNotifications(ctx context.Context) ([]*Notification, error) {
	now := time.Now()
	notifications, err := n.storage.GetDueNotifications(ctx, now)
	if err != nil {
		return nil, err
	}

	var notificationResponses []*Notification
	for _, notification := range notifications {
		notificationResponses = append(notificationResponses, &Notification{
			ID:         notification.ID,
			UserID:     notification.UserID,
			Message:    notification.Message,
			NotifyTime: notification.NotifyTime,
		})
	}

	return notificationResponses, nil
}

func (n *NotificationService) DeleteNotification(ctx context.Context, notificationID int64) error {
	err := n.storage.DeleteNotification(ctx, notificationID)
	if err != nil {
		return err
	}
	return nil
}
