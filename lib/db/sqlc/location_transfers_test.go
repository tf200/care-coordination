package db

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================
// Helper: Create test location transfer dependencies
// ============================================================

type locationTransferDeps struct {
	ClientID             string
	FromLocationID       string
	ToLocationID         string
	CurrentCoordinatorID string
	NewCoordinatorID     string
}

func createLocationTransferDeps(t *testing.T, q *Queries) locationTransferDeps {
	t.Helper()

	// Create locations
	fromLocationID := CreateTestLocation(t, q, CreateTestLocationOptions{Name: strPtr("From Location")})
	toLocationID := CreateTestLocation(t, q, CreateTestLocationOptions{Name: strPtr("To Location")})

	// Create coordinators
	user1ID := CreateTestUser(t, q, CreateTestUserOptions{})
	currentCoordinatorID := CreateTestEmployee(t, q, CreateTestEmployeeOptions{
		UserID:     user1ID,
		FirstName:  strPtr("Current"),
		LastName:   strPtr("Coordinator"),
		LocationID: &fromLocationID,
	})

	user2ID := CreateTestUser(t, q, CreateTestUserOptions{})
	newCoordinatorID := CreateTestEmployee(t, q, CreateTestEmployeeOptions{
		UserID:     user2ID,
		FirstName:  strPtr("New"),
		LastName:   strPtr("Coordinator"),
		LocationID: &toLocationID,
	})

	// Create client
	regFormID := CreateTestRegistrationForm(t, q, CreateTestRegistrationFormOptions{
		FirstName: strPtr("John"),
		LastName:  strPtr("Client"),
	})
	intakeFormID := CreateTestIntakeForm(t, q, CreateTestIntakeFormOptions{
		RegistrationFormID: regFormID,
		LocationID:         fromLocationID,
		CoordinatorID:      currentCoordinatorID,
	})
	clientID := CreateTestClient(t, q, CreateTestClientOptions{
		RegistrationFormID: regFormID,
		IntakeFormID:       intakeFormID,
		AssignedLocationID: fromLocationID,
		CoordinatorID:      currentCoordinatorID,
		FirstName:          strPtr("John"),
		LastName:           strPtr("Client"),
	})

	return locationTransferDeps{
		ClientID:             clientID,
		FromLocationID:       fromLocationID,
		ToLocationID:         toLocationID,
		CurrentCoordinatorID: currentCoordinatorID,
		NewCoordinatorID:     newCoordinatorID,
	}
}

// ============================================================
// Test: CreateLocationTransfer
// ============================================================

