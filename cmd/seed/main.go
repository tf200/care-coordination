package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"

	"care-cordination/lib/config"
	db "care-cordination/lib/db/sqlc"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	gonanoid "github.com/matoous/go-nanoid/v2"
	"golang.org/x/crypto/bcrypt"
)

// Sample data for generating random employees
var (
	firstNames = []string{
		"Jan", "Piet", "Sophie", "Emma", "Lucas",
		"Daan", "Lotte", "Finn", "Julia", "Sem",
		"Anna", "Lars", "Sara", "Thomas", "Eva",
		"Tim", "Lisa", "Niels", "Fleur", "Ruben",
	}

	lastNames = []string{
		"de Vries", "Jansen", "van den Berg", "Bakker", "Visser",
		"Smit", "Meijer", "de Groot", "Mulder", "de Boer",
		"Vos", "Peters", "Hendriks", "van Leeuwen", "Dekker",
		"Brouwer", "de Wit", "Dijkstra", "Smits", "de Graaf",
	}

	roles = []string{
		"coordinator",
		"admin",
		"manager",
		"therapist",
		"nurse",
	}

	genders = []db.GenderEnum{
		db.GenderEnumMale,
		db.GenderEnumFemale,
		db.GenderEnumOther,
	}
)

func main() {
	ctx := context.Background()

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Connect to database
	connPool, err := pgxpool.New(ctx, cfg.DBSource)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer connPool.Close()

	// Create store
	store := db.NewStore(connPool)

	// Seed employees
	if err := seedEmployees(ctx, store, 20); err != nil {
		log.Fatalf("Failed to seed employees: %v", err)
	}

	fmt.Println("✅ Successfully seeded database!")
}

func seedEmployees(ctx context.Context, store *db.Store, count int) error {
	fmt.Printf("🌱 Seeding %d employees...\n", count)

	for i := 0; i < count; i++ {
		employee, err := createRandomEmployee(ctx, store)
		if err != nil {
			return fmt.Errorf("failed to create employee %d: %w", i+1, err)
		}
		fmt.Printf("  ✓ Created employee: %s %s (%s)\n", employee.FirstName, employee.LastName, employee.Role)
	}

	fmt.Printf("✅ Successfully seeded %d employees\n", count)
	return nil
}

type EmployeeInfo struct {
	FirstName string
	LastName  string
	Role      string
}

func createRandomEmployee(ctx context.Context, store *db.Store) (*EmployeeInfo, error) {
	// Generate random data
	firstName := randomElement(firstNames)
	lastName := randomElement(lastNames)
	email := generateEmail(firstName, lastName)
	role := randomElement(roles)

	// Generate IDs
	userID, err := gonanoid.New()
	if err != nil {
		return nil, fmt.Errorf("failed to generate user ID: %w", err)
	}
	employeeID, err := gonanoid.New()
	if err != nil {
		return nil, fmt.Errorf("failed to generate employee ID: %w", err)
	}

	// Hash a default password
	passwordHash, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create the employee using the transaction
	err = store.CreateEmployeeTx(ctx, db.CreateEmployeeTxParams{
		User: db.CreateUserParams{
			ID:           userID,
			Email:        email,
			PasswordHash: string(passwordHash),
		},
		Emp: db.CreateEmployeeParams{
			ID:          employeeID,
			UserID:      userID, // Will be overwritten in tx, but need to provide
			FirstName:   firstName,
			LastName:    lastName,
			Bsn:         generateBSN(),
			DateOfBirth: generateRandomDateOfBirth(),
			PhoneNumber: generatePhoneNumber(),
			Gender:      randomGender(),
			Role:        role,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create employee tx: %w", err)
	}

	return &EmployeeInfo{
		FirstName: firstName,
		LastName:  lastName,
		Role:      role,
	}, nil
}

// Helper functions

func randomElement[T any](slice []T) T {
	return slice[rand.Intn(len(slice))]
}

func randomGender() db.GenderEnum {
	return genders[rand.Intn(len(genders))]
}

func generateEmail(firstName, lastName string) string {
	// Create a unique email using nanoid suffix
	suffix, _ := gonanoid.Generate("0123456789", 4)
	return fmt.Sprintf("%s.%s.%s@example.com",
		normalizeForEmail(firstName),
		normalizeForEmail(lastName),
		suffix)
}

func normalizeForEmail(s string) string {
	// Simple normalization - lowercase and replace spaces
	result := ""
	for _, c := range s {
		if c >= 'a' && c <= 'z' {
			result += string(c)
		} else if c >= 'A' && c <= 'Z' {
			result += string(c + 32) // Convert to lowercase
		}
	}
	return result
}

func generateBSN() string {
	// Generate a random 9-digit BSN (Dutch social security number)
	// Note: This doesn't follow the actual BSN validation rules, just random digits
	return fmt.Sprintf("%09d", rand.Intn(1000000000))
}

func generateRandomDateOfBirth() pgtype.Date {
	// Generate a date of birth between 20 and 60 years ago
	minAge, maxAge := 20, 60
	yearsAgo := minAge + rand.Intn(maxAge-minAge)

	dob := time.Now().AddDate(-yearsAgo, -rand.Intn(12), -rand.Intn(28))

	return pgtype.Date{
		Time:  dob,
		Valid: true,
	}
}

func generatePhoneNumber() string {
	// Generate a Dutch mobile phone number format
	return fmt.Sprintf("06%08d", rand.Intn(100000000))
}
