package db

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	gonanoid "github.com/matoous/go-nanoid/v2"
)

// ============================================================
// Error Helpers
// ============================================================

// isPgError checks if the error is a PostgreSQL error with the given code.
// Common codes:
//   - "23505" = unique_violation
//   - "23503" = foreign_key_violation
//   - "23514" = check_violation
//   - "23502" = not_null_violation
func isPgError(err error, code string) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == code
}

// IsUniqueViolation checks if the error is a PostgreSQL unique constraint violation.
func IsUniqueViolation(err error) bool {
	return isPgError(err, "23505")
}

// IsForeignKeyViolation checks if the error is a PostgreSQL foreign key constraint violation.
func IsForeignKeyViolation(err error) bool {
	return isPgError(err, "23503")
}

// IsCheckViolation checks if the error is a PostgreSQL check constraint violation.
func IsCheckViolation(err error) bool {
	return isPgError(err, "23514")
}

// IsNotNullViolation checks if the error is a PostgreSQL not null constraint violation.
func IsNotNullViolation(err error) bool {
	return isPgError(err, "23502")
}

// ============================================================
// ID Generation
// ============================================================

// generateTestID creates a unique nanoid for test entities.
func generateTestID() string {
	id, _ := gonanoid.New()
	return id
}

// ============================================================
// pgtype Helpers
// ============================================================

// toPgDate converts a time.Time to pgtype.Date.
func toPgDate(t time.Time) pgtype.Date {
	return pgtype.Date{Time: t, Valid: true}
}

// toPgTimestamp converts a time.Time to pgtype.Timestamp.
func toPgTimestamp(t time.Time) pgtype.Timestamp {
	return pgtype.Timestamp{Time: t, Valid: true}
}

// toPgTime converts a time.Time to pgtype.Time (extracts only time portion).
func toPgTime(t time.Time) pgtype.Time {
	microseconds := int64(
		t.Hour(),
	)*3600*1e6 + int64(
		t.Minute(),
	)*60*1e6 + int64(
		t.Second(),
	)*1e6 + int64(
		t.Nanosecond()/1000,
	)
	return pgtype.Time{Microseconds: microseconds, Valid: true}
}

// strPtr returns a pointer to the given string.
func strPtr(s string) *string {
	return &s
}

// int32Ptr returns a pointer to the given int32.
func int32Ptr(i int32) *int32 {
	return &i
}

// ============================================================
// Factory: User
// ============================================================

// CreateTestUserOptions configures a test user.
type CreateTestUserOptions struct {
	ID           *string
	Email        *string
	PasswordHash *string
}

// CreateTestUser creates a user for testing. Returns the created user's ID.
func CreateTestUser(t *testing.T, q *Queries, opts CreateTestUserOptions) string {
	t.Helper()
	ctx := context.Background()

	id := generateTestID()
	if opts.ID != nil {
		id = *opts.ID
	}

	email := fmt.Sprintf("test-%s@example.com", id)
	if opts.Email != nil {
		email = *opts.Email
	}

	passwordHash := "$2a$10$testhashedpassword"
	if opts.PasswordHash != nil {
		passwordHash = *opts.PasswordHash
	}

	createdID, err := q.CreateUser(ctx, CreateUserParams{
		ID:           id,
		Email:        email,
		PasswordHash: passwordHash,
	})
	if err != nil {
		t.Fatalf("CreateTestUser failed: %v", err)
	}

	return createdID
}

// ============================================================
// Factory: Location
// ============================================================

// CreateTestLocationOptions configures a test location.
type CreateTestLocationOptions struct {
	ID         *string
	Name       *string
	PostalCode *string
	Address    *string
	Capacity   *int32
	Occupied   *int32
}

