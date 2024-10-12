package accounts

import (
	"bytes"
	"encoding/json"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"identity-server/internal/accounts/handlers/signup"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSignupEmailHandler(t *testing.T) {

	handler := signup.SignUp(Deps.AccountRepo, Deps.TimeProvider, Deps.Hasher, Deps.Bus, Deps.TokenManager)

	e := echo.New()
	rec := httptest.NewRecorder()

	t.Run("SignUp_with_email_should_return_ok_when_data_is_valid", func(t *testing.T) {
		reqData, _ := json.Marshal(signup.SignUpEmailReq{Email: "test@testing.com", Password: "test-password"})
		req := httptest.NewRequest(http.MethodPost, "/sign-up/email", bytes.NewBuffer(reqData))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		c := e.NewContext(req, rec)

		if assert.NoError(t, handler(c)) {
			assert.Equal(t, http.StatusAccepted, rec.Code)
		}
	})
}
