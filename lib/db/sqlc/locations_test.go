package db

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================
// Test: CreateLocation
// ============================================================

func TestCreateLocation(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(t *testing.T, q *Queries) CreateLocationParams
		wantErr  bool
		checkErr func(t *testing.T, err error)
		validate func(t *testing.T, q *Queries, params CreateLocationParams)
	}{
		{
			name: "success",
			setup: func(t *testing.T, q *Queries) CreateLocationParams {
				return CreateLocationParams{
					ID:         generateTestID(),
					Name:       "Main Office",
					PostalCode: "1234AB",
					Address:    "123 Test Street",
					Capacity:   20,
					Occupied:   5,
				}
			},
			wantErr: false,
			validate: func(t *testing.T, q *Queries, params CreateLocationParams) {
				ctx := context.Background()
				results, err := q.ListLocations(ctx, ListLocationsParams{
					Limit:  10,
					Offset: 0,
					Search: &params.Name,
				})
				require.NoError(t, err)
				require.Len(t, results, 1)
				assert.Equal(t, params.Name, results[0].Name)
				assert.Equal(t, params.PostalCode, results[0].PostalCode)
				assert.Equal(t, params.Address, results[0].Address)
				assert.Equal(t, params.Capacity, results[0].Capacity)
				assert.Equal(t, params.Occupied, results[0].Occupied)
			},
		},
		{
			name: "success_zero_occupied",
			setup: func(t *testing.T, q *Queries) CreateLocationParams {
				return CreateLocationParams{
					ID:         generateTestID(),
					Name:       "Empty Location",
					PostalCode: "5678CD",
					Address:    "456 Empty Road",
					Capacity:   10,
					Occupied:   0,
				}
			},
			wantErr: false,
		},
		{
			name: "duplicate_id",
			setup: func(t *testing.T, q *Queries) CreateLocationParams {
				existingID := generateTestID()
				CreateTestLocation(t, q, CreateTestLocationOptions{ID: &existingID})
				return CreateLocationParams{
					ID:         existingID,
					Name:       "Different Name",
					PostalCode: "9999XX",
					Address:    "Different Address",
					Capacity:   15,
					Occupied:   0,
				}
			},
			wantErr: true,
			checkErr: func(t *testing.T, err error) {
				assert.True(t, IsUniqueViolation(err), "expected unique violation, got: %v", err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runTestWithTx(t, func(t *testing.T, q *Queries) {
				ctx := context.Background()
				params := tt.setup(t, q)

				err := q.CreateLocation(ctx, params)

				if tt.wantErr {
					require.Error(t, err)
					if tt.checkErr != nil {
						tt.checkErr(t, err)
					}
					return
				}

				require.NoError(t, err)
				if tt.validate != nil {
					tt.validate(t, q, params)
				}
			})
		})
	}
}

// ============================================================
// Test: ListLocations
// ============================================================

