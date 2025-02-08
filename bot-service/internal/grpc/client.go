package grpc

import (
	"context"
	"log"
	"os"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pbTask "github.com/Task-bot/bot-service/internal/generated/task"
	pbUser "github.com/Task-bot/bot-service/internal/generated/user"
)

func ConnectClients() (pbTask.TaskServiceClient, pbUser.UserServiceClient) {
	taskServiceAddr := os.Getenv("GRPC_TASK_SERVICE_ADDR")
	userServiceAddr := os.Getenv("GRPC_USER_SERVICE_ADDR")

	if taskServiceAddr == "" || userServiceAddr == "" {
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

	log.Println("Successfully connected to gRPC services")
	return pbTask.NewTaskServiceClient(taskConn), pbUser.NewUserServiceClient(userConn)
}
