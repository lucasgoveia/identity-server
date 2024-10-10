package providers

import (
	"fmt"
	"go.uber.org/zap"
	"identity-server/config"
	accRepos "identity-server/internal/accounts/repositories"
	authRepos "identity-server/internal/auth/repositories"
	"identity-server/pkg/providers/cache"
	"identity-server/pkg/providers/database"
	"identity-server/pkg/providers/hashing"
	"identity-server/pkg/providers/mailing"
	"identity-server/pkg/providers/messaging"
	"identity-server/pkg/providers/time"
)

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
