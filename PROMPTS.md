Write integration tests for @[lib/db/sqlc/locations.go]

Follow the testing patterns in @[lib/db/sqlc/README.md]

Use the existing infrastructure:
- @[lib/db/sqlc/main_test.go] - runTestWithTx helper
- @[lib/db/sqlc/testutil.go] - factory functions and error helpers
- @[lib/db/sqlc/users_test.go] - example test file with table-driven tests

Requirements:
1. Create `locations.go` in the same directory
2. Use **idiomatic Go table-driven tests** with test struct containing: name, setup, wantErr, checkErr, validate
3. Test every query function in the source file
4. For each query, include test cases for:
   - Success case (happy path)
   - Not found case (for Get queries) using pgx.ErrNoRows
   - Unique constraint violations using IsUniqueViolation
   - FK violations using IsForeignKeyViolation (if applicable)
5. Use runTestWithTx for test isolation
6. Use factory functions for creating test data
7. Run tests to verify they pass: `go test -v ./lib/db/sqlc/... -run TestRegistration -count=1`