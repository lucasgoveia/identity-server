package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"identity-server/internal/accounts/domain"
	"identity-server/pkg/providers/database"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/lib/pq"
	"github.com/oklog/ulid/v2"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
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

	accountManager := NewPostgresAccountRepository(&database.Db{Db: db})

	t.Run("Successfully save user and identity", func(t *testing.T) {
		ctx := context.Background()
		userId := ulid.Make()
		user := domain.NewUser(userId, "John Doe", nil, time.Now(), time.Now())

		identityId := ulid.Make()
		identity := domain.NewEmailIdentity(identityId, userId, "johndoe@example.com", "hashed-password", time.Now(), time.Now())

		err := accountManager.Save(ctx, user, identity)
		assert.NoError(t, err, "Saving user and identity should not return an error")
	})

	t.Run("Duplicate email", func(t *testing.T) {
		ctx := context.Background()
		userId := ulid.Make()
		user := domain.NewUser(userId, "Jane Doe", nil, time.Now(), time.Now())

		identityId := ulid.Make()
		identity := domain.NewEmailIdentity(identityId, userId, "johndoe@example.com", "hashed-password", time.Now(), time.Now()) // Duplicate email

		err := accountManager.Save(ctx, user, identity)
		assert.Error(t, err, "Saving a duplicate email should return an error")
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			assert.Equal(t, "unique_violation", pqErr.Code.Name())
		}
	})

	t.Run("Transaction rollback on user insert failure", func(t *testing.T) {
		ctx := context.Background()
		// Simulate a failure on user insert
		_, err := db.Exec("DROP TABLE user_identities") // Force an error by dropping the table
		assert.NoError(t, err)

		userId := ulid.Make()
		user := domain.NewUser(userId, "User With Error", nil, time.Now(), time.Now())

		identityId := ulid.Make()
		identity := domain.NewEmailIdentity(identityId, userId, "error@example.com", "hashed-password", time.Now(), time.Now())

		err = accountManager.Save(ctx, user, identity)
		assert.Error(t, err, "Inserting user with error should return an error")
	})

}

func TestAccountManager_IdentityExists(t *testing.T) {
	db, teardown := setupTestDb(t)
	defer teardown()

	accountManager := NewPostgresAccountRepository(&database.Db{Db: db})

	t.Run("Identity exists", func(t *testing.T) {
		ctx := context.Background()
		// Create a user and identity first
		userId := ulid.Make()
		user := domain.NewUser(userId, "John Doe", nil, time.Now(), time.Now())

		identityId := ulid.Make()
		identity := domain.NewEmailIdentity(identityId, userId, "existing@example.com", "hashed-password", time.Now(), time.Now())

		err := accountManager.Save(ctx, user, identity)
		assert.NoError(t, err)

		// Check if the identity exists
		exists, err := accountManager.IdentityExists(ctx, string(identity.Type), identity.Value)
		assert.NoError(t, err)
		assert.True(t, exists, "Identity should exist")
	})

	t.Run("Identity does not exist", func(t *testing.T) {
		ctx := context.Background()
		// Check for a non-existing identity
		exists, err := accountManager.IdentityExists(ctx, "email", "nonexistent@example.com")
		assert.NoError(t, err)
		assert.False(t, exists, "Identity should not exist")
	})

	t.Run("Error in query execution", func(t *testing.T) {
		ctx := context.Background()
		// Simulate an error by dropping the `user_identities` table
		_, err := db.Exec("DROP TABLE user_identities")
		assert.NoError(t, err)

		// Attempt to check for an identity, which should now fail
		_, err = accountManager.IdentityExists(ctx, "email", "error@example.com")
		assert.Error(t, err, "Query execution should return an error")
	})
}
