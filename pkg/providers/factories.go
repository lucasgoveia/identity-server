package providers

import (
	"fmt"
	"go.uber.org/zap"
	"identity-server/config"
	"identity-server/internal/app/accounts"
	"identity-server/internal/database/postgres"
	"identity-server/internal/mailing"
	"identity-server/pkg/providers/database"
	postgresDb "identity-server/pkg/providers/database/postgres"
	"identity-server/pkg/providers/hashing"
	"identity-server/pkg/providers/messaging"
	"identity-server/pkg/providers/time"
)

func CreateDatabase(config *config.AppConfig) (database.Database, error) {
	switch config.Database.Provider {
	case "postgres":
		return postgresDb.NewPostgresDb(config.Postgres.URL)
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

func CreateAccountManager(db database.Database) (accounts.AccountManager, error) {
	switch db.GetProviderType() {
	case "postgres":
		return postgres.NewAccountManager(db.(*postgresDb.Db)), nil
	default:
		return nil, fmt.Errorf("unsupported database provider: %s", db.GetProviderType())
	}
}

func CreateMessageBus(logger *zap.Logger) messaging.MessageBus {

	return messaging.NewInMemoryMessageBus(logger)
}

func CreateMailSender(config *config.AppConfig, logger *zap.Logger) mailing.Sender {
	switch config.Mailer.Provider {
	case "smtp":
		return mailing.NewSmtpSender(&config.Smtp, logger)
	default:
		return mailing.NewStubSender(logger)
	}
}