// CreateTestLocation creates a location for testing.
func CreateTestLocation(t *testing.T, q *Queries, opts CreateTestLocationOptions) string {
	t.Helper()
	ctx := context.Background()

	id := generateTestID()
	if opts.ID != nil {
		id = *opts.ID
	}

	name := fmt.Sprintf("Test Location %s", id[:8])
	if opts.Name != nil {
		name = *opts.Name
	}

	postalCode := "1234AB"
	if opts.PostalCode != nil {
		postalCode = *opts.PostalCode
	}

	address := "123 Test Street"
	if opts.Address != nil {
		address = *opts.Address
	}

	capacity := int32(10)
	if opts.Capacity != nil {
		capacity = *opts.Capacity
	}

	occupied := int32(0)
	if opts.Occupied != nil {
		occupied = *opts.Occupied
	}

	err := q.CreateLocation(ctx, CreateLocationParams{
		ID:         id,
		Name:       name,
		PostalCode: postalCode,
		Address:    address,
		Capacity:   capacity,
		Occupied:   occupied,
	})
	if err != nil {
		t.Fatalf("CreateTestLocation failed: %v", err)
	}

	return id
}

// ============================================================
// Factory: ReferringOrg
// ============================================================

// CreateTestReferringOrgOptions configures a test referring organization.
type CreateTestReferringOrgOptions struct {
	ID            *string
	Name          *string
	ContactPerson *string
	PhoneNumber   *string
	Email         *string
}

// CreateTestReferringOrg creates a referring organization for testing.
func CreateTestReferringOrg(t *testing.T, q *Queries, opts CreateTestReferringOrgOptions) string {
	t.Helper()
	ctx := context.Background()

	id := generateTestID()
	if opts.ID != nil {
		id = *opts.ID
	}

	name := fmt.Sprintf("Test Org %s", id[:8])
	if opts.Name != nil {
		name = *opts.Name
	}

	contactPerson := "John Doe"
	if opts.ContactPerson != nil {
		contactPerson = *opts.ContactPerson
	}

	phoneNumber := "+31612345678"
	if opts.PhoneNumber != nil {
		phoneNumber = *opts.PhoneNumber
	}

	email := fmt.Sprintf("org-%s@example.com", id[:8])
	if opts.Email != nil {
		email = *opts.Email
	}

	err := q.CreateReferringOrg(ctx, CreateReferringOrgParams{
		ID:            id,
		Name:          name,
		ContactPerson: contactPerson,
		PhoneNumber:   phoneNumber,
		Email:         email,
	})
	if err != nil {
		t.Fatalf("CreateTestReferringOrg failed: %v", err)
	}

	return id
}

// ============================================================
// Factory: Employee
// ============================================================

// CreateTestEmployeeOptions configures a test employee.
// UserID is required - the user must be created first.
type CreateTestEmployeeOptions struct {
	ID          *string
	UserID      string  // Required
	LocationID  *string // Optional - will be created if not provided
	FirstName   *string
	LastName    *string
	Bsn         *string
	DateOfBirth *time.Time
	PhoneNumber *string
	Gender      *GenderEnum
}

// CreateTestEmployee creates an employee for testing.
// Requires a User to be created first.
func CreateTestEmployee(t *testing.T, q *Queries, opts CreateTestEmployeeOptions) string {
	t.Helper()
	ctx := context.Background()

	if opts.UserID == "" {
		t.Fatal("CreateTestEmployee requires UserID")
	}

	id := generateTestID()
	if opts.ID != nil {
		id = *opts.ID
	}

	firstName := "Test"
	if opts.FirstName != nil {
		firstName = *opts.FirstName
	}

	lastName := "Employee"
	if opts.LastName != nil {
		lastName = *opts.LastName
	}

	// Generate unique BSN for each test
	bsn := fmt.Sprintf("EMP%s", id[:8])
	if opts.Bsn != nil {
		bsn = *opts.Bsn
	}

	dob := time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC)
	if opts.DateOfBirth != nil {
		dob = *opts.DateOfBirth
	}

	phoneNumber := "+31687654321"
	if opts.PhoneNumber != nil {
		phoneNumber = *opts.PhoneNumber
	}

	gender := GenderEnumOther
	if opts.Gender != nil {
		gender = *opts.Gender
	}

	// Create or use provided location
	locationID := ""
	if opts.LocationID != nil {
		locationID = *opts.LocationID
	} else {
		locationID = CreateTestLocation(t, q, CreateTestLocationOptions{})
	}

	err := q.CreateEmployee(ctx, CreateEmployeeParams{
		ID:          id,
		UserID:      opts.UserID,
		FirstName:   firstName,
		LastName:    lastName,
		Bsn:         bsn,
		DateOfBirth: toPgDate(dob),
		PhoneNumber: phoneNumber,
		Gender:      gender,
		LocationID:  locationID,
	})
	if err != nil {
		t.Fatalf("CreateTestEmployee failed: %v", err)
	}

	return id
}

