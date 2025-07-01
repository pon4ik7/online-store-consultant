package main

import (
	"database/sql"
	"fmt"
	"github.com/google/uuid"
	"log"
	"time"
)

func startSessionChecker() {
	ticker := time.NewTicker(3 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			checkInactiveSessions()
		}
	}
}

// Every three minutes checks is there an inactive session
func checkInactiveSessions() {
	rows, err := db.Query(`
        SELECT session_id, last_active
        FROM anonymous_sessions
    `)
	if err != nil {
		log.Printf("Error fetching anonymous_sessions: %v", err)
	} else {
		defer rows.Close()
		for rows.Next() {
			var sessionID uuid.UUID
			var lastActive time.Time
			if err := rows.Scan(&sessionID, &lastActive); err != nil {
				log.Printf("Error scanning anonymous_sessions: %v", err)
				continue
			}
			if time.Since(lastActive) > 15*time.Minute {
				sessionIDStr := sessionID.String()
				log.Printf("Anonymous session %s has been inactive for 15 minutes, deleting it", sessionIDStr)
				if _, err := db.Exec(`DELETE FROM anonymous_sessions WHERE session_id = $1`, sessionID); err != nil {
					log.Printf("Error deleting anonymous session %s: %v", sessionIDStr, err)
				}
			}
		}
	}
	rows, err = db.Query(`
        SELECT session_id, last_active, was_context_updated
        FROM user_sessions
    `)
	if err != nil {
		log.Printf("Error fetching sessions: %v", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var sessionID uuid.UUID
		var lastActive time.Time
		var wasContextUpdated bool
		err := rows.Scan(&sessionID, &lastActive, &wasContextUpdated)
		if err != nil {
			log.Printf("Error reading session data: %v", err)
			continue
		}
		// If last active time is older than 15 minutes and context is empty, save context
		if time.Since(lastActive) > 15*time.Minute && !wasContextUpdated {
			sessionIDStr := sessionID.String()
			log.Printf("User session %s has been inactive for 15 minutes, saving context", sessionID)
			SaveDialogueContext(sessionIDStr, db)
		}
	}
}

// Create user_messages table for each new session
func createUserMessagesTable(sessionID string) error {
	query := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS "user_messages_%s" (
			message_id SERIAL PRIMARY KEY,
			message TEXT,
			response TEXT
		);`, sessionID)
	_, err := db.Exec(query)
	if err != nil {
		log.Printf("Error creating user messages table for session %s: %v", sessionID, err)
	}
	return err
}

func createAnonymousMessagesTable(sessionID string) error {
	query := fmt.Sprintf(`
        CREATE TABLE IF NOT EXISTS "anonymous_messages_%s" (
            message_id SERIAL PRIMARY KEY,
            message TEXT,
            response TEXT
        );
    `, sessionID)
	_, err := db.Exec(query)
	if err != nil {
		log.Printf("Error creating anonymous messages table for session %s: %v", sessionID, err)
	}
	return err
}

// Updates "last_active" field when there is a new message from the user
func updateLastActive(sessionID string, isRegistered bool) {
	table := "anonymous_sessions"
	if isRegistered {
		table = "user_sessions"
	}
	_, err := db.Exec(fmt.Sprintf(
		`UPDATE %s SET last_active = NOW() WHERE session_id = $1`, table,
	), sessionID)
	if err != nil {
		log.Printf("Error updating last_active for %s %s: %v", table, sessionID, err)
	}
}

// Save all questions and responses from the session
func saveUserMessage(sessionID string, userMessage string, response string) error {
	query := fmt.Sprintf(`
        INSERT INTO "user_messages_%s" (message, response)
        VALUES ($1, $2)
    `, sessionID)
	_, err := db.Exec(query, userMessage, response)
	if err != nil {
		log.Printf("Error saving user message for session %s: %v", sessionID, err)
		return err
	}
	log.Printf("Message and response saved for session %s", sessionID)
	return nil
}

func saveAnonymousMessage(sessionID string, userMessage string, response string) error {
	query := fmt.Sprintf(`
        INSERT INTO "anonymous_messages_%s" (message, response)
        VALUES ($1, $2)
    `, sessionID)
	_, err := db.Exec(query, userMessage, response)
	if err != nil {
		log.Printf("Error saving anonymous message for session %s: %v", sessionID, err)
		return err
	}
	log.Printf("Message and response saved for session %s", sessionID)
	return nil
}

// Save popular products info from the website
func savePopularProductsData() {
	insertProduct := `INSERT INTO popular_products (name, description, price, rating, category, product_url, image_url) 
	VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	// Test values
	_, err := db.Exec(insertProduct, "Smartphone", "A high-end smartphone with great features", 999.99, 4.7, "Electronics", "http://example.com/product/1", "http://example.com/images/product1.jpg")
	if err != nil {
		log.Printf("Error inserting product: %v", err)
	}
	log.Println("Test data inserted successfully!")
}

// Returns the context within a particular session
func returnSessionMessages(sessionID string, registered bool) ([]string, error) {
	table := "anonymous_messages_" + sessionID
	if registered {
		table = "user_messages_" + sessionID
	}
	query := fmt.Sprintf(`SELECT message, response FROM "%s"`, table)
	rows, err := db.Query(query)
	if err != nil {
		log.Printf("Error fetching session messages for session %s: %v", sessionID, err)
		return nil, err
	}
	defer rows.Close()

	messagesCache := make([]string, 0)
	for rows.Next() {
		var msg, response string
		if err := rows.Scan(&msg, &response); err != nil {
			log.Printf("Error reading messages for session %s: %v", sessionID, err)
		}
		messagesCache = append(messagesCache, msg+":"+response)
	}

	return messagesCache, nil
}

func registerUser(credentials string) (string, error) {
	userID := uuid.New().String()
	_, err := db.Exec(`
		INSERT INTO users (user_id, credentials)
		VALUES ($1, $2)
	`, userID, credentials)
	if err != nil {
		log.Printf("Error inserting user: %v", err)
		return "", err
	}
	return userID, nil
}

func loginUser(credentials string) (string, error) {
	var userID, sessionID string
	err := db.QueryRow(`
		SELECT user_id, session_id FROM users WHERE credentials = $1
	`, credentials).Scan(&userID, &sessionID)

	if err != nil {
		if err == sql.ErrNoRows {
			log.Println("User not found")
			return "", ErrNotExistingUser
		}
		log.Printf("Error querying user: %v", err)
		return "", err
	}
	if _, err := db.Exec(`
        UPDATE user_sessions
        SET last_active = NOW(),
            was_context_updated = FALSE
        WHERE session_id = $1
    `, sessionID); err != nil {
		log.Printf("Error updating last_active: %v", err)
	}
	if err := createUserMessagesTable(sessionID); err != nil {
		log.Printf("Error ensuring user_messages_%s table: %v", sessionID, err)
	}
	return sessionID, nil
}

func deleteAnonymousSession(sessionID string) error {
	if _, err := db.Exec(fmt.Sprintf(`DROP TABLE IF EXISTS "anonymous_messages_%s"`, sessionID)); err != nil {
		return fmt.Errorf("delete anonymous messages table %s: %w", sessionID, err)
	}
	if _, err := db.Exec(`DELETE FROM anonymous_sessions WHERE session_id = $1`, sessionID); err != nil {
		return fmt.Errorf("delete anonymous session %s: %w", sessionID, err)
	}
	return nil
}
