package exchange

import (
	"github.com/labstack/echo/v4"
	"identity-server/internal/auth/services"
	"net/http"
)

type ExchangeTokenData struct {
	CodeVerifier string `json:"code_verifier"`
	Code         string `json:"code"`
	RedirectUri  string `json:"redirect_uri"`
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func Token(authServ *services.AuthService) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req ExchangeTokenData
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, err)
		}

		device := services.IdentifyDevice(c.Request())
		aud := c.Request().Header.Get("Origin")
		res, err := authServ.Authenticate(c.Request().Context(), device, aud, req.Code, req.CodeVerifier, req.RedirectUri)

		if err != nil {
			return c.JSON(http.StatusUnauthorized, err)
		}

		return c.JSON(http.StatusOK, TokenResponse{
			AccessToken:  res.AccessToken,
			RefreshToken: res.RefreshToken,
		})
	}
}