// ============================================================
// Factory: RegistrationForm
// ============================================================

// CreateTestRegistrationFormOptions configures a test registration form.
type CreateTestRegistrationFormOptions struct {
	ID                 *string
	FirstName          *string
	LastName           *string
	Bsn                *string
	DateOfBirth        *time.Time
	Gender             *GenderEnum
	ReferringOrgID     *string // Optional FK
	CareType           *CareTypeEnum
	RegistrationReason *string
	AdditionalNotes    *string
}

// CreateTestRegistrationForm creates a registration form for testing.
func CreateTestRegistrationForm(
	t *testing.T,
	q *Queries,
	opts CreateTestRegistrationFormOptions,
) string {
	t.Helper()
	ctx := context.Background()

	id := generateTestID()
	if opts.ID != nil {
		id = *opts.ID
	}

	firstName := "Test"
	if opts.FirstName != nil {
		firstName = *opts.FirstName
	}

	lastName := "Client"
	if opts.LastName != nil {
		lastName = *opts.LastName
	}

	// Generate unique BSN
	bsn := fmt.Sprintf("REG%s", id[:8])
	if opts.Bsn != nil {
		bsn = *opts.Bsn
	}

	dob := time.Date(1985, 6, 15, 0, 0, 0, 0, time.UTC)
	if opts.DateOfBirth != nil {
		dob = *opts.DateOfBirth
	}

	gender := GenderEnumOther
	if opts.Gender != nil {
		gender = *opts.Gender
	}

	careType := CareTypeEnumProtectedLiving
	if opts.CareType != nil {
		careType = *opts.CareType
	}

	reason := "Test registration reason"
	if opts.RegistrationReason != nil {
		reason = *opts.RegistrationReason
	}

	err := q.CreateRegistrationForm(ctx, CreateRegistrationFormParams{
		ID:                 id,
		FirstName:          firstName,
		LastName:           lastName,
		Bsn:                bsn,
		DateOfBirth:        toPgDate(dob),
		Gender:             gender,
		RefferingOrgID:     opts.ReferringOrgID,
		CareType:           careType,
		RegistrationReason: reason,
		AdditionalNotes:    opts.AdditionalNotes,
	})
	if err != nil {
		t.Fatalf("CreateTestRegistrationForm failed: %v", err)
	}

	return id
}

// ============================================================
// Factory: IntakeForm
// ============================================================

// CreateTestIntakeFormOptions configures a test intake form.
// RegistrationFormID, LocationID, and CoordinatorID are required.
type CreateTestIntakeFormOptions struct {
	ID                 *string
	RegistrationFormID string // Required
	LocationID         string // Required
	CoordinatorID      string // Required
	IntakeDate         *time.Time
	IntakeTime         *time.Time
	FamilySituation    *string
	MainProvider       *string
	Limitations        *string
	FocusAreas         *string
	Goals              *string
	Notes              *string
}

