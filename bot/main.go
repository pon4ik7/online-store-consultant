package main

import (
	"flag"
	"github.com/joho/godotenv"
	"log"
	"os"
	"regexp"

	tgClient "bot/clients/telegram"
	"bot/consumer/event-consumer"
	"bot/events/telegram"
)

const (
	tgBotHost = "api.telegram.org"
	batchSize = 100
)

const projectDirName = "online-store-consultant"

func loadEnv() {
	projectName := regexp.MustCompile(`^(.*` + projectDirName + `)`)
	currentWorkDirectory, _ := os.Getwd()
	rootPath := projectName.Find([]byte(currentWorkDirectory))
	err := godotenv.Load(string(rootPath) + `/.env`)

	if err != nil {
		log.Fatalf("Error loading .env file")
	}
}

func main() {
	loadEnv()
	botToken, exists := os.LookupEnv("BOT_TOKEN")
	if !exists {
		log.Fatal("BOT_TOKEN environment variable not set")
	}

	eventsProcessor := telegram.New(
		tgClient.New(tgBotHost, botToken),
	)

	log.Print("service started")

	consumer := event_consumer.New(eventsProcessor, eventsProcessor, batchSize)

	if err := consumer.Start(); err != nil {
		log.Fatal("service is stopped", err)
	}
}

func mustToken() string {
	token := flag.String(
		"tg-bot-token",
		"",
		"token for access to telegram bot",
	)

	flag.Parse()

	if *token == "" {
		log.Fatal("token is not specified")
	}

	return *token
}
