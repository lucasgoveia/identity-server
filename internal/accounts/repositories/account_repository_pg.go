package repositories

import (
	"context"
	"errors"
	"fmt"
	"github.com/Masterminds/squirrel"
	"github.com/lib/pq"
	"github.com/oklog/ulid/v2"
	domain2 "identity-server/internal/domain"
	"identity-server/pkg/providers/database"
)

type PostgresAccountRepository struct {
	db *database.Db
}

func NewPostgresAccountRepository(db *database.Db) AccountRepository {
	return &PostgresAccountRepository{db: db}
}

func (r *PostgresAccountRepository) Save(ctx context.Context, user *domain2.User, identity *domain2.Identity) (err error) {
	tx, err := r.db.Db.BeginTx(ctx, nil)
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
	_, err = tx.ExecContext(ctx, insertCmd, args...)

	if err != nil {
		return errors.New(fmt.Sprintf("failed to insert user: %v", err))
	}

	insertCmd, args, err = psql.Insert("user_identities").
		Columns("id", "user_id", "type", "value", "credential", "provider", "verified", "created_at", "updated_at").
		Values(identity.Id.String(), identity.UserId.String(), identity.Type, identity.Value, identity.Credential, identity.Provider, identity.Verified, identity.CreatedAt, identity.UpdatedAt).
		ToSql()

	_, err = tx.ExecContext(ctx, insertCmd, args...)

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

func (r *PostgresAccountRepository) IdentityExists(ctx context.Context, identityType string, value string) (bool, error) {
	var exists bool
	err := r.db.Db.QueryRowContext(ctx, "SELECT EXISTS(select 1 from user_identities where type = $1 and value = $2)", identityType, value).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

func (r *PostgresAccountRepository) SetIdentityVerified(ctx context.Context, userId ulid.ULID, identityId ulid.ULID) error {
	_, err := r.db.Db.ExecContext(ctx, "UPDATE user_identities SET verified = true WHERE user_id = $1 AND id = $2", userId.String(), identityId.String())
	if err != nil {
		return err
	}
	return nil
}
