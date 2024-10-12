package signup

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"identity-server/pkg/providers"
	"identity-server/tests/fakes"
	"identity-server/tests/setup/containers"
	"identity-server/tests/setup/dependencies"
	"identity-server/tests/setup/respawn"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

var Deps *providers.DependencyContainer

func TestMain(m *testing.M) {
	appConfig, teardown := dependencies.SetupPostgresRedisConfig()
	Deps = providers.CreateDependencyContainer(appConfig)

	db, err := sql.Open("postgres", Deps.Config.Postgres.URL)
	if err != nil {
		log.Fatalf("failed to open connection to db: %s", err)
	}

	err = containers.RunMigrations(db, "../../../../atlas/migrations")
	if err != nil {
		log.Fatalf("failed to run migrations: %s", err)
	}

	Deps.Bus = &fakes.FakeMessageBus{}
	Deps.Logger = zap.NewNop()

	code := m.Run()

	Deps.Destroy()
	teardown()
	os.Exit(code)
}

func TestSignupEmailHandler(t *testing.T) {

	handler := SignUp(Deps.AccountRepo, Deps.TimeProvider, Deps.Hasher, Deps.Bus, Deps.TokenManager)

	respawner := respawn.NewPostgresRespawner([]string{"public"})

	e := echo.New()

	t.Run("SignUp_with_email_should_return_ok_when_data_is_valid", func(t *testing.T) {
		err := respawner.Respawn(Deps.Config.Postgres.URL)
		assert.NoError(t, err)
		reqData, _ := json.Marshal(SignUpEmailReq{Email: "test@testing.com", Password: "test-password"})
		req := httptest.NewRequest(http.MethodPost, "/sign-up/email", bytes.NewBuffer(reqData))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		if assert.NoError(t, handler(c)) {
			assert.Equal(t, http.StatusAccepted, rec.Code)
		}
	})

	t.Run("SignUp_with_email_should_not_allow_duplicated_emails", func(t *testing.T) {
		err := respawner.Respawn(Deps.Config.Postgres.URL)
		assert.NoError(t, err)

		reqData, _ := json.Marshal(SignUpEmailReq{Email: "test@testing.com", Password: "test-password"})
		req := httptest.NewRequest(http.MethodPost, "/sign-up/email", bytes.NewBuffer(reqData))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		// First req creates the user with email 'test@testing.com'
		assert.NoError(t, handler(c))
		assert.Equal(t, http.StatusAccepted, rec.Code)

		// Second req tries to create the user with the same email but receives a conflict status code
		req = httptest.NewRequest(http.MethodPost, "/sign-up/email", bytes.NewBuffer(reqData))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec = httptest.NewRecorder()
		c = e.NewContext(req, rec)
		if assert.NoError(t, handler(c)) {
			assert.Equal(t, http.StatusConflict, rec.Code)
		}
	})

	t.Run("SignUp_with_email_should_only_accept_valid_emails", func(t *testing.T) {
		err := respawner.Respawn(Deps.Config.Postgres.URL)
		assert.NoError(t, err)
		reqData, _ := json.Marshal(SignUpEmailReq{Email: "test", Password: "test-password"})
		req := httptest.NewRequest(http.MethodPost, "/sign-up/email", bytes.NewBuffer(reqData))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		if assert.NoError(t, handler(c)) {
			assert.Equal(t, http.StatusBadRequest, rec.Code)
		}
	})

	t.Run("SignUp_with_email_should_not_allow_password_with_less_than_8_chars", func(t *testing.T) {
		err := respawner.Respawn(Deps.Config.Postgres.URL)
		assert.NoError(t, err)
		reqData, _ := json.Marshal(SignUpEmailReq{Email: "test@testing.com", Password: "test"})
		req := httptest.NewRequest(http.MethodPost, "/sign-up/email", bytes.NewBuffer(reqData))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		if assert.NoError(t, handler(c)) {
			assert.Equal(t, http.StatusBadRequest, rec.Code)
		}
	})

}
