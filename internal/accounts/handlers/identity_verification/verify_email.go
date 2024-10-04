package identity_verification

import (
	"github.com/labstack/echo/v4"
	"identity-server/internal/accounts/repositories"
	"identity-server/internal/accounts/services"
	"identity-server/pkg/middlewares"
	"identity-server/pkg/security"
	"net/http"
)

type VerifyEmailReq struct {
	Code string `json:"code"`
}

func VerifyEmail(accManager repositories.AccountRepository, tokenMge *security.TokenManager, verificationManager *services.IdentityVerificationManager) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req VerifyEmailReq
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, err)
		}

		// Get user and identity id from context
		user := c.Get("user").(middlewares.LoggedInUser)

		// Check if code is valid
		verified, err := verificationManager.VerifyEmailOtp(user.UserId, user.IdentityId, req.Code)

		if err != nil {
			return c.JSON(http.StatusInternalServerError, err)
		}

		if !verified {
			return c.JSON(http.StatusUnauthorized, "Invalid code")
		}

		err = accManager.SetIdentityVerified(user.UserId, user.IdentityId)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, err)
		}

		err = verificationManager.RevokeEmailOtp(user.UserId, user.IdentityId)
		if err != nil {
			// FIXME: Retry?
			return c.JSON(http.StatusInternalServerError, err)
		}

		err = tokenMge.RevokeVerifyIdentityToken(user.TokenId)
		if err != nil {
			// FIXME: Retry?
			return c.JSON(http.StatusInternalServerError, err)
		}

		return c.JSON(http.StatusOK, "Email verified")
	}
}
