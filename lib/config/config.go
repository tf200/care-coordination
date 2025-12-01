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

	// Rate Limiting
	RedisURL                  string
	RateLimitEnabled          bool
	LoginRateLimitPerIP       int
	LoginRateLimitWindowIP    time.Duration
	LoginRateLimitPerEmail    int
	LoginRateLimitWindowEmail time.Duration
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

	// Parse rate limit windows with defaults
	loginRateLimitWindowIP := 15 * time.Minute
	if val := os.Getenv("LOGIN_RATE_LIMIT_WINDOW_IP"); val != "" {
		if parsed, err := time.ParseDuration(val); err == nil {
			loginRateLimitWindowIP = parsed
		}
	}

	loginRateLimitWindowEmail := 15 * time.Minute
	if val := os.Getenv("LOGIN_RATE_LIMIT_WINDOW_EMAIL"); val != "" {
		if parsed, err := time.ParseDuration(val); err == nil {
			loginRateLimitWindowEmail = parsed
		}
	}

	// Parse rate limit counts with defaults
	loginRateLimitPerIP := 5
	if val := os.Getenv("LOGIN_RATE_LIMIT_PER_IP"); val != "" {
		if parsed, err := time.ParseDuration(val); err == nil {
			loginRateLimitPerIP = int(parsed)
		}
	}

	loginRateLimitPerEmail := 3
	if val := os.Getenv("LOGIN_RATE_LIMIT_PER_EMAIL"); val != "" {
		if parsed, err := time.ParseDuration(val); err == nil {
			loginRateLimitPerEmail = int(parsed)
		}
	}

	rateLimitEnabled := true
	if val := os.Getenv("RATE_LIMIT_ENABLED"); val == "false" {
		rateLimitEnabled = false
	}

	config := &Config{
		DBSource:           os.Getenv("DB_SOURCE"),
		AccessTokenSecret:  os.Getenv("ACCESS_TOKEN_SECRET"),
		RefreshTokenSecret: os.Getenv("REFRESH_TOKEN_SECRET"),
		AccessTokenTTL:     accessTokenTTL,
		RefreshTokenTTL:    refreshTokenTTL,

		// Rate Limiting
		RedisURL:                  os.Getenv("REDIS_URL"),
		RateLimitEnabled:          rateLimitEnabled,
		LoginRateLimitPerIP:       loginRateLimitPerIP,
		LoginRateLimitWindowIP:    loginRateLimitWindowIP,
		LoginRateLimitPerEmail:    loginRateLimitPerEmail,
		LoginRateLimitWindowEmail: loginRateLimitWindowEmail,
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

	// Rate limiting validation (only if enabled)
	if c.RateLimitEnabled && c.RedisURL == "" {
		return errors.New("REDIS_URL is required when rate limiting is enabled")
	}

	return nil
}
