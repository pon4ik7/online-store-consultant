package repositories

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"log"
	"time"
)

type Repository struct {
	db           *pgx.Conn
	SessionStore map[string]Session
}

type SessionContextSaver interface {
	SaveDialogueContext(sessionID string) error
}

type Session struct {
	ID           string
	LastActive   time.Time
	Context      string
	IsRegistered bool
}

// NewRepository - initializes a new Repository instance
// and establishes a connection to the database
func NewRepository(ctx context.Context, connString string) (*Repository, error) {
	conn, err := pgx.Connect(ctx, connString)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %v", err)
	}
	return &Repository{db: conn}, nil
}

// Bellow are placed functions from AIHandler.go for their
// testing using integration tests
func (r *Repository) CreateSessionMessagesTable(sessionID string) error {
	_, err := r.db.Exec(context.Background(), fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS "session_messages_%s" (
			message_id SERIAL PRIMARY KEY,
			session_id UUID NOT NULL,
			message TEXT,
			response TEXT
		)
	`, sessionID))
	return err
}

func (r *Repository) UpdateLastActive(sessionID string) {
	_, err := r.db.Exec(context.Background(), `
		UPDATE sessions
		SET last_active = NOW()
		WHERE session_id = $1
	`, sessionID)
	if err != nil {
		log.Printf("Error updating last_active for session %s: %v", sessionID, err)
	}
}

func (r *Repository) SaveMessage(sessionID string, userMessage string, response string) error {
	query := fmt.Sprintf(`
        INSERT INTO "session_messages_%s" (session_id, message, response)
        VALUES ($1, $2, $3)
    `, sessionID)

	_, err := r.db.Exec(context.Background(), query, sessionID, userMessage, response)
	if err != nil {
		log.Printf("Error saving message for session %s: %v", sessionID, err)
		return err
	}
	log.Printf("Message and response saved for session %s", sessionID)
	return nil
}

func (r *Repository) ReturnSessionMessages(sessionID string) ([]string, error) {
	query := fmt.Sprintf(`SELECT message, response FROM "session_messages_%s"`, sessionID)
	rows, err := r.db.Query(context.Background(), query)
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
			continue
		}
		messagesCache = append(messagesCache, msg+":"+response)
	}

	return messagesCache, nil
}

func (r *Repository) CheckInactiveSessions(sessionStore map[string]Session, contextSaver SessionContextSaver) error {
	rows, err := r.db.Query(context.Background(), `
        SELECT session_id, last_active, context
        FROM sessions;
    `)

	if err != nil {
		return fmt.Errorf("error fetching sessions: %w", err)
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

		sessionIDStr := sessionID.String()

		if time.Since(lastActive) > 1*time.Millisecond && context == "" {
			log.Printf("Session %s has been inactive, saving context", sessionIDStr)

			session, exists := sessionStore[sessionIDStr]
			if !exists || !session.IsRegistered {
				log.Printf("Unregistered session %s will be deleted", sessionIDStr)
				delete(sessionStore, sessionIDStr)
			} else {
				log.Printf("Registered session %s, context not deleted", sessionIDStr)
			}

			_ = contextSaver.SaveDialogueContext(sessionIDStr)
		}
	}

	return nil
}

// Mocked function to test CheckInactiveSessions(),
// will be implemented later
func (r *Repository) SaveDialogueContext(sessionID string) error {
	return nil
}
