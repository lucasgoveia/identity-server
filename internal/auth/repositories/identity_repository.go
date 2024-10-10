package repositories

import (
	"context"
	"github.com/oklog/ulid/v2"
	"time"
)

type EmailIdentityInfoForLogin struct {
	Email        string
	PasswordHash string
	UserId       ulid.ULID
	IdentityId   ulid.ULID
	LockedOut    bool
	Verified     bool
}

type IdentityRepository interface {
	GetEmailIdentityInfoForLogin(ctx context.Context, email string, now time.Time) (*EmailIdentityInfoForLogin, error)
}
