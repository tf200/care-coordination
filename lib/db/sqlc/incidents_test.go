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
// Test: CreateIncident
// ============================================================

func TestCreateIncident(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(t *testing.T, q *Queries) CreateIncidentParams
		wantErr  bool
		checkErr func(t *testing.T, err error)
		validate func(t *testing.T, q *Queries, params CreateIncidentParams)
	}{
		{
			name: "success",
			setup: func(t *testing.T, q *Queries) CreateIncidentParams {
				clientID, deps := CreateTestClientWithDependencies(t, q)
				return CreateIncidentParams{
					ID:                  generateTestID(),
					ClientID:            clientID,
					IncidentDate:        toPgDate(time.Now()),
					IncidentTime:        toPgTime(time.Now()),
					IncidentType:        IncidentTypeEnumAggression,
					IncidentSeverity:    IncidentSeverityEnumMinor,
					LocationID:          deps.LocationID,
					CoordinatorID:       deps.EmployeeID,
					IncidentDescription: "Test description",
					ActionTaken:         "Test action",
					Status:              IncidentStatusEnumPending,
				}
			},
			wantErr: false,
			validate: func(t *testing.T, q *Queries, params CreateIncidentParams) {
				ctx := context.Background()
				incident, err := q.GetIncident(ctx, params.ID)
				require.NoError(t, err)
				assert.Equal(t, params.ID, incident.ID)
				assert.Equal(t, params.ClientID, incident.ClientID)
				assert.Equal(t, params.IncidentDescription, incident.IncidentDescription)
			},
		},
		{
			name: "duplicate_id",
			setup: func(t *testing.T, q *Queries) CreateIncidentParams {
				clientID, deps := CreateTestClientWithDependencies(t, q)
				id := generateTestID()
				CreateTestIncident(t, q, CreateTestIncidentOptions{
					ID:            &id,
					ClientID:      clientID,
					LocationID:    deps.LocationID,
					CoordinatorID: deps.EmployeeID,
				})
				return CreateIncidentParams{
					ID:               id, // Duplicate
					ClientID:         clientID,
					IncidentDate:     toPgDate(time.Now()),
					IncidentTime:     toPgTime(time.Now()),
					IncidentType:     IncidentTypeEnumOther,
					IncidentSeverity: IncidentSeverityEnumMinor,
					LocationID:       deps.LocationID,
					CoordinatorID:    deps.EmployeeID,
					Status:           IncidentStatusEnumPending,
				}
			},
			wantErr: true,
			checkErr: func(t *testing.T, err error) {
				assert.True(t, IsUniqueViolation(err))
			},
		},
		{
			name: "invalid_client_fk",
			setup: func(t *testing.T, q *Queries) CreateIncidentParams {
				deps := CreateFullClientDependencyChain(t, q)
				return CreateIncidentParams{
					ID:               generateTestID(),
					ClientID:         "non-existent-client",
					IncidentDate:     toPgDate(time.Now()),
					IncidentTime:     toPgTime(time.Now()),
					IncidentType:     IncidentTypeEnumOther,
					IncidentSeverity: IncidentSeverityEnumMinor,
					LocationID:       deps.LocationID,
					CoordinatorID:    deps.EmployeeID,
					Status:           IncidentStatusEnumPending,
				}
			},
			wantErr: true,
			checkErr: func(t *testing.T, err error) {
				assert.True(t, IsForeignKeyViolation(err))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runTestWithTx(t, func(t *testing.T, q *Queries) {
				ctx := context.Background()
				params := tt.setup(t, q)

				err := q.CreateIncident(ctx, params)

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
// Test: GetIncident
// ============================================================

func TestGetIncident(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(t *testing.T, q *Queries) string // returns ID
		wantErr  bool
		checkErr func(t *testing.T, err error)
		validate func(t *testing.T, incident GetIncidentRow)
	}{
		{
			name: "found",
			setup: func(t *testing.T, q *Queries) string {
				clientID, deps := CreateTestClientWithDependencies(t, q)
				return CreateTestIncident(t, q, CreateTestIncidentOptions{
					ClientID:      clientID,
					LocationID:    deps.LocationID,
					CoordinatorID: deps.EmployeeID,
				})
			},
			wantErr: false,
			validate: func(t *testing.T, incident GetIncidentRow) {
				assert.NotEmpty(t, incident.ID)
				assert.NotEmpty(t, incident.ClientFirstName)
				assert.NotEmpty(t, incident.LocationName)
				assert.NotEmpty(t, incident.CoordinatorFirstName)
			},
		},
		{
			name: "not_found",
			setup: func(t *testing.T, q *Queries) string {
				return "non-existent-id"
			},
			wantErr: true,
			checkErr: func(t *testing.T, err error) {
				assert.True(t, errors.Is(err, pgx.ErrNoRows))
			},
		},
		{
			name: "soft_deleted_not_found",
			setup: func(t *testing.T, q *Queries) string {
				clientID, deps := CreateTestClientWithDependencies(t, q)
				id := CreateTestIncident(t, q, CreateTestIncidentOptions{
					ClientID:      clientID,
					LocationID:    deps.LocationID,
					CoordinatorID: deps.EmployeeID,
				})
				err := q.SoftDeleteIncident(context.Background(), id)
				if err != nil {
					t.Fatalf("failed to soft delete: %v", err)
				}
				return id
			},
			wantErr: true,
			checkErr: func(t *testing.T, err error) {
				assert.True(t, errors.Is(err, pgx.ErrNoRows))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runTestWithTx(t, func(t *testing.T, q *Queries) {
				ctx := context.Background()
				id := tt.setup(t, q)

				incident, err := q.GetIncident(ctx, id)

				if tt.wantErr {
					require.Error(t, err)
					if tt.checkErr != nil {
						tt.checkErr(t, err)
					}
					return
				}

				require.NoError(t, err)
				if tt.validate != nil {
					tt.validate(t, incident)
				}
			})
		})
	}
}

// ============================================================
// Test: GetIncidentStats
// ============================================================

func TestGetIncidentStats(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(t *testing.T, q *Queries)
		validate func(t *testing.T, stats GetIncidentStatsRow)
	}{
		{
			name:  "empty",
			setup: func(t *testing.T, q *Queries) {},
			validate: func(t *testing.T, stats GetIncidentStatsRow) {
				assert.Equal(t, int64(0), stats.TotalCount)
			},
		},
		{
			name: "mixed_data",
			setup: func(t *testing.T, q *Queries) {
				clientID, deps := CreateTestClientWithDependencies(t, q)

				// 1 Minor, Pending, Aggression
				CreateTestIncident(t, q, CreateTestIncidentOptions{
					ClientID:         clientID,
					LocationID:       deps.LocationID,
					CoordinatorID:    deps.EmployeeID,
					IncidentSeverity: func() *IncidentSeverityEnum { e := IncidentSeverityEnumMinor; return &e }(),
					Status:           func() *IncidentStatusEnum { e := IncidentStatusEnumPending; return &e }(),
					IncidentType:     func() *IncidentTypeEnum { e := IncidentTypeEnumAggression; return &e }(),
				})

				// 1 Moderate, Completed, Medical Emergency
				CreateTestIncident(t, q, CreateTestIncidentOptions{
					ClientID:         clientID,
					LocationID:       deps.LocationID,
					CoordinatorID:    deps.EmployeeID,
					IncidentSeverity: func() *IncidentSeverityEnum { e := IncidentSeverityEnumModerate; return &e }(),
					Status:           func() *IncidentStatusEnum { e := IncidentStatusEnumCompleted; return &e }(),
					IncidentType:     func() *IncidentTypeEnum { e := IncidentTypeEnumMedicalEmergency; return &e }(),
				})

				// 1 Severe, Under Investigation, Safety Concern
				CreateTestIncident(t, q, CreateTestIncidentOptions{
					ClientID:         clientID,
					LocationID:       deps.LocationID,
					CoordinatorID:    deps.EmployeeID,
					IncidentSeverity: func() *IncidentSeverityEnum { e := IncidentSeverityEnumSevere; return &e }(),
					Status:           func() *IncidentStatusEnum { e := IncidentStatusEnumUnderInvestigation; return &e }(),
					IncidentType:     func() *IncidentTypeEnum { e := IncidentTypeEnumSafetyConcern; return &e }(),
				})

				// 1 deleted (should not be counted)
				id := CreateTestIncident(t, q, CreateTestIncidentOptions{
					ClientID:      clientID,
					LocationID:    deps.LocationID,
					CoordinatorID: deps.EmployeeID,
				})
				q.SoftDeleteIncident(context.Background(), id)

				// 1 Unwanted Behavior
				CreateTestIncident(t, q, CreateTestIncidentOptions{
					ClientID:      clientID,
					LocationID:    deps.LocationID,
					CoordinatorID: deps.EmployeeID,
					IncidentType:  func() *IncidentTypeEnum { e := IncidentTypeEnumUnwantedBehavior; return &e }(),
				})

				// 1 Other Type
				CreateTestIncident(t, q, CreateTestIncidentOptions{
					ClientID:      clientID,
					LocationID:    deps.LocationID,
					CoordinatorID: deps.EmployeeID,
					IncidentType:  func() *IncidentTypeEnum { e := IncidentTypeEnumOther; return &e }(),
				})
			},
			validate: func(t *testing.T, stats GetIncidentStatsRow) {
				assert.Equal(t, int64(5), stats.TotalCount)
				assert.Equal(t, int64(3), stats.MinorCount) // Defaults in factory
				assert.Equal(t, int64(1), stats.ModerateCount)
				assert.Equal(t, int64(1), stats.SevereCount)
				assert.Equal(t, int64(3), stats.PendingCount) // Defaults in factory
				assert.Equal(t, int64(1), stats.CompletedCount)
				assert.Equal(t, int64(1), stats.UnderInvestigationCount)
				assert.Equal(t, int64(1), stats.AggressionCount)
				assert.Equal(t, int64(1), stats.MedicalEmergencyCount)
				assert.Equal(t, int64(1), stats.SafetyConcernCount)
				assert.Equal(t, int64(1), stats.UnwantedBehaviorCount)
				assert.Equal(t, int64(1), stats.OtherTypeCount)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runTestWithTx(t, func(t *testing.T, q *Queries) {
				tt.setup(t, q)
				stats, err := q.GetIncidentStats(context.Background())
				require.NoError(t, err)
				tt.validate(t, stats)
			})
		})
	}
}

// ============================================================
// Test: ListIncidents
// ============================================================

func TestListIncidents(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(t *testing.T, q *Queries)
		params   ListIncidentsParams
		validate func(t *testing.T, results []ListIncidentsRow)
	}{
		{
			name:   "empty",
			setup:  func(t *testing.T, q *Queries) {},
			params: ListIncidentsParams{Limit: 10, Offset: 0},
			validate: func(t *testing.T, results []ListIncidentsRow) {
				assert.Len(t, results, 0)
			},
		},
		{
			name: "with_pagination",
			setup: func(t *testing.T, q *Queries) {
				clientID, deps := CreateTestClientWithDependencies(t, q)
				for i := 0; i < 5; i++ {
					CreateTestIncident(t, q, CreateTestIncidentOptions{
						ClientID:      clientID,
						LocationID:    deps.LocationID,
						CoordinatorID: deps.EmployeeID,
					})
				}
			},
			params: ListIncidentsParams{Limit: 2, Offset: 0},
			validate: func(t *testing.T, results []ListIncidentsRow) {
				assert.Len(t, results, 2)
				assert.Equal(t, int64(5), results[0].TotalCount)
			},
		},
		{
			name: "with_search",
			setup: func(t *testing.T, q *Queries) {
				// Client 1: Alpha User
				c1, d1 := CreateTestClientWithDependencies(t, q)
				q.UpdateClient(context.Background(), UpdateClientParams{
					ID:        c1,
					FirstName: strPtr("Alpha"),
					LastName:  strPtr("User"),
				})
				CreateTestIncident(t, q, CreateTestIncidentOptions{
					ClientID:      c1,
					LocationID:    d1.LocationID,
					CoordinatorID: d1.EmployeeID,
				})

				// Client 2: Beta User
				c2, d2 := CreateTestClientWithDependencies(t, q)
				q.UpdateClient(context.Background(), UpdateClientParams{
					ID:        c2,
					FirstName: strPtr("Beta"),
					LastName:  strPtr("User"),
				})
				CreateTestIncident(t, q, CreateTestIncidentOptions{
					ClientID:      c2,
					LocationID:    d2.LocationID,
					CoordinatorID: d2.EmployeeID,
				})
			},
			params: ListIncidentsParams{Limit: 10, Offset: 0, Search: strPtr("Alpha")},
			validate: func(t *testing.T, results []ListIncidentsRow) {
				assert.Len(t, results, 1)
				assert.Equal(t, "Alpha", results[0].ClientFirstName)
			},
		},
		{
			name: "excludes_deleted",
			setup: func(t *testing.T, q *Queries) {
				clientID, deps := CreateTestClientWithDependencies(t, q)
				id := CreateTestIncident(t, q, CreateTestIncidentOptions{
					ClientID:      clientID,
					LocationID:    deps.LocationID,
					CoordinatorID: deps.EmployeeID,
				})
				q.SoftDeleteIncident(context.Background(), id)
			},
			params: ListIncidentsParams{Limit: 10, Offset: 0},
			validate: func(t *testing.T, results []ListIncidentsRow) {
				assert.Len(t, results, 0)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runTestWithTx(t, func(t *testing.T, q *Queries) {
				tt.setup(t, q)
				results, err := q.ListIncidents(context.Background(), tt.params)
				require.NoError(t, err)
				tt.validate(t, results)
			})
		})
	}
}

// ============================================================
// Test: SoftDeleteIncident
// ============================================================

func TestSoftDeleteIncident(t *testing.T) {
	runTestWithTx(t, func(t *testing.T, q *Queries) {
		clientID, deps := CreateTestClientWithDependencies(t, q)
		id := CreateTestIncident(t, q, CreateTestIncidentOptions{
			ClientID:      clientID,
			LocationID:    deps.LocationID,
			CoordinatorID: deps.EmployeeID,
		})

		err := q.SoftDeleteIncident(context.Background(), id)
		require.NoError(t, err)

		// Verify it's not returned by GetIncident
		_, err = q.GetIncident(context.Background(), id)
		assert.True(t, errors.Is(err, pgx.ErrNoRows))
	})
}

// ============================================================
// Test: UpdateIncident
// ============================================================

func TestUpdateIncident(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(t *testing.T, q *Queries) UpdateIncidentParams
		wantErr  bool
		validate func(t *testing.T, q *Queries, id string)
	}{
		{
			name: "success_multiple_fields",
			setup: func(t *testing.T, q *Queries) UpdateIncidentParams {
				clientID, deps := CreateTestClientWithDependencies(t, q)
				id := CreateTestIncident(t, q, CreateTestIncidentOptions{
					ClientID:            clientID,
					LocationID:          deps.LocationID,
					CoordinatorID:       deps.EmployeeID,
					IncidentDescription: strPtr("Original description"),
					ActionTaken:         strPtr("Original action"),
				})
				return UpdateIncidentParams{
					ID:                  id,
					IncidentDescription: strPtr("Updated description"),
					ActionTaken:         strPtr("Updated action"),
					Status:              NullIncidentStatusEnum{IncidentStatusEnum: IncidentStatusEnumCompleted, Valid: true},
					IncidentSeverity:    NullIncidentSeverityEnum{IncidentSeverityEnum: IncidentSeverityEnumSevere, Valid: true},
				}
			},
			wantErr: false,
			validate: func(t *testing.T, q *Queries, id string) {
				incident, err := q.GetIncident(context.Background(), id)
				require.NoError(t, err)
				assert.Equal(t, "Updated description", incident.IncidentDescription)
				assert.Equal(t, "Updated action", incident.ActionTaken)
				assert.Equal(t, IncidentStatusEnumCompleted, incident.Status)
				assert.Equal(t, IncidentSeverityEnumSevere, incident.IncidentSeverity)
			},
		},
		{
			name: "not_found",
			setup: func(t *testing.T, q *Queries) UpdateIncidentParams {
				return UpdateIncidentParams{
					ID:                  "non-existent-id",
					IncidentDescription: strPtr("No update"),
				}
			},
			wantErr: false, // Update on non-existent row in sqlc usually doesn't error unless it's :execrows and checked
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runTestWithTx(t, func(t *testing.T, q *Queries) {
				params := tt.setup(t, q)
				err := q.UpdateIncident(context.Background(), params)

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
