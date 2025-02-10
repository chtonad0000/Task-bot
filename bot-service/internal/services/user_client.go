package services

import (
	userpb "github.com/Task-bot/bot-service/internal/generated/user"
)

type UserServiceClient struct {
	Client userpb.UserServiceClient
}

func NewUserServiceClient(c userpb.UserServiceClient) *UserServiceClient {
	return &UserServiceClient{Client: c}
}

func (t *UserServiceClient) GetName() string {
	return "user"
}
