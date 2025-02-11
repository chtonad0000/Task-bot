package grpc

import (
	"context"
	"github.com/Task-bot/bot-service/internal/services"
	"log"
	"os"
	"time"

	pbNotification "github.com/Task-bot/bot-service/internal/generated/notification"
	pbScheduler "github.com/Task-bot/bot-service/internal/generated/scheduler"
	pbTask "github.com/Task-bot/bot-service/internal/generated/task"
	pbUser "github.com/Task-bot/bot-service/internal/generated/user"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func ConnectClients() *services.ServiceRegistry {
	taskServiceAddr := os.Getenv("GRPC_TASK_SERVICE_ADDR")
	userServiceAddr := os.Getenv("GRPC_USER_SERVICE_ADDR")
	schedulerServiceAddr := os.Getenv("GRPC_SCHEDULER_SERVICE_ADDR")
	notificationServiceAddr := os.Getenv("GRPC_NOTIFICATION_SERVICE_ADDR")

	if taskServiceAddr == "" || userServiceAddr == "" || schedulerServiceAddr == "" || notificationServiceAddr == "" {
		log.Fatal("gRPC service addresses are not set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	taskConn, err := grpc.DialContext(ctx, taskServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		log.Fatalf("Failed to connect to Task service: %v", err)
	}

	userConn, err := grpc.DialContext(ctx, userServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		log.Fatalf("Failed to connect to User service: %v", err)
	}

	schedulerConn, err := grpc.DialContext(ctx, schedulerServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		log.Fatalf("Failed to connect to Scheduler service: %v", err)
	}

	notificationConn, err := grpc.DialContext(ctx, notificationServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		log.Fatalf("Failed to connect to Scheduler service: %v", err)
	}

	log.Println("Successfully connected to gRPC services")
	registry := services.NewServiceRegistry()
	registry.RegisterService(services.NewUserServiceClient(pbUser.NewUserServiceClient(userConn)))
	registry.RegisterService(services.NewTaskServiceClient(pbTask.NewTaskServiceClient(taskConn)))
	registry.RegisterService(services.NewSchedulerServiceClient(pbScheduler.NewSchedulerClient(schedulerConn)))
	registry.RegisterService(services.NewNotificationServiceClient(pbNotification.NewNotificationServiceClient(notificationConn)))

	return registry
}
