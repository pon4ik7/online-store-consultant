package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	ErrExistingUser    = errors.New("Пользователь с такими данными уже зарегистрирован")
	ErrNotExistingUser = errors.New("Пользователя с такими данными не существует. Сначала завершите регистрацию")
	ErrNotInfoAboutReg = errors.New("no information about the registered user")
)

// Session data type for managing sessions
type Session struct {
	ID           string
	LastActive   time.Time
	isRegistered bool
	Context      string
}

// Function that initializes the session with the consultant and attach the unique identifier to it
func startHandler(w http.ResponseWriter, r *http.Request) {
	// We should check that client only send data
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	session := getInitialSession(w, r)

	log.Println("New unauthorized session: " + session.ID)
	log.Print("Caching the user messages")
	updateLastActive(session.ID, session.isRegistered)
}

// Function to handle the user register action
func registerHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Calling /api/register")
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
	}
	anonSession := getInitialSession(w, r)
	oldAnonID := anonSession.ID
	var clientMsg struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&clientMsg); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	key := fmt.Sprintf("%s_%s", strings.TrimSpace(clientMsg.Login), strings.TrimSpace(clientMsg.Password))
	isReg, err := isRegistered(r)
	if err != nil {
		log.Println("Error while checking if user has been registered: " + err.Error())
	}
	resp := make(map[string]string)
	w.Header().Set("Content-Type", "application/json")
	if isReg {
		resp["response"] = ErrExistingUser.Error()
		json.NewEncoder(w).Encode(resp)
		log.Printf("Session for %s already registered", key)
		return
	}
	var userID string
	err = db.QueryRow(`SELECT user_id FROM users WHERE credentials = $1`, key).Scan(&userID)
	if err == nil {
		http.Error(w, "User already exists", http.StatusConflict)
		return
	}
	_, err = registerUser(key)
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}
	session, err := createAuthorizedSession(w, key)
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}
	if _, err := db.Exec(`
        UPDATE users
        SET session_id = $2
        WHERE credentials = $1
    `, key, session.ID); err != nil {
		log.Printf("Error updating user.session_id: %v", err)
	}
	if _, err := db.Exec(fmt.Sprintf(`
        INSERT INTO "user_messages_%s" (message, response)
        SELECT message, response
        FROM "anonymous_messages_%s";`,
		session.ID, oldAnonID,
	)); err != nil {
		log.Printf("Error migrating messages: %v", err)
	}
	if err := deleteAnonymousSession(oldAnonID); err != nil {
		log.Println("Warning:", err)
	}
	log.Println(fmt.Sprintf("New user %s is registered: %s", strings.TrimSpace(clientMsg.Login), session.ID))
	resp["response"] = "Вы успешно зарегистрировались и вошли в аккаунт"
	json.NewEncoder(w).Encode(resp)

}

// Function to handle signing in
func loginHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Calling /api/login")
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
	sessionID, err := loginUser(key)
	resp := make(map[string]string)
	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		resp["response"] = err.Error()
		json.NewEncoder(w).Encode(resp)
		log.Printf("Error encountered while signing in the user %s: %v", key, "user is not registered")
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    sessionID,
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
	anonSession := getInitialSession(w, r)
	oldAnonID := anonSession.ID
	if err := deleteAnonymousSession(oldAnonID); err != nil {
		log.Println("Warning:", err)
	}
	log.Println(fmt.Sprintf("The user %s is loged in for the session: %s", clientMsg.Login, sessionID))
	resp["response"] = "Вы успешно вошли в систему"
	json.NewEncoder(w).Encode(resp)

}

// Function that handles the end of the session
func endHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}
	resp := make(map[string]string)
	cookie, err := r.Cookie("session_id")
	if err != nil {
		log.Printf("Error getting session cookie: %v", err)
		return
	}
	if cookie.Value == "" {
		resp["response"] = "У вас нет никаких запущенных сессий"
		log.Printf("The user does not have any session running")
	} else {
		sessionID := cookie.Value
		isReg := false
		if regCookie, err := r.Cookie("isRegistered"); err == nil && regCookie.Value == "true" {
			isReg = true
		}
		if !isReg {
			log.Printf("Ending anonymous session %s, deleting it", sessionID)
			_, err := db.Exec(`
			DELETE FROM anonymous_sessions
        	WHERE session_id = $1
		`, sessionID)
			if err != nil {
				log.Printf("Error deleting anonymous session %s: %v", sessionID, err)
			}
		} else {
			log.Printf("Ending registered session %s, saving context", sessionID)
			SaveDialogueContext(sessionID, db)
			_, err := db.Exec(`
			UPDATE user_sessions
			SET last_active = NOW() - INTERVAL '15 minutes'
			WHERE session_id = $1
		`, sessionID)
			if err != nil {
				log.Printf("Error expiring user session %s: %v", sessionID, err)
			}
		}
		resp["response"] = "Спасибо, что воспользовались нашим консультантом." +
			" Пожалуйста, оцените сессию "
	}
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		log.Fatalf("Error encounter while responsing from api/end: %v", err)
	}

}

