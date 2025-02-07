package service

import (
	"context"
	pb "github.com/Task-bot/task-service/internal/generated/task"
	"google.golang.org/protobuf/types/known/timestamppb"
	"time"
)

type Task struct {
	ID       int64
	UserID   int64
	Text     string
	Priority int32
	Deadline time.Time
	Progress int32
}

type TaskStorage interface {
	CreateTask(ctx context.Context, userID int64, text string, priority int32, deadline time.Time) (*Task, error)
	GetTasks(ctx context.Context, userID int64) ([]Task, error)
	DeleteTask(ctx context.Context, taskID int64) error
	UpdateTaskStatus(ctx context.Context, taskID int64, progress int32) error
}

type TaskService struct {
	storage TaskStorage
	pb.UnimplementedTaskServiceServer
}

func NewTaskService(storage TaskStorage) *TaskService {
	return &TaskService{storage: storage}
}
func (t *TaskService) CreateTask(ctx context.Context, req *pb.CreateTaskRequest) (*pb.TaskResponse, error) {
	task, err := t.storage.CreateTask(ctx, req.UserId, req.TaskText, req.Priority, req.Deadline.AsTime())
	if err != nil {
		return nil, err
	}
	return &pb.TaskResponse{
		UserId:   task.UserID,
		TaskText: task.Text,
		Priority: task.Priority,
		Deadline: timestamppb.New(task.Deadline),
		Progress: task.Progress,
	}, nil
}

func (t *TaskService) GetTasks(ctx context.Context, req *pb.GetTaskRequest) (*pb.GetTaskResponse, error) {
	tasks, err := t.storage.GetTasks(ctx, req.UserId)
	if err != nil {
		return nil, err
	}
	var taskResponses []*pb.TaskResponse
	for _, task := range tasks {
		taskResponses = append(taskResponses, &pb.TaskResponse{
			UserId:   task.UserID,
			TaskText: task.Text,
			Priority: task.Priority,
			Deadline: timestamppb.New(task.Deadline),
			Progress: task.Progress,
		})
	}
	return &pb.GetTaskResponse{
		Tasks: taskResponses,
	}, nil
}

func (t *TaskService) DeleteTask(ctx context.Context, req *pb.DeleteTaskRequest) (*pb.DeleteTaskResponse, error) {
	err := t.storage.DeleteTask(ctx, req.TaskId)
	if err != nil {
		return &pb.DeleteTaskResponse{Success: false, Message: err.Error()}, nil
	}

	return &pb.DeleteTaskResponse{Success: true, Message: "Deleted successfully"}, nil
}

func (t *TaskService) UpdateTaskStatus(ctx context.Context, req *pb.UpdateTaskStatusRequest) (*pb.UpdateTaskStatusResponse, error) {
	err := t.storage.UpdateTaskStatus(ctx, req.TaskId, req.Progress)
	if err != nil {
		return &pb.UpdateTaskStatusResponse{Success: false, Message: err.Error()}, nil
	}

	return &pb.UpdateTaskStatusResponse{Success: true, Message: "Updated successfully"}, nil
}
