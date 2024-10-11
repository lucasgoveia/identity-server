package providers

import (
	"fmt"
	"go.uber.org/zap"
	"identity-server/config"
	accRepos "identity-server/internal/accounts/repositories"
	accServices "identity-server/internal/accounts/services"
	authRepos "identity-server/internal/auth/repositories"
	authServices "identity-server/internal/auth/services"
	"identity-server/pkg/providers/cache"
	"identity-server/pkg/providers/database"
	"identity-server/pkg/providers/hashing"
	"identity-server/pkg/providers/mailing"
	"identity-server/pkg/providers/messaging"
	"identity-server/pkg/providers/time"
	"identity-server/pkg/security"
	"log"
)

type DependencyContainer struct {
	Database                    database.Database
	Logger                      *zap.Logger
	Hasher                      hashing.Hasher
	TimeProvider                time.Provider
	Bus                         messaging.MessageBus
	Mailer                      mailing.Sender
	SecureKeyGen                *security.SecureKeyGenerator
	OTPGen                      *security.OTPGenerator
	RsaHolder                   *security.RSAKeyHolder
	TokenManager                *security.TokenManager
	pckeManager                 *authServices.PCKEManager
	IdentityVerificationManager *accServices.IdentityVerificationManager
	AccountRepo                 accRepos.AccountRepository
	SessionRepo                 authRepos.SessionRepository
	IdentityRepo                authRepos.IdentityRepository
	AuthService                 *authServices.AuthService
	Config                      *config.AppConfig
}

func (c *DependencyContainer) Destroy() {
	if err := c.Database.Close(); err != nil {
		log.Fatalf("Failed to close database: %v", err)
	}
	if err := c.Logger.Sync(); err != nil {
		log.Fatalf("failed to flush logs on shutdown %s", err)
	}

	c.Bus.Stop()
}

func CreateDependencyContainer(config *config.AppConfig) *DependencyContainer {

	logger, err := zap.NewDevelopment()
	db, err := CreateDatabase(config)

	if err != nil {
		log.Fatalf("Failed to create database: %v", err)
	}

	accRepo, err := CreateAccountRepository(db)
	timeProvider := CreateDefaultTimeProvider()
	hasher, err := CreateHasher(config)
	bus := CreateMessageBus(logger)

	mailer := CreateMailSender(config, logger)

	cacher, err := CreateCache(config)

	secureKeyGen := security.NewSecureKeyGenerator()

	otpGen := security.NewOTPGenerator(secureKeyGen)

	identityVerificationManager := accServices.NewIdentityVerificationManager(otpGen, cacher, hasher, logger, config.Auth.CredentialVerificationConfig)

	rsaHolder, err := security.NewRSAKeyHolder(config.Auth.AccessTokenConfig.PrivateKey, config.Auth.AccessTokenConfig.PublicKey)

	tokenManager := security.NewTokenManager(config.Auth, timeProvider, logger, cacher, rsaHolder)

	sessionRepo, err := CreateSessionRepository(db)
	identityRepo, err := CreateIdentityRepository(db)

	pcke := authServices.NewPCKEManager(secureKeyGen, cacher)

	authService := authServices.NewAuthService(logger, tokenManager, sessionRepo, timeProvider, config.Auth.SessionConfig, pcke)

	return &DependencyContainer{
		Config:                      config,
		Database:                    db,
		AccountRepo:                 accRepo,
		AuthService:                 authService,
		Bus:                         bus,
		Hasher:                      hasher,
		IdentityRepo:                identityRepo,
		IdentityVerificationManager: identityVerificationManager,
		TokenManager:                tokenManager,
		Logger:                      logger,
		OTPGen:                      otpGen,
		SecureKeyGen:                secureKeyGen,
		pckeManager:                 pcke,
		RsaHolder:                   rsaHolder,
		SessionRepo:                 sessionRepo,
		TimeProvider:                timeProvider,
		Mailer:                      mailer,
	}
}

func CreateDatabase(config *config.AppConfig) (database.Database, error) {
	switch config.Database.Provider {
	case "postgres":
		return database.NewPostgresDb(config.Postgres.URL)
	default:
		return nil, fmt.Errorf("unsupported database provider: %s", config.Database.Provider)
	}
}

func CreateHasher(config *config.AppConfig) (hashing.Hasher, error) {
	switch config.Hashing.Algorithm {
	case "argon2":
		return &hashing.Argon2Hasher{}, nil
	default:
		return nil, fmt.Errorf("unsupported hashing algorithm: %s", config.Hashing.Algorithm)
	}
}

func CreateDefaultTimeProvider() time.Provider {
	return &time.DefaultTimeProvider{}
}

func CreateAccountRepository(db database.Database) (accRepos.AccountRepository, error) {
	switch db.GetProviderType() {
	case "postgres":
		return accRepos.NewPostgresAccountRepository(db.(*database.Db)), nil
	default:
		return nil, fmt.Errorf("unsupported database provider: %s", db.GetProviderType())
	}
}

func CreateCache(config *config.AppConfig) (cache.Cache, error) {
	switch config.Cache.Provider {
	case "inmemory":
		return cache.NewInMemory(), nil
	case "redis":
		return cache.NewRedisCache(config.Redis), nil
	default:
		return nil, fmt.Errorf("unssuported caching provider %s", config.Cache.Provider)
	}
}

func CreateMessageBus(logger *zap.Logger) messaging.MessageBus {

	return messaging.NewInMemoryMessageBus(logger)
}

func CreateMailSender(config *config.AppConfig, logger *zap.Logger) mailing.Sender {
	switch config.Mailer.Provider {
	case "smtp":
		return mailing.NewSmtpSender(config.Smtp, logger)
	default:
		return mailing.NewStubSender(logger)
	}
}

func CreateIdentityRepository(db database.Database) (authRepos.IdentityRepository, error) {
	switch db.GetProviderType() {
	case "postgres":
		return authRepos.NewPostgresIdentityRepository(db.(*database.Db)), nil
	default:
		return nil, fmt.Errorf("unsupported database provider: %s", db.GetProviderType())
	}
}

func CreateSessionRepository(db database.Database) (authRepos.SessionRepository, error) {
	switch db.GetProviderType() {
	case "postgres":
		return authRepos.NewPostgresSessionRepository(db.(*database.Db)), nil
	default:
		return nil, fmt.Errorf("unsupported database provider: %s", db.GetProviderType())
	}
}
