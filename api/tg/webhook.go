package handler

import (
	"net/http"

	"github.com/harnyk/tgvercel-example/pkg/botlogic"
	"github.com/harnyk/tgvercelbot"
)

var tgv = tgvercelbot.New(tgvercelbot.DefaultOptions())

func WebhookHandler(w http.ResponseWriter, r *http.Request) {
	tgv.HandleWebhook(r, botlogic.OnUpdate)
}
