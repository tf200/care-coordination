package db

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

// testStore is the shared Store instance used by all tests.
// It is initialized in TestMain and available to all test functions.
var testStore *Store

// TestMain sets up the test environment by starting a PostgreSQL container,
// running migrations, and initializing the shared store.
func TestMain(m *testing.M) {
	ctx := context.Background()

	// Start PostgreSQL container
	container, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	if err != nil {
		panic("failed to start postgres container: " + err.Error())
	}

	// Ensure cleanup on exit
	defer func() {
		if err := container.Terminate(ctx); err != nil {
			panic("failed to terminate container: " + err.Error())
		}
	}()

	// Get connection string
	connStr, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		panic("failed to get connection string: " + err.Error())
	}

	// Create connection pool
	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		panic("failed to create connection pool: " + err.Error())
	}
	defer pool.Close()

	// Run migrations
	if err := runMigrations(ctx, pool); err != nil {
		panic("failed to run migrations: " + err.Error())
	}

	// Initialize the shared store
	testStore = NewStore(pool)

	// Run tests
	os.Exit(m.Run())
}

// runMigrations executes the SQL migration files against the database.
func runMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	// Get the path to the migrations directory
	// This assumes tests are run from the project root or lib/db/sqlc directory
	migrationPaths := []string{
		"../migrations/000001_init.up.sql",           // When running from lib/db/sqlc
		"lib/db/migrations/000001_init.up.sql",       // When running from project root
		"../../lib/db/migrations/000001_init.up.sql", // Alternative path
	}

	var migrationSQL []byte
	var err error

	for _, path := range migrationPaths {
		migrationSQL, err = os.ReadFile(path)
		if err == nil {
			break
		}
		// Try absolute path resolution
		absPath, _ := filepath.Abs(path)
		migrationSQL, err = os.ReadFile(absPath)
		if err == nil {
			break
		}
	}

	if err != nil {
		return err
	}

	_, err = pool.Exec(ctx, string(migrationSQL))
	return err
}

// runTestWithTx runs a test function within a transaction that is always rolled back.
// This ensures complete test isolation - each test starts with a clean database state.
//
// Usage:
//
//	func TestCreateUser_Success(t *testing.T) {
//	    runTestWithTx(t, func(t *testing.T, q *Queries) {
//	        // Test code here - any data created will be rolled back
//	    })
//	}
func runTestWithTx(t *testing.T, fn func(t *testing.T, q *Queries)) {
	t.Helper()
	ctx := context.Background()

	tx, err := testStore.ConnPool.Begin(ctx)
	if err != nil {
		t.Fatalf("failed to begin transaction: %v", err)
	}
	defer tx.Rollback(ctx) // Always rollback - never commit

	fn(t, New(tx))
}

// runTestWithTxAndStore runs a test function with both a transactional Queries
// instance and access to a Store-like transaction executor.
// Use this when testing code that needs ExecTx functionality.
func runTestWithTxAndStore(t *testing.T, fn func(t *testing.T, q *Queries, execTx func(fn func(*Queries) error) error)) {
	t.Helper()
	ctx := context.Background()

	tx, err := testStore.ConnPool.Begin(ctx)
	if err != nil {
		t.Fatalf("failed to begin transaction: %v", err)
	}
	defer tx.Rollback(ctx) // Always rollback

	q := New(tx)

	// Create a nested transaction executor using savepoints
	execTx := func(fn func(*Queries) error) error {
		// Use savepoint for nested "transaction"
		_, err := tx.Exec(ctx, "SAVEPOINT test_savepoint")
		if err != nil {
			return err
		}

		if err := fn(q); err != nil {
			tx.Exec(ctx, "ROLLBACK TO SAVEPOINT test_savepoint")
			return err
		}

		_, err = tx.Exec(ctx, "RELEASE SAVEPOINT test_savepoint")
		return err
	}

	fn(t, q, execTx)
}
