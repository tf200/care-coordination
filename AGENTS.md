# Agent Guide: Care Coordination Backend

This guide provides instructions for development, testing, and code style conventions for the Care Coordination backend.

## 🛠 Build and Development Commands

### Core Commands
- **Run Application**: `go run cmd/app/main.go`
- **Build**: `go build -o server cmd/app/main.go`
- **SQL Generation**: `make sqlc` (Uses `sqlc.yaml` to generate type-safe DB code)
- **Swagger Generation**: `make swagger` (Generates API docs from annotations)
- **Add Feature**: `make add-feature NAME=<feature_name>` (Scaffolds a new feature directory)
- **Dependency Update**: `go mod tidy`

### Database & Migrations
- **Up**: `make migrate-up`
- **Down**: `make migrate-down`
- **Version**: `make migrate-version`
- **Force Version**: `make migrate-force VERSION=<v>`
- **Seed Data**: `make seed` or `make admin`

### Docker Support
- **Up**: `make docker-up` or `docker compose up -d`
- **Down**: `make docker-down`
- **Rebuild**: `make docker-rebuild`

## 🧪 Testing Guidelines

### Unit Testing
Use standard Go `testing` package with table-driven tests.
- **Run All Tests**: `go test ./...`
- **Run Single Test**: `go test -v -run <TestName> <PackagePath>`
- **Generate Mocks**: `go generate ./...` (Uses `mockgen`)

**Patterns:**
- Use `github.com/stretchr/testify/assert` and `require`.
- Use `go.uber.org/mock/gomock` for mocking interfaces.
- Mock files are located in `internal/mocks/` or package-local `mocks/` directories.

### Integration Testing (Database)
Follow patterns in `lib/db/sqlc/` as specified in `PROMPTS.md`.
- **Isolation**: Use `runTestWithTx` helper for test isolation.
- **Factories**: Use factory functions in `testutil.go` for creating test data.
- **Constraints**: Test for `pgx.ErrNoRows`, `IsUniqueViolation`, and `IsForeignKeyViolation`.
- **Command**: `go test -v ./lib/db/sqlc/... -run <TestName> -count=1`

## 🗄 Database Structure & RLS

### Schema Design
The database uses a feature-based relational schema managed by `sqlc` and `golang-migrate`.
- **Core Tables**: `users`, `employees`, `clients`, `locations`.
- **Access Control**: `roles`, `permissions`, `user_roles`, `role_permissions`.
- **Features**: `appointments`, `notifications`, `incidents`, `intake_forms`, `client_evaluations`.
- **Primary Keys**: Nanoids are used for IDs (generated via `lib/nanoid`).

### Row Level Security (RLS)
Postgres RLS is enabled on sensitive tables (e.g., `clients`, `appointments`, `client_goals`).
- **Mechanism**: Policies rely on the `app.current_user_id` session variable.
- **Policy Patterns**:
    - `admin`: Full access via `admin_all` policies.
    - `coordinator`: Limited to assigned records (e.g., `coordinator_own_clients` checks `coordinator_id`).
    - `user`: Limited to own records (e.g., `user_own_notifications`).
- **Enforcement in Go**:
    The setting MUST be applied per-transaction. Use `store.ExecTx` which automatically sets the context:
    ```go
    // Inside lib/db/sqlc/store.go
    tx.Exec(ctx, "SELECT set_config('app.current_user_id', $1, true)", userID)
    ```
- **Constraint**: Direct use of `store.Queries` outside of `ExecTx` will NOT enforce RLS if `userID` is missing from the context. Always ensure `context` contains the `user_id`.

## 📏 Code Style & Conventions

### Architecture
Feature-based architecture in `features/`. Each feature typically contains:
- `interface.go`: Defines the Service interface and DTOs.
- `service.go`: Business logic implementation (`xxxService` private struct).
- `handler.go`: Gin HTTP handlers (`XxxHandler` struct).
- `dto.go`: Data transfer objects.

### Naming Conventions
- **Handlers**: `XxxHandler`, factory `NewXxxHandler(...)`.
- **Services**: Interface `XxxService`, implementation `xxxService`, factory `NewXxxService(...)`.
- **Routes**: Method `SetupXxxRoutes(router *gin.Engine, ...)`.
- **Files**: Use `snake_case` for filenames. Standard files: `handler.go`, `service.go`.

### Imports
Group imports into:
1. Standard library
2. Project dependencies (Alias generated DB code: `db "care-cordination/lib/db/sqlc"`)
3. Third-party packages

### Error Handling
- Define package-level error variables (e.g., `ErrNotFound = errors.New("...")`).
- In handlers, use `switch` on errors to return appropriate HTTP status codes via `lib/resp`.
- Use `lib/resp` helpers: `resp.Success(data, msg)`, `resp.Error(err)`, `resp.PagResp(...)`.

### Logging
- Use `zap` logger via the `logger.Logger` interface.
- Pattern: `s.logger.Error(ctx, "Operation", "Message", zap.Error(err), zap.String("key", value))`.

### API Documentation
- Use `swag` annotations on handler methods.
- Required fields: `@Summary`, `@Description`, `@Tags`, `@Accept`, `@Produce`, `@Param`, `@Success`, `@Failure`, `@Router`.
