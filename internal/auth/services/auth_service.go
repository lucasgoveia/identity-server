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
	pcke          *PCKEManager
}

func NewAuthService(logger *zap.Logger, tokenManager *security.TokenManager, sessionRepo repositories.SessionRepository, timeProvider timeProvider.Provider, sessionConfig *config.SessionConfig, pcke *PCKEManager) *AuthService {
	return &AuthService{
		logger:        logger,
		tokenManager:  tokenManager,
		sessionRepo:   sessionRepo,
		timeProvider:  timeProvider,
		sessionConfig: sessionConfig,
		pcke:          pcke,
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

func (a *AuthService) InitiateAuthentication(ctx context.Context, userId ulid.ULID, identityId ulid.ULID, rememberMe bool, codeChallenge string, codeChallengeMethod string, redirectUri string) (string, error) {
	return a.pcke.New(ctx, userId, identityId, codeChallenge, codeChallengeMethod, redirectUri, rememberMe)
}

func (a *AuthService) Authenticate(ctx context.Context, device *domain.Device, aud string, code string, codeVerifier string, redirectUri string) (*AuthenticateResponse, error) {
	res, err := a.pcke.Exchange(ctx, code, codeVerifier, redirectUri)

	if err != nil {
		return nil, err
	}

	sessionId := ulid.Make()
	now := a.timeProvider.UtcNow()
	session := domain.NewUserSession(res.UserId, res.CredentialId, sessionId, device, a.timeProvider.UtcNow(), now.Add(a.getSessionDuration(res.RememberMe)))

	err = a.sessionRepo.Save(ctx, session)

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
