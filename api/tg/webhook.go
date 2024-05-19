package handler

import (
	"log"
	"net/http"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	token := os.Getenv("TELEGRAM_TOKEN")
	if token == "" {
		log.Fatal("TELEGRAM_TOKEN is not set")
	}

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatal(err)
	}

	upd, err := bot.HandleUpdate(r)
	if err != nil {
		log.Fatal(err)
	}

	if upd.Message != nil {
		log.Printf("[%s] %s", upd.Message.From.UserName, upd.Message.Text)
		msg := tgbotapi.NewMessage(upd.Message.Chat.ID, "echo: "+upd.Message.Text)
		_, err := bot.Send(msg)
		if err != nil {
			log.Fatal(err)
		}
	}
}
