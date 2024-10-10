package services

import (
	"context"
	"github.com/oklog/ulid/v2"
	"go.uber.org/zap"
	"identity-server/config"
	"identity-server/internal/auth/repositories"
	"identity-server/internal/domain"
	timeProvider "identity-server/pkg/providers/time"
	"identity-server/pkg/security"
	"time"
)

type AuthService struct {
	logger        *zap.Logger
	tokenManager  *security.TokenManager
	sessionRepo   repositories.SessionRepository
	timeProvider  timeProvider.Provider
	sessionConfig *config.SessionConfig
}

func NewAuthService(logger *zap.Logger, tokenManager *security.TokenManager, sessionRepo repositories.SessionRepository, timeProvider timeProvider.Provider, sessionConfig *config.SessionConfig) *AuthService {
	return &AuthService{
		logger:        logger,
		tokenManager:  tokenManager,
		sessionRepo:   sessionRepo,
		timeProvider:  timeProvider,
		sessionConfig: sessionConfig,
	}
}

type AuthenticateResponse struct {
	AccessToken  string
	RefreshToken string
}

func (a *AuthService) getSessionDuration(rememberMe bool) time.Duration {
	if rememberMe {
		return time.Duration(a.sessionConfig.TrustedLifetimeHours) * time.Hour
	}
	return time.Duration(a.sessionConfig.LifetimeHours) * time.Hour
}

func (a *AuthService) Authenticate(ctx context.Context, userId ulid.ULID, identityId ulid.ULID, device *domain.Device, rememberMe bool, aud string) (*AuthenticateResponse, error) {
	sessionId := ulid.Make()
	now := a.timeProvider.UtcNow()
	session := domain.NewUserSession(userId, identityId, sessionId, device, a.timeProvider.UtcNow(), now.Add(a.getSessionDuration(rememberMe)))

	err := a.sessionRepo.Save(ctx, session)

	if err != nil {
		return nil, err
	}

	accessToken, err := a.tokenManager.GenerateAccessToken(session.UserId, session.IdentityId, sessionId, aud)

	if err != nil {
		a.logger.Error("Failed to generate access token", zap.Error(err))
		return nil, err
	}

	refreshToken, err := a.tokenManager.GenerateRefreshToken(session.UserId, session.IdentityId, sessionId, aud)

	if err != nil {
		a.logger.Error("Failed to generate refresh token", zap.Error(err))
		return nil, err
	}

	return &AuthenticateResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}
