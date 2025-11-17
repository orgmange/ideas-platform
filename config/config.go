package config

import (
	"fmt"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type Config struct {
	Server     ServerConfig
	DB         DBConfig
	App        AppConfig
	AuthConfig AuthConfig
}

type ServerConfig struct {
	Host string `env:"SERVER_HOST" envDefault:"localhost"`
	Port int    `env:"SERVER_PORT" envDefault:"8080"`
}

type DBConfig struct {
	Host     string `env:"DB_HOST" envDefault:"localhost"`
	Port     int    `env:"DB_PORT" envDefault:"5432"`
	User     string `env:"DB_USER,required"`
	Password string `env:"DB_PASSWORD,required"`
	Name     string `env:"DB_NAME" envDefault:"ideas_db"`
	SSLMode  string `env:"DB_SSLMODE" envDefault:"disable"`
}

type AppConfig struct {
	Env     string `env:"APP_ENV" envDefault:"development"`
	Version string `env:"APP_VERSION,required"`
}

type AuthConfig struct {
	OTPConfig OTPConfig
	JWTConfig JWTConfig
}

type OTPConfig struct {
	ExpiresAtTimer        time.Duration
	AttemptsLeft          int
	ResetResendCountTimer time.Duration
	SoftAttemptsCount     int
	HardAttemptsCount     int
	SubSoftAttemptsTimer  time.Duration
	SubHardAttemptsTimer  time.Duration
	PostHardAttemptsCount time.Duration
}

type JWTConfig struct {
	RefreshTokenTimer time.Duration
	JWTTokenTimer     time.Duration
}

func Load() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("No .env file found, proceeding with environment variables. ", err)
	}

	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
