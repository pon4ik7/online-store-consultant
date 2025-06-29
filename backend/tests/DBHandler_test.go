package tests

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"log"
	"online-store-consultant/backend/repositories"
	"online-store-consultant/backend/testhelpers"
	"testing"
	"time"
)

type Session struct {
	ID           string
	LastActive   time.Time
	isRegistered bool
	Context      string
}

// Defines the test suite structure for testing repository functions
type SessionRepoTestSuite struct {
	suite.Suite
	pgContainer *testhelpers.PostgresContainer
	ctx         context.Context
	repository  *repositories.Repository
	conn        *pgx.Conn
}

// Initializes the PostgreSQL container and repository before any tests are run
func (suite *SessionRepoTestSuite) SetupSuite() {
	suite.ctx = context.Background()
	pgContainer, err := testhelpers.CreatePostgresContainer(suite.ctx)
	if err != nil {
		log.Fatal(err)
	}
	suite.pgContainer = pgContainer
	connStr := suite.pgContainer.ConnectionString
	repository, err := repositories.NewRepository(suite.ctx, connStr)
	if err != nil {
		log.Fatal(err)
	}
	suite.repository = repository
	suite.conn, err = pgx.Connect(suite.ctx, connStr)
	if err != nil {
		log.Fatal(err)
	}
	// Generate a session ID and create the session_messages table in the database
	sessionID := uuid.New().String()
	err = suite.repository.CreateSessionMessagesTable(sessionID)
	if err != nil {
		log.Fatal(err)
	}
}

// This function is called after all tests are run to clean up the resources
func (suite *SessionRepoTestSuite) TearDownSuite() {
	if suite.conn != nil {
		suite.conn.Close(suite.ctx)
	}
	if err := suite.pgContainer.Terminate(suite.ctx); err != nil {
		log.Fatalf("Error terminating postgres container: %s", err)
	}
}

// A simple example of an integration test: it tests
// the functionality of creating a session table using DB
func (suite *SessionRepoTestSuite) TestCreateSessionMessagesTable() {
	t := suite.T()
	sessionID := "7cb63410-8b72-459b-b20a-93989a55c361"
	err := suite.repository.CreateSessionMessagesTable(sessionID)
	assert.NoError(t, err)
	var exists bool
	err = suite.conn.QueryRow(suite.ctx, `SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = $1)`, "session_messages_"+sessionID).Scan(&exists)
	assert.NoError(t, err)
	assert.True(t, exists, "Table session_messages_%s should exist", sessionID)
}

// An integration test for testing the updating the last time of
// activity in session
func (suite *SessionRepoTestSuite) TestUpdateLastActive() {
	t := suite.T()

	// Create sessions table
	_, err := suite.conn.Exec(suite.ctx, `
		CREATE TABLE IF NOT EXISTS sessions (
			session_id UUID PRIMARY KEY,
			is_registered BOOLEAN NOT NULL,
			last_active TIMESTAMP NOT NULL
		)
	`)
	assert.NoError(t, err)

	// Create session in database
	sessionID := uuid.New().String()
	_, err = suite.conn.Exec(suite.ctx, `
		INSERT INTO sessions (session_id, is_registered, last_active)
		VALUES ($1, false, NOW() - INTERVAL '1 hour')
	`, sessionID)
	assert.NoError(t, err)

	// Save first last_active value
	var oldTime time.Time
	err = suite.conn.QueryRow(suite.ctx, `
	SELECT last_active FROM sessions WHERE session_id = $1
`, sessionID).Scan(&oldTime)
	assert.NoError(t, err)

	// Wait for 2 seconds for checking the update of time
	time.Sleep(2 * time.Second)

	// Call the function
	suite.repository.UpdateLastActive(sessionID)

	// Save updated last_active value
	var newTime time.Time
	err = suite.conn.QueryRow(suite.ctx, `
	SELECT last_active FROM sessions WHERE session_id = $1
`, sessionID).Scan(&newTime)

	// Compare new and old last_active's
	assert.True(t, newTime.After(oldTime), "last_active must be updated")
}

