package signup

import (
	"github.com/labstack/echo/v4"
	"github.com/oklog/ulid/v2"
	"identity-server/internal/accounts/messages/commands"
	"identity-server/internal/accounts/repositories"
	domain2 "identity-server/internal/domain"
	"identity-server/pkg/providers/hashing"
	"identity-server/pkg/providers/messaging"
	tprovider "identity-server/pkg/providers/time"
	"identity-server/pkg/security"
	"net/http"
)

type SignUpEmailReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func SignUp(accManager repositories.AccountRepository, timeProvider tprovider.Provider, hash hashing.Hasher, bus messaging.MessageBus, tokenMge *security.TokenManager) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req SignUpEmailReq
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, err)
		}

		exists, err := accManager.IdentityExists(c.Request().Context(), domain2.IdentityEmail.String(), req.Email)

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
		user := domain2.NewUser(ulid.Make(), req.Email, nil, now, now)
		identity := domain2.NewEmailIdentity(ulid.Make(), user.Id, req.Email, hashedPassword, now, now)

		if err := accManager.Save(c.Request().Context(), user, identity); err != nil {
			if err.Error() == "duplicated email" {
				return c.JSON(http.StatusBadRequest, err)
			}
			return c.JSON(http.StatusInternalServerError, err)
		}

		// Authenticate with limited scope just for email verification
		token, err := tokenMge.GenerateVerifyIdentityToken(user.Id, identity.Id)

		if err != nil {
			return c.JSON(http.StatusInternalServerError, err)
		}

		// Send verification email
		bus.Publish(c.Request().Context(), commands.SendVerificationEmail{
			Email:      identity.Value,
			IdentityId: identity.Id,
			UserId:     user.Id,
		})

		return c.JSON(http.StatusAccepted, token)
	}
}
