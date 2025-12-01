
sqlc:
	@echo "Generating SQL queries..."
	sqlc generate

swagger:
	swag init --parseDependency -g server.go --dir ./api
.PHONY: sqlc swagger