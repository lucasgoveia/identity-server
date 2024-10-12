package repositories

import (
	"context"
	"database/sql"
	"errors"
	"github.com/labstack/gommon/log"
	domain2 "identity-server/internal/domain"
	"identity-server/pkg/providers/database"
	"identity-server/tests/setup/containers"
	"testing"
	"time"

	"github.com/lib/pq"
	"github.com/oklog/ulid/v2"
	"github.com/stretchr/testify/assert"
)

func setupDb(t *testing.T) (*sql.DB, func()) {
	dbConn, teardown, err := containers.SetupTestPostgresDb()

	assert.NoError(t, err)

	db, err := sql.Open("postgres", dbConn)

	assert.NoError(t, err)

	return db, func() {
		err := db.Close()
		if err != nil {
			log.Error("Failed to close db connection")
		}
		teardown()
	}
}

func TestAccountManager_Save(t *testing.T) {
	db, teardown := setupDb(t)
	defer teardown()

	accountManager := NewPostgresAccountRepository(&database.Db{Db: db})

	t.Run("Successfully save user and identity", func(t *testing.T) {
		ctx := context.Background()
		userId := ulid.Make()
		user := domain2.NewUser(userId, "John Doe", nil, time.Now(), time.Now())

		identityId := ulid.Make()
		identity := domain2.NewEmailIdentity(identityId, userId, "johndoe@example.com", "hashed-password", time.Now(), time.Now())

		err := accountManager.Save(ctx, user, identity)
		assert.NoError(t, err, "Saving user and identity should not return an error")
	})

	t.Run("Duplicate email", func(t *testing.T) {
		ctx := context.Background()
		userId := ulid.Make()
		user := domain2.NewUser(userId, "Jane Doe", nil, time.Now(), time.Now())

		identityId := ulid.Make()
		identity := domain2.NewEmailIdentity(identityId, userId, "johndoe@example.com", "hashed-password", time.Now(), time.Now()) // Duplicate email

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
		user := domain2.NewUser(userId, "User With Error", nil, time.Now(), time.Now())

		identityId := ulid.Make()
		identity := domain2.NewEmailIdentity(identityId, userId, "error@example.com", "hashed-password", time.Now(), time.Now())

		err = accountManager.Save(ctx, user, identity)
		assert.Error(t, err, "Inserting user with error should return an error")
	})

}

func TestAccountManager_IdentityExists(t *testing.T) {
	db, teardown := setupDb(t)
	defer teardown()

	accountManager := NewPostgresAccountRepository(&database.Db{Db: db})

	t.Run("Identity exists", func(t *testing.T) {
		ctx := context.Background()
		// Create a user and identity first
		userId := ulid.Make()
		user := domain2.NewUser(userId, "John Doe", nil, time.Now(), time.Now())

		identityId := ulid.Make()
		identity := domain2.NewEmailIdentity(identityId, userId, "existing@example.com", "hashed-password", time.Now(), time.Now())

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
