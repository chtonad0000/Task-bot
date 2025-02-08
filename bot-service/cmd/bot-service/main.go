package main

import (
	"flag"
	"github.com/Task-bot/bot-service/internal/bot"
	"github.com/Task-bot/bot-service/internal/grpc"
	"github.com/joho/godotenv"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	envPath := flag.String("env", "", "Path to the .env file")
	flag.Parse()

	if err := godotenv.Load(*envPath); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	taskClient, userClient := grpc.ConnectClients()

	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	if botToken == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN is not set")
	}

	go bot.StartBot(botToken, taskClient, userClient)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	log.Println("Shutting down bot-service...")
}
