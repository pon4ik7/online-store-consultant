package main

import (
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
		SELECT session_id, last_active, context
		FROM sessions;
	`)
	if err != nil {
		log.Printf("Error fetching sessions: %v", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var sessionID uuid.UUID
		var lastActive time.Time
		var context string
		err := rows.Scan(&sessionID, &lastActive, &context)
		if err != nil {
			log.Printf("Error reading session data: %v", err)
			continue
		}

		// If last active time is older than 15 minutes and context is empty, save context
		if time.Since(lastActive) > 15*time.Minute && context == "" {
			log.Printf("Session %s has been inactive for 15 minutes, saving context", sessionID)
			SaveDialogueContext(sessionID.String(), db)
		}
	}
}

// Create session_messages table for each new session
func createSessionMessagesTable(sessionID string) error {
	query := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS "session_messages_%s" (
			message_id SERIAL PRIMARY KEY,
			session_id UUID REFERENCES sessions (session_id) ON DELETE CASCADE,
			message TEXT,
			response TEXT
		);`, sessionID)
	_, err := db.Exec(query)
	if err != nil {
		log.Printf("Error creating session messages table for session %s: %v", sessionID, err)
	}
	return err
}

// Updates "last_active" field when there is a new message from the user
func updateLastActive(sessionID string) {
	_, err := db.Exec(`
		UPDATE sessions
		SET last_active = NOW()
		WHERE session_id = $1
	`, sessionID)
	if err != nil {
		log.Printf("Error updating last_active for session %s: %v", sessionID, err)
	}
}

// Save all questions and responses from the session
func saveMessage(sessionID string, userMessage string, response string) error {
	query := fmt.Sprintf(`
        INSERT INTO "session_messages_%s" (session_id, message, response)
        VALUES ($1, $2, $3)
    `, sessionID)
	_, err := db.Exec(query, sessionID, userMessage, response)
	if err != nil {
		log.Printf("Error saving message for session %s: %v", sessionID, err)
		return err
	}
	log.Printf("Message and response saved for session %s", sessionID)
	return nil
}

// Save popular products info from the website
// TODO: Use queries to get information about popular products on the site
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
func returnSessionMessages(sessionID string) ([]string, error) {
	query := fmt.Sprintf(`SELECT message, response FROM "session_messages_%s"`, sessionID)
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
