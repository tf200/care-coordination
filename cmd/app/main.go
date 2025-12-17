package main

import (
	"care-cordination/api"
	"care-cordination/features/attachments"
	"care-cordination/features/auth"
	"care-cordination/features/client"
	"care-cordination/features/employee"
	"care-cordination/features/incident"
	"care-cordination/features/intake"
	locTransfer "care-cordination/features/location_transfer"
	"care-cordination/features/locations"
	"care-cordination/features/middleware"
	"care-cordination/features/rbac"
	referringOrgs "care-cordination/features/referring_orgs"
	"care-cordination/features/registration"
	"care-cordination/lib/bucket"
	"care-cordination/lib/config"
	db "care-cordination/lib/db/sqlc"
	"care-cordination/lib/logger"
	"care-cordination/lib/ratelimit"
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

	// Initialize Object Storage
	bucketClient, err := bucket.NewObjectStorageClient(
		cfg.MinioEndpoint,
		cfg.MinioAccessKeyID,
		cfg.MinioSecretAccessKey,
		cfg.MinioUseSSL,
		cfg.MinioBucketName,
	)
	if err != nil {
		l.Error(ctx, "main", "cannot create object storage client", zap.Error(err))
		os.Exit(1)
	}

	// Create bucket if it doesn't exist
	if err := bucketClient.GetOrCreateBucket(ctx); err != nil {
		l.Error(ctx, "main", "cannot create bucket", zap.Error(err))
		os.Exit(1)
	}

	// Initialize Rate Limiter
	var rateLimiter ratelimit.RateLimiter
	if cfg.RateLimitEnabled {
		rlConfig := &ratelimit.Config{
			RedisURL:       cfg.RedisURL,
			IPLimit:        cfg.LoginRateLimitPerIP,
			IPWindow:       cfg.LoginRateLimitWindowIP,
			EmailLimit:     cfg.LoginRateLimitPerEmail,
			EmailWindow:    cfg.LoginRateLimitWindowEmail,
			EnableFallback: true, // Use in-memory fallback if Redis fails
		}

		rateLimiter, err = ratelimit.NewRateLimiter(rlConfig)
		if err != nil {
			l.Warn(ctx, "main", "failed to initialize rate limiter, continuing without rate limiting", zap.Error(err))
			rateLimiter = nil
		} else {
			l.Info(ctx, "main", "rate limiter initialized successfully")
			defer func() {
				if err := rateLimiter.Close(); err != nil {
					l.Error(ctx, "main", "failed to close rate limiter", zap.Error(err))
				}
			}()
		}
	} else {
		l.Info(ctx, "main", "rate limiting is disabled")
	}

	// 5. Initialize Features
	mdw := middleware.NewMiddleware(tokenManager, rateLimiter, l, store)

	authService := auth.NewAuthService(store, tokenManager, l)
	authHandler := auth.NewAuthHandler(authService, mdw)

	employeeService := employee.NewEmployeeService(store, l)
	employeeHandler := employee.NewEmployeeHandler(employeeService, mdw)

	registrationService := registration.NewRegistrationService(store, l)
	registrationHandler := registration.NewRegistrationHandler(registrationService, mdw)

	attachmentsService := attachments.NewAttachmentsService(store, bucketClient, l)
	attachmentsHandler := attachments.NewAttachmentsHandler(attachmentsService, mdw)

	referringOrgService := referringOrgs.NewReferringOrgService(store, l)
	referringOrgHandler := referringOrgs.NewReferringOrgHandler(referringOrgService, mdw)

	locTransferService := locTransfer.NewLocationTransferService(store, l)
	locTransferHandler := locTransfer.NewLocTransferHandler(locTransferService, mdw)

	locationService := locations.NewLocationService(store, l)
	locationHandler := locations.NewLocationHandler(locationService, mdw)

	intakeService := intake.NewIntakeService(store, l)
	intakeHandler := intake.NewIntakeHandler(intakeService, mdw)

	incidentService := incident.NewIncidentService(store, l)
	incidentHandler := incident.NewIncidentHandler(incidentService, mdw)

	clientService := client.NewClientService(store, l)
	clientHandler := client.NewClientHandler(clientService, mdw)

	rbacService := rbac.NewRBACService(store, l)
	rbacHandler := rbac.NewRBACHandler(rbacService, mdw)

	// 6. Initialize Server
	server := api.NewServer(
		l,
		cfg.Environment,
		authHandler,
		employeeHandler,
		registrationHandler,
		attachmentsHandler,
		locationHandler,
		intakeHandler,
		incidentHandler,
		clientHandler,
		referringOrgHandler,
		locTransferHandler,
		rbacHandler,
		rateLimiter,
		cfg.ServerAddress,
		cfg.Url,
	)

	// 7. Start Server
	go func() {
		l.Info(ctx, "main", "starting server", zap.String("address", cfg.ServerAddress))
		if err := server.Start(); err != nil && err != http.ErrServerClosed {
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
