package bot

import (
	"log"

	"github.com/Task-bot/bot-service/internal/services"
	tgbot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
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
