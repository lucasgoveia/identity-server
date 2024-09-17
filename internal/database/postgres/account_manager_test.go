package postgres_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"identity-server/pkg/providers/database/postgres"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/lib/pq"
	"github.com/oklog/ulid/v2"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	am "identity-server/internal/database/postgres"
	"identity-server/internal/domain/entities"
)

func setupTestDb(t *testing.T) (*sql.DB, func()) {
	ctx := context.Background()

	// Start a PostgreSQL container using Testcontainers
	req := testcontainers.ContainerRequest{
		Image:        "postgres:latest",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "test",
			"POSTGRES_PASSWORD": "test",
			"POSTGRES_DB":       "testdb",
		},
		WaitingFor: wait.ForListeningPort("5432/tcp").WithStartupTimeout(30 * time.Second),
	}
	postgresContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})

	assert.NoError(t, err)

	host, err := postgresContainer.Host(ctx)
	assert.NoError(t, err)
	port, err := postgresContainer.MappedPort(ctx, "5432")
	assert.NoError(t, err)

	dsn := fmt.Sprintf("postgres://test:test@%s:%s/testdb?sslmode=disable", host, port.Port())
	db, err := sql.Open("postgres", dsn)
	assert.NoError(t, err)

	// Run schema migrations here to setup the database schema for testing
	err = runMigrations(db, "../../../atlas/migrations")
	assert.NoError(t, err)

	// Return a teardown function to stop the container
	return db, func() {
		_ = db.Close()
		_ = postgresContainer.Terminate(ctx)
	}
}

func runMigrations(db *sql.DB, migrationsDir string) error {
	files, err := os.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %v", err)
	}

	// Sort the migration files by name to ensure they're applied sequentially
	for _, file := range files {
		if file.IsDir() || filepath.Ext(file.Name()) != ".sql" {
			continue
		}

		filePath := filepath.Join(migrationsDir, file.Name())
		err = applyMigration(db, filePath)
		if err != nil {
			return fmt.Errorf("failed to apply migration %s: %v", file.Name(), err)
		}
	}
	return nil
}

// Function to apply a single migration file to the database
func applyMigration(db *sql.DB, migrationFile string) error {
	migrationSQL, err := os.ReadFile(migrationFile)
	if err != nil {
		return fmt.Errorf("failed to read migration file %s: %v", migrationFile, err)
	}

	_, err = db.Exec(string(migrationSQL))
	if err != nil {
		return fmt.Errorf("failed to execute migration %s: %v", migrationFile, err)
	}

	return nil
}

func TestAccountManager_Save(t *testing.T) {
	db, teardown := setupTestDb(t)
	defer teardown()

	accountManager := am.NewAccountManager(&postgres.Db{Db: db})

	t.Run("Successfully save user and identity", func(t *testing.T) {
		userId := ulid.Make()
		user := entities.NewUser(userId, "John Doe", nil, time.Now(), time.Now())

		identityId := ulid.Make()
		identity := entities.NewEmailIdentity(identityId, userId, "johndoe@example.com", "hashed-password", time.Now(), time.Now())

		err := accountManager.Save(user, identity)
		assert.NoError(t, err, "Saving user and identity should not return an error")
	})

	t.Run("Duplicate email", func(t *testing.T) {
		userId := ulid.Make()
		user := entities.NewUser(userId, "Jane Doe", nil, time.Now(), time.Now())

		identityId := ulid.Make()
		identity := entities.NewEmailIdentity(identityId, userId, "johndoe@example.com", "hashed-password", time.Now(), time.Now()) // Duplicate email

		err := accountManager.Save(user, identity)
		assert.Error(t, err, "Saving a duplicate email should return an error")
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			assert.Equal(t, "unique_violation", pqErr.Code.Name())
		}
	})

	t.Run("Transaction rollback on user insert failure", func(t *testing.T) {
		// Simulate a failure on user insert
		_, err := db.Exec("DROP TABLE user_identities") // Force an error by dropping the table
		assert.NoError(t, err)

		userId := ulid.Make()
		user := entities.NewUser(userId, "User With Error", nil, time.Now(), time.Now())

		identityId := ulid.Make()
		identity := entities.NewEmailIdentity(identityId, userId, "error@example.com", "hashed-password", time.Now(), time.Now())

		err = accountManager.Save(user, identity)
		assert.Error(t, err, "Inserting user with error should return an error")
	})
}
