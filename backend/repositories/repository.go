package repositories

import (
	"context"
	"errors"
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

func (r *Repository) CheckInactiveSessions(sessionStore map[string]Session) error {
	rows, err := r.db.Query(context.Background(), `
        SELECT session_id, last_active, context, is_registered
        FROM sessions;
    `)
	if err != nil {
		return fmt.Errorf("error fetching sessions: %w", err)
	}
	defer rows.Close()

	type dbSession struct {
		ID           uuid.UUID
		LastActive   time.Time
		Context      string
		IsRegistered bool
	}

	var sessions []dbSession

	// 1. Читаем все строки в память
	for rows.Next() {
		var s dbSession
		if err := rows.Scan(&s.ID, &s.LastActive, &s.Context, &s.IsRegistered); err != nil {
			log.Printf("Error reading session data: %v", err)
			continue
		}
		sessions = append(sessions, s)
	}

	// 2. Обрабатываем сессии после rows.Close()
	// (defer rows.Close() уже есть выше)
	for _, s := range sessions {
		sessionIDStr := s.ID.String()

		if time.Since(s.LastActive) > 1*time.Minute {
			log.Printf("Session %s has been inactive, saving context", sessionIDStr)

			session, exists := sessionStore[sessionIDStr]
			regStatus := s.IsRegistered
			if exists {
				regStatus = session.IsRegistered
			}

			err := r.SaveDialogueContext(sessionIDStr, s.Context, regStatus)
			if err != nil {
				return fmt.Errorf("failed to save context for session %s: %w", sessionIDStr, err)
			}

			if !regStatus && exists {
				log.Printf("Unregistered session %s will be deleted", sessionIDStr)
				delete(sessionStore, sessionIDStr)
			} else {
				log.Printf("Registered session %s, context not deleted", sessionIDStr)
			}
		}
	}

	return nil
}

func (r *Repository) SaveDialogueContext(sessionIDStr string, keyWords string, isRegistered bool) error {
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		return fmt.Errorf("invalid session ID: %w", err)
	}

	now := time.Now()
	_, err = r.db.Exec(context.Background(), `
		INSERT INTO sessions (session_id, context, last_active, is_registered)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (session_id) DO UPDATE 
		SET context = EXCLUDED.context,
		    last_active = EXCLUDED.last_active,
		    is_registered = EXCLUDED.is_registered;
	`, sessionID, keyWords, now, isRegistered)
	if err != nil {
		return fmt.Errorf("failed to save context: %w", err)
	}

	log.Printf("Context for session %s has been saved successfully", sessionID)
	return nil
}

func (r *Repository) DeleteAnonymousSession(sessionID string) error {
	if _, err := r.db.Exec(context.Background(), fmt.Sprintf(
		`DROP TABLE IF EXISTS "anonymous_messages_%s"`, sessionID,
	)); err != nil {
		return fmt.Errorf("delete anonymous messages table %s: %w", sessionID, err)
	}

	if _, err := r.db.Exec(context.Background(),
		`DELETE FROM anonymous_sessions WHERE session_id = $1`, sessionID,
	); err != nil {
		return fmt.Errorf("delete anonymous session %s: %w", sessionID, err)
	}

	return nil
}

func (r *Repository) RegisterUser(ctx context.Context, credentials string) (string, error) {
	userID := uuid.New().String()
	_, err := r.db.Exec(ctx, `
		INSERT INTO users (user_id, credentials)
		VALUES ($1, $2)
	`, userID, credentials)
	if err != nil {
		log.Printf("Error inserting user: %v", err)
		return "", err
	}
	return userID, nil
}

func (r *Repository) СreateUserMessagesTable(ctx context.Context, sessionID string) error {
	query := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS "user_messages_%s" (
			message_id SERIAL PRIMARY KEY,
			message TEXT,
			response TEXT
		);`, sessionID)

	_, err := r.db.Exec(ctx, query)
	if err != nil {
		log.Printf("Error creating user messages table for session %s: %v", sessionID, err)
	}
	return err
}

func (r *Repository) LoginUser(ctx context.Context, credentials string) (string, error) {
	var userID, sessionID string
	err := r.db.QueryRow(ctx, `
		SELECT user_id, session_id FROM users WHERE credentials = $1
	`, credentials).Scan(&userID, &sessionID)

	if err != nil {
		if err == pgx.ErrNoRows {
			log.Println("User not found")
			return "", errors.New("Пользователя с такими данными не существует. Сначала завершите регистрацию")
		}
		log.Printf("Error querying user: %v", err)
		return "", err
	}

	_, err = r.db.Exec(ctx, `
		UPDATE user_sessions
		SET last_active = NOW(),
		    was_context_updated = FALSE
		WHERE session_id = $1
	`, sessionID)
	if err != nil {
		log.Printf("Error updating last_active: %v", err)
	}

	if err := r.СreateUserMessagesTable(ctx, sessionID); err != nil {
		log.Printf("Error ensuring user_messages_%s table: %v", sessionID, err)
	}

	return sessionID, nil
}
