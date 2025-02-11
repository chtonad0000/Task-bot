package main

import (
	"context"
	"flag"
	"fmt"
	"google.golang.org/grpc"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	pb "github.com/Task-bot/notification-service/internal/generated/notification"
	"github.com/Task-bot/notification-service/internal/kafka"
	"github.com/Task-bot/notification-service/internal/service"
	"github.com/Task-bot/notification-service/internal/storage"
	"github.com/Task-bot/notification-service/internal/worker"
)

func main() {
	dbURL := flag.String("database-url", "postgres://root:root@localhost:5432/task_manager?sslmode=disable", "PostgreSQL database connection string")
	kafkaBroker := flag.String("kafka-broker", "localhost:9092", "Kafka broker address")
	kafkaTopic := flag.String("kafka-topic", "notifications", "Kafka topic for notifications")
	grpcPort := flag.String("grpc-port", "50054", "gRPC server port")
	flag.Parse()

	db, err := storage.NewPostgresStorage(*dbURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}

	kafkaProducer, err := kafka.NewKafkaProducer(*kafkaBroker, *kafkaTopic)
	if err != nil {
		log.Fatalf("Error creating Kafka producer: %v", err)
	}
	defer kafkaProducer.Close()

	notificationService := service.NewNotificationService(db)
	notificationWorker := worker.NewNotificationWorker(db, kafkaProducer)

	go func() {
		log.Println("Starting notification worker...")
		notificationWorker.Start(context.Background())
	}()

	grpcServer := grpc.NewServer()
	pb.RegisterNotificationServiceServer(grpcServer, notificationService)
	listener, err := net.Listen("tcp", ":"+*grpcPort)
	if err != nil {
		log.Fatalf("Ошибка при запуске сервера: %v", err)
	}

	fmt.Println("gRPC сервер запущен на порту", *grpcPort)
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Ошибка работы gRPC сервера: %v", err)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigChan
	log.Printf("Received signal %v, shutting down gracefully...", sig)

	log.Println("Gracefully shutting down the service...")
}
