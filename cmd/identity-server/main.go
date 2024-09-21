package main

import (
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
	"identity-server/config"
	"identity-server/features/signup/email"
	"identity-server/internal/cache"
	"identity-server/internal/consumers"
	"identity-server/internal/messages/commands"
	"identity-server/internal/security"
	"identity-server/pkg/providers"
	"log"
	"os"
	"os/signal"
	"reflect"
	"syscall"
)

func main() {
	e := echo.New()

	// todo: we should probably not pass around zap, maybe create a wrapper with less methods
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()

	if err := godotenv.Load(); err != nil {
		e.Logger.Debug("Error loading .env file: %v", err)
	}

	appConfig, err := config.LoadConfig()

	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	db, err := providers.CreateDatabase(appConfig)

	defer func() {
		if err := db.Close(); err != nil {
			log.Fatalf("Failed to close database: %v", err)
		}
	}()

	if err != nil {
		log.Fatalf("Failed to create database: %v", err)
	}

	accountManager, err := providers.CreateAccountManager(db)
	timeProvider := providers.CreateDefaultTimeProvider()
	hasher, err := providers.CreateHasher(appConfig)

	bus := providers.CreateMessageBus(logger)

	mailer := providers.CreateMailSender(appConfig, logger)

	consumer := consumers.NewSendVerificationEmailConsumer(security.NewOTPGenerator(security.NewSecureKeyGenerator()),
		cache.NewInMemory(),
		logger,
		mailer)
	bus.RegisterConsumer(reflect.TypeOf(commands.SendVerificationEmail{}), consumer.Handle)

	bus.Start()

	// TODO: Change: instead of using hasher directly, create an wrapper for password hashing
	// because, for example, totp secret does not have the same security requirements as password
	e.POST("/sign-up/email", email.SignUp(accountManager, timeProvider, hasher, bus))

	go func() {
		// Start the server
		if err := e.Start(":1323"); err != nil {
			logger.Sugar().Fatalf("Shutting down the server: %v", err)
		}
	}()

	// Gracefully handle OS signals
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, syscall.SIGINT, syscall.SIGTERM)
	<-signalChannel

	// Stop the server and clean up
	bus.Stop()
}
