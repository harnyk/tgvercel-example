package tgvercel

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func DefaultOptions() Options {
	return Options{
		WebhookRelativeUrl:   "/api/tg/webhook",
		TelegramTokenEnvName: "TELEGRAM_TOKEN",
		VercelUrlEnvName:     "VERCEL_URL",
		KeyEnvName:           "TGVERCEL_KEY",
		KeyParamName:         "key",
	}
}

type Options struct {
	WebhookRelativeUrl   string
	TelegramTokenEnvName string
	VercelUrlEnvName     string
	KeyEnvName           string
	KeyParamName         string
}

func (o *Options) Validate() error {
	if o.WebhookRelativeUrl == "" {
		return fmt.Errorf("WebhookRelativeUrl must be set")
	}
	if o.TelegramTokenEnvName == "" {
		return fmt.Errorf("TelegramTokenEnvName must be set")
	}
	if o.VercelUrlEnvName == "" {
		return fmt.Errorf("VercelUrlEnvName must be set")
	}
	if o.KeyEnvName == "" {
		return fmt.Errorf("KeyEnvName must be set")
	}
	if o.KeyParamName == "" {
		return fmt.Errorf("KeyParamName must be set")
	}
	return nil
}

type TgVercel struct {
	options Options
}

func New(options Options) *TgVercel {
	err := options.Validate()
	if err != nil {
		log.Fatal(err)
	}

	return &TgVercel{
		options: options,
	}
}

func setJsonType(w http.ResponseWriter) {
	w.Header().Add("Content-Type", "application/json")
}

func errorResponse(w http.ResponseWriter, err error) {
	setJsonType(w)
	w.WriteHeader(http.StatusInternalServerError)
	errorJson := struct {
		Error string `json:"error"`
	}{
		Error: err.Error(),
	}
	jsonData, err := json.Marshal(errorJson)
	if err != nil {
		log.Fatal(err)
	}
	w.Write(jsonData)
}

func unauthorizedResponse(w http.ResponseWriter, err error) {
	setJsonType(w)
	w.WriteHeader(http.StatusUnauthorized)
	errorJson := struct {
		Error string `json:"error"`
	}{
		Error: err.Error(),
	}
	jsonData, err := json.Marshal(errorJson)
	if err != nil {
		log.Fatal(err)
	}
	w.Write(jsonData)
}

func okResponse(w http.ResponseWriter, data interface{}) {
	setJsonType(w)
	w.WriteHeader(http.StatusOK)
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Fatal(err)
	}
	w.Write(jsonData)
}

func (t *TgVercel) Handler(w http.ResponseWriter, r *http.Request) {
	o := t.options

	tgBotServiceKey := os.Getenv(o.KeyEnvName)
	if tgBotServiceKey == "" {
		errorResponse(w, fmt.Errorf("%s is not set", o.KeyEnvName))
		return
	}

	token := os.Getenv(o.TelegramTokenEnvName)
	if token == "" {
		errorResponse(w, fmt.Errorf("%s is not set", o.TelegramTokenEnvName))
		return
	}

	vercelUrl := os.Getenv(o.VercelUrlEnvName)
	if vercelUrl == "" {
		errorResponse(w, fmt.Errorf("%s is not set", o.VercelUrlEnvName))
		return
	}

	key := r.URL.Query().Get(o.KeyParamName)
	if key != tgBotServiceKey {
		unauthorizedResponse(w, fmt.Errorf("invalid key"))
		return
	}

	webhookUri := fmt.Sprintf("https://%s%s", vercelUrl, o.WebhookRelativeUrl)

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		errorResponse(w, err)
		return
	}

	bot.Debug = true
	log.Printf("Authorized on account %s", bot.Self.UserName)

	wh, err := tgbotapi.NewWebhook(webhookUri)
	if err != nil {
		log.Fatal(err)
	}

	apiResponse, err := bot.Request(wh)
	if err != nil {
		errorResponse(w, fmt.Errorf("failed to set webhook: %w", err))
		return
	}

	if !apiResponse.Ok {
		errorResponse(w, fmt.Errorf("failed to set webhook: %s", apiResponse.Description))
		return
	}

	okResponse(w, apiResponse.Description)
}
