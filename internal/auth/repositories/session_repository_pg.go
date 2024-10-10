package repositories

import (
	"context"
	"github.com/Masterminds/squirrel"
	"identity-server/internal/domain"
	"identity-server/pkg/providers/database"
)

type PostgresSessionRepository struct {
	db *database.Db
}

func NewPostgresSessionRepository(db *database.Db) SessionRepository {
	return &PostgresSessionRepository{db: db}
}

func (r *PostgresSessionRepository) Save(ctx context.Context, session *domain.UserSession) error {
	psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	insertCmd, args, err := psql.Insert("user_sessions").
		Columns("session_id", "user_id", "identity_id", "ip_address", "user_agent", "created_at", "expires_at").
		Values(session.SessionId.String(), session.UserId.String(), session.IdentityId.String(), session.Device.IpAddress, session.Device.UserAgent, session.CreatedAt, session.ExpiresAt).
		ToSql()

	if err != nil {
		return err
	}

	_, err = r.db.Db.ExecContext(ctx, insertCmd, args...)

	return err
}
