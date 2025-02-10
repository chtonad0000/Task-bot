package filter

import (
	taskpb "github.com/Task-bot/scheduler-service/internal/generated/task_service"
	"time"
)

func ValidTaskFilter(tasks []*taskpb.TaskResponse) []*taskpb.TaskResponse {
	now := time.Now()
	result := make([]*taskpb.TaskResponse, 0)
	for _, task := range tasks {
		if task.Deadline.AsTime().After(now) && task.Progress != 100 {
			result = append(result, task)
		}
	}
	return result
}