func TestListLocations(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(t *testing.T, q *Queries)
		params   ListLocationsParams
		validate func(t *testing.T, results []ListLocationsRow)
	}{
		{
			name:   "empty",
			setup:  func(t *testing.T, q *Queries) {},
			params: ListLocationsParams{Limit: 10, Offset: 0},
			validate: func(t *testing.T, results []ListLocationsRow) {
				assert.Len(t, results, 0)
			},
		},
		{
			name: "with_data",
			setup: func(t *testing.T, q *Queries) {
				CreateTestLocation(t, q, CreateTestLocationOptions{Name: strPtr("Location A")})
				CreateTestLocation(t, q, CreateTestLocationOptions{Name: strPtr("Location B")})
			},
			params: ListLocationsParams{Limit: 10, Offset: 0},
			validate: func(t *testing.T, results []ListLocationsRow) {
				assert.Len(t, results, 2)
				assert.Equal(t, int64(2), results[0].TotalCount)
			},
		},
		{
			name: "with_pagination",
			setup: func(t *testing.T, q *Queries) {
				for i := 0; i < 5; i++ {
					CreateTestLocation(t, q, CreateTestLocationOptions{})
				}
			},
			params: ListLocationsParams{Limit: 2, Offset: 0},
			validate: func(t *testing.T, results []ListLocationsRow) {
				assert.Len(t, results, 2)
				assert.Equal(t, int64(5), results[0].TotalCount)
			},
		},
		{
			name: "with_offset",
			setup: func(t *testing.T, q *Queries) {
				for i := 0; i < 5; i++ {
					CreateTestLocation(t, q, CreateTestLocationOptions{})
				}
			},
			params: ListLocationsParams{Limit: 10, Offset: 3},
			validate: func(t *testing.T, results []ListLocationsRow) {
				assert.Len(t, results, 2) // 5 total - 3 offset = 2 remaining
			},
		},
		{
			name: "search_by_name",
			setup: func(t *testing.T, q *Queries) {
				CreateTestLocation(t, q, CreateTestLocationOptions{Name: strPtr("Main Office")})
				CreateTestLocation(t, q, CreateTestLocationOptions{Name: strPtr("Branch Office")})
				CreateTestLocation(t, q, CreateTestLocationOptions{Name: strPtr("Warehouse")})
			},
			params: ListLocationsParams{Limit: 10, Offset: 0, Search: strPtr("Office")},
			validate: func(t *testing.T, results []ListLocationsRow) {
				assert.Len(t, results, 2)
				// Both contain "Office"
				for _, r := range results {
					assert.Contains(t, r.Name, "Office")
				}
			},
		},
		{
			name: "search_by_postal_code",
			setup: func(t *testing.T, q *Queries) {
				CreateTestLocation(t, q, CreateTestLocationOptions{
					Name:       strPtr("Location 1"),
					PostalCode: strPtr("1234AB"),
				})
				CreateTestLocation(t, q, CreateTestLocationOptions{
					Name:       strPtr("Location 2"),
					PostalCode: strPtr("5678CD"),
				})
			},
			params: ListLocationsParams{Limit: 10, Offset: 0, Search: strPtr("1234")},
			validate: func(t *testing.T, results []ListLocationsRow) {
				assert.Len(t, results, 1)
				assert.Equal(t, "1234AB", results[0].PostalCode)
			},
		},
		{
			name: "search_by_address",
			setup: func(t *testing.T, q *Queries) {
				CreateTestLocation(t, q, CreateTestLocationOptions{
					Name:    strPtr("Location 1"),
					Address: strPtr("123 Main Street"),
				})
				CreateTestLocation(t, q, CreateTestLocationOptions{
					Name:    strPtr("Location 2"),
					Address: strPtr("456 Side Avenue"),
				})
			},
			params: ListLocationsParams{Limit: 10, Offset: 0, Search: strPtr("Main Street")},
			validate: func(t *testing.T, results []ListLocationsRow) {
				assert.Len(t, results, 1)
				assert.Contains(t, results[0].Address, "Main Street")
			},
		},
		{
			name: "search_case_insensitive",
			setup: func(t *testing.T, q *Queries) {
				CreateTestLocation(t, q, CreateTestLocationOptions{Name: strPtr("UPPER CASE")})
				CreateTestLocation(t, q, CreateTestLocationOptions{Name: strPtr("lower case")})
			},
			params: ListLocationsParams{Limit: 10, Offset: 0, Search: strPtr("upper")},
			validate: func(t *testing.T, results []ListLocationsRow) {
				assert.Len(t, results, 1)
				assert.Equal(t, "UPPER CASE", results[0].Name)
			},
		},
		{
			name: "search_no_match",
			setup: func(t *testing.T, q *Queries) {
				CreateTestLocation(t, q, CreateTestLocationOptions{Name: strPtr("Test Location")})
			},
			params: ListLocationsParams{Limit: 10, Offset: 0, Search: strPtr("Zzzzzzz")},
			validate: func(t *testing.T, results []ListLocationsRow) {
				assert.Len(t, results, 0)
			},
		},
		{
			name: "excludes_soft_deleted",
			setup: func(t *testing.T, q *Queries) {
				ctx := context.Background()
				// Create and soft delete one
				id := CreateTestLocation(t, q, CreateTestLocationOptions{Name: strPtr("Deleted")})
				q.SoftDeleteLocation(ctx, id)

				// Create active one
				CreateTestLocation(t, q, CreateTestLocationOptions{Name: strPtr("Active")})
			},
			params: ListLocationsParams{Limit: 10, Offset: 0},
			validate: func(t *testing.T, results []ListLocationsRow) {
				assert.Len(t, results, 1)
				assert.Equal(t, "Active", results[0].Name)
			},
		},
		{
			name: "ordered_by_name",
			setup: func(t *testing.T, q *Queries) {
				CreateTestLocation(t, q, CreateTestLocationOptions{Name: strPtr("Charlie")})
				CreateTestLocation(t, q, CreateTestLocationOptions{Name: strPtr("Alpha")})
				CreateTestLocation(t, q, CreateTestLocationOptions{Name: strPtr("Beta")})
			},
			params: ListLocationsParams{Limit: 10, Offset: 0},
			validate: func(t *testing.T, results []ListLocationsRow) {
				assert.Len(t, results, 3)
				assert.Equal(t, "Alpha", results[0].Name)
				assert.Equal(t, "Beta", results[1].Name)
				assert.Equal(t, "Charlie", results[2].Name)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runTestWithTx(t, func(t *testing.T, q *Queries) {
				ctx := context.Background()
				tt.setup(t, q)

				results, err := q.ListLocations(ctx, tt.params)

				require.NoError(t, err)
				tt.validate(t, results)
			})
		})
	}
}

