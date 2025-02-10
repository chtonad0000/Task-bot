package services

import (
	taskpb "github.com/Task-bot/bot-service/internal/generated/task"
)

type TaskServiceClient struct {
	Client taskpb.TaskServiceClient
}

func NewTaskServiceClient(c taskpb.TaskServiceClient) *TaskServiceClient {
	return &TaskServiceClient{Client: c}
}

func (t *TaskServiceClient) GetName() string {
	return "task"
}
