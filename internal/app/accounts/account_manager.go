package accounts

import "identity-server/internal/domain/entities"

type AccountManager interface {
	Save(user *entities.User, identity *entities.Identity) error
	IdentityExists(identityType string, value string) (bool, error)
}
