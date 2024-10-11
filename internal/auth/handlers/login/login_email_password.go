package login

import (
	"github.com/labstack/echo/v4"
	"identity-server/internal/auth/repositories"
	"identity-server/internal/auth/services"
	"identity-server/pkg/providers/hashing"
	timeProvider "identity-server/pkg/providers/time"
	"net/http"
)

type EmailPasswordLogin struct {
	Email      string `json:"email"`
	Password   string `json:"password"`
	RememberMe bool   `json:"remember_me"`
}

type Response struct {
	Code string `json:"code"`
}

func Login(repo repositories.IdentityRepository, hash hashing.Hasher, timeProvider timeProvider.Provider, authServ *services.AuthService) echo.HandlerFunc {
	return func(c echo.Context) error {
		codeChallenge := c.QueryParam("code_challenge")
		codeChallengeMethod := c.QueryParam("code_challenge_method")
		redirectUri := c.QueryParam("redirect_uri")

		if codeChallenge == "" || codeChallengeMethod == "" || redirectUri == "" {
			return c.JSON(http.StatusBadRequest, "Missing required parameters")
		}

		var req EmailPasswordLogin
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, err)
		}

		info, err := repo.GetEmailIdentityInfoForLogin(c.Request().Context(), req.Email, timeProvider.UtcNow())

		if err != nil {
			return c.JSON(http.StatusInternalServerError, err)
		}

		if info.LockedOut {
			return c.JSON(http.StatusUnauthorized, "Account is locked")
		}

		if !info.Verified {
			return c.JSON(http.StatusUnauthorized, "Account is not verified")
		}

		verified, err := hash.Verify(req.Password, info.PasswordHash)
		if err != nil {
			return err
		}

		if !verified {
			return c.JSON(http.StatusUnauthorized, "Invalid email or password")
		}

		code, err := authServ.InitiateAuthentication(c.Request().Context(), info.UserId, info.IdentityId, req.RememberMe, codeChallenge, codeChallengeMethod, redirectUri)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, Response{
			Code: code,
		})
	}
}