// CreateTestIntakeForm creates an intake form for testing.
// Requires RegistrationForm, Location, and Employee (coordinator) to be created first.
func CreateTestIntakeForm(t *testing.T, q *Queries, opts CreateTestIntakeFormOptions) string {
	t.Helper()
	ctx := context.Background()

	if opts.RegistrationFormID == "" {
		t.Fatal("CreateTestIntakeForm requires RegistrationFormID")
	}
	if opts.LocationID == "" {
		t.Fatal("CreateTestIntakeForm requires LocationID")
	}
	if opts.CoordinatorID == "" {
		t.Fatal("CreateTestIntakeForm requires CoordinatorID")
	}

	id := generateTestID()
	if opts.ID != nil {
		id = *opts.ID
	}

	intakeDate := time.Now()
	if opts.IntakeDate != nil {
		intakeDate = *opts.IntakeDate
	}

	intakeTime := time.Date(0, 1, 1, 10, 0, 0, 0, time.UTC)
	if opts.IntakeTime != nil {
		intakeTime = *opts.IntakeTime
	}

	err := q.CreateIntakeForm(ctx, CreateIntakeFormParams{
		ID:                 id,
		RegistrationFormID: opts.RegistrationFormID,
		IntakeDate:         toPgDate(intakeDate),
		IntakeTime:         toPgTime(intakeTime),
		LocationID:         opts.LocationID,
		CoordinatorID:      opts.CoordinatorID,
		FamilySituation:    opts.FamilySituation,
		MainProvider:       opts.MainProvider,
		Limitations:        opts.Limitations,
		FocusAreas:         opts.FocusAreas,
		Notes:              opts.Notes,
	})
	if err != nil {
		t.Fatalf("CreateTestIntakeForm failed: %v", err)
	}

	return id
}

// ============================================================
// Factory: Client
// ============================================================

// CreateTestClientOptions configures a test client.
// RegistrationFormID, IntakeFormID, AssignedLocationID, and CoordinatorID are required.
type CreateTestClientOptions struct {
	ID                  *string
	FirstName           *string
	LastName            *string
	Bsn                 *string
	DateOfBirth         *time.Time
	PhoneNumber         *string
	Gender              *GenderEnum
	RegistrationFormID  string // Required
	IntakeFormID        string // Required
	CareType            *CareTypeEnum
	ReferringOrgID      *string
	WaitingListPriority *WaitingListPriorityEnum
	Status              *ClientStatusEnum
	AssignedLocationID  string // Required
	CoordinatorID       string // Required
	FamilySituation     *string
	Limitations         *string
	FocusAreas          *string
	Goals               []string
	Notes               *string
	CareStartDate       *time.Time
	CareEndDate         *time.Time
	DischargeDate       *time.Time
	ReasonForDischarge  *DischargeReasonEnum
	DischargeStatus     *DischargeStatusEnum
}

