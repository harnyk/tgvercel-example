package main

import (
	"log"
	"os"
	"telegram-bot/pkg/botlogic"

	"github.com/harnyk/tgvercel"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("error loading .env file: %v", err)
	}

	token := os.Getenv("TELEGRAM_TOKEN")

	err = tgvercel.RunLocal(token, botlogic.OnUpdate)
	if err != nil {
		log.Fatalf("failed to run locally: %v", err)
	}

}
