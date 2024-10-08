package domain

import (
	"github.com/oklog/ulid/v2"
	"time"
)

type User struct {
	Id         ulid.ULID
	Name       string
	AvatarLink *string
	CreatedAt  time.Time
	UpdatedAt  time.Time
	DeletedAt  *time.Time
}

func NewUser(id ulid.ULID, name string, avatarUrl *string, createdAt time.Time, updatedAt time.Time) *User {
	return &User{
		Id:         id,
		Name:       name,
		AvatarLink: avatarUrl,
		CreatedAt:  createdAt,
		UpdatedAt:  updatedAt,
		DeletedAt:  nil,
	}
}
