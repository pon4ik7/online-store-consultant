package main

import (
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/api/hello", messageHandler)
	log.Println("Сервер запущен на http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
