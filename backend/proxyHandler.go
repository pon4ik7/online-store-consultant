package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Session data type for managing sessions
type Session struct {
	ID         string
	LastActive time.Time
	Context    string // Save some data about this dialog mb
}

var (
	sessionStore  = make(map[string]Session)
	messagesCache map[string]map[string]string
	storeMu       sync.Mutex // For locking/unlocking sessionStore
)

// First (start) button for starting dialog with AIHelper
func startHandler(w http.ResponseWriter, r *http.Request) {
	// We should check that client only send data
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	session := getOrCreateSession(w, r)
	messagesCache = make(map[string]map[string]string)

	log.Println("Новая сессия: " + session.ID)
	log.Print("Кэшируем сообщения пользователя")
	resp := map[string]string{
		"message": "Привет! Я твой AI-консультант. Задавай вопросы!"}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
	updateLastActive(session.ID)
}

func messageHandler(w http.ResponseWriter, r *http.Request) {
	// We should check that client only send data
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	session := getOrCreateSession(w, r)

	var clientMsg struct {
		Message string `json:"message"`
	}

	if err := json.NewDecoder(r.Body).Decode(&clientMsg); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	aiResponse, ok := HandleUserQuery(messagesCache, clientMsg.Message, false, session.ID)
	log.Printf("Сообщение от %s: %s", session.ID, clientMsg.Message)
	resp := make(map[string]string)
	if ok == nil {
		resp["response"] = aiResponse
	} else {
		resp["response"] = "ai could not handle the request, please, ask the support team"
		log.Println(ok)
	}

	err := saveMessage(session.ID, clientMsg.Message, resp["response"])
	if err != nil {
		http.Error(w, "Error saving message", http.StatusInternalServerError)
		return
	}
	updateLastActive(session.ID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func getOrCreateSession(w http.ResponseWriter, r *http.Request) Session {
	cookie, err := r.Cookie("session_id") //The user was in our cite?
	if err != nil || cookie.Value == "" { //If no create new ID
		newID := uuid.New().String()
		session := Session{
			ID:      newID,
			Context: "",
		}

		storeMu.Lock()
		sessionStore[newID] = session //Put the session into map
		storeMu.Unlock()

		http.SetCookie(w, &http.Cookie{ //Set cookie for client for future work with new ID
			Name:     "session_id",
			Value:    newID,
			Path:     "/",
			HttpOnly: true,
			Secure:   false,
		})

		return session
	}

	storeMu.Lock()
	session, exists := sessionStore[cookie.Value] //Get the ID from map if user already was on our cite
	storeMu.Unlock()

	if !exists {
		log.Printf("Старая сессия %s не найдена, создаём новую", cookie.Value)
		return createNewSession(w)
	}

	return session
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
	err = createSessionMessagesTable(newID)
	if err != nil {
		log.Printf("Error creating table for session %s", newID)
	}
	return session
}
