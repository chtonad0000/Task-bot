package services

import (
	notificationpb "github.com/Task-bot/bot-service/internal/generated/notification"
)

type NotificationServiceClient struct {
	Client notificationpb.NotificationServiceClient
}

func NewNotificationServiceClient(c notificationpb.NotificationServiceClient) *NotificationServiceClient {
	return &NotificationServiceClient{Client: c}
}

func (t *NotificationServiceClient) GetName() string {
	return "notification"
}
