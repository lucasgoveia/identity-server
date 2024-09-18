package main

import (
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"identity-server/config"
	"identity-server/features/signup/email"
	"identity-server/internal/consumers"
	"identity-server/internal/messages/commands"
	"identity-server/pkg/providers"
	"log"
	"os"
	"os/signal"
	"reflect"
	"syscall"
)

func main() {
	e := echo.New()

	if err := godotenv.Load(); err != nil {
		e.Logger.Debug("Error loading .env file: %v", err)
	}

	appConfig, err := config.LoadConfig()

	e.Use(middleware.Logger())
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

	bus := providers.CreateMessageBus()

	consumer := consumers.SendVerificationEmailConsumer{}
	bus.RegisterConsumer(reflect.TypeOf(commands.SendVerificationEmail{}), consumer.Handle)

	bus.Start()

	// TODO: Change: instead of using hasher directly, create an wrapper for password hashing
	// because, for example, totp secret does not have the same security requirements as password
	e.POST("/sign-up/email", email.SignUp(accountManager, timeProvider, hasher, bus))

	go func() {
		// Start the server
		if err := e.Start(":1323"); err != nil {
			log.Printf("Shutting down the server: %v", err)
		}
	}()

	// Gracefully handle OS signals
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, syscall.SIGINT, syscall.SIGTERM)
	<-signalChannel

	// Stop the server and clean up
	bus.Stop()
}
