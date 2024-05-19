package handler

import (
	"net/http"
	"telegram-bot/pkg/botlogic"
)

func WebhookHandler(w http.ResponseWriter, r *http.Request) {
	tgv.HandleWebhook(r, botlogic.OnUpdate)
}
