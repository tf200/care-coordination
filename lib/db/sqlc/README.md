# SQLC Integration Tests

## Quick Start

```bash
go test -v ./lib/db/sqlc/... -count=1
```

---

## AI Agent Instructions

### Your Task
Write integration tests for each `*.sql.go` file using **idiomatic Go table-driven tests**.

### Files Needing Tests

| Query File | Test File (create this) |
|------------|------------------------|
| `clients.sql.go` | `clients_test.go` |
| `employees.sql.go` | `employees_test.go` |
| `locations.sql.go` | `locations_test.go` |
| `referring_orgs.sql.go` | `referring_orgs_test.go` |
| `registration_forms.sql.go` | `registration_forms_test.go` |
| `intake_forms.sql.go` | `intake_forms_test.go` |
| `incidents.sql.go` | `incidents_test.go` |
| `location_transfers.sql.go` | `location_transfers_test.go` |
| `attachments.sql.go` | `attachments_test.go` |

---

## Table-Driven Test Pattern (REQUIRED)

All tests MUST use the idiomatic Go table-driven test pattern:

```go
package db

import (
    "context"
    "errors"
    "testing"

    "github.com/jackc/pgx/v5"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestCreateXxx(t *testing.T) {
    tests := []struct {
        name     string
        setup    func(t *testing.T, q *Queries) CreateXxxParams
        wantErr  bool
        checkErr func(t *testing.T, err error)
    }{
        {
            name: "success",
            setup: func(t *testing.T, q *Queries) CreateXxxParams {
                // Create any FK dependencies here
                return CreateXxxParams{
                    ID:   generateTestID(),
                    Name: "Test Name",
                }
            },
            wantErr: false,
        },
        {
            name: "duplicate_unique_field",
            setup: func(t *testing.T, q *Queries) CreateXxxParams {
                // Create first record
                CreateTestXxx(t, q, CreateTestXxxOptions{Name: strPtr("duplicate")})
                // Return params that will fail
                return CreateXxxParams{
                    ID:   generateTestID(),
                    Name: "duplicate",
                }
            },
            wantErr: true,
            checkErr: func(t *testing.T, err error) {
                assert.True(t, IsUniqueViolation(err), "expected unique violation, got: %v", err)
            },
        },
        {
            name: "invalid_foreign_key",
            setup: func(t *testing.T, q *Queries) CreateXxxParams {
                return CreateXxxParams{
                    ID:        generateTestID(),
                    ParentID:  "non-existent-id",
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

                result, err := q.CreateXxx(ctx, params)

                if tt.wantErr {
                    require.Error(t, err)
                    if tt.checkErr != nil {
                        tt.checkErr(t, err)
                    }
                    return
                }

                require.NoError(t, err)
                assert.Equal(t, params.ID, result.ID)
            })
        })
    }
}
```

---

## Pattern for Different Query Types

### Get Queries (returns single row)

```go
func TestGetXxxByID(t *testing.T) {
    tests := []struct {
        name     string
        setup    func(t *testing.T, q *Queries) string // returns ID to query
        wantErr  bool
        checkErr func(t *testing.T, err error)
        validate func(t *testing.T, result Xxx)
    }{
        {
            name: "found",
            setup: func(t *testing.T, q *Queries) string {
                id := CreateTestXxx(t, q, CreateTestXxxOptions{})
                return id
            },
            wantErr: false,
            validate: func(t *testing.T, result Xxx) {
                assert.NotEmpty(t, result.ID)
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
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            runTestWithTx(t, func(t *testing.T, q *Queries) {
                ctx := context.Background()
                id := tt.setup(t, q)

                result, err := q.GetXxxByID(ctx, id)

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
```

### List Queries (returns multiple rows)

```go
func TestListXxx(t *testing.T) {
    tests := []struct {
        name     string
        setup    func(t *testing.T, q *Queries) // create test data
        params   ListXxxParams
        validate func(t *testing.T, results []ListXxxRow)
    }{
        {
            name:   "empty",
            setup:  func(t *testing.T, q *Queries) {},
            params: ListXxxParams{Limit: 10, Offset: 0},
            validate: func(t *testing.T, results []ListXxxRow) {
                assert.Len(t, results, 0)
            },
        },
        {
            name: "with_pagination",
            setup: func(t *testing.T, q *Queries) {
                for i := 0; i < 5; i++ {
                    CreateTestXxx(t, q, CreateTestXxxOptions{})
                }
            },
            params: ListXxxParams{Limit: 2, Offset: 0},
            validate: func(t *testing.T, results []ListXxxRow) {
                assert.Len(t, results, 2)
                assert.Equal(t, int64(5), results[0].TotalCount)
            },
        },
        {
            name: "with_search",
            setup: func(t *testing.T, q *Queries) {
                CreateTestXxx(t, q, CreateTestXxxOptions{Name: strPtr("Alpha")})
                CreateTestXxx(t, q, CreateTestXxxOptions{Name: strPtr("Beta")})
                CreateTestXxx(t, q, CreateTestXxxOptions{Name: strPtr("Gamma")})
            },
            params: ListXxxParams{Limit: 10, Offset: 0, Search: strPtr("Beta")},
            validate: func(t *testing.T, results []ListXxxRow) {
                assert.Len(t, results, 1)
                assert.Equal(t, "Beta", results[0].Name)
            },
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            runTestWithTx(t, func(t *testing.T, q *Queries) {
                ctx := context.Background()
                tt.setup(t, q)

                results, err := q.ListXxx(ctx, tt.params)

                require.NoError(t, err)
                tt.validate(t, results)
            })
        })
    }
}
```

