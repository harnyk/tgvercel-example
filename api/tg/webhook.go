package handler

import (
	"net/http"

	"github.com/harnyk/tgvercel-example/pkg/botlogic"
)

func WebhookHandler(w http.ResponseWriter, r *http.Request) {
	tgv.HandleWebhook(r, botlogic.OnUpdate)
}
