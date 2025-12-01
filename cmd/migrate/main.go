package main

import (
	"fmt"
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	dbURL := os.Getenv("DB_MIGRATE_URL")
	if dbURL == "" {
		log.Fatal("DB_MIGRATE_URL is not set in environment")
	}

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	migrationsPath := "file://lib/db/migrations"
	m, err := migrate.New(migrationsPath, dbURL)
	if err != nil {
		log.Fatalf("Failed to create migrate instance: %v", err)
	}
	defer m.Close()

	command := os.Args[1]

	switch command {
	case "up":
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("Failed to run migrations up: %v", err)
		}
		fmt.Println("✓ Migrations applied successfully")

	case "down":
		if err := m.Down(); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("Failed to run migrations down: %v", err)
		}
		fmt.Println("✓ All migrations rolled back successfully")

	case "up1":
		if err := m.Steps(1); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("Failed to migrate up 1 step: %v", err)
		}
		fmt.Println("✓ Migrated up 1 step successfully")

	case "down1":
		if err := m.Steps(-1); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("Failed to migrate down 1 step: %v", err)
		}
		fmt.Println("✓ Migrated down 1 step successfully")

	case "version":
		version, dirty, err := m.Version()
		if err != nil {
			log.Fatalf("Failed to get migration version: %v", err)
		}
		fmt.Printf("Current version: %d (dirty: %v)\n", version, dirty)

	case "force":
		if len(os.Args) < 3 {
			log.Fatal("Usage: migrate force <version>")
		}
		var version int
		_, err := fmt.Sscanf(os.Args[2], "%d", &version)
		if err != nil {
			log.Fatalf("Invalid version number: %v", err)
		}
		if err := m.Force(version); err != nil {
			log.Fatalf("Failed to force version: %v", err)
		}
		fmt.Printf("✓ Forced version to %d\n", version)

	default:
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Usage: migrate <command>")
	fmt.Println("\nCommands:")
	fmt.Println("  up       - Apply all pending migrations")
	fmt.Println("  down     - Rollback all migrations")
	fmt.Println("  up1      - Apply the next migration")
	fmt.Println("  down1    - Rollback the last migration")
	fmt.Println("  version  - Show current migration version")
	fmt.Println("  force <version> - Force set migration version (use with caution)")
}
