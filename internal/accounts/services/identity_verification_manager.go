package services

import (
	"errors"
	"fmt"
	"github.com/oklog/ulid/v2"
	"go.uber.org/zap"
	"identity-server/pkg/providers/cache"
	"identity-server/pkg/providers/hashing"
	"identity-server/pkg/security"
	"time"
)

type IdentityVerificationManager struct {
	otpGen *security.OTPGenerator
	cache  cache.Cache
	hasher hashing.Hasher
	logger *zap.Logger
}

func NewIdentityVerificationManager(otpGenerator *security.OTPGenerator, cache cache.Cache, hasher hashing.Hasher, logger *zap.Logger) *IdentityVerificationManager {
	return &IdentityVerificationManager{otpGen: otpGenerator, cache: cache, hasher: hasher, logger: logger}
}

func buildOtpCacheKey(userId ulid.ULID, identityId ulid.ULID) string {
	return fmt.Sprintf("users:%s:identities:%s:email", userId.String(), identityId.String())
}

func (m *IdentityVerificationManager) GenerateEmailOTP(userId ulid.ULID, identityId ulid.ULID) (string, error) {
	otp, err := m.otpGen.GenerateOTP()

	if err != nil {
		m.logger.Error("Failed to generate OTP", zap.Error(err))
		return "", err
	}

	cacheKey := buildOtpCacheKey(userId, identityId)

	hashedOtp, err := m.hasher.Hash(otp)

	if err != nil {
		m.logger.Error("Failed to hash OTP", zap.Error(err))
		return "", err
	}

	err = m.cache.Set(cacheKey, hashedOtp, time.Hour*1)

	if err != nil {
		m.logger.Error("Failed to set verification code to cache", zap.Error(err))
		return "", err
	}

	return otp, nil
}

func (m *IdentityVerificationManager) VerifyEmailOtp(userId ulid.ULID, identityId ulid.ULID, code string) (bool, error) {
	cacheKey := buildOtpCacheKey(userId, identityId)

	hashedOtp, exists := m.cache.Get(cacheKey)

	if !exists {
		m.logger.Error("Failed to get verification code from cache")
		return false, errors.New("code not found")
	}

	verified, err := m.hasher.Verify(code, hashedOtp.(string))
	if err != nil {
		m.logger.Error("Failed to verify code", zap.Error(err))
		return false, err
	}

	return verified, nil
}

func (m *IdentityVerificationManager) RevokeEmailOtp(userId ulid.ULID, identityId ulid.ULID) error {
	cacheKey := buildOtpCacheKey(userId, identityId)
	err := m.cache.Remove(cacheKey)
	if err != nil {
		return err
	}

	return nil
}
