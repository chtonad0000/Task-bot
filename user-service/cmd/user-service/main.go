package main

import (
	"flag"
	"fmt"
	pb "github.com/Task-bot/user-service/internal/generated/user"
	"github.com/Task-bot/user-service/internal/service"
	"github.com/Task-bot/user-service/internal/storage"
	"google.golang.org/grpc"
	"log"
	"net"
)

func main() {
	dbURL := flag.String("db-url", "postgres://root:root@localhost:5432/task_manager?sslmode=disable", "Database connection URL")
	grpcPort := flag.String("grpc-port", "50052", "gRPC server port")
	flag.Parse()

	db, err := storage.NewPostgresStorage(*dbURL)
	if err != nil {
		log.Fatalf("Ошибка подключения к БД: %v", err)
	}
	userService := service.NewUserService(db)
	grpcServer := grpc.NewServer()
	pb.RegisterUserServiceServer(grpcServer, userService)
	listener, err := net.Listen("tcp", ":"+*grpcPort)
	if err != nil {
		log.Fatalf("Ошибка при запуске сервера: %v", err)
	}

	fmt.Println("gRPC сервер запущен на порту", *grpcPort)
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Ошибка работы gRPC сервера: %v", err)
	}
}