func TestCreateLocationTransfer(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(t *testing.T, q *Queries) CreateLocationTransferParams
		wantErr  bool
		checkErr func(t *testing.T, err error)
		validate func(t *testing.T, result CreateLocationTransferRow)
	}{
		{
			name: "success_with_all_fields",
			setup: func(t *testing.T, q *Queries) CreateLocationTransferParams {
				deps := createLocationTransferDeps(t, q)
				return CreateLocationTransferParams{
					ID:                   generateTestID(),
					ClientID:             deps.ClientID,
					FromLocationID:       &deps.FromLocationID,
					ToLocationID:         deps.ToLocationID,
					CurrentCoordinatorID: deps.CurrentCoordinatorID,
					NewCoordinatorID:     deps.NewCoordinatorID,
					TransferDate:         toPgTimestamp(time.Now().Add(24 * time.Hour)),
					Reason:               strPtr("Client requested transfer"),
				}
			},
			wantErr: false,
			validate: func(t *testing.T, result CreateLocationTransferRow) {
				assert.NotEmpty(t, result.ID)
				assert.NotEmpty(t, result.ClientID)
				assert.NotNil(t, result.FromLocationID)
				assert.NotEmpty(t, result.ToLocationID)
				assert.NotEmpty(t, result.CurrentCoordinatorID)
				assert.NotEmpty(t, result.NewCoordinatorID)
			},
		},
		{
			name: "success_without_from_location",
			setup: func(t *testing.T, q *Queries) CreateLocationTransferParams {
				deps := createLocationTransferDeps(t, q)
				return CreateLocationTransferParams{
					ID:                   generateTestID(),
					ClientID:             deps.ClientID,
					FromLocationID:       nil, // No from location (initial assignment)
					ToLocationID:         deps.ToLocationID,
					CurrentCoordinatorID: deps.CurrentCoordinatorID,
					NewCoordinatorID:     deps.NewCoordinatorID,
					TransferDate:         toPgTimestamp(time.Now()),
				}
			},
			wantErr: false,
			validate: func(t *testing.T, result CreateLocationTransferRow) {
				assert.Nil(t, result.FromLocationID)
			},
		},
		{
			name: "success_without_reason",
			setup: func(t *testing.T, q *Queries) CreateLocationTransferParams {
				deps := createLocationTransferDeps(t, q)
				return CreateLocationTransferParams{
					ID:                   generateTestID(),
					ClientID:             deps.ClientID,
					FromLocationID:       &deps.FromLocationID,
					ToLocationID:         deps.ToLocationID,
					CurrentCoordinatorID: deps.CurrentCoordinatorID,
					NewCoordinatorID:     deps.NewCoordinatorID,
					TransferDate:         toPgTimestamp(time.Now()),
					Reason:               nil,
				}
			},
			wantErr: false,
		},
		{
			name: "invalid_client_fk",
			setup: func(t *testing.T, q *Queries) CreateLocationTransferParams {
				deps := createLocationTransferDeps(t, q)
				return CreateLocationTransferParams{
					ID:                   generateTestID(),
					ClientID:             "nonexistent-client",
					FromLocationID:       &deps.FromLocationID,
					ToLocationID:         deps.ToLocationID,
					CurrentCoordinatorID: deps.CurrentCoordinatorID,
					NewCoordinatorID:     deps.NewCoordinatorID,
					TransferDate:         toPgTimestamp(time.Now()),
				}
			},
			wantErr: true,
			checkErr: func(t *testing.T, err error) {
				assert.True(t, IsForeignKeyViolation(err), "expected FK violation, got: %v", err)
			},
		},
		{
			name: "invalid_to_location_fk",
			setup: func(t *testing.T, q *Queries) CreateLocationTransferParams {
				deps := createLocationTransferDeps(t, q)
				return CreateLocationTransferParams{
					ID:                   generateTestID(),
					ClientID:             deps.ClientID,
					FromLocationID:       &deps.FromLocationID,
					ToLocationID:         "nonexistent-location",
					CurrentCoordinatorID: deps.CurrentCoordinatorID,
					NewCoordinatorID:     deps.NewCoordinatorID,
					TransferDate:         toPgTimestamp(time.Now()),
				}
			},
			wantErr: true,
			checkErr: func(t *testing.T, err error) {
				assert.True(t, IsForeignKeyViolation(err), "expected FK violation, got: %v", err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runTestWithTx(t, func(t *testing.T, q *Queries) {
				ctx := context.Background()
				params := tt.setup(t, q)

				result, err := q.CreateLocationTransfer(ctx, params)

				if tt.wantErr {
					require.Error(t, err)
					if tt.checkErr != nil {
						tt.checkErr(t, err)
					}
					return
				}

				require.NoError(t, err)
				if tt.validate != nil {
					tt.validate(t, result)
				}
			})
		})
	}
}

// ============================================================
// Test: GetLocationTransferByID
// ============================================================

func TestGetLocationTransferByID(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(t *testing.T, q *Queries) string // returns ID to query
		wantErr  bool
		checkErr func(t *testing.T, err error)
		validate func(t *testing.T, result GetLocationTransferByIDRow)
	}{
		{
			name: "found_with_all_details",
			setup: func(t *testing.T, q *Queries) string {
				ctx := context.Background()
				deps := createLocationTransferDeps(t, q)
				id := generateTestID()
				_, err := q.CreateLocationTransfer(ctx, CreateLocationTransferParams{
					ID:                   id,
					ClientID:             deps.ClientID,
					FromLocationID:       &deps.FromLocationID,
					ToLocationID:         deps.ToLocationID,
					CurrentCoordinatorID: deps.CurrentCoordinatorID,
					NewCoordinatorID:     deps.NewCoordinatorID,
					TransferDate:         toPgTimestamp(time.Now()),
					Reason:               strPtr("Test reason"),
				})
				require.NoError(t, err)
				return id
			},
			wantErr: false,
			validate: func(t *testing.T, result GetLocationTransferByIDRow) {
				assert.NotEmpty(t, result.ID)
				assert.Equal(t, "John", result.ClientFirstName)
				assert.Equal(t, "Client", result.ClientLastName)
				assert.Equal(t, "From Location", *result.FromLocationName)
				assert.Equal(t, "To Location", *result.ToLocationName)
				assert.Equal(t, "Current", *result.CurrentCoordinatorFirstName)
				assert.Equal(t, "New", *result.NewCoordinatorFirstName)
				assert.Equal(t, "Test reason", *result.Reason)
				assert.Equal(t, LocationTransferStatusEnumPending, result.Status)
			},
		},
		{
			name: "not_found",
			setup: func(t *testing.T, q *Queries) string {
				return "nonexistent-id"
			},
			wantErr: true,
			checkErr: func(t *testing.T, err error) {
				assert.True(t, errors.Is(err, pgx.ErrNoRows), "expected ErrNoRows, got: %v", err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runTestWithTx(t, func(t *testing.T, q *Queries) {
				ctx := context.Background()
				id := tt.setup(t, q)

				result, err := q.GetLocationTransferByID(ctx, id)

				if tt.wantErr {
					require.Error(t, err)
					if tt.checkErr != nil {
						tt.checkErr(t, err)
					}
					return
				}

				require.NoError(t, err)
				if tt.validate != nil {
					tt.validate(t, result)
				}
			})
		})
	}
}

