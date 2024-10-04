package domain

import (
	"github.com/oklog/ulid/v2"
	"time"
)

type IdentityType string

const (
	IdentityEmail    IdentityType = "email"
	IdentityUsername IdentityType = "username"
	IdentityPhone    IdentityType = "phone"
	IdentitySocial   IdentityType = "social"
	IdentityB2B      IdentityType = "b2b"
	IdentityPasskey  IdentityType = "passkey"
)

// String method for IdentityType
func (i IdentityType) String() string {
	return string(i)
}

type Identity struct {
	Id         ulid.ULID
	UserId     ulid.ULID
	Type       IdentityType
	Value      string
	Credential string
	Provider   *string
	Verified   bool
	CreatedAt  time.Time
	UpdatedAt  time.Time
	DeletedAt  *time.Time
}

func NewEmailIdentity(id ulid.ULID, userId ulid.ULID, email string, password string, createdAt time.Time, updatedAt time.Time) *Identity {
	return &Identity{
		Id:         id,
		UserId:     userId,
		Type:       IdentityEmail,
		Value:      email,
		Credential: password,
		Provider:   nil,
		Verified:   false,
		CreatedAt:  createdAt,
		UpdatedAt:  updatedAt,
		DeletedAt:  nil,
	}
}
