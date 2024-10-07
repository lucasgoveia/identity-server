package repositories

import (
	"context"
	"github.com/oklog/ulid/v2"
	"identity-server/internal/accounts/domain"
)

type AccountRepository interface {
	Save(ctx context.Context, user *domain.User, identity *domain.Identity) error
	IdentityExists(ctx context.Context, identityType string, value string) (bool, error)
	SetIdentityVerified(ctx context.Context, userId ulid.ULID, identityId ulid.ULID) error
}
