package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"

	pb "github.com/Task-bot/scheduler-service/internal/generated/scheduler_service"
	"github.com/Task-bot/scheduler-service/internal/service"
	"google.golang.org/grpc"
)

func main() {
	grpcPort := flag.String("grpc-port", "50053", "gRPC server port")
	taskServiceAddr := flag.String("task-service", "localhost:50051", "Task service address")
	flag.Parse()

	ctx := context.Background()
	schedulerService, err := service.NewSchedulerService(ctx, *taskServiceAddr)
	if err != nil {
		log.Fatalf("Ошибка инициализации scheduler-service: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterSchedulerServer(grpcServer, schedulerService)

	listener, err := net.Listen("tcp", ":"+*grpcPort)
	if err != nil {
		log.Fatalf("Ошибка при запуске сервера: %v", err)
	}

	fmt.Println("Scheduler service запущен на порту", *grpcPort)
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Ошибка работы gRPC сервера: %v", err)
	}
}
