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

type AppConfig struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Hashing  HashingConfig  `mapstructure:"hashing"`
	Postgres PostgresConfig `mapstructure:"postgres"`
}

func LoadConfig() (*AppConfig, error) {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Printf("Error loading .env file: %v", err)
	}

	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.host", "0.0.0.0")
	_ = viper.BindEnv("server.port", "SERVER_PORT")
	_ = viper.BindEnv("server.host", "SERVER_HOST")
	_ = viper.BindEnv("postgres.url", "POSTGRES_URL")

	viper.AutomaticEnv()

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
