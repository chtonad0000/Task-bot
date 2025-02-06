package main

import (
	"flag"
	"fmt"
	"log"
	"net"

	pb "github.com/Task-bot/task-service/internal/generated/task"
	"github.com/Task-bot/task-service/internal/service"
	"github.com/Task-bot/task-service/internal/storage"
	"google.golang.org/grpc"
)

func main() {
	dbURL := flag.String("db-url", "postgres://root:root@localhost:5432/task_manager?sslmode=disable", "Database connection URL")
	grpcPort := flag.String("grpc-port", "50051", "gRPC server port")
	flag.Parse()

	db, err := storage.NewPostgresStorage(*dbURL)
	if err != nil {
		log.Fatalf("Ошибка подключения к БД: %v", err)
	}
	taskService := service.NewTaskService(db)

	grpcServer := grpc.NewServer()
	pb.RegisterTaskServiceServer(grpcServer, taskService)
	listener, err := net.Listen("tcp", ":"+*grpcPort)
	if err != nil {
		log.Fatalf("Ошибка при запуске сервера: %v", err)
	}

	fmt.Println("gRPC сервер запущен на порту", *grpcPort)
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Ошибка работы gRPC сервера: %v", err)
	}
}
