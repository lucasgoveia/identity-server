package providers

import (
	"fmt"
	"go.uber.org/zap"
	"identity-server/config"
	"identity-server/internal/accounts/repositories"
	cache2 "identity-server/pkg/providers/cache"
	"identity-server/pkg/providers/database"
	"identity-server/pkg/providers/hashing"
	mailing2 "identity-server/pkg/providers/mailing"
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

func CreateAccountRepository(db database.Database) (repositories.AccountRepository, error) {
	switch db.GetProviderType() {
	case "postgres":
		return repositories.NewPostgresAccountRepository(db.(*database.Db)), nil
	default:
		return nil, fmt.Errorf("unsupported database provider: %s", db.GetProviderType())
	}
}

func CreateCache(config *config.AppConfig) (cache2.Cache, error) {
	switch config.Cache.Provider {
	case "inmemory":
		return cache2.NewInMemory(), nil
	case "redis":
		return cache2.NewRedisCache(config.Redis), nil
	default:
		return nil, fmt.Errorf("unssuported caching provider %s", config.Cache.Provider)
	}
}

func CreateMessageBus(logger *zap.Logger) messaging.MessageBus {

	return messaging.NewInMemoryMessageBus(logger)
}

func CreateMailSender(config *config.AppConfig, logger *zap.Logger) mailing2.Sender {
	switch config.Mailer.Provider {
	case "smtp":
		return mailing2.NewSmtpSender(config.Smtp, logger)
	default:
		return mailing2.NewStubSender(logger)
	}
}