### Update Queries

```go
func TestUpdateXxx(t *testing.T) {
    tests := []struct {
        name     string
        setup    func(t *testing.T, q *Queries) UpdateXxxParams
        wantErr  bool
        checkErr func(t *testing.T, err error)
    }{
        {
            name: "success",
            setup: func(t *testing.T, q *Queries) UpdateXxxParams {
                id := CreateTestXxx(t, q, CreateTestXxxOptions{})
                return UpdateXxxParams{
                    ID:   id,
                    Name: "Updated Name",
                }
            },
            wantErr: false,
        },
        {
            name: "not_found",
            setup: func(t *testing.T, q *Queries) UpdateXxxParams {
                return UpdateXxxParams{
                    ID:   "non-existent-id",
                    Name: "Won't Update",
                }
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
                params := tt.setup(t, q)

                _, err := q.UpdateXxx(ctx, params)

                if tt.wantErr {
                    require.Error(t, err)
                    if tt.checkErr != nil {
                        tt.checkErr(t, err)
                    }
                    return
                }

                require.NoError(t, err)
            })
        })
    }
}
```

### Delete Queries

```go
func TestDeleteXxx(t *testing.T) {
    tests := []struct {
        name  string
        setup func(t *testing.T, q *Queries) string // returns ID to delete
    }{
        {
            name: "existing",
            setup: func(t *testing.T, q *Queries) string {
                return CreateTestXxx(t, q, CreateTestXxxOptions{})
            },
        },
        {
            name: "non_existent",
            setup: func(t *testing.T, q *Queries) string {
                return "non-existent-id"
            },
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            runTestWithTx(t, func(t *testing.T, q *Queries) {
                ctx := context.Background()
                id := tt.setup(t, q)

                err := q.DeleteXxx(ctx, id)

                // DELETE is idempotent - should never error
                require.NoError(t, err)
            })
        })
    }
}
```

---

## Test Struct Fields Reference

| Field | Type | Purpose |
|-------|------|---------|
| `name` | `string` | Test case name (use snake_case) |
| `setup` | `func(t, q) → params/id` | Create test data, return input for query |
| `wantErr` | `bool` | Whether an error is expected |
| `checkErr` | `func(t, err)` | Assert specific error type |
| `validate` | `func(t, result)` | Assert result fields |
| `params` | `XxxParams` | Direct params for List queries |

---

## Available Factories (in testutil.go)

```go
CreateTestUser(t, q, CreateTestUserOptions{})
CreateTestLocation(t, q, CreateTestLocationOptions{})
CreateTestReferringOrg(t, q, CreateTestReferringOrgOptions{})
CreateTestEmployee(t, q, CreateTestEmployeeOptions{UserID: userID})
CreateTestRegistrationForm(t, q, CreateTestRegistrationFormOptions{})
CreateTestIntakeForm(t, q, CreateTestIntakeFormOptions{...})
CreateTestClient(t, q, CreateTestClientOptions{...})
CreateTestSession(t, q, CreateTestSessionOptions{UserID: userID})

// Quick helpers
deps := CreateFullClientDependencyChain(t, q)
clientID, deps := CreateTestClientWithDependencies(t, q)
```

---

## Error Assertions

```go
IsUniqueViolation(err)      // Duplicate unique field
IsForeignKeyViolation(err)  // Invalid FK reference  
IsCheckViolation(err)       // CHECK constraint failed
IsNotNullViolation(err)     // Required field missing
errors.Is(err, pgx.ErrNoRows) // Record not found
```

---

## FK Dependency Order

```
User → Employee
Location (standalone)
ReferringOrg (standalone)
RegistrationForm → IntakeForm → Client → Incident
```

---

## Client Status Constraints

| Status | Must Have | Must NOT Have |
|--------|-----------|---------------|
| `waiting_list` | - | care_start_date, discharge_date, discharge_status |
| `in_care` | care_start_date | discharge_date, discharge_status |
| `discharged` | care_start_date, discharge_date, discharge_status, reason_for_discharge | - |

---

## Example: See users_test.go

Reference [users_test.go](./users_test.go) for complete working examples using table-driven tests.
