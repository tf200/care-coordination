
sqlc:
	@echo "Generating SQL queries..."
	sqlc generate

swagger:
	swag init --parseDependency -g server.go --dir ./api,./features

# Migration commands
migrate-up:
	@echo "Running all migrations..."
	go run cmd/migrate/main.go up

migrate-down:
	@echo "Rolling back all migrations..."
	go run cmd/migrate/main.go down

migrate-up1:
	@echo "Running next migration..."
	go run cmd/migrate/main.go up1

migrate-down1:
	@echo "Rolling back last migration..."
	go run cmd/migrate/main.go down1

migrate-version:
	@echo "Checking migration version..."
	go run cmd/migrate/main.go version

migrate-force:
	@echo "Force migration version (use: make migrate-force VERSION=1)..."
	go run cmd/migrate/main.go force $(VERSION)

.PHONY: sqlc swagger migrate-up migrate-down migrate-up1 migrate-down1 migrate-version migrate-force