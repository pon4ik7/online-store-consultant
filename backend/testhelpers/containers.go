package testhelpers

import (
	"context"
	"fmt"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"time"

	"github.com/jackc/pgx/v5"
)

// PostgresContainer - defines a container for running database in a Docker container
type PostgresContainer struct {
	testcontainers.Container
	ConnectionString string
}

// CreatePostgresContainer - creates and starts a PostgreSQL container
func CreatePostgresContainer(ctx context.Context) (*PostgresContainer, error) {
	req := testcontainers.ContainerRequest{
		Image:        "postgres:15.3-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "postgres",
			"POSTGRES_PASSWORD": "postgres",
			"POSTGRES_DB":       "test-db",
		},
		WaitingFor: wait.ForLog("database system is ready to accept connections").
			WithOccurrence(2).
			WithStartupTimeout(5 * time.Second),
	}

	// Starting the container
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, err
	}
	host, err := container.Host(ctx)
	if err != nil {
		return nil, err
	}
	port, err := container.MappedPort(ctx, "5432")
	if err != nil {
		return nil, err
	}
	connStr := fmt.Sprintf("postgres://postgres:postgres@%s:%s/test-db?sslmode=disable", host, port.Port())
	return &PostgresContainer{
		Container:        container,
		ConnectionString: connStr,
	}, nil
}

// GetDBConnection - retrieves a database connection
func (p *PostgresContainer) GetDBConnection(ctx context.Context) (*pgx.Conn, error) {
	conn, err := pgx.Connect(ctx, p.ConnectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}
	return conn, nil
}

// CheckConnection - checks if a valid connection to PostgreSQL can be established
func (p *PostgresContainer) CheckConnection(ctx context.Context) error {
	conn, err := pgx.Connect(ctx, p.ConnectionString)
	if err != nil {
		return err
	}
	defer conn.Close(ctx)
	return nil
}

// Close - terminates the container
func (p *PostgresContainer) Close(ctx context.Context) error {
	return p.Terminate(ctx)
}
