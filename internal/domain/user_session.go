package domain

import (
	"github.com/oklog/ulid/v2"
	"time"
)

type UserSession struct {
	UserId     ulid.ULID
	IdentityId ulid.ULID
	SessionId  ulid.ULID
	Device     *Device
	CreatedAt  time.Time
	ExpiresAt  time.Time
}

func NewUserSession(userId ulid.ULID, identityId ulid.ULID, sessionId ulid.ULID, device *Device, createdAt time.Time, expiresAt time.Time) *UserSession {
	return &UserSession{
		UserId:     userId,
		IdentityId: identityId,
		SessionId:  sessionId,
		Device:     device,
		CreatedAt:  createdAt,
		ExpiresAt:  expiresAt,
	}
}
