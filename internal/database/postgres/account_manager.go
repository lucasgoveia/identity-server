package postgres

import (
	"github.com/Masterminds/squirrel"
	"identity-server/internal/domain/entities"
	"identity-server/pkg/providers/database/postgres"
	"log"
)

type AccountManager struct {
	db *postgres.Db
}

func NewAccountManager(db *postgres.Db) *AccountManager {
	return &AccountManager{db: db}
}

func (r *AccountManager) Save(user *entities.User, identity *entities.Identity) error {
	tx, err := r.db.Db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			err := tx.Rollback()
			if err != nil {
				log.Fatalf("Transaction rollback failed: %v", err)
			}
			log.Fatalf("Transaction rolled back due to error: %v", err)
		}
		err = tx.Commit()
		if err != nil {
			log.Fatalf("failed to commit transaction: %v", err)
		}
	}()

	insertCmd, args, err := squirrel.Insert("users").
		Columns("id", "name", "avatar_url", "created_at", "updated_at").
		Values(user.Id, user.Name, user.AvatarUrl, user.CreatedAt, user.UpdatedAt).
		ToSql()
	_, err = tx.Exec(insertCmd, args...)

	if err != nil {
		log.Fatalf("failed to insert user: %v", err)
	}

	insertCmd, args, err = squirrel.Insert("users_identities").
		Columns("id", "user_id", "type", "value", "credential", "provider", "verified", "created_at", "updated_at").
		Values(identity.Id, identity.UserId, identity.Type, identity.Value, identity.Credential, identity.Provider, identity.Verified, identity.CreatedAt, identity.UpdatedAt).
		ToSql()

	_, err = tx.Exec(insertCmd, args...)

	if err != nil {
		log.Fatalf("failed to insert identity: %v", err)
	}

	return nil
}
