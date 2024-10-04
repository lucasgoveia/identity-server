package main

import (
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
	"identity-server/config"
	"identity-server/internal/accounts/consumers"
	"identity-server/internal/accounts/handlers/identity_verification"
	"identity-server/internal/accounts/handlers/signup"
	"identity-server/internal/accounts/messages/commands"
	"identity-server/internal/accounts/services"
	"identity-server/pkg/middlewares"
	"identity-server/pkg/providers"
	"identity-server/pkg/security"
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

	accountManager, err := providers.CreateAccountRepository(db)
	timeProvider := providers.CreateDefaultTimeProvider()
	hasher, err := providers.CreateHasher(appConfig)

	bus := providers.CreateMessageBus(logger)

	mailer := providers.CreateMailSender(appConfig, logger)

	cache, err := providers.CreateCache(appConfig)

	secureKeyGen := security.NewSecureKeyGenerator()

	otpGen := security.NewOTPGenerator(secureKeyGen)

	identityVerificationManager := services.NewIdentityVerificationManager(otpGen, cache, hasher, logger)

	tm := security.NewTokenManager(appConfig.Auth, timeProvider, logger, cache)

	consumer := consumers.NewSendVerificationEmailConsumer(identityVerificationManager, logger, mailer)
	bus.RegisterConsumer(reflect.TypeOf(commands.SendVerificationEmail{}), consumer.Handle)

	bus.Start()

	// TODO: Change: instead of using hasher directly, create an wrapper for password hashing
	// because, for example, totp secret does not have the same security requirements as password
	e.POST("/sign-up/email", signup.SignUp(accountManager, timeProvider, hasher, bus, tm))

	verificationRoutes := e.Group("/verify")

	verificationRoutes.Use(middlewares.VerifyIdentityAuth(tm))

	verificationRoutes.POST("/email", identity_verification.VerifyEmail(accountManager, tm, identityVerificationManager))

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
