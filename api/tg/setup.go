package handler

import (
	"net/http"
	"telegram-bot/pkg/tgvercel"
)

var tgv = tgvercel.New(tgvercel.DefaultOptions())

func TgWebhookHandler(w http.ResponseWriter, r *http.Request) {
	tgv.Handler(w, r)
}
