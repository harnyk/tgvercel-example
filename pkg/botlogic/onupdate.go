package botlogic

import (
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func OnUpdate(
	bot *tgbotapi.BotAPI,
	update *tgbotapi.Update) {
	if update.Message != nil {
		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "echo: "+update.Message.Text)
		_, err := bot.Send(msg)
		if err != nil {
			log.Fatal(err)
		}
	}
}