// ============================================================
// Test: ListLocationTransfers
// ============================================================

func TestListLocationTransfers(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(t *testing.T, q *Queries)
		params   ListLocationTransfersParams
		validate func(t *testing.T, results []ListLocationTransfersRow)
	}{
		{
			name:   "empty",
			setup:  func(t *testing.T, q *Queries) {},
			params: ListLocationTransfersParams{Limit: 10, Offset: 0},
			validate: func(t *testing.T, results []ListLocationTransfersRow) {
				assert.Len(t, results, 0)
			},
		},
		{
			name: "with_data",
			setup: func(t *testing.T, q *Queries) {
				ctx := context.Background()
				deps := createLocationTransferDeps(t, q)
				q.CreateLocationTransfer(ctx, CreateLocationTransferParams{
					ID:                   generateTestID(),
					ClientID:             deps.ClientID,
					FromLocationID:       &deps.FromLocationID,
					ToLocationID:         deps.ToLocationID,
					CurrentCoordinatorID: deps.CurrentCoordinatorID,
					NewCoordinatorID:     deps.NewCoordinatorID,
					TransferDate:         toPgTimestamp(time.Now()),
				})
			},
			params: ListLocationTransfersParams{Limit: 10, Offset: 0},
			validate: func(t *testing.T, results []ListLocationTransfersRow) {
				assert.Len(t, results, 1)
				assert.Equal(t, int64(1), results[0].TotalCount)
			},
		},
		{
			name: "with_pagination",
			setup: func(t *testing.T, q *Queries) {
				ctx := context.Background()
				for i := 0; i < 5; i++ {
					deps := createLocationTransferDeps(t, q)
					q.CreateLocationTransfer(ctx, CreateLocationTransferParams{
						ID:                   generateTestID(),
						ClientID:             deps.ClientID,
						FromLocationID:       &deps.FromLocationID,
						ToLocationID:         deps.ToLocationID,
						CurrentCoordinatorID: deps.CurrentCoordinatorID,
						NewCoordinatorID:     deps.NewCoordinatorID,
						TransferDate:         toPgTimestamp(time.Now()),
					})
				}
			},
			params: ListLocationTransfersParams{Limit: 2, Offset: 0},
			validate: func(t *testing.T, results []ListLocationTransfersRow) {
				assert.Len(t, results, 2)
				assert.Equal(t, int64(5), results[0].TotalCount)
			},
		},
		{
			name: "with_search_by_client_name",
			setup: func(t *testing.T, q *Queries) {
				ctx := context.Background()

				// Create transfer for "Alice Wonderland"
				loc1 := CreateTestLocation(t, q, CreateTestLocationOptions{})
				loc2 := CreateTestLocation(t, q, CreateTestLocationOptions{})
				user1 := CreateTestUser(t, q, CreateTestUserOptions{})
				emp1 := CreateTestEmployee(t, q, CreateTestEmployeeOptions{UserID: user1, LocationID: &loc1})
				user2 := CreateTestUser(t, q, CreateTestUserOptions{})
				emp2 := CreateTestEmployee(t, q, CreateTestEmployeeOptions{UserID: user2, LocationID: &loc2})
				reg1 := CreateTestRegistrationForm(t, q, CreateTestRegistrationFormOptions{
					FirstName: strPtr("Alice"),
					LastName:  strPtr("Wonderland"),
				})
				intake1 := CreateTestIntakeForm(t, q, CreateTestIntakeFormOptions{
					RegistrationFormID: reg1, LocationID: loc1, CoordinatorID: emp1,
				})
				client1 := CreateTestClient(t, q, CreateTestClientOptions{
					RegistrationFormID: reg1, IntakeFormID: intake1,
					AssignedLocationID: loc1, CoordinatorID: emp1,
					FirstName: strPtr("Alice"),
					LastName:  strPtr("Wonderland"),
				})
				q.CreateLocationTransfer(ctx, CreateLocationTransferParams{
					ID:                   generateTestID(),
					ClientID:             client1,
					FromLocationID:       &loc1,
					ToLocationID:         loc2,
					CurrentCoordinatorID: emp1,
					NewCoordinatorID:     emp2,
					TransferDate:         toPgTimestamp(time.Now()),
				})

				// Create transfer for "Bob Builder"
				reg2 := CreateTestRegistrationForm(t, q, CreateTestRegistrationFormOptions{
					FirstName: strPtr("Bob"),
					LastName:  strPtr("Builder"),
				})
				intake2 := CreateTestIntakeForm(t, q, CreateTestIntakeFormOptions{
					RegistrationFormID: reg2, LocationID: loc1, CoordinatorID: emp1,
				})
				client2 := CreateTestClient(t, q, CreateTestClientOptions{
					RegistrationFormID: reg2, IntakeFormID: intake2,
					AssignedLocationID: loc1, CoordinatorID: emp1,
					FirstName: strPtr("Bob"),
					LastName:  strPtr("Builder"),
				})
				q.CreateLocationTransfer(ctx, CreateLocationTransferParams{
					ID:                   generateTestID(),
					ClientID:             client2,
					FromLocationID:       &loc1,
					ToLocationID:         loc2,
					CurrentCoordinatorID: emp1,
					NewCoordinatorID:     emp2,
					TransferDate:         toPgTimestamp(time.Now()),
				})
			},
			params: ListLocationTransfersParams{Limit: 10, Offset: 0, Search: strPtr("Alice")},
			validate: func(t *testing.T, results []ListLocationTransfersRow) {
				assert.Len(t, results, 1)
				assert.Equal(t, "Alice", results[0].ClientFirstName)
			},
		},
		{
			name: "search_no_match",
			setup: func(t *testing.T, q *Queries) {
				ctx := context.Background()
				deps := createLocationTransferDeps(t, q)
				q.CreateLocationTransfer(ctx, CreateLocationTransferParams{
					ID:                   generateTestID(),
					ClientID:             deps.ClientID,
					FromLocationID:       &deps.FromLocationID,
					ToLocationID:         deps.ToLocationID,
					CurrentCoordinatorID: deps.CurrentCoordinatorID,
					NewCoordinatorID:     deps.NewCoordinatorID,
					TransferDate:         toPgTimestamp(time.Now()),
				})
			},
			params: ListLocationTransfersParams{Limit: 10, Offset: 0, Search: strPtr("Zzzzzzzz")},
			validate: func(t *testing.T, results []ListLocationTransfersRow) {
				assert.Len(t, results, 0)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runTestWithTx(t, func(t *testing.T, q *Queries) {
				ctx := context.Background()
				tt.setup(t, q)

				results, err := q.ListLocationTransfers(ctx, tt.params)

				require.NoError(t, err)
				tt.validate(t, results)
			})
		})
	}
}

