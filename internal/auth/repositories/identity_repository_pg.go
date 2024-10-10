package repositories

import (
	"context"
	"github.com/oklog/ulid/v2"
	"identity-server/pkg/providers/database"
	"time"
)

type PostgresIdentityRepository struct {
	db *database.Db
}

type EmailIdentityInfoForLoginInternal struct {
	Email        string
	PasswordHash string
	UserId       string
	IdentityId   string
	LockedOut    bool
	Verified     bool
}

func NewPostgresIdentityRepository(db *database.Db) IdentityRepository {
	return &PostgresIdentityRepository{db: db}
}

func (r *PostgresIdentityRepository) GetEmailIdentityInfoForLogin(ctx context.Context, email string, now time.Time) (*EmailIdentityInfoForLogin, error) {
	var emailIdentityInfo EmailIdentityInfoForLoginInternal

	query := `
SELECT i.id, i.user_id, i.credential,  (u.lockout_end_date IS NOT NULL AND u.lockout_end_date < $2) AS locked_out, i.verified
                FROM user_identities i
                INNER JOIN users u ON i.user_id = u.id
                WHERE  i.value = $1 AND i.type = 'email'::identity_type AND u.deleted_at IS NULL 
                    AND i.deleted_at IS NULL
	`

	err := r.db.Db.QueryRowContext(ctx, query, email, now).Scan(&emailIdentityInfo.IdentityId, &emailIdentityInfo.UserId, &emailIdentityInfo.PasswordHash, &emailIdentityInfo.LockedOut, &emailIdentityInfo.Verified)

	if err != nil {
		return nil, err
	}

	return &EmailIdentityInfoForLogin{
		Email:        emailIdentityInfo.Email,
		PasswordHash: emailIdentityInfo.PasswordHash,
		UserId:       ulid.MustParse(emailIdentityInfo.UserId),
		IdentityId:   ulid.MustParse(emailIdentityInfo.IdentityId),
		LockedOut:    emailIdentityInfo.LockedOut,
		Verified:     emailIdentityInfo.Verified,
	}, nil
}
