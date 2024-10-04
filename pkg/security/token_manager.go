package security

import (
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
}

func NewTokenManager(config *config.AuthConfig, timeProvider tprovider.Provider, logger *zap.Logger, cache cache.Cache) *TokenManager {
	return &TokenManager{
		config:       config,
		timeProvider: timeProvider,
		logger:       logger,
		cache:        cache,
	}
}

const (
	ClaimIssuer     = "iss"
	ClaimSubject    = "sub"
	ClaimIdentityId = "identity_id"
	ClaimAudience   = "aud"
	ClaimExpiration = "exp"
	ClaimNotBefore  = "nbf"
	ClaimIssuedAt   = "iat"
	ClaimJWTID      = "jti"
)

func (m *TokenManager) GenerateVerifyIdentityToken(userId ulid.ULID, identityId ulid.ULID, duration time.Duration) (string, error) {
	now := m.timeProvider.UtcNow()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		ClaimSubject:    userId.String(),
		ClaimIdentityId: identityId.String(),
		ClaimIssuedAt:   now.Unix(),
		ClaimExpiration: now.Add(duration).Unix(),
		ClaimJWTID:      ulid.Make().String(),
		ClaimNotBefore:  now.Unix(),
	})

	secret := []byte(m.config.VerificationJwtSecret)
	tokenString, err := token.SignedString(secret)

	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (m *TokenManager) CheckVerifyIdentityToken(tokenString string) (jwt.MapClaims, error) {
	claims := jwt.MapClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(m.config.VerificationJwtSecret), nil
	}, jwt.WithTimeFunc(m.timeProvider.UtcNow))

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, jwt.ErrSignatureInvalid
	}

	exists := m.cache.Exists(buildVerifyIdentityCacheKey(ulid.MustParse(claims[ClaimJWTID].(string))))

	if exists {
		return nil, jwt.ErrTokenInvalidId
	}

	return claims, nil
}

func buildVerifyIdentityCacheKey(tokenId ulid.ULID) string {
	return fmt.Sprintf("revoked-tokens:verify-identity:%s", tokenId.String())
}

func (m *TokenManager) RevokeVerifyIdentityToken(tokenId ulid.ULID) error {
	return m.cache.Set(buildVerifyIdentityCacheKey(tokenId), true, time.Hour*1)
}
