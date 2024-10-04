package repositories

import (
	"github.com/oklog/ulid/v2"
	domain2 "identity-server/internal/accounts/domain"
)

type AccountRepository interface {
	Save(user *domain2.User, identity *domain2.Identity) error
	IdentityExists(identityType string, value string) (bool, error)
	SetIdentityVerified(userId ulid.ULID, identityId ulid.ULID) error
}
