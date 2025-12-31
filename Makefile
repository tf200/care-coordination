
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

admin:
	@echo "Seeding admin user..."
	go run cmd/admin/main.go

docker-rebuild:
	@echo "Rebuilding docker image..."
	docker compose rm -s -f app
	docker compose up --build -d app

add-feature:
	@echo "Adding new feature (use: make add-feature NAME=feature_name)..."
	mkdir -p features/$(NAME)
	touch features/$(NAME)/interface.go
	touch features/$(NAME)/service.go
	touch features/$(NAME)/dto.go
	touch features/$(NAME)/handler.go
seed:
	@echo "Seeding admin user..."
	go run cmd/seed/main.go
deploy:
	@echo "Deploying..."
	./scripts/deploy.sh

.PHONY: sqlc swagger migrate-up migrate-down migrate-up1 migrate-down1 migrate-version migrate-force admin dokcer-rebuild add-feature seed deploy