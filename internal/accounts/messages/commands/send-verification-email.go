package commands

import "github.com/oklog/ulid/v2"

type SendVerificationEmail struct {
	UserId     ulid.ULID
	IdentityId ulid.ULID
	Email      string
}
