package middlewares

import (
	"github.com/labstack/echo/v4"
	"github.com/oklog/ulid/v2"
	"identity-server/pkg/security"
)

type LoggedInUser struct {
	UserId     ulid.ULID
	IdentityId ulid.ULID
	TokenId    ulid.ULID
}

func VerifyIdentityAuth(tokenMge *security.TokenManager) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {

			// Extract token from request
			token := c.Request().Header.Get("Authorization")

			if len(token) > 7 {
				token = token[7:]
			} else {
				return c.JSON(401, "Unauthorized")
			}

			// Check if token is valid
			claims, err := tokenMge.CheckVerifyIdentityToken(c.Request().Context(), token)

			if err != nil {
				return c.JSON(401, "Unauthorized")
			}

			// Set user id and identity id in context
			c.Set("user", LoggedInUser{
				UserId:     ulid.MustParse(claims[security.ClaimSubject].(string)),
				IdentityId: ulid.MustParse(claims[security.ClaimIdentityId].(string)),
			})

			return next(c)
		}
	}
}
