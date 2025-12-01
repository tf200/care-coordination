package config

import (
	"errors"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	DBSource           string
	AccessTokenSecret  string
	RefreshTokenSecret string
	AccessTokenTTL     time.Duration
	RefreshTokenTTL    time.Duration
	Environment        string
	ServerAddress      string
}

func LoadConfig() (*Config, error) {
	godotenv.Load()

	// PARSE DURATION
	accessTokenTTL, err := time.ParseDuration(os.Getenv("ACCESS_TOKEN_TTL"))
	if err != nil {
		return nil, err
	}
	refreshTokenTTL, err := time.ParseDuration(os.Getenv("REFRESH_TOKEN_TTL"))
	if err != nil {
		return nil, err
	}

	config := &Config{
		DBSource:           os.Getenv("DB_SOURCE"),
		AccessTokenSecret:  os.Getenv("ACCESS_TOKEN_SECRET"),
		RefreshTokenSecret: os.Getenv("REFRESH_TOKEN_SECRET"),
		AccessTokenTTL:     accessTokenTTL,
		RefreshTokenTTL:    refreshTokenTTL,
	}

	if err := config.validate(); err != nil {
		return nil, err
	}

	return config, nil
}

func (c *Config) validate() error {
	if c.DBSource == "" {
		return errors.New("DB_SOURCE is not set")
	}
	if c.AccessTokenSecret == "" {
		return errors.New("ACCESS_TOKEN_SECRET is not set")
	}
	if c.RefreshTokenSecret == "" {
		return errors.New("REFRESH_TOKEN_SECRET is not set")
	}
	if c.AccessTokenTTL == 0 {
		return errors.New("ACCESS_TOKEN_TTL is not set")
	}
	if c.RefreshTokenTTL == 0 {
		return errors.New("REFRESH_TOKEN_TTL is not set")
	}
	return nil
}
