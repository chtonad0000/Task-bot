package service

import (
	"context"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"time"

	"github.com/Task-bot/scheduler-service/internal/algorithm"
	"github.com/Task-bot/scheduler-service/internal/filter"
	pb "github.com/Task-bot/scheduler-service/internal/generated/scheduler_service"
	taskpb "github.com/Task-bot/scheduler-service/internal/generated/task_service"
	"google.golang.org/grpc"
)

type SchedulerService struct {
	taskClient taskpb.TaskServiceClient
	pb.UnimplementedSchedulerServer
}

func NewSchedulerService(ctx context.Context, taskServiceAddr string) (*SchedulerService, error) {
	dialCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(dialCtx, taskServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		return nil, err
	}

	taskClient := taskpb.NewTaskServiceClient(conn)
	return &SchedulerService{taskClient: taskClient}, nil
}

func (s *SchedulerService) CalculateOptimalPlan(ctx context.Context, req *pb.CalculatePlanRequest) (*pb.CalculatePlanResponse, error) {
	log.Printf("Calculating optimal plan for user: %s", req.UserId)

	tasksResp, err := s.taskClient.GetTasks(ctx, &taskpb.GetTaskRequest{UserId: req.UserId})
	if err != nil {
		return nil, err
	}

	validTasks := filter.ValidTaskFilter(tasksResp.Tasks)

	sortedTasks := algorithm.CalculateOptimalOrder(validTasks)

	taskResponses := convertToProto(sortedTasks)

	return &pb.CalculatePlanResponse{Tasks: taskResponses}, nil
}

func convertToProto(tasks []*taskpb.TaskResponse) []*pb.TaskInfo {
	result := make([]*pb.TaskInfo, len(tasks))
	for i, task := range tasks {
		result[i] = &pb.TaskInfo{
			TaskId:   task.TaskId,
			TaskText: task.TaskText,
			Priority: task.Priority,
			Deadline: task.Deadline,
			Progress: task.Progress,
		}
	}
	return result
}
