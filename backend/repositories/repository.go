package repositories

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
)

type Repository struct {
	db *pgx.Conn
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

// CreateSessionMessagesTable - a simple test function to demonstrate integration tests
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
