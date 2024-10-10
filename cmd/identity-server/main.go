package main

import (
	"context"
	"fmt"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"identity-server/config"
	"identity-server/internal/accounts/consumers"
	"identity-server/internal/accounts/handlers/identity_verification"
	"identity-server/internal/accounts/handlers/signup"
	"identity-server/internal/accounts/messages/commands"
	accServices "identity-server/internal/accounts/services"
	"identity-server/internal/auth/handlers/login"
	authServices "identity-server/internal/auth/services"
	"identity-server/pkg/middlewares"
	"identity-server/pkg/providers"
	"identity-server/pkg/security"
	"log"
	"os"
	"os/signal"
	"reflect"
	"syscall"
)

var serviceName = semconv.ServiceNameKey.String("identity-service")

func main() {
	e := echo.New()

	// todo: we should probably not pass around zap, maybe create a wrapper with less methods
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := logger.Sync(); err != nil {
			log.Fatalf("failed to flush logs on shutdown %s", err)
		}
	}()

	if err := godotenv.Load(); err != nil {
		e.Logger.Debug("Error loading .env file: %v", err)
	}

	appConfig, err := config.LoadConfig()

	conn, _ := initConn()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	e.Use(otelecho.Middleware(serviceName.Value.AsString()))
	res, err := resource.New(ctx,
		resource.WithAttributes(
			// The service name used to display traces in backends
			serviceName,
		),
	)
	if err != nil {
		logger.Error("error while creating resource", zap.Error(err))
	}
	shutdownTracerProvider, err := initTracerProvider(ctx, res, conn)
	defer func() {
		if err := shutdownTracerProvider(ctx); err != nil {
			log.Fatalf("failed to shutdown TracerProvider: %s", err)
		}
	}()
	if err != nil {
		logger.Error("error while initializing tracing provider", zap.Error(err))
	}

	shutdownMeterProvider, err := initMeterProvider(ctx, res, conn)
	if err != nil {
		logger.Error("error while initializing metric provider", zap.Error(err))
	}
	defer func() {
		if err := shutdownMeterProvider(ctx); err != nil {
			log.Fatalf("failed to shutdown MeterProvider: %s", err)
		}
	}()

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

	identityVerificationManager := accServices.NewIdentityVerificationManager(otpGen, cache, hasher, logger)

	rsaHolder, err := security.NewRSAKeyHolder(appConfig.Auth.AccessTokenConfig.PrivateKey, appConfig.Auth.AccessTokenConfig.PublicKey)

	tm := security.NewTokenManager(appConfig.Auth, timeProvider, logger, cache, rsaHolder)

	sessionRepo, err := providers.CreateSessionRepository(db)
	identityRepo, err := providers.CreateIdentityRepository(db)

	authService := authServices.NewAuthService(logger, tm, sessionRepo, timeProvider, appConfig.Auth.SessionConfig)

	consumer := consumers.NewSendVerificationEmailConsumer(identityVerificationManager, logger, mailer)
	bus.RegisterConsumer(reflect.TypeOf(commands.SendVerificationEmail{}), consumer.Handle)

	bus.Start()

	// TODO: Change: instead of using hasher directly, create an wrapper for password hashing
	// because, for example, totp secret does not have the same security requirements as password
	e.POST("/sign-up/email", signup.SignUp(accountManager, timeProvider, hasher, bus, tm))

	e.POST("login/email", login.Login(identityRepo, hasher, timeProvider, authService))

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

func initConn() (*grpc.ClientConn, error) {
	// It connects the OpenTelemetry Collector through local gRPC connection.
	// You may replace `localhost:4317` with your endpoint.
	conn, err := grpc.NewClient("localhost:4317",
		// Note the use of insecure transport here. TLS is recommended in production.
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC connection to collector: %w", err)
	}

	return conn, err
}

func initTracerProvider(ctx context.Context, res *resource.Resource, conn *grpc.ClientConn) (func(context.Context) error, error) {
	// Set up a trace exporter
	traceExporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(conn))
	if err != nil {
		return nil, fmt.Errorf("failed to create trace exporter: %w", err)
	}

	// Register the trace exporter with a TracerProvider, using a batch
	// span processor to aggregate spans before export.
	bsp := sdktrace.NewBatchSpanProcessor(traceExporter)
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(bsp),
	)
	otel.SetTracerProvider(tracerProvider)

	// Set global propagator to tracecontext (the default is no-op).
	otel.SetTextMapPropagator(propagation.TraceContext{})

	// Shutdown will flush any remaining spans and shut down the exporter.
	return tracerProvider.Shutdown, nil
}

// Initializes an OTLP exporter, and configures the corresponding meter provider.
func initMeterProvider(ctx context.Context, res *resource.Resource, conn *grpc.ClientConn) (func(context.Context) error, error) {
	metricExporter, err := otlpmetricgrpc.New(ctx, otlpmetricgrpc.WithGRPCConn(conn))
	if err != nil {
		return nil, fmt.Errorf("failed to create metrics exporter: %w", err)
	}

	meterProvider := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExporter)),
		sdkmetric.WithResource(res),
	)
	otel.SetMeterProvider(meterProvider)

	return meterProvider.Shutdown, nil
}
