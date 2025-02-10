package algorithm

import (
	"sort"
	"time"

	taskpb "github.com/Task-bot/scheduler-service/internal/generated/task_service"
)

func CalculateOptimalOrder(tasks []*taskpb.TaskResponse) []*taskpb.TaskResponse {
	Tmax := 48.0
	now := time.Now()

	type scoredTask struct {
		task  *taskpb.TaskResponse
		score float64
	}

	var scored []scoredTask

	for _, t := range tasks {

		var deadlineScore float64
		delta := t.Deadline.AsTime().Sub(now).Hours()
		if delta <= 0 {
			deadlineScore = 1.0
		} else {
			normalized := delta / Tmax
			if normalized > 1 {
				normalized = 1
			}
			deadlineScore = 1 - normalized
		}

		priorityScore := float64(t.Priority-1) / 4.0

		progressScore := 1 - float64(t.Progress)/100.0

		urgency := 0.5*deadlineScore + 0.3*priorityScore + 0.2*progressScore

		scored = append(scored, scoredTask{
			task:  t,
			score: urgency,
		})
	}

	sort.SliceStable(scored, func(i, j int) bool {
		return scored[i].score > scored[j].score
	})

	sortedTasks := make([]*taskpb.TaskResponse, len(scored))
	for i, st := range scored {
		sortedTasks[i] = st.task
	}

	return sortedTasks
}