// ============================================================
// Test: UpdateLocation
// ============================================================

func TestUpdateLocation(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(t *testing.T, q *Queries) (string, UpdateLocationParams)
		wantErr  bool
		validate func(t *testing.T, q *Queries, id string)
	}{
		{
			name: "update_name",
			setup: func(t *testing.T, q *Queries) (string, UpdateLocationParams) {
				id := CreateTestLocation(t, q, CreateTestLocationOptions{Name: strPtr("Original")})
				return id, UpdateLocationParams{
					ID:   id,
					Name: strPtr("Updated"),
				}
			},
			wantErr: false,
			validate: func(t *testing.T, q *Queries, id string) {
				ctx := context.Background()
				results, err := q.ListLocations(ctx, ListLocationsParams{
					Limit:  10,
					Offset: 0,
					Search: strPtr("Updated"),
				})
				require.NoError(t, err)
				require.Len(t, results, 1)
				assert.Equal(t, "Updated", results[0].Name)
			},
		},
		{
			name: "update_postal_code",
			setup: func(t *testing.T, q *Queries) (string, UpdateLocationParams) {
				id := CreateTestLocation(t, q, CreateTestLocationOptions{PostalCode: strPtr("1111AA")})
				return id, UpdateLocationParams{
					ID:         id,
					PostalCode: strPtr("9999ZZ"),
				}
			},
			wantErr: false,
			validate: func(t *testing.T, q *Queries, id string) {
				ctx := context.Background()
				results, err := q.ListLocations(ctx, ListLocationsParams{
					Limit:  10,
					Offset: 0,
					Search: strPtr("9999ZZ"),
				})
				require.NoError(t, err)
				require.Len(t, results, 1)
				assert.Equal(t, "9999ZZ", results[0].PostalCode)
			},
		},
		{
			name: "update_capacity",
			setup: func(t *testing.T, q *Queries) (string, UpdateLocationParams) {
				id := CreateTestLocation(t, q, CreateTestLocationOptions{Capacity: int32Ptr(10)})
				return id, UpdateLocationParams{
					ID:       id,
					Capacity: int32Ptr(50),
				}
			},
			wantErr: false,
			validate: func(t *testing.T, q *Queries, id string) {
				ctx := context.Background()
				results, err := q.ListLocations(ctx, ListLocationsParams{Limit: 100, Offset: 0})
				require.NoError(t, err)
				var found *ListLocationsRow
				for i := range results {
					if results[i].ID == id {
						found = &results[i]
						break
					}
				}
				require.NotNil(t, found)
				assert.Equal(t, int32(50), found.Capacity)
			},
		},
		{
			name: "update_occupied",
			setup: func(t *testing.T, q *Queries) (string, UpdateLocationParams) {
				id := CreateTestLocation(t, q, CreateTestLocationOptions{Occupied: int32Ptr(0)})
				return id, UpdateLocationParams{
					ID:       id,
					Occupied: int32Ptr(15),
				}
			},
			wantErr: false,
			validate: func(t *testing.T, q *Queries, id string) {
				ctx := context.Background()
				results, err := q.ListLocations(ctx, ListLocationsParams{Limit: 100, Offset: 0})
				require.NoError(t, err)
				var found *ListLocationsRow
				for i := range results {
					if results[i].ID == id {
						found = &results[i]
						break
					}
				}
				require.NotNil(t, found)
				assert.Equal(t, int32(15), found.Occupied)
			},
		},
		{
			name: "update_multiple_fields",
			setup: func(t *testing.T, q *Queries) (string, UpdateLocationParams) {
				id := CreateTestLocation(t, q, CreateTestLocationOptions{
					Name:       strPtr("Old Name"),
					PostalCode: strPtr("0000AA"),
					Address:    strPtr("Old Address"),
				})
				return id, UpdateLocationParams{
					ID:         id,
					Name:       strPtr("New Name"),
					PostalCode: strPtr("1111BB"),
					Address:    strPtr("New Address"),
				}
			},
			wantErr: false,
			validate: func(t *testing.T, q *Queries, id string) {
				ctx := context.Background()
				results, err := q.ListLocations(ctx, ListLocationsParams{
					Limit:  10,
					Offset: 0,
					Search: strPtr("New Name"),
				})
				require.NoError(t, err)
				require.Len(t, results, 1)
				assert.Equal(t, "New Name", results[0].Name)
				assert.Equal(t, "1111BB", results[0].PostalCode)
				assert.Equal(t, "New Address", results[0].Address)
			},
		},
		{
			name: "partial_update_preserves_other_fields",
			setup: func(t *testing.T, q *Queries) (string, UpdateLocationParams) {
				id := CreateTestLocation(t, q, CreateTestLocationOptions{
					Name:       strPtr("Keep This Name"),
					PostalCode: strPtr("1234AB"),
					Capacity:   int32Ptr(25),
				})
				// Only update postal code
				return id, UpdateLocationParams{
					ID:         id,
					PostalCode: strPtr("9999ZZ"),
				}
			},
			wantErr: false,
			validate: func(t *testing.T, q *Queries, id string) {
				ctx := context.Background()
				results, err := q.ListLocations(ctx, ListLocationsParams{
					Limit:  10,
					Offset: 0,
					Search: strPtr("Keep This Name"),
				})
				require.NoError(t, err)
				require.Len(t, results, 1)
				assert.Equal(t, "Keep This Name", results[0].Name) // Preserved
				assert.Equal(t, "9999ZZ", results[0].PostalCode)   // Updated
				assert.Equal(t, int32(25), results[0].Capacity)    // Preserved
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runTestWithTx(t, func(t *testing.T, q *Queries) {
				ctx := context.Background()
				id, params := tt.setup(t, q)

				err := q.UpdateLocation(ctx, params)

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
// Test: SoftDeleteLocation
// ============================================================

func TestSoftDeleteLocation(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(t *testing.T, q *Queries) string // returns ID to delete
		wantErr  bool
		validate func(t *testing.T, q *Queries, id string)
	}{
		{
			name: "success",
			setup: func(t *testing.T, q *Queries) string {
				return CreateTestLocation(t, q, CreateTestLocationOptions{Name: strPtr("To Delete")})
			},
			wantErr: false,
			validate: func(t *testing.T, q *Queries, id string) {
				ctx := context.Background()
				// Should not appear in list (is_deleted = TRUE)
				results, err := q.ListLocations(ctx, ListLocationsParams{
					Limit:  10,
					Offset: 0,
					Search: strPtr("To Delete"),
				})
				require.NoError(t, err)
				assert.Len(t, results, 0)
			},
		},
		{
			name: "nonexistent_id_no_error",
			setup: func(t *testing.T, q *Queries) string {
				return "nonexistent-id"
			},
			wantErr: false, // UPDATE on non-existent row doesn't error
			validate: func(t *testing.T, q *Queries, id string) {
				// Nothing to validate
			},
		},
		{
			name: "already_deleted",
			setup: func(t *testing.T, q *Queries) string {
				ctx := context.Background()
				id := CreateTestLocation(t, q, CreateTestLocationOptions{})
				q.SoftDeleteLocation(ctx, id)
				return id
			},
			wantErr: false,
			validate: func(t *testing.T, q *Queries, id string) {
				ctx := context.Background()
				// Still doesn't appear
				results, err := q.ListLocations(ctx, ListLocationsParams{Limit: 100, Offset: 0})
				require.NoError(t, err)
				for _, r := range results {
					assert.NotEqual(t, id, r.ID)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runTestWithTx(t, func(t *testing.T, q *Queries) {
				ctx := context.Background()
				id := tt.setup(t, q)

				err := q.SoftDeleteLocation(ctx, id)

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
// Test: IncrementLocationOccupied
// ============================================================

func TestIncrementLocationOccupied(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(t *testing.T, q *Queries) string // returns ID
		wantErr  bool
		validate func(t *testing.T, q *Queries, id string)
	}{
		{
			name: "success",
			setup: func(t *testing.T, q *Queries) string {
				return CreateTestLocation(t, q, CreateTestLocationOptions{
					Occupied: int32Ptr(5),
				})
			},
			wantErr: false,
			validate: func(t *testing.T, q *Queries, id string) {
				ctx := context.Background()
				results, err := q.ListLocations(ctx, ListLocationsParams{Limit: 100, Offset: 0})
				require.NoError(t, err)
				var found *ListLocationsRow
				for i := range results {
					if results[i].ID == id {
						found = &results[i]
						break
					}
				}
				require.NotNil(t, found)
				assert.Equal(t, int32(6), found.Occupied) // 5 + 1
			},
		},
		{
			name: "from_zero",
			setup: func(t *testing.T, q *Queries) string {
				return CreateTestLocation(t, q, CreateTestLocationOptions{
					Occupied: int32Ptr(0),
				})
			},
			wantErr: false,
			validate: func(t *testing.T, q *Queries, id string) {
				ctx := context.Background()
				results, err := q.ListLocations(ctx, ListLocationsParams{Limit: 100, Offset: 0})
				require.NoError(t, err)
				var found *ListLocationsRow
				for i := range results {
					if results[i].ID == id {
						found = &results[i]
						break
					}
				}
				require.NotNil(t, found)
				assert.Equal(t, int32(1), found.Occupied) // 0 + 1
			},
		},
		{
			name: "multiple_increments",
			setup: func(t *testing.T, q *Queries) string {
				ctx := context.Background()
				id := CreateTestLocation(t, q, CreateTestLocationOptions{
					Occupied: int32Ptr(0),
				})
				// Increment twice before test
				q.IncrementLocationOccupied(ctx, id)
				q.IncrementLocationOccupied(ctx, id)
				return id
			},
			wantErr: false,
			validate: func(t *testing.T, q *Queries, id string) {
				ctx := context.Background()
				results, err := q.ListLocations(ctx, ListLocationsParams{Limit: 100, Offset: 0})
				require.NoError(t, err)
				var found *ListLocationsRow
				for i := range results {
					if results[i].ID == id {
						found = &results[i]
						break
					}
				}
				require.NotNil(t, found)
				assert.Equal(t, int32(3), found.Occupied) // 0 + 2 (setup) + 1 (test)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runTestWithTx(t, func(t *testing.T, q *Queries) {
				ctx := context.Background()
				id := tt.setup(t, q)

				err := q.IncrementLocationOccupied(ctx, id)

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
// Test: DecrementLocationOccupied
// ============================================================

func TestDecrementLocationOccupied(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(t *testing.T, q *Queries) string // returns ID
		wantErr  bool
		validate func(t *testing.T, q *Queries, id string)
	}{
		{
			name: "success",
			setup: func(t *testing.T, q *Queries) string {
				return CreateTestLocation(t, q, CreateTestLocationOptions{
					Occupied: int32Ptr(5),
				})
			},
			wantErr: false,
			validate: func(t *testing.T, q *Queries, id string) {
				ctx := context.Background()
				results, err := q.ListLocations(ctx, ListLocationsParams{Limit: 100, Offset: 0})
				require.NoError(t, err)
				var found *ListLocationsRow
				for i := range results {
					if results[i].ID == id {
						found = &results[i]
						break
					}
				}
				require.NotNil(t, found)
				assert.Equal(t, int32(4), found.Occupied) // 5 - 1
			},
		},
		{
			name: "does_not_go_below_zero",
			setup: func(t *testing.T, q *Queries) string {
				return CreateTestLocation(t, q, CreateTestLocationOptions{
					Occupied: int32Ptr(0),
				})
			},
			wantErr: false,
			validate: func(t *testing.T, q *Queries, id string) {
				ctx := context.Background()
				results, err := q.ListLocations(ctx, ListLocationsParams{Limit: 100, Offset: 0})
				require.NoError(t, err)
				var found *ListLocationsRow
				for i := range results {
					if results[i].ID == id {
						found = &results[i]
						break
					}
				}
				require.NotNil(t, found)
				assert.Equal(t, int32(0), found.Occupied) // Should stay at 0
			},
		},
		{
			name: "from_one_to_zero",
			setup: func(t *testing.T, q *Queries) string {
				return CreateTestLocation(t, q, CreateTestLocationOptions{
					Occupied: int32Ptr(1),
				})
			},
			wantErr: false,
			validate: func(t *testing.T, q *Queries, id string) {
				ctx := context.Background()
				results, err := q.ListLocations(ctx, ListLocationsParams{Limit: 100, Offset: 0})
				require.NoError(t, err)
				var found *ListLocationsRow
				for i := range results {
					if results[i].ID == id {
						found = &results[i]
						break
					}
				}
				require.NotNil(t, found)
				assert.Equal(t, int32(0), found.Occupied) // 1 - 1 = 0
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runTestWithTx(t, func(t *testing.T, q *Queries) {
				ctx := context.Background()
				id := tt.setup(t, q)

				err := q.DecrementLocationOccupied(ctx, id)

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
// Test: GetLocationCapacityStats
// ============================================================

func TestGetLocationCapacityStats(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(t *testing.T, q *Queries)
		validate func(t *testing.T, stats GetLocationCapacityStatsRow)
	}{
		{
			name:  "empty_database",
			setup: func(t *testing.T, q *Queries) {},
			validate: func(t *testing.T, stats GetLocationCapacityStatsRow) {
				// When empty, totals should be 0
				assert.Equal(t, int32(0), stats.FreeCapacity)
			},
		},
		{
			name: "locations_no_clients",
			setup: func(t *testing.T, q *Queries) {
				CreateTestLocation(t, q, CreateTestLocationOptions{Capacity: int32Ptr(10)})
				CreateTestLocation(t, q, CreateTestLocationOptions{Capacity: int32Ptr(20)})
			},
			validate: func(t *testing.T, stats GetLocationCapacityStatsRow) {
				// Total capacity is 30, no clients in care
				assert.Equal(t, int32(30), stats.FreeCapacity)
			},
		},
		{
			name: "with_clients_in_care",
			setup: func(t *testing.T, q *Queries) {
				ctx := context.Background()

				// Create location with capacity 10
				locationID := CreateTestLocation(t, q, CreateTestLocationOptions{Capacity: int32Ptr(10)})

				// Create a client in care at this location
				// Pass locationID to employee to avoid creating an additional location
				userID := CreateTestUser(t, q, CreateTestUserOptions{})
				employeeID := CreateTestEmployee(t, q, CreateTestEmployeeOptions{
					UserID:     userID,
					LocationID: &locationID,
				})
				regFormID := CreateTestRegistrationForm(t, q, CreateTestRegistrationFormOptions{})
				intakeFormID := CreateTestIntakeForm(t, q, CreateTestIntakeFormOptions{
					RegistrationFormID: regFormID,
					LocationID:         locationID,
					CoordinatorID:      employeeID,
				})

				clientID := CreateTestClient(t, q, CreateTestClientOptions{
					RegistrationFormID: regFormID,
					IntakeFormID:       intakeFormID,
					AssignedLocationID: locationID,
					CoordinatorID:      employeeID,
				})

				// Move client to in_care status
				careStartDate := toPgDate(time.Now())
				_, err := q.UpdateClient(ctx, UpdateClientParams{
					ID:            clientID,
					Status:        NullClientStatusEnum{ClientStatusEnum: ClientStatusEnumInCare, Valid: true},
					CareStartDate: careStartDate,
				})
				require.NoError(t, err)
			},
			validate: func(t *testing.T, stats GetLocationCapacityStatsRow) {
				// 1 client in care, capacity 10, so free = 9
				assert.Equal(t, int32(9), stats.FreeCapacity)
			},
		},
		{
			name: "excludes_soft_deleted_locations",
			setup: func(t *testing.T, q *Queries) {
				ctx := context.Background()

				// Create and soft delete location
				deletedID := CreateTestLocation(t, q, CreateTestLocationOptions{Capacity: int32Ptr(100)})
				q.SoftDeleteLocation(ctx, deletedID)

				// Create active location
				CreateTestLocation(t, q, CreateTestLocationOptions{Capacity: int32Ptr(20)})
			},
			validate: func(t *testing.T, stats GetLocationCapacityStatsRow) {
				// Only the active location with capacity 20 should be counted
				assert.Equal(t, int32(20), stats.FreeCapacity)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runTestWithTx(t, func(t *testing.T, q *Queries) {
				ctx := context.Background()
				tt.setup(t, q)

				stats, err := q.GetLocationCapacityStats(ctx)

				require.NoError(t, err)
				tt.validate(t, stats)
			})
		})
	}
}
