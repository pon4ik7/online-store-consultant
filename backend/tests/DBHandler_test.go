package tests

import (
	"context"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"log"
	"online-store-consultant/backend/repositories"
	"online-store-consultant/backend/testhelpers"
	"testing"
)

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

// Runs the test suite
func TestSessionRepoTestSuite(t *testing.T) {
	suite.Run(t, new(SessionRepoTestSuite))
}
