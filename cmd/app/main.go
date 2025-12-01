package main

import (
	"care-cordination/api"
	"care-cordination/features/auth"
	"care-cordination/lib/config"
	db "care-cordination/lib/db/sqlc"
	"care-cordination/lib/logger"
	"care-cordination/lib/middleware"
	"care-cordination/lib/token"
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

func main() {
	// 1. Load Configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("cannot load config: %v", err)
	}

	// 2. Initialize Logger
	l := logger.NewLogger(cfg.Environment)
	// Create a background context that we can use for startup operations
	// and that we can cancel on shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// 3. Initialize Database Connection
	connPool, err := pgxpool.New(ctx, cfg.DBSource)
	if err != nil {
		l.Error(ctx, "main", "cannot connect to db", zap.Error(err))
		os.Exit(1)
	}
	defer connPool.Close()

	// 4. Initialize Dependencies
	store := db.NewStore(connPool)
	tokenManager := token.NewTokenManager(
		cfg.AccessTokenSecret,
		cfg.RefreshTokenSecret,
		cfg.AccessTokenTTL,
		cfg.RefreshTokenTTL,
	)
	mdw := middleware.NewMiddleware(tokenManager)

	// 5. Initialize Features
	authService := auth.NewAuthService(store, tokenManager, l)
	authHandler := auth.NewAuthHandler(authService, mdw)

	// 6. Initialize Server
	server := api.NewServer(l, cfg.Environment, authHandler)

	// 7. Start Server
	go func() {
		l.Info(ctx, "main", "starting server", zap.String("address", cfg.ServerAddress))
		if err := server.Start(cfg.ServerAddress); err != nil && err != http.ErrServerClosed {
			l.Error(ctx, "main", "cannot start server", zap.Error(err))
		}
	}()

	// 8. Graceful Shutdown
	<-ctx.Done()
	l.Info(ctx, "main", "shutdown signal received")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		l.Error(ctx, "main", "server shutdown failed", zap.Error(err))
	}

	l.Info(ctx, "main", "server exited properly")
}
