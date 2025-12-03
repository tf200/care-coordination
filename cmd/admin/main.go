package main

import (
	"care-cordination/lib/config"
	db "care-cordination/lib/db/sqlc"
	"care-cordination/lib/nanoid"
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("cannot load config: %v", err)
	}

	if cfg.AdminEmail == "" || cfg.AdminPassword == "" {
		log.Fatal("ADMIN_EMAIL and ADMIN_PASSWORD must be set")
	}

	connPool, err := pgxpool.New(context.Background(), cfg.DBSource)
	if err != nil {
		log.Fatalf("cannot connect to db: %v", err)
	}
	defer connPool.Close()

	store := db.NewStore(connPool)
	ctx := context.Background()

	// Check if admin user already exists
	_, err = store.GetUserByEmail(ctx, cfg.AdminEmail)
	if err == nil {
		log.Printf("Admin user %s already exists", cfg.AdminEmail)
		return
	}

	// Create admin user
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(cfg.AdminPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("cannot hash password: %v", err)
	}

	arg := db.CreateUserParams{
		ID:           nanoid.Generate(),
		Email:        cfg.AdminEmail,
		PasswordHash: string(hashedPassword),
	}

	userID, err := store.CreateUser(ctx, arg)
	if err != nil {
		log.Fatalf("cannot create admin user: %v", err)
	}

	log.Printf("Admin user created with ID: %s", userID)
}
