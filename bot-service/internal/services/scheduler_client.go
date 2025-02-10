package services

import (
	schedulerpb "github.com/Task-bot/bot-service/internal/generated/scheduler"
)

type SchedulerServiceClient struct {
	Client schedulerpb.SchedulerClient
}

func NewSchedulerServiceClient(c schedulerpb.SchedulerClient) *SchedulerServiceClient {
	return &SchedulerServiceClient{Client: c}
}

func (s *SchedulerServiceClient) GetName() string {
	return "scheduler"
}