// Function to handle the main flow of the user dialogue
func messageHandler(w http.ResponseWriter, r *http.Request) {
	// We should check that client only send data
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	session := getInitialSession(w, r)
	var clientMsg struct {
		Message   string `json:"message"`
		ProductID string `json:"productID"`
	}

	if err := json.NewDecoder(r.Body).Decode(&clientMsg); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	productID := strings.TrimSpace(clientMsg.ProductID)
	log.Printf("Message from %s: %s", session.ID, clientMsg.Message)
	aiResponse, ok := HandleUserQuery(clientMsg.Message, false, session.ID, productID, session.isRegistered)
	resp := make(map[string]string)
	if ok == nil {
		resp["response"] = aiResponse
	} else {
		resp["response"] = "Консультант не может помочь с этим вопросом, пожалуйста, обратитесь к технической поддержке через /help"
		log.Println(ok)
	}
	updateLastActive(session.ID, session.isRegistered)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// Function to handle the GET requests about product data
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

	log.Println("Successfully opened the product file:", filename)
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

// Function used to differentiate authorized and non-authorized users
func isRegistered(r *http.Request) (bool, error) {
	cookie, err := r.Cookie("isRegistered")
	if err != nil {
		return false, ErrNotExistingUser
	}

	if cookie.Value == "true" {
		return true, nil
	}

	return false, nil
}

// Function to create a session for a newly registered user
func createAuthorizedSession(w http.ResponseWriter, key string) (Session, error) {
	newID := uuid.New().String()
	if _, err := db.Exec(`
        INSERT INTO user_sessions (session_id, user_id)
        SELECT $1, user_id FROM users WHERE credentials = $2
    `, newID, key); err != nil {
		return Session{}, fmt.Errorf("cannot insert user_session: %w", err)
	}
	newSession := Session{ID: newID, LastActive: time.Now(), isRegistered: true}

	if err := createUserMessagesTable(newID); err != nil {
		return Session{}, fmt.Errorf("cannot create user messages table: %w", err)
	}
	newSession.isRegistered = true
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    newSession.ID,
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
	return newSession, nil
}

// Function to create or get the general session without information regarding authorization
func getInitialSession(w http.ResponseWriter, r *http.Request) Session {

	cookie, err := r.Cookie("session_id") //The user was in our cite?
	if err != nil || cookie.Value == "" { //If no create new ID
		return createNewAnonymousSession(w)
	}
	sessionID := cookie.Value
	var lastActive time.Time
	var context sql.NullString
	var wasUpdated bool
	err = db.QueryRow(`
        SELECT last_active, context, was_context_updated
        FROM user_sessions
        WHERE session_id = $1
    `, sessionID).Scan(&lastActive, &context, &wasUpdated)
	if err == nil {
		return Session{
			ID:           sessionID,
			LastActive:   lastActive,
			isRegistered: true,
			Context:      context.String,
		}
	}
	if err != sql.ErrNoRows {
		log.Printf("Error fetching user session %s: %v", sessionID, err)
	}
	err = db.QueryRow(`
        SELECT last_active
        FROM anonymous_sessions
        WHERE session_id = $1
    `, sessionID).Scan(&lastActive)

	if err == nil {
		return Session{
			ID:           sessionID,
			LastActive:   lastActive,
			isRegistered: false,
			Context:      "",
		}
	}
	if err != sql.ErrNoRows {
		log.Printf("Error fetching anonymous session %s: %v", sessionID, err)
	}
	log.Printf("The old session %s is not found, creating a new one", cookie.Value)
	return createNewAnonymousSession(w)
}

// Function used to fetch product data available at the shop from json files
func getProductFromSite(productID string) (map[string]interface{}, error) {
	url := fmt.Sprintf("http://localhost:8080/api/products/%s", productID)
	log.Println("GET-request on URL:", url)
	resp, err := http.Get(url)
	if err != nil {
		log.Println("Error while requesting:", err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Server returned status: %d\n", resp.StatusCode)
		return nil, fmt.Errorf("Server returned: %d", resp.StatusCode)
	}
	// Читаем тело ответа в map
	var data map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}
	return data, nil
}

// Function to initialize the general session with the client
func createNewAnonymousSession(w http.ResponseWriter) Session {
	newID := uuid.New().String()
	session := Session{
		ID:           newID,
		LastActive:   time.Now(),
		isRegistered: false,
		Context:      "",
	}
	_, err := db.Exec(`
		INSERT INTO anonymous_sessions (session_id, last_active)
		VALUES ($1, $2)
	`, newID, session.LastActive)
	if err != nil {
		log.Printf("Error inserting session into database: %v", err)
	}

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

	err = createAnonymousMessagesTable(newID)

	return session
}
