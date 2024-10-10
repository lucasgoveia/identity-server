package security

import (
	"context"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/oklog/ulid/v2"
	"go.uber.org/zap"
	"identity-server/config"
	"identity-server/pkg/providers/cache"
	tprovider "identity-server/pkg/providers/time"
	"time"
)

type TokenManager struct {
	config       *config.AuthConfig
	timeProvider tprovider.Provider
	logger       *zap.Logger
	cache        cache.Cache
	rsaHolder    *RSAKeyHolder
}

func NewTokenManager(config *config.AuthConfig, timeProvider tprovider.Provider, logger *zap.Logger, cache cache.Cache, rsaHolder *RSAKeyHolder) *TokenManager {
	return &TokenManager{
		config:       config,
		timeProvider: timeProvider,
		logger:       logger,
		cache:        cache,
		rsaHolder:    rsaHolder,
	}
}

const (
	ClaimIssuer       = "iss"
	ClaimSubject      = "sub"
	ClaimCredentialId = "cid"
	ClaimAudience     = "aud"
	ClaimExpiration   = "exp"
	ClaimNotBefore    = "nbf"
	ClaimIssuedAt     = "iat"
	ClaimJWTID        = "jti"
	ClaimSessionId    = "sid"
)

func (m *TokenManager) GenerateVerifyIdentityToken(userId ulid.ULID, identityId ulid.ULID) (string, error) {
	now := m.timeProvider.UtcNow()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		ClaimSubject:      userId.String(),
		ClaimCredentialId: identityId.String(),
		ClaimIssuedAt:     now.Unix(),
		ClaimExpiration:   now.Add(time.Duration(m.config.CredentialVerificationConfig.LifetimeMinutes) * time.Minute).Unix(),
		ClaimJWTID:        ulid.Make().String(),
		ClaimNotBefore:    now.Unix(),
	})

	secret := []byte(m.config.CredentialVerificationConfig.Secret)
	tokenString, err := token.SignedString(secret)

	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (m *TokenManager) CheckVerifyIdentityToken(ctx context.Context, tokenString string) (jwt.MapClaims, error) {
	claims := jwt.MapClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(m.config.CredentialVerificationConfig.Secret), nil
	}, jwt.WithTimeFunc(m.timeProvider.UtcNow))

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, jwt.ErrSignatureInvalid
	}

	exists := m.cache.Exists(ctx, buildVerifyIdentityCacheKey(ulid.MustParse(claims[ClaimJWTID].(string))))

	if exists {
		return nil, jwt.ErrTokenInvalidId
	}

	return claims, nil
}

func buildVerifyIdentityCacheKey(tokenId ulid.ULID) string {
	return fmt.Sprintf("revoked-tokens:verify-identity:%s", tokenId.String())
}

func (m *TokenManager) RevokeVerifyIdentityToken(ctx context.Context, tokenId ulid.ULID) error {
	return m.cache.Set(ctx, buildVerifyIdentityCacheKey(tokenId), true, time.Minute*time.Duration(m.config.CredentialVerificationConfig.LifetimeMinutes))
}

func (m *TokenManager) GenerateAccessToken(userId ulid.ULID, identityId ulid.ULID, sessionId ulid.ULID, audience string) (string, error) {
	now := m.timeProvider.UtcNow()

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		ClaimSubject:      userId.String(),
		ClaimCredentialId: identityId.String(),
		ClaimAudience:     audience,
		ClaimIssuedAt:     now.Unix(),
		ClaimExpiration:   now.Add(time.Duration(m.config.AccessTokenConfig.LifetimeMinutes) * time.Minute).Unix(),
		ClaimJWTID:        ulid.MustNew(ulid.Timestamp(now), nil).String(),
		ClaimNotBefore:    now.Unix(),
		ClaimSessionId:    sessionId.String(),
	})

	tokenString, err := token.SignedString(m.rsaHolder.PrivateKey)

	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (m *TokenManager) GenerateRefreshToken(userId ulid.ULID, identityId ulid.ULID, sessionId ulid.ULID, audience string) (string, error) {
	now := m.timeProvider.UtcNow()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		ClaimSubject:      userId.String(),
		ClaimCredentialId: identityId.String(),
		ClaimSessionId:    sessionId.String(),
		ClaimAudience:     audience,
		ClaimIssuedAt:     now.Unix(),
		ClaimExpiration:   now.Add(time.Duration(m.config.SessionConfig.LifetimeHours) * time.Hour).Unix(),
		ClaimJWTID:        ulid.MustNew(ulid.Timestamp(now), nil).String(),
		ClaimNotBefore:    now.Unix(),
	})

	secret := []byte(m.config.RefreshTokenConfig.Secret)
	tokenString, err := token.SignedString(secret)

	if err != nil {
		return "", err
	}

	return tokenString, nil
}
