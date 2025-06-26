package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

var (
	ErrExistingUser    = errors.New("the user already exists")
	ErrNotExistingUser = errors.New("the user does not exist")
	ErrNotInfoAboutReg = errors.New("no information about the registered user")
)

// Session data type for managing sessions
type Session struct {
	ID         string
	LastActive time.Time
	Context    string // Save some data about this dialog mb
}

var (
	sessionStore = make(map[string]Session)
	storeMu      sync.Mutex // For locking/unlocking sessionStore
)

// TODO create the DB table with users
var (
	logSessionStore = make(map[string]Session)
	logStoreMu      sync.RWMutex
)

// First (start) button for starting dialog with AIHelper
func startHandler(w http.ResponseWriter, r *http.Request) {
	// We should check that client only send data
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	session := getOrCreateSession(w, r)

	log.Println("Новая сессия: " + session.ID)
	log.Print("Кэшируем сообщения пользователя")
	resp := map[string]string{
		"message": "Привет! Я твой AI-консультант. Задавай вопросы!"}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
	updateLastActive(session.ID)
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Call /api/register")
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
	}
	var clientMsg struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&clientMsg); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	key := fmt.Sprintf("%s_%s", strings.TrimSpace(clientMsg.Login), strings.TrimSpace(clientMsg.Password))

	session, err := createLogSession(w, key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Println(fmt.Sprintf("New user %s register: %s", strings.TrimSpace(clientMsg.Login), session.ID))
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Call /api/login")
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
	}
	var clientMsg struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&clientMsg); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	key := fmt.Sprintf("%s_%s", strings.TrimSpace(clientMsg.Login), strings.TrimSpace(clientMsg.Password))
	session, err := getLogSession(w, key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Println(fmt.Sprintf("The user %s login: %s", clientMsg.Login, session.ID))
}

// Function that handles the end of the session
func endHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	cookie, err := r.Cookie("session_id")
	if err != nil {
		log.Printf("Error getting session cookie: %v", err)
		return
	}

	resp := make(map[string]string)

	_, ok := sessionStore[cookie.Value]

	if cookie.Value == "" || !ok {
		resp["response"] = "У вас нет никаких запущенных сессий"
		log.Printf("The user does not have any session running")
	} else {
		sessionID := sessionStore[cookie.Value].ID
		log.Printf("The session %s has been ended by the user", sessionID)

		resp["response"] = "Спасибо, что воспользовались нашим консультантом." +
			" Пожалуйста, оцените сессию "

		SaveDialogueContext(sessionID, db)
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(resp)

	if err != nil {
		log.Fatalf("Error encounter while responsing from api/end: %v", err)
	}

}

func messageHandler(w http.ResponseWriter, r *http.Request) {
	// We should check that client only send data
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	session := getOrCreateSession(w, r)
	// TODO create the different logic for register and not register users
	//isRegistered := isRegister(r)
	var clientMsg struct {
		Message   string `json:"message"`
		ProductID string `json:"productID"`
	}

	if err := json.NewDecoder(r.Body).Decode(&clientMsg); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	productID := strings.TrimSpace(clientMsg.ProductID)
	log.Printf("Сообщение от %s: %s", session.ID, clientMsg.Message)
	aiResponse, ok := HandleUserQuery(clientMsg.Message, false, session.ID, productID)
	resp := make(map[string]string)
	if ok == nil {
		resp["response"] = aiResponse
	} else {
		resp["response"] = "The consultant could not handle the request, please, ask the support team"
		log.Println(ok)
	}

	updateLastActive(session.ID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func isRegister(r *http.Request) (bool, error) {
	cookie, err := r.Cookie("isRegister")
	if err != nil {
		return false, ErrNotExistingUser
	}
	if cookie.Value == "true" {
		return true, nil
	}
	return false, nil
}

func getLogSession(w http.ResponseWriter, key string) (Session, error) {
	logStoreMu.Lock()
	session, exists := logSessionStore[key]
	logStoreMu.Unlock()
	if !exists {
		return Session{}, ErrNotExistingUser
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    session.ID,
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "isRegistered",
		Value:    "true",
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
	})

	return session, nil
}

func createLogSession(w http.ResponseWriter, key string) (Session, error) {
	logStoreMu.Lock()
	_, exists := logSessionStore[key]
	logStoreMu.Unlock()
	if exists {
		return Session{}, ErrExistingUser
	}
	session := createNewSession(w)
	logStoreMu.Lock()
	logSessionStore[key] = session
	logStoreMu.Unlock()
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    session.ID,
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "isRegistered",
		Value:    "true",
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
	})
	return session, nil
}

func getOrCreateSession(w http.ResponseWriter, r *http.Request) Session {
	cookie, err := r.Cookie("session_id") //The user was in our cite?

	if err != nil || cookie.Value == "" { //If no create new ID
		return createNewSession(w)
	}

	storeMu.Lock()
	session, exists := sessionStore[cookie.Value] //Get the ID from map if user already was on our cite
	storeMu.Unlock()

	if !exists {
		log.Printf("The old session %s is not found, creating a new one", cookie.Value)
		return createNewSession(w)
	}

	return session
}

func getProductFromSite(productID string) (map[string]interface{}, error) {
	url := fmt.Sprintf("http://localhost:8080/api/products/%s", productID)
	log.Println("GET-запрос на URL:", url)
	resp, err := http.Get(url)
	if err != nil {
		log.Println("Ошибка при запросе:", err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Сервер вернул статус: %d\n", resp.StatusCode)
		return nil, fmt.Errorf("сервер вернул %d", resp.StatusCode)
	}
	// Читаем тело ответа в map
	var data map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}
	return data, nil
}

func productsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET allowed", http.StatusMethodNotAllowed)
		return
	}

	id := strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/products/"), "/")
	filename := filepath.Join("/app/data", fmt.Sprintf("product%s.json", id))
	log.Println("Trying to open file:", filename)

	file, err := os.Open(filename)
	if err != nil {
		log.Println("Product not found:", err)
		http.Error(w, "Product not found", http.StatusNotFound)
		return
	}
	defer file.Close()

	var product map[string]interface{}
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&product)
	if err != nil {
		log.Println("Ошибка при декодировании файла:", err)
		http.Error(w, "Failed to decode product data", http.StatusInternalServerError)
		return
	}

	// Возвращаем JSON как ответ
	w.Header().Set("Content-Type", "application/json")
	file.Seek(0, 0)
	http.ServeFile(w, r, filename)
}

func createNewSession(w http.ResponseWriter) Session {
	newID := uuid.New().String()
	session := Session{
		ID:         newID,
		LastActive: time.Now(),
		Context:    "",
	}

	_, err := db.Exec(`
		INSERT INTO sessions (session_id, context, last_active)
		VALUES ($1, $2, $3)
	`, newID, session.Context, session.LastActive)
	if err != nil {
		log.Printf("Error inserting session into database: %v", err)
	}

	storeMu.Lock()
	sessionStore[newID] = session
	storeMu.Unlock()

	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    newID,
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "isRegistered",
		Value:    "false",
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
	})

	err = createSessionMessagesTable(newID)
	if err != nil {
		log.Printf("Error creating table for session %s", newID)
	}

	return session
}
