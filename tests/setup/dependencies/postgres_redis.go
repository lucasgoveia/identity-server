package dependencies

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"identity-server/config"
	"identity-server/tests/setup/containers"
	"log"
)

func encodePrivateKeyToBase64(privateKey *rsa.PrivateKey) string {
	privBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	return base64.StdEncoding.EncodeToString(privBytes)
}

func encodePublicKeyToBase64(publicKey *rsa.PublicKey) string {
	pubBytes := x509.MarshalPKCS1PublicKey(publicKey)
	return base64.StdEncoding.EncodeToString(pubBytes)
}

func SetupPostgresRedisConfig() (*config.AppConfig, func()) {
	dbconn, pgTeardown, err := containers.SetupTestPostgresDb()

	if err != nil {
		if pgTeardown != nil {
			pgTeardown()
		}
		log.Fatal("Failed to setup postgres database")
	}

	rdConn, redisTeardown, err := containers.SetupTestRedis()

	if err != nil {
		if redisTeardown != nil {
			redisTeardown()
		}
		log.Fatal("Failed to setup redis")
	}

	teardown := func() {
		pgTeardown()
		redisTeardown()
	}

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)

	if err != nil {
		teardown()
		log.Fatalf("error generating RSA key: %v", err)
	}

	privateKeyB64 := encodePrivateKeyToBase64(privateKey)
	publicKeyB64 := encodePublicKeyToBase64(&privateKey.PublicKey)

	appconfig := config.AppConfig{
		Server: &config.ServerConfig{Host: "test", Port: 80},
		Database: &config.DatabaseConfig{
			Provider: "postgres",
		},
		Hashing:  &config.HashingConfig{Algorithm: "argon2"},
		Postgres: &config.PostgresConfig{URL: dbconn},
		Mailer:   &config.MailerConfig{Provider: "stub"},
		Smtp:     nil,
		Cache:    &config.CacheConfig{Provider: "redis"},
		Redis:    rdConn,
		Auth: &config.AuthConfig{
			CredentialVerificationConfig: &config.CredentialVerificationConfig{LifetimeMinutes: 30, Secret: "my-credential-verification-test-secret"},
			SessionConfig: &config.SessionConfig{
				LifetimeHours:        24,
				TrustedLifetimeHours: 720,
			},
			AccessTokenConfig: &config.AccessTokenConfig{
				PublicKey:       publicKeyB64,
				PrivateKey:      privateKeyB64,
				LifetimeMinutes: 5,
				Issuer:          "testing",
			},
			RefreshTokenConfig: &config.RefreshTokenConfig{Secret: "my-refresh-token-test-secret"},
		},
	}

	return &appconfig, teardown
}
