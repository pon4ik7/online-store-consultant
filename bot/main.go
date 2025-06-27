package main

import (
	tgClient "bot/clients/telegram"
	"bot/consumer/event-consumer"
	"bot/events/telegram"
	"github.com/joho/godotenv"
	"log"
	"os"
)

const (
	tgBotHost = "api.telegram.org"
	batchSize = 100
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	botToken, exists := os.LookupEnv("BOT_TOKEN")
	if !exists {
		log.Fatal("BOT_TOKEN environment variable not set")
	}

	eventsProcessor := telegram.New(
		tgClient.New(tgBotHost, botToken),
	)

	log.Print("telegram bot started")

	consumer := event_consumer.New(eventsProcessor, eventsProcessor, batchSize)

	if err := consumer.Start(); err != nil {
		log.Fatal("service is stopped", err)
	}
}