// CreateTestClient creates a client for testing.
// Requires RegistrationForm, IntakeForm, Location, and Employee (coordinator) to be created first.
func CreateTestClient(t *testing.T, q *Queries, opts CreateTestClientOptions) string {
	t.Helper()
	ctx := context.Background()

	if opts.RegistrationFormID == "" {
		t.Fatal("CreateTestClient requires RegistrationFormID")
	}
	if opts.IntakeFormID == "" {
		t.Fatal("CreateTestClient requires IntakeFormID")
	}
	if opts.AssignedLocationID == "" {
		t.Fatal("CreateTestClient requires AssignedLocationID")
	}
	if opts.CoordinatorID == "" {
		t.Fatal("CreateTestClient requires CoordinatorID")
	}

	id := generateTestID()
	if opts.ID != nil {
		id = *opts.ID
	}

	firstName := "Test"
	if opts.FirstName != nil {
		firstName = *opts.FirstName
	}

	lastName := "Client"
	if opts.LastName != nil {
		lastName = *opts.LastName
	}

	// Generate unique BSN
	bsn := fmt.Sprintf("CLI%s", id[:8])
	if opts.Bsn != nil {
		bsn = *opts.Bsn
	}

	dob := time.Date(1985, 6, 15, 0, 0, 0, 0, time.UTC)
	if opts.DateOfBirth != nil {
		dob = *opts.DateOfBirth
	}

	gender := GenderEnumOther
	if opts.Gender != nil {
		gender = *opts.Gender
	}

	careType := CareTypeEnumProtectedLiving
	if opts.CareType != nil {
		careType = *opts.CareType
	}

	priority := WaitingListPriorityEnumNormal
	if opts.WaitingListPriority != nil {
		priority = *opts.WaitingListPriority
	}

	_, err := q.CreateClient(ctx, CreateClientParams{
		ID:                  id,
		FirstName:           firstName,
		LastName:            lastName,
		Bsn:                 bsn,
		DateOfBirth:         toPgDate(dob),
		PhoneNumber:         opts.PhoneNumber,
		Gender:              gender,
		RegistrationFormID:  opts.RegistrationFormID,
		IntakeFormID:        opts.IntakeFormID,
		CareType:            careType,
		ReferringOrgID:      opts.ReferringOrgID,
		WaitingListPriority: priority,
		Status:              ClientStatusEnumWaitingList, // Always start as waiting_list to satisfy constraints
		AssignedLocationID:  opts.AssignedLocationID,
		CoordinatorID:       opts.CoordinatorID,
		FamilySituation:     opts.FamilySituation,
		Limitations:         opts.Limitations,
		FocusAreas:          opts.FocusAreas,
		Notes:               opts.Notes,
	})
	if err != nil {
		t.Fatalf("CreateTestClient failed: %v", err)
	}

	// Update with extra fields or different status if provided
	if (opts.Status != nil && *opts.Status != ClientStatusEnumWaitingList) ||
		opts.CareStartDate != nil || opts.CareEndDate != nil || opts.DischargeDate != nil ||
		opts.ReasonForDischarge != nil || opts.DischargeStatus != nil {
		updateParams := UpdateClientParams{
			ID: id,
		}
		if opts.Status != nil {
			updateParams.Status = NullClientStatusEnum{
				ClientStatusEnum: *opts.Status,
				Valid:            true,
			}
		}
		if opts.CareStartDate != nil {
			updateParams.CareStartDate = toPgDate(*opts.CareStartDate)
		}
		if opts.CareEndDate != nil {
			updateParams.CareEndDate = toPgDate(*opts.CareEndDate)
		}
		if opts.DischargeDate != nil {
			updateParams.DischargeDate = toPgDate(*opts.DischargeDate)
		}
		if opts.ReasonForDischarge != nil {
			updateParams.ReasonForDischarge = NullDischargeReasonEnum{
				DischargeReasonEnum: *opts.ReasonForDischarge,
				Valid:               true,
			}
		}
		if opts.DischargeStatus != nil {
			updateParams.DischargeStatus = NullDischargeStatusEnum{
				DischargeStatusEnum: *opts.DischargeStatus,
				Valid:               true,
			}
		}
		_, err = q.UpdateClient(ctx, updateParams)
		if err != nil {
			t.Fatalf("UpdateClient in CreateTestClient failed: %v", err)
		}
	}

	return id
}

// ============================================================
// Composite Factory: CreateFullClientDependencyChain
// ============================================================

// ClientDependencies holds all IDs needed to create a client.
type ClientDependencies struct {
	UserID             string
	EmployeeID         string
	LocationID         string
	RegistrationFormID string
	IntakeFormID       string
}

// CreateFullClientDependencyChain creates all the dependencies needed for a client.
// This is useful when you just need a client quickly without caring about the specifics.
func CreateFullClientDependencyChain(t *testing.T, q *Queries) ClientDependencies {
	t.Helper()

	// 1. Create User
	userID := CreateTestUser(t, q, CreateTestUserOptions{})

	// 2. Create Location
	locationID := CreateTestLocation(t, q, CreateTestLocationOptions{})

	// 3. Create Employee
	employeeID := CreateTestEmployee(t, q, CreateTestEmployeeOptions{
		UserID: userID,
	})

	// 4. Create RegistrationForm
	registrationFormID := CreateTestRegistrationForm(t, q, CreateTestRegistrationFormOptions{})

	// 5. Create IntakeForm
	intakeFormID := CreateTestIntakeForm(t, q, CreateTestIntakeFormOptions{
		RegistrationFormID: registrationFormID,
		LocationID:         locationID,
		CoordinatorID:      employeeID,
	})

	return ClientDependencies{
		UserID:             userID,
		EmployeeID:         employeeID,
		LocationID:         locationID,
		RegistrationFormID: registrationFormID,
		IntakeFormID:       intakeFormID,
	}
}

