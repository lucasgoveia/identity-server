package postgres

import (
	"errors"
	"fmt"
	"github.com/Masterminds/squirrel"
	"github.com/lib/pq"
	"identity-server/internal/domain/entities"
	"identity-server/pkg/providers/database/postgres"
)

type AccountManager struct {
	db *postgres.Db
}

func NewAccountManager(db *postgres.Db) *AccountManager {
	return &AccountManager{db: db}
}

func (r *AccountManager) Save(user *entities.User, identity *entities.Identity) (err error) {
	tx, err := r.db.Db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			err := tx.Rollback()
			if err != nil {
				err = fmt.Errorf("transaction rollback failed: %v", err)
			}
			err = fmt.Errorf("transaction rolled back due to error: %v", err)
		}
		err = tx.Commit()
		if err != nil {
			err = errors.New(fmt.Sprintf("failed to commit transaction: %v", err))
		}
	}()

	psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	insertCmd, args, err := psql.Insert("users").
		Columns("id", "name", "avatar_link", "created_at", "updated_at").
		Values(user.Id.String(), user.Name, user.AvatarLink, user.CreatedAt, user.UpdatedAt).
		ToSql()
	_, err = tx.Exec(insertCmd, args...)

	if err != nil {
		return errors.New(fmt.Sprintf("failed to insert user: %v", err))
	}

	insertCmd, args, err = psql.Insert("user_identities").
		Columns("id", "user_id", "type", "value", "credential", "provider", "verified", "created_at", "updated_at").
		Values(identity.Id.String(), identity.UserId.String(), identity.Type, identity.Value, identity.Credential, identity.Provider, identity.Verified, identity.CreatedAt, identity.UpdatedAt).
		ToSql()

	_, err = tx.Exec(insertCmd, args...)

	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			if pqErr.Code.Name() == "unique_violation" {
				return fmt.Errorf("duplicated email: %w", err)
			}
		}
		return errors.New(fmt.Sprintf("failed to insert identity: %v", err))
	}

	return nil
}
