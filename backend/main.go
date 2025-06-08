package main

import (
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/api/start", startHandler)
	http.HandleFunc("/api/message", messageHandler)

	log.Println("Сервер запущен на :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