// CreateTestClientWithDependencies creates a client along with all its dependencies.
// Returns the client ID and the dependency chain.
func CreateTestClientWithDependencies(t *testing.T, q *Queries) (string, ClientDependencies) {
	t.Helper()

	deps := CreateFullClientDependencyChain(t, q)

	clientID := CreateTestClient(t, q, CreateTestClientOptions{
		RegistrationFormID: deps.RegistrationFormID,
		IntakeFormID:       deps.IntakeFormID,
		AssignedLocationID: deps.LocationID,
		CoordinatorID:      deps.EmployeeID,
	})

	return clientID, deps
}

// ============================================================
// Factory: Session
// ============================================================

// CreateTestSessionOptions configures a test session.
// UserID is required.
type CreateTestSessionOptions struct {
	ID          *string
	UserID      string // Required
	TokenFamily *string
	TokenHash   *string
	ExpiresAt   *time.Time
	UserAgent   *string
	IpAddress   *string
}

// CreateTestSession creates a session for testing.
func CreateTestSession(t *testing.T, q *Queries, opts CreateTestSessionOptions) string {
	t.Helper()
	ctx := context.Background()

	if opts.UserID == "" {
		t.Fatal("CreateTestSession requires UserID")
	}

	id := generateTestID()
	if opts.ID != nil {
		id = *opts.ID
	}

	tokenFamily := generateTestID()
	if opts.TokenFamily != nil {
		tokenFamily = *opts.TokenFamily
	}

	tokenHash := fmt.Sprintf("hash-%s", generateTestID())
	if opts.TokenHash != nil {
		tokenHash = *opts.TokenHash
	}

	expiresAt := time.Now().Add(24 * time.Hour)
	if opts.ExpiresAt != nil {
		expiresAt = *opts.ExpiresAt
	}

	err := q.CreateUserSession(ctx, CreateUserSessionParams{
		ID:          id,
		UserID:      opts.UserID,
		TokenFamily: tokenFamily,
		TokenHash:   tokenHash,
		ExpiresAt:   pgtype.Timestamptz{Time: expiresAt, Valid: true},
		UserAgent:   opts.UserAgent,
		IpAddress:   opts.IpAddress,
	})
	if err != nil {
		t.Fatalf("CreateTestSession failed: %v", err)
	}

	return id
}

// ============================================================
// Factory: Permission
// ============================================================

// CreateTestPermissionOptions configures a test permission.
type CreateTestPermissionOptions struct {
	ID          *string
	Resource    *string
	Action      *string
	Description *string
}

// CreateTestPermission creates a permission for testing.
func CreateTestPermission(t *testing.T, q *Queries, opts CreateTestPermissionOptions) string {
	t.Helper()
	ctx := context.Background()

	id := generateTestID()
	if opts.ID != nil {
		id = *opts.ID
	}

	resource := fmt.Sprintf("resource_%s", id[:8])
	if opts.Resource != nil {
		resource = *opts.Resource
	}

	action := "read"
	if opts.Action != nil {
		action = *opts.Action
	}

	permission, err := q.CreatePermission(ctx, CreatePermissionParams{
		ID:          id,
		Resource:    resource,
		Action:      action,
		Description: opts.Description,
	})
	if err != nil {
		t.Fatalf("CreateTestPermission failed: %v", err)
	}

	return permission.ID
}

// ============================================================
// Factory: Role
// ============================================================

// CreateTestRoleOptions configures a test role.
type CreateTestRoleOptions struct {
	ID          *string
	Name        *string
	Description *string
}

// CreateTestRole creates a role for testing.
func CreateTestRole(t *testing.T, q *Queries, opts CreateTestRoleOptions) string {
	t.Helper()
	ctx := context.Background()

	id := generateTestID()
	if opts.ID != nil {
		id = *opts.ID
	}

	name := fmt.Sprintf("role_%s", id[:8])
	if opts.Name != nil {
		name = *opts.Name
	}

	role, err := q.CreateRole(ctx, CreateRoleParams{
		ID:          id,
		Name:        name,
		Description: opts.Description,
	})
	if err != nil {
		t.Fatalf("CreateTestRole failed: %v", err)
	}

	return role.ID
}

