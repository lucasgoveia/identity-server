package email

import (
	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
	"github.com/oklog/ulid/v2"
	"identity-server/internal/app/accounts"
	"identity-server/internal/domain/entities"
	"identity-server/internal/messages/commands"
	"identity-server/pkg/providers/hashing"
	"identity-server/pkg/providers/messaging"
	tprovider "identity-server/pkg/providers/time"
	"net/http"
	"time"
)

type SignUpEmailReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

const (
	ClaimIssuer     = "iss"
	ClaimSubject    = "sub"
	ClaimAudience   = "aud"
	ClaimExpiration = "exp"
	ClaimNotBefore  = "nbf"
	ClaimIssuedAt   = "iat"
	ClaimJWTID      = "jti"
)

func SignUp(accManager accounts.AccountManager, timeProvider tprovider.Provider, hash hashing.Hasher, bus messaging.MessageBus) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req SignUpEmailReq
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, err)
		}

		exists, err := accManager.IdentityExists(entities.IdentityEmail.String(), req.Email)

		if err != nil {
			return c.JSON(http.StatusInternalServerError, err)
		}
		if exists {
			return c.JSON(http.StatusConflict, "Email already in use")
		}

		hashedPassword, err := hash.Hash(req.Password)

		if err != nil {
			return c.JSON(http.StatusInternalServerError, err)
		}

		now := timeProvider.UtcNow()
		user := entities.NewUser(ulid.Make(), req.Email, nil, now, now)
		identity := entities.NewEmailIdentity(ulid.Make(), user.Id, req.Email, hashedPassword, now, now)

		if err := accManager.Save(user, identity); err != nil {
			if err.Error() == "duplicated email" {
				return c.JSON(http.StatusBadRequest, err)
			}
			return c.JSON(http.StatusInternalServerError, err)
		}

		// Authenticate with limited scope just for email verification
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			ClaimSubject:    user.Id.String(),
			ClaimIssuedAt:   now,
			ClaimExpiration: now.Add(2 * time.Hour),
			ClaimJWTID:      ulid.Make().String(),
		})

		secret := []byte("very-hard-to-guess-secret")
		tokenString, err := token.SignedString(secret)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, err)
		}

		// Send verification email
		bus.Publish(commands.SendVerificationEmail{
			Email:      identity.Value,
			IdentityId: identity.Id,
			UserId:     user.Id,
		})

		return c.JSON(http.StatusAccepted, tokenString)
	}
}
