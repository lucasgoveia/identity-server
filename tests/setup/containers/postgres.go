package containers

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/labstack/gommon/log"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"os"
	"path/filepath"
	"time"
)

func SetupTestPostgresDb() (string, func(), error) {
	ctx := context.Background()

	// Start a PostgreSQL container using Testcontainers
	postgresContainer, err := postgres.Run(ctx, "postgres:latest",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second)))

	if err != nil {
		return "", nil, err
	}

	teardown := func() {
		if err := postgresContainer.Terminate(ctx); err != nil {
			log.Errorf("Failed to terminate container: %s", err)
		}
	}

	dsn, err := postgresContainer.ConnectionString(ctx, "sslmode=disable")

	if err != nil {
		log.Error("failed to get conn string")
		return "", teardown, err
	}

	db, err := sql.Open("postgres", dsn)

	defer func() {
		err := db.Close()
		if err != nil {
			log.Error("Failed to close db connection")
		}
	}()

	if err != nil {
		return "", teardown, err
	}

	// Run schema migrations here to setup the database schema for testing
	err = runMigrations(db, "../../../atlas/migrations")

	if err != nil {
		return "", teardown, err
	}

	// Return a teardown function to stop the container
	return dsn, teardown, nil
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