// ============================================================
// Test: ConfirmLocationTransfer
// ============================================================

func TestConfirmLocationTransfer(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(t *testing.T, q *Queries) string // returns ID
		wantErr  bool
		validate func(t *testing.T, q *Queries, id string)
	}{
		{
			name: "success",
			setup: func(t *testing.T, q *Queries) string {
				ctx := context.Background()
				deps := createLocationTransferDeps(t, q)
				id := generateTestID()
				q.CreateLocationTransfer(ctx, CreateLocationTransferParams{
					ID:                   id,
					ClientID:             deps.ClientID,
					FromLocationID:       &deps.FromLocationID,
					ToLocationID:         deps.ToLocationID,
					CurrentCoordinatorID: deps.CurrentCoordinatorID,
					NewCoordinatorID:     deps.NewCoordinatorID,
					TransferDate:         toPgTimestamp(time.Now()),
				})
				return id
			},
			wantErr: false,
			validate: func(t *testing.T, q *Queries, id string) {
				ctx := context.Background()
				result, err := q.GetLocationTransferByID(ctx, id)
				require.NoError(t, err)
				assert.Equal(t, LocationTransferStatusEnumApproved, result.Status)
				assert.True(t, result.TransferDate.Valid)
			},
		},
		{
			name: "nonexistent_id_no_error",
			setup: func(t *testing.T, q *Queries) string {
				return "nonexistent-id"
			},
			wantErr: false, // UPDATE on non-matching row doesn't error
		},
		{
			name: "already_approved",
			setup: func(t *testing.T, q *Queries) string {
				ctx := context.Background()
				deps := createLocationTransferDeps(t, q)
				id := generateTestID()
				q.CreateLocationTransfer(ctx, CreateLocationTransferParams{
					ID:                   id,
					ClientID:             deps.ClientID,
					FromLocationID:       &deps.FromLocationID,
					ToLocationID:         deps.ToLocationID,
					CurrentCoordinatorID: deps.CurrentCoordinatorID,
					NewCoordinatorID:     deps.NewCoordinatorID,
					TransferDate:         toPgTimestamp(time.Now()),
				})
				// Approve first
				q.ConfirmLocationTransfer(ctx, id)
				return id
			},
			wantErr: false, // No error, but status won't change (WHERE status = 'pending')
			validate: func(t *testing.T, q *Queries, id string) {
				ctx := context.Background()
				result, err := q.GetLocationTransferByID(ctx, id)
				require.NoError(t, err)
				assert.Equal(t, LocationTransferStatusEnumApproved, result.Status)
			},
		},
		{
			name: "rejected_cannot_be_confirmed",
			setup: func(t *testing.T, q *Queries) string {
				ctx := context.Background()
				deps := createLocationTransferDeps(t, q)
				id := generateTestID()
				q.CreateLocationTransfer(ctx, CreateLocationTransferParams{
					ID:                   id,
					ClientID:             deps.ClientID,
					FromLocationID:       &deps.FromLocationID,
					ToLocationID:         deps.ToLocationID,
					CurrentCoordinatorID: deps.CurrentCoordinatorID,
					NewCoordinatorID:     deps.NewCoordinatorID,
					TransferDate:         toPgTimestamp(time.Now()),
				})
				// Reject first
				q.RefuseLocationTransfer(ctx, RefuseLocationTransferParams{
					ID:              id,
					RejectionReason: strPtr("Rejected"),
				})
				return id
			},
			wantErr: false,
			validate: func(t *testing.T, q *Queries, id string) {
				ctx := context.Background()
				result, err := q.GetLocationTransferByID(ctx, id)
				require.NoError(t, err)
				// Status should still be rejected (confirm only works on pending)
				assert.Equal(t, LocationTransferStatusEnumRejected, result.Status)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runTestWithTx(t, func(t *testing.T, q *Queries) {
				ctx := context.Background()
				id := tt.setup(t, q)

				err := q.ConfirmLocationTransfer(ctx, id)

				if tt.wantErr {
					require.Error(t, err)
					return
				}

				require.NoError(t, err)
				if tt.validate != nil {
					tt.validate(t, q, id)
				}
			})
		})
	}
}