// ============================================================
// Helper: Role-Permission Assignment
// ============================================================

// AssignTestPermissionToRole assigns a permission to a role for testing.
func AssignTestPermissionToRole(t *testing.T, q *Queries, roleID, permissionID string) {
	t.Helper()
	ctx := context.Background()

	err := q.AssignPermissionToRole(ctx, AssignPermissionToRoleParams{
		RoleID:       roleID,
		PermissionID: permissionID,
	})
	if err != nil {
		t.Fatalf("AssignTestPermissionToRole failed: %v", err)
	}
}

// ============================================================
// Helper: User-Role Assignment
// ============================================================

// AssignTestRoleToUser assigns a role to a user for testing.
func AssignTestRoleToUser(t *testing.T, q *Queries, userID, roleID string) {
	t.Helper()
	ctx := context.Background()

	err := q.AssignRoleToUser(ctx, AssignRoleToUserParams{
		UserID: userID,
		RoleID: roleID,
	})
	if err != nil {
		t.Fatalf("AssignTestRoleToUser failed: %v", err)
	}
}

// ============================================================
// Factory: Incident
// ============================================================

// CreateTestIncidentOptions configures a test incident.
// ClientID, LocationID, and CoordinatorID are required.
type CreateTestIncidentOptions struct {
	ID                  *string
	ClientID            string // Required
	IncidentDate        *time.Time
	IncidentTime        *time.Time
	IncidentType        *IncidentTypeEnum
	IncidentSeverity    *IncidentSeverityEnum
	LocationID          string // Required
	CoordinatorID       string // Required
	IncidentDescription *string
	ActionTaken         *string
	OtherParties        *string
	Status              *IncidentStatusEnum
}

// CreateTestIncident creates an incident for testing.
// Requires Client, Location, and Employee (coordinator) to be created first.
func CreateTestIncident(t *testing.T, q *Queries, opts CreateTestIncidentOptions) string {
	t.Helper()
	ctx := context.Background()

	if opts.ClientID == "" {
		t.Fatal("CreateTestIncident requires ClientID")
	}
	if opts.LocationID == "" {
		t.Fatal("CreateTestIncident requires LocationID")
	}
	if opts.CoordinatorID == "" {
		t.Fatal("CreateTestIncident requires CoordinatorID")
	}

	id := generateTestID()
	if opts.ID != nil {
		id = *opts.ID
	}

	incidentDate := time.Now()
	if opts.IncidentDate != nil {
		incidentDate = *opts.IncidentDate
	}

	incidentTime := time.Date(0, 1, 1, 14, 30, 0, 0, time.UTC)
	if opts.IncidentTime != nil {
		incidentTime = *opts.IncidentTime
	}

	incidentType := IncidentTypeEnumOther
	if opts.IncidentType != nil {
		incidentType = *opts.IncidentType
	}

	incidentSeverity := IncidentSeverityEnumMinor
	if opts.IncidentSeverity != nil {
		incidentSeverity = *opts.IncidentSeverity
	}

	description := "Test incident description"
	if opts.IncidentDescription != nil {
		description = *opts.IncidentDescription
	}

	actionTaken := "Test action taken"
	if opts.ActionTaken != nil {
		actionTaken = *opts.ActionTaken
	}

	status := IncidentStatusEnumPending
	if opts.Status != nil {
		status = *opts.Status
	}

	err := q.CreateIncident(ctx, CreateIncidentParams{
		ID:                  id,
		ClientID:            opts.ClientID,
		IncidentDate:        toPgDate(incidentDate),
		IncidentTime:        toPgTime(incidentTime),
		IncidentType:        incidentType,
		IncidentSeverity:    incidentSeverity,
		LocationID:          opts.LocationID,
		CoordinatorID:       opts.CoordinatorID,
		IncidentDescription: description,
		ActionTaken:         actionTaken,
		OtherParties:        opts.OtherParties,
		Status:              status,
	})
	if err != nil {
		t.Fatalf("CreateTestIncident failed: %v", err)
	}

	return id
}

// ============================================================
// Factory: Calendar
// ============================================================