// An integration test for testing the saving questions and responses from the session
func (suite *SessionRepoTestSuite) TestSaveMessage() {
	t := suite.T()

	// Create unique session ID
	sessionID := uuid.New().String()

	// Create table for sessionID
	createTableQuery := fmt.Sprintf(`
		CREATE TABLE "session_messages_%s" (
			id SERIAL PRIMARY KEY,
			session_id UUID NOT NULL,
			message TEXT NOT NULL,
			response TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT NOW()
		)
	`, sessionID)

	_, err := suite.conn.Exec(suite.ctx, createTableQuery)
	require.NoError(t, err, "Could not create table session_messages_%s", sessionID)

	// Paste the message
	message := "Hello"
	response := "Hello, how can I help you?"

	err = suite.repository.SaveMessage(sessionID, message, response)
	require.NoError(t, err, "saveMessage returned an error")

	// Check the saving of the message
	query := fmt.Sprintf(`SELECT session_id, message, response FROM "session_messages_%s"`, sessionID)
	row := suite.conn.QueryRow(suite.ctx, query)

	var gotSessionID uuid.UUID
	var gotMessage, gotResponse string
	err = row.Scan(&gotSessionID, &gotMessage, &gotResponse)
	require.NoError(t, err, "Could not scan the message")

	assert.Equal(t, sessionID, gotSessionID.String(), "session_id must be equal")
	assert.Equal(t, message, gotMessage, "message must be equal")
	assert.Equal(t, response, gotResponse, "response must be equal")
}

// An integration test for testing the returning of the context within a particular session
func (suite *SessionRepoTestSuite) TestReturnSessionMessages() {
	sessionID := uuid.New().String()

	// Create a table for the messages
	createTableQuery := fmt.Sprintf(`
        CREATE TABLE IF NOT EXISTS "session_messages_%s" (
            id SERIAL PRIMARY KEY,
            session_id UUID,
            message TEXT,
            response TEXT
        );
    `, sessionID)
	_, err := suite.conn.Exec(context.Background(), createTableQuery)
	suite.Require().NoError(err, "The table must be created")

	// Insert the test messages
	insertQuery := fmt.Sprintf(`
        INSERT INTO "session_messages_%s" (session_id, message, response)
        VALUES ($1, $2, $3), ($1, $4, $5)
    `, sessionID)
	_, err = suite.conn.Exec(context.Background(), insertQuery,
		sessionID,
		"hello", "hi there",
		"what's your name?", "I'm a bot",
	)
	suite.Require().NoError(err, "должны вставить тестовые сообщения")

	// Test the function
	messages, err := suite.repository.ReturnSessionMessages(sessionID)
	suite.Require().NoError(err, "Expected no error from ReturnSessionMessages. Got: %s", err)

	expected := []string{
		"hello:hi there",
		"what's your name?:I'm a bot",
	}
	suite.Equal(expected, messages, "Messages must be equal")
}

// For mocking
type mockContextSaver struct {
	CalledWith string
}

// Mock the call of SaveDialogueContext()
func (m *mockContextSaver) SaveDialogueContext(sessionID string) error {
	m.CalledWith = sessionID
	return nil
}

func (suite *SessionRepoTestSuite) TestCheckInactiveSessions() {
	// Create sessions
	_, err := suite.conn.Exec(suite.ctx, `
        CREATE TABLE IF NOT EXISTS sessions (
            session_id TEXT PRIMARY KEY,
            last_active TIMESTAMP NOT NULL,
            is_registered BOOLEAN NOT NULL,
            context TEXT
        )
    `)
	suite.Require().NoError(err)

	// Testing sessions
	activeSessionID := uuid.New().String()
	inactiveSessionID := uuid.New().String()

	_, err = suite.conn.Exec(suite.ctx, `
        INSERT INTO sessions (session_id, last_active, is_registered, context)
        VALUES 
            ($1, NOW(), true, ''),
            ($2, NOW() - INTERVAL '1 hour', false, '')
    `, activeSessionID, inactiveSessionID)
	suite.Require().NoError(err)

	// Create sessionStore
	sessionStore := map[string]repositories.Session{
		activeSessionID:   {ID: activeSessionID, IsRegistered: true},
		inactiveSessionID: {ID: inactiveSessionID, IsRegistered: false},
	}

	// Create mock
	mockSaver := &mockContextSaver{}

	// Call the method
	err = suite.repository.CheckInactiveSessions(sessionStore, mockSaver)
	suite.Require().NoError(err)

	// Checks
	suite.Equal(inactiveSessionID, mockSaver.CalledWith, "SaveDialogueContext должен быть вызван для неактивной сессии")

	_, exists := sessionStore[inactiveSessionID]
	suite.False(exists, "Неактивная сессия должна быть удалена")
}

// Runs the test suite
func TestSessionRepoTestSuite(t *testing.T) {
	suite.Run(t, new(SessionRepoTestSuite))
}