// ============================================================
// Test: RefuseLocationTransfer
// ============================================================

func TestRefuseLocationTransfer(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(t *testing.T, q *Queries) RefuseLocationTransferParams
		wantErr  bool
		validate func(t *testing.T, q *Queries, id string)
	}{
		{
			name: "success_with_reason",
			setup: func(t *testing.T, q *Queries) RefuseLocationTransferParams {
				ctx := context.Background()
				deps := createLocationTransferDeps(t, q)
				id := generateTestID()
				q.CreateLocationTransfer(ctx, CreateLocationTransferParams{
					ID:                   id,
					ClientID:             deps.ClientID,
					FromLocationID:       &deps.FromLocationID,
					ToLocationID:         deps.ToLocationID,
					CurrentCoordinatorID: deps.CurrentCoordinatorID,
					NewCoordinatorID:     deps.NewCoordinatorID,
					TransferDate:         toPgTimestamp(time.Now()),
				})
				return RefuseLocationTransferParams{
					ID:              id,
					RejectionReason: strPtr("Location at capacity"),
				}
			},
			wantErr: false,
			validate: func(t *testing.T, q *Queries, id string) {
				ctx := context.Background()
				result, err := q.GetLocationTransferByID(ctx, id)
				require.NoError(t, err)
				assert.Equal(t, LocationTransferStatusEnumRejected, result.Status)
				assert.Equal(t, "Location at capacity", *result.RejectionReason)
			},
		},
		{
			name: "success_without_reason",
			setup: func(t *testing.T, q *Queries) RefuseLocationTransferParams {
				ctx := context.Background()
				deps := createLocationTransferDeps(t, q)
				id := generateTestID()
				q.CreateLocationTransfer(ctx, CreateLocationTransferParams{
					ID:                   id,
					ClientID:             deps.ClientID,
					FromLocationID:       &deps.FromLocationID,
					ToLocationID:         deps.ToLocationID,
					CurrentCoordinatorID: deps.CurrentCoordinatorID,
					NewCoordinatorID:     deps.NewCoordinatorID,
					TransferDate:         toPgTimestamp(time.Now()),
				})
				return RefuseLocationTransferParams{
					ID:              id,
					RejectionReason: nil,
				}
			},
			wantErr: false,
			validate: func(t *testing.T, q *Queries, id string) {
				ctx := context.Background()
				result, err := q.GetLocationTransferByID(ctx, id)
				require.NoError(t, err)
				assert.Equal(t, LocationTransferStatusEnumRejected, result.Status)
				assert.Nil(t, result.RejectionReason)
			},
		},
		{
			name: "already_rejected",
			setup: func(t *testing.T, q *Queries) RefuseLocationTransferParams {
				ctx := context.Background()
				deps := createLocationTransferDeps(t, q)
				id := generateTestID()
				q.CreateLocationTransfer(ctx, CreateLocationTransferParams{
					ID:                   id,
					ClientID:             deps.ClientID,
					FromLocationID:       &deps.FromLocationID,
					ToLocationID:         deps.ToLocationID,
					CurrentCoordinatorID: deps.CurrentCoordinatorID,
					NewCoordinatorID:     deps.NewCoordinatorID,
					TransferDate:         toPgTimestamp(time.Now()),
				})
				// Reject first
				q.RefuseLocationTransfer(ctx, RefuseLocationTransferParams{
					ID:              id,
					RejectionReason: strPtr("First rejection"),
				})
				return RefuseLocationTransferParams{
					ID:              id,
					RejectionReason: strPtr("Second rejection"),
				}
			},
			wantErr: false,
			validate: func(t *testing.T, q *Queries, id string) {
				ctx := context.Background()
				result, err := q.GetLocationTransferByID(ctx, id)
				require.NoError(t, err)
				assert.Equal(t, LocationTransferStatusEnumRejected, result.Status)
				// Should keep first rejection reason (WHERE status = 'pending')
				assert.Equal(t, "First rejection", *result.RejectionReason)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runTestWithTx(t, func(t *testing.T, q *Queries) {
				ctx := context.Background()
				params := tt.setup(t, q)

				err := q.RefuseLocationTransfer(ctx, params)

				if tt.wantErr {
					require.Error(t, err)
					return
				}

				require.NoError(t, err)
				if tt.validate != nil {
					tt.validate(t, q, params.ID)
				}
			})
		})
	}
}