// CreateTestAppointmentOptions configures a test appointment.
// OrganizerID is required.
type CreateTestAppointmentOptions struct {
	ID             *string
	Title          *string
	Description    *string
	StartTime      *time.Time
	EndTime        *time.Time
	Location       *string
	OrganizerID    string // Required
	Status         *AppointmentStatusEnum
	Type           *AppointmentTypeEnum
	RecurrenceRule *string
}

// CreateTestAppointment creates an appointment for testing.
func CreateTestAppointment(t *testing.T, q *Queries, opts CreateTestAppointmentOptions) string {
	t.Helper()
	ctx := context.Background()

	if opts.OrganizerID == "" {
		t.Fatal("CreateTestAppointment requires OrganizerID")
	}

	id := generateTestID()
	if opts.ID != nil {
		id = *opts.ID
	}

	title := "Test Appointment"
	if opts.Title != nil {
		title = *opts.Title
	}

	startTime := time.Now().Add(time.Hour)
	if opts.StartTime != nil {
		startTime = *opts.StartTime
	}

	endTime := startTime.Add(time.Hour)
	if opts.EndTime != nil {
		endTime = *opts.EndTime
	}

	status := AppointmentStatusEnumConfirmed
	if opts.Status != nil {
		status = *opts.Status
	}

	appType := AppointmentTypeEnumGeneral
	if opts.Type != nil {
		appType = *opts.Type
	}

	appointment, err := q.CreateAppointment(ctx, CreateAppointmentParams{
		ID:             id,
		Title:          title,
		Description:    opts.Description,
		StartTime:      pgtype.Timestamptz{Time: startTime, Valid: true},
		EndTime:        pgtype.Timestamptz{Time: endTime, Valid: true},
		Location:       opts.Location,
		OrganizerID:    opts.OrganizerID,
		Status:         NullAppointmentStatusEnum{AppointmentStatusEnum: status, Valid: true},
		Type:           appType,
		RecurrenceRule: opts.RecurrenceRule,
	})
	if err != nil {
		t.Fatalf("CreateTestAppointment failed: %v", err)
	}

	return appointment.ID
}

// CreateTestAppointmentParticipant creates a participant for an appointment.
func CreateTestAppointmentParticipant(t *testing.T, q *Queries, appointmentID, participantID string, pType ParticipantTypeEnum) {
	t.Helper()
	ctx := context.Background()

	err := q.AddAppointmentParticipant(ctx, AddAppointmentParticipantParams{
		AppointmentID:   appointmentID,
		ParticipantID:   participantID,
		ParticipantType: pType,
	})
	if err != nil {
		t.Fatalf("CreateTestAppointmentParticipant failed: %v", err)
	}
}

// CreateTestReminderOptions configures a test reminder.
// UserID is required.
type CreateTestReminderOptions struct {
	ID          *string
	UserID      string // Required (Employee ID)
	Title       *string
	Description *string
	DueTime     *time.Time
	IsCompleted *bool
}

// CreateTestReminder creates a reminder for testing.
func CreateTestReminder(t *testing.T, q *Queries, opts CreateTestReminderOptions) string {
	t.Helper()
	ctx := context.Background()

	if opts.UserID == "" {
		t.Fatal("CreateTestReminder requires UserID")
	}

	id := generateTestID()
	if opts.ID != nil {
		id = *opts.ID
	}

	title := "Test Reminder"
	if opts.Title != nil {
		title = *opts.Title
	}

	dueTime := time.Now().Add(time.Hour)
	if opts.DueTime != nil {
		dueTime = *opts.DueTime
	}

	isCompleted := false
	if opts.IsCompleted != nil {
		isCompleted = *opts.IsCompleted
	}

	reminder, err := q.CreateReminder(ctx, CreateReminderParams{
		ID:          id,
		UserID:      opts.UserID,
		Title:       title,
		Description: opts.Description,
		DueTime:     pgtype.Timestamptz{Time: dueTime, Valid: true},
		IsCompleted: &isCompleted,
	})
	if err != nil {
		t.Fatalf("CreateTestReminder failed: %v", err)
	}

	return reminder.ID
}
