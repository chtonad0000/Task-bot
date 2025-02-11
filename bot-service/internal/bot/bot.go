package bot

import (
	"github.com/Task-bot/bot-service/internal/services"
	tgbot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"os"
)

func StartBot(token string, registry *services.ServiceRegistry) {
	bot, err := tgbot.NewBotAPI(token)
	if err != nil {
		log.Fatalf("Failed to create bot: %v", err)
	}

	bot.Debug = true
	log.Printf("Authorized on account %s", bot.Self.UserName)
	u := tgbot.NewUpdate(0)
	u.Timeout = 60
	kafkaBroker := os.Getenv("KAFKA_BROKER")
	kafkaTopic := os.Getenv("KAFKA_TOPIC")
	if kafkaBroker == "" || kafkaTopic == "" {
		log.Fatal("KAFKA_BROKER or KAFKA_TOPIC is not set")
	}

	consumer, err := NewKafkaConsumer(kafkaBroker, kafkaTopic, bot)
	if err != nil {
		log.Fatalf("Failed to create Kafka consumer: %v", err)
	}
	go consumer.Start()

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		go func(upd tgbot.Update) {
			if upd.Message != nil {
				handleMessage(bot, upd.Message, registry)
			} else if upd.CallbackQuery != nil {
				handleCallback(bot, upd.CallbackQuery, registry)
			}
		}(update)
	}
}
