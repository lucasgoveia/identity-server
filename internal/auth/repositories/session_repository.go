package repositories

import (
	"context"
	"identity-server/internal/domain"
)

type SessionRepository interface {
	Save(ctx context.Context, session *domain.UserSession) error
}
