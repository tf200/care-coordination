package main

import (
	"care-cordination/lib/config"
	db "care-cordination/lib/db/sqlc"
	"care-cordination/lib/nanoid"
	"context"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
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

	userID := nanoid.Generate()
	employeeID := nanoid.Generate()

	// Create user
	_, err = store.CreateUser(ctx, db.CreateUserParams{
		ID:           userID,
		Email:        cfg.AdminEmail,
		PasswordHash: string(hashedPassword),
	})
	if err != nil {
		log.Fatalf("cannot create admin user: %v", err)
	}
	log.Printf("Admin user created with ID: %s", userID)

	// Create employee record for admin
	err = store.CreateEmployee(ctx, db.CreateEmployeeParams{
		ID:          employeeID,
		UserID:      userID,
		FirstName:   "System",
		LastName:    "Admin",
		Bsn:         "000000000",
		DateOfBirth: pgtype.Date{Time: time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC), Valid: true},
		PhoneNumber: "0000000000",
		Gender:      db.GenderEnumOther,
	})
	if err != nil {
		log.Fatalf("cannot create admin employee: %v", err)
	}
	log.Printf("Admin employee created with ID: %s", employeeID)

	// Assign admin role (role_admin is preset in the migration)
	err = store.AssignRoleToUser(ctx, db.AssignRoleToUserParams{
		UserID: userID,
		RoleID: "role_admin",
	})
	if err != nil {
		log.Fatalf("cannot assign admin role: %v", err)
	}
	log.Printf("Admin role assigned to user %s", cfg.AdminEmail)

	log.Println("Admin setup complete!")
}