// ============================================================
// Test: UpdateLocationTransfer
// ============================================================

func TestUpdateLocationTransfer(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(t *testing.T, q *Queries) (string, UpdateLocationTransferParams)
		wantErr  bool
		validate func(t *testing.T, q *Queries, id string)
	}{
		{
			name: "update_reason",
			setup: func(t *testing.T, q *Queries) (string, UpdateLocationTransferParams) {
				ctx := context.Background()
				deps := createLocationTransferDeps(t, q)
				id := generateTestID()
				q.CreateLocationTransfer(ctx, CreateLocationTransferParams{
					ID:                   id,
					ClientID:             deps.ClientID,
					FromLocationID:       &deps.FromLocationID,
					ToLocationID:         deps.ToLocationID,
					CurrentCoordinatorID: deps.CurrentCoordinatorID,
					NewCoordinatorID:     deps.NewCoordinatorID,
					TransferDate:         toPgTimestamp(time.Now()),
					Reason:               strPtr("Original reason"),
				})
				return id, UpdateLocationTransferParams{
					ID:     id,
					Reason: strPtr("Updated reason"),
				}
			},
			wantErr: false,
			validate: func(t *testing.T, q *Queries, id string) {
				ctx := context.Background()
				result, err := q.GetLocationTransferByID(ctx, id)
				require.NoError(t, err)
				assert.Equal(t, "Updated reason", *result.Reason)
			},
		},
		{
			name: "update_to_location",
			setup: func(t *testing.T, q *Queries) (string, UpdateLocationTransferParams) {
				ctx := context.Background()
				deps := createLocationTransferDeps(t, q)
				newLocationID := CreateTestLocation(t, q, CreateTestLocationOptions{Name: strPtr("New Destination")})

				id := generateTestID()
				q.CreateLocationTransfer(ctx, CreateLocationTransferParams{
					ID:                   id,
					ClientID:             deps.ClientID,
					FromLocationID:       &deps.FromLocationID,
					ToLocationID:         deps.ToLocationID,
					CurrentCoordinatorID: deps.CurrentCoordinatorID,
					NewCoordinatorID:     deps.NewCoordinatorID,
					TransferDate:         toPgTimestamp(time.Now()),
				})
				return id, UpdateLocationTransferParams{
					ID:           id,
					ToLocationID: &newLocationID,
				}
			},
			wantErr: false,
			validate: func(t *testing.T, q *Queries, id string) {
				ctx := context.Background()
				result, err := q.GetLocationTransferByID(ctx, id)
				require.NoError(t, err)
				assert.Equal(t, "New Destination", *result.ToLocationName)
			},
		},
		{
			name: "update_new_coordinator",
			setup: func(t *testing.T, q *Queries) (string, UpdateLocationTransferParams) {
				ctx := context.Background()
				deps := createLocationTransferDeps(t, q)

				// Create another coordinator
				user3 := CreateTestUser(t, q, CreateTestUserOptions{})
				newCoordinator := CreateTestEmployee(t, q, CreateTestEmployeeOptions{
					UserID:     user3,
					FirstName:  strPtr("Another"),
					LastName:   strPtr("Coordinator"),
					LocationID: &deps.ToLocationID,
				})

				id := generateTestID()
				q.CreateLocationTransfer(ctx, CreateLocationTransferParams{
					ID:                   id,
					ClientID:             deps.ClientID,
					FromLocationID:       &deps.FromLocationID,
					ToLocationID:         deps.ToLocationID,
					CurrentCoordinatorID: deps.CurrentCoordinatorID,
					NewCoordinatorID:     deps.NewCoordinatorID,
					TransferDate:         toPgTimestamp(time.Now()),
				})
				return id, UpdateLocationTransferParams{
					ID:               id,
					NewCoordinatorID: &newCoordinator,
				}
			},
			wantErr: false,
			validate: func(t *testing.T, q *Queries, id string) {
				ctx := context.Background()
				result, err := q.GetLocationTransferByID(ctx, id)
				require.NoError(t, err)
				assert.Equal(t, "Another", *result.NewCoordinatorFirstName)
			},
		},
		{
			name: "cannot_update_approved",
			setup: func(t *testing.T, q *Queries) (string, UpdateLocationTransferParams) {
				ctx := context.Background()
				deps := createLocationTransferDeps(t, q)
				id := generateTestID()
				q.CreateLocationTransfer(ctx, CreateLocationTransferParams{
					ID:                   id,
					ClientID:             deps.ClientID,
					FromLocationID:       &deps.FromLocationID,
					ToLocationID:         deps.ToLocationID,
					CurrentCoordinatorID: deps.CurrentCoordinatorID,
					NewCoordinatorID:     deps.NewCoordinatorID,
					TransferDate:         toPgTimestamp(time.Now()),
					Reason:               strPtr("Original"),
				})
				// Approve it
				q.ConfirmLocationTransfer(ctx, id)
				return id, UpdateLocationTransferParams{
					ID:     id,
					Reason: strPtr("Cannot update this"),
				}
			},
			wantErr: false,
			validate: func(t *testing.T, q *Queries, id string) {
				ctx := context.Background()
				result, err := q.GetLocationTransferByID(ctx, id)
				require.NoError(t, err)
				// Should still have original reason (update only works on pending)
				assert.Equal(t, "Original", *result.Reason)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runTestWithTx(t, func(t *testing.T, q *Queries) {
				ctx := context.Background()
				id, params := tt.setup(t, q)

				err := q.UpdateLocationTransfer(ctx, params)

				if tt.wantErr {
					require.Error(t, err)
					return
				}

				require.NoError(t, err)
				if tt.validate != nil {
					tt.validate(t, q, id)
				}
			})
		})
	}
}

