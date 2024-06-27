package main

import (
	"log"
	"os"

	"github.com/harnyk/tgvercel-example/pkg/botlogic"
	"github.com/harnyk/tgvercelbot"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("error loading .env file: %v", err)
	}

	token := os.Getenv("TELEGRAM_TOKEN")

	err = tgvercelbot.RunLocal(token, botlogic.OnUpdate)
	if err != nil {
		log.Fatalf("failed to run locally: %v", err)
	}

}
