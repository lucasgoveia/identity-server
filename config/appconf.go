package config

import (
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
	"log"
)

type ServerConfig struct {
	Port int    `mapstructure:"port"`
	Host string `mapstructure:"host"`
}

type DatabaseConfig struct {
	Provider string `mapstructure:"provider"`
}

type HashingConfig struct {
	Algorithm string `mapstructure:"algorithm"`
}

type PostgresConfig struct {
	URL string `mapstructure:"url"`
}

type SmtpConfig struct {
	Host               string `mapstructure:"host"`
	Port               int    `mapstructure:"port"`
	Username           string `mapstructure:"username"`
	Password           string `mapstructure:"password"`
	From               string `mapstructure:"from"`
	FromName           string `mapstructure:"from_name"`
	TLS                bool   `mapstructure:"tls"`
	DefaultCredentials bool   `mapstructure:"default_credentials"`
}

type MailerConfig struct {
	Provider string `mapstructure:"provider"`
}

type AppConfig struct {
	Server   *ServerConfig   `mapstructure:"server"`
	Database *DatabaseConfig `mapstructure:"database"`
	Hashing  *HashingConfig  `mapstructure:"hashing"`
	Postgres *PostgresConfig `mapstructure:"postgres"`
	Mailer   *MailerConfig   `mapstructure:"mailer"`
	Smtp     *SmtpConfig     `mapstructure:"smtp"`
	Cache    *CacheConfig    `mapstructure:"cache"`
	Redis    *RedisConfig    `mapstructure:"redis"`
}

type CacheConfig struct {
	Provider string `mapstructure:"provider"`
}

type RedisConfig struct {
	Url      string `mapstructure:"url"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
}

func LoadConfig() (*AppConfig, error) {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Printf("Error loading .env file: %v", err)
	}

	viper.AutomaticEnv()

	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("cache.provider", "inmemory")
	_ = viper.BindEnv("cache.provider", "CACHE_PROVIDER")
	_ = viper.BindEnv("mailer.provider", "MAILER_PROVIDER")
	_ = viper.BindEnv("server.port", "SERVER_PORT")
	_ = viper.BindEnv("server.host", "SERVER_HOST")
	_ = viper.BindEnv("postgres.url", "POSTGRES_URL")
	_ = viper.BindEnv("smtp.host", "SMTP_HOST")
	_ = viper.BindEnv("smtp.port", "SMTP_PORT")
	_ = viper.BindEnv("smtp.username", "SMTP_USERNAME")
	_ = viper.BindEnv("smtp.password", "SMTP_PASSWORD")
	_ = viper.BindEnv("smtp.from", "SMTP_FROM")
	_ = viper.BindEnv("smtp.from_name", "SMTP_FROM_NAME")
	_ = viper.BindEnv("smtp.tls", "SMTP_TLS")
	_ = viper.BindEnv("smtp.default_credentials", "SMTP_DEFAULT_CREDENTIALS")
	_ = viper.BindEnv("redis.url", "REDIS_URL")
	_ = viper.BindEnv("redis.username", "REDIS_USERNAME")
	_ = viper.BindEnv("redis.password", "REDIS_PASSWORD")

	// Read the configuration file
	viper.SetConfigFile("config/config.yaml")
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file: %v", err)
	}

	var config AppConfig
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil

	return &config, nil
}