// ============================================================
// Test: GetLocationTransferStats
// ============================================================

func TestGetLocationTransferStats(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(t *testing.T, q *Queries)
		validate func(t *testing.T, stats GetLocationTransferStatsRow)
	}{
		{
			name:  "empty_database",
			setup: func(t *testing.T, q *Queries) {},
			validate: func(t *testing.T, stats GetLocationTransferStatsRow) {
				assert.Equal(t, int64(0), stats.TotalCount)
				assert.Equal(t, int64(0), stats.PendingCount)
				assert.Equal(t, int64(0), stats.ApprovedCount)
				assert.Equal(t, int64(0), stats.RejectedCount)
				assert.Equal(t, int32(0), stats.ApprovalRate)
			},
		},
		{
			name: "with_pending_only",
			setup: func(t *testing.T, q *Queries) {
				ctx := context.Background()
				for i := 0; i < 3; i++ {
					deps := createLocationTransferDeps(t, q)
					q.CreateLocationTransfer(ctx, CreateLocationTransferParams{
						ID:                   generateTestID(),
						ClientID:             deps.ClientID,
						FromLocationID:       &deps.FromLocationID,
						ToLocationID:         deps.ToLocationID,
						CurrentCoordinatorID: deps.CurrentCoordinatorID,
						NewCoordinatorID:     deps.NewCoordinatorID,
						TransferDate:         toPgTimestamp(time.Now()),
					})
				}
			},
			validate: func(t *testing.T, stats GetLocationTransferStatsRow) {
				assert.Equal(t, int64(3), stats.TotalCount)
				assert.Equal(t, int64(3), stats.PendingCount)
				assert.Equal(t, int64(0), stats.ApprovedCount)
				assert.Equal(t, int64(0), stats.RejectedCount)
				assert.Equal(t, int32(0), stats.ApprovalRate) // No approved/rejected yet
			},
		},
		{
			name: "with_mixed_statuses",
			setup: func(t *testing.T, q *Queries) {
				ctx := context.Background()

				// Create 2 approved
				for i := 0; i < 2; i++ {
					deps := createLocationTransferDeps(t, q)
					id := generateTestID()
					q.CreateLocationTransfer(ctx, CreateLocationTransferParams{
						ID:                   id,
						ClientID:             deps.ClientID,
						FromLocationID:       &deps.FromLocationID,
						ToLocationID:         deps.ToLocationID,
						CurrentCoordinatorID: deps.CurrentCoordinatorID,
						NewCoordinatorID:     deps.NewCoordinatorID,
						TransferDate:         toPgTimestamp(time.Now()),
					})
					q.ConfirmLocationTransfer(ctx, id)
				}

				// Create 2 rejected (makes 50% approval rate - a whole number)
				for i := 0; i < 2; i++ {
					deps := createLocationTransferDeps(t, q)
					rejectedID := generateTestID()
					q.CreateLocationTransfer(ctx, CreateLocationTransferParams{
						ID:                   rejectedID,
						ClientID:             deps.ClientID,
						FromLocationID:       &deps.FromLocationID,
						ToLocationID:         deps.ToLocationID,
						CurrentCoordinatorID: deps.CurrentCoordinatorID,
						NewCoordinatorID:     deps.NewCoordinatorID,
						TransferDate:         toPgTimestamp(time.Now()),
					})
					q.RefuseLocationTransfer(ctx, RefuseLocationTransferParams{
						ID:              rejectedID,
						RejectionReason: strPtr("Rejected"),
					})
				}
			},
			validate: func(t *testing.T, stats GetLocationTransferStatsRow) {
				assert.Equal(t, int64(4), stats.TotalCount)
				assert.Equal(t, int64(0), stats.PendingCount)
				assert.Equal(t, int64(2), stats.ApprovedCount)
				assert.Equal(t, int64(2), stats.RejectedCount)
				// Approval rate: 2 approved / 4 (approved + rejected) = 50%
				assert.Equal(t, int32(50), stats.ApprovalRate)
			},
		},
		{
			name: "all_approved",
			setup: func(t *testing.T, q *Queries) {
				ctx := context.Background()
				for i := 0; i < 3; i++ {
					deps := createLocationTransferDeps(t, q)
					id := generateTestID()
					q.CreateLocationTransfer(ctx, CreateLocationTransferParams{
						ID:                   id,
						ClientID:             deps.ClientID,
						FromLocationID:       &deps.FromLocationID,
						ToLocationID:         deps.ToLocationID,
						CurrentCoordinatorID: deps.CurrentCoordinatorID,
						NewCoordinatorID:     deps.NewCoordinatorID,
						TransferDate:         toPgTimestamp(time.Now()),
					})
					q.ConfirmLocationTransfer(ctx, id)
				}
			},
			validate: func(t *testing.T, stats GetLocationTransferStatsRow) {
				assert.Equal(t, int64(3), stats.TotalCount)
				assert.Equal(t, int64(0), stats.PendingCount)
				assert.Equal(t, int64(3), stats.ApprovedCount)
				assert.Equal(t, int64(0), stats.RejectedCount)
				assert.Equal(t, int32(100), stats.ApprovalRate)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runTestWithTx(t, func(t *testing.T, q *Queries) {
				ctx := context.Background()
				tt.setup(t, q)

				stats, err := q.GetLocationTransferStats(ctx)

				require.NoError(t, err)
				tt.validate(t, stats)
			})
		})
	}
}
