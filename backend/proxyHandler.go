package main

import (
	"encoding/json"
	"github.com/golang-jwt/jwt/v5"
	"log"
	"net/http"
	"strings"
)

type RequestMessage struct {
	Message string `json:"message"`
}

type ResponseMessage struct {
	Message string `json:"message"`
}

type User struct {
	Username string `json:"user"`
	Password string `json:"password"`
}

var users = map[string]User{
	"admin": {Username: "admin", Password: "1234"},
}

func startHandler(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		http.Error(writer, "Only POST allowed", http.StatusMethodNotAllowed)
	}

	log.Println("The helper was started")

	resp := ResponseMessage{Message: "Hello, I'm your personal AI helper, sing up or login"}

	writer.Header().Set("Content-Type", "application/json")

	json.NewEncoder(writer).Encode(resp)

	http.HandleFunc("/api/login", loginHandler)

}

func loginHandler(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		http.Error(writer, "Only POST", http.StatusMethodNotAllowed)
		return
	}

	var creds struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	json.NewDecoder(request.Body).Decode(&creds)

	user, exists := users[creds.Username]
	if !exists || user.Password != creds.Password {
		http.Error(writer, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	//TODO create a JWT-token and return to client
	resp := map[string]string{"message": "Вы вошли как " + creds.Username}
	json.NewEncoder(writer).Encode(resp)
}

var jwtKey = []byte("my_secret_key") // TODO save in
// Функция для валидации JWT
func validateJWT(request *http.Request) (string, error) {
	authHeader := request.Header.Get("Authorization")
	if authHeader == "" {
		return "", http.ErrNoCookie
	}

	// Ожидаем: "Bearer <токен>"
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", http.ErrNoCookie
	}

	tokenStr := parts[1]

	claims := &jwt.RegisteredClaims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	if err != nil || !token.Valid {
		return "", err
	}

	// Возвращаем имя пользователя (если добавлено в Subject)
	return claims.Subject, nil
}

func messageHandler(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		http.Error(writer, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	// 🔒 Проверяем токен
	username, err := validateJWT(request)
	if err != nil {
		http.Error(writer, "Unauthorized: invalid or missing token", http.StatusUnauthorized)
		return
	}

	log.Println("Authorized user:", username)

	var req RequestMessage
	err = json.NewDecoder(request.Body).Decode(&req)
	if err != nil {
		http.Error(writer, "Invalid JSON", http.StatusBadRequest)
		return
	}

	resp := ResponseMessage{Message: "Got, " + req.Message + " from user " + username}

	writer.Header().Set("Content-Type", "application/json")
	json.NewEncoder(writer).Encode(resp)
}
