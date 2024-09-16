package email

import (
	"github.com/labstack/echo/v4"
	"github.com/oklog/ulid/v2"
	"identity-server/internal/app/accounts"
	"identity-server/internal/domain/entities"
	"identity-server/pkg/providers/hashing"
	"identity-server/pkg/providers/time"
	"net/http"
)

type SignUpEmailReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func SignUp(accManager accounts.AccountManager, timeProvider time.Provider, hash hashing.Hasher) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req SignUpEmailReq
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, err)
		}

		hashedPassword, err := hash.Hash(req.Password)

		if err != nil {
			return c.JSON(http.StatusInternalServerError, err)
		}

		user := entities.NewUser(ulid.Make(), req.Email, nil, timeProvider.Now(), timeProvider.Now())
		identity := entities.NewEmailIdentity(ulid.Make(), user.Id, req.Email, hashedPassword, timeProvider.Now(), timeProvider.Now())

		if err := accManager.Save(user, identity); err != nil {
			if err.Error() == "duplicated email" {
				return c.JSON(http.StatusBadRequest, err)
			}
			return c.JSON(http.StatusInternalServerError, err)
		}

		// Authenticate with limited scope just for email verification
		// Send verification email
		// verification will be handled by another route
		// After verification, user will be able to log in with full scope

		return c.JSON(http.StatusAccepted, nil)
	}
}
