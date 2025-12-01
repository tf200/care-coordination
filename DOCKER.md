# Docker Setup Guide

This guide will help you run the Care Coordination application using Docker.

## Prerequisites

- Docker (version 20.10 or higher)
- Docker Compose (version 2.0 or higher)

## Quick Start

1. **Copy the environment file:**
   ```bash
   cp .env.example .env
   ```

2. **Update the `.env` file with your configuration:**
   - Change `ACCESS_TOKEN_SECRET` and `REFRESH_TOKEN_SECRET` to secure random strings
   - Adjust database credentials if needed
   - For production, set `ENVIRONMENT=production`

3. **Build and start the services:**
   ```bash
   make docker-up
   ```
   
   Or without make:
   ```bash
   docker-compose up -d
   ```

4. **Check if services are running:**
   ```bash
   make docker-ps
   ```

5. **View logs:**
   ```bash
   make docker-logs
   ```

## Available Services

The Docker Compose setup includes:

- **PostgreSQL** (port 5432): Database service
- **Redis** (port 6379): Cache and rate limiting service
- **App** (port 8080): Your Go application

## Makefile Commands

Run `make help` to see all available commands:

### Code Generation
- `make sqlc` - Generate SQL queries with sqlc
- `make swagger` - Generate Swagger documentation

### Docker Management
- `make docker-build` - Build Docker images
- `make docker-up` - Start all services in detached mode
- `make docker-down` - Stop all services
- `make docker-restart` - Restart all services
- `make docker-clean` - Stop and remove all containers, networks, and volumes
- `make docker-rebuild` - Rebuild and restart all services (no cache)

### Logs
- `make docker-logs` - View logs from all services
- `make docker-logs-app` - View logs from app service only
- `make docker-logs-postgres` - View logs from postgres service only
- `make docker-logs-redis` - View logs from redis service only

### Shell Access
- `make docker-shell-app` - Open shell in app container
- `make docker-shell-postgres` - Open psql shell in postgres container
- `make docker-shell-redis` - Open redis-cli shell in redis container

### Development
- `make docker-dev` - Start services in development mode (with logs in foreground)
- `make docker-ps` - Show running containers

## Environment Variables

### Database Configuration
- `DB_USER` - PostgreSQL username (default: postgres)
- `DB_PASSWORD` - PostgreSQL password (default: postgres)
- `DB_NAME` - Database name (default: care_coordination)
- `DB_HOST` - Database host (use 'postgres' for Docker, 'localhost' for local)
- `DB_PORT` - Database port (default: 5432)
- `DB_SOURCE` - Full PostgreSQL connection string

### Redis Configuration
- `REDIS_HOST` - Redis host (use 'redis' for Docker, 'localhost' for local)
- `REDIS_PORT` - Redis port (default: 6379)
- `REDIS_URL` - Full Redis connection string

### Server Configuration
- `SERVER_ADDRESS` - Server bind address (default: 0.0.0.0:8080)
- `SERVER_PORT` - Server port (default: 8080)
- `ENVIRONMENT` - Environment mode (development/production)

### JWT Configuration
- `ACCESS_TOKEN_SECRET` - Secret for access tokens (change in production!)
- `REFRESH_TOKEN_SECRET` - Secret for refresh tokens (change in production!)
- `ACCESS_TOKEN_TTL` - Access token time-to-live (default: 15m)
- `REFRESH_TOKEN_TTL` - Refresh token time-to-live (default: 168h)

### Rate Limiting Configuration
- `RATE_LIMIT_ENABLED` - Enable/disable rate limiting (default: true)
- `LOGIN_RATE_LIMIT_PER_IP` - Max login attempts per IP (default: 5)
- `LOGIN_RATE_LIMIT_WINDOW_IP` - IP rate limit window (default: 15m)
- `LOGIN_RATE_LIMIT_PER_EMAIL` - Max login attempts per email (default: 3)
- `LOGIN_RATE_LIMIT_WINDOW_EMAIL` - Email rate limit window (default: 15m)

## Troubleshooting

### Services won't start
1. Check if ports 5432, 6379, or 8080 are already in use:
   ```bash
   sudo lsof -i :5432
   sudo lsof -i :6379
   sudo lsof -i :8080
   ```

2. Check Docker logs:
   ```bash
   make docker-logs
   ```

### Database connection issues
1. Ensure PostgreSQL is healthy:
   ```bash
   docker-compose ps
   ```

2. Check PostgreSQL logs:
   ```bash
   make docker-logs-postgres
   ```

3. Try connecting manually:
   ```bash
   make docker-shell-postgres
   ```

### Redis connection issues
1. Check Redis is running:
   ```bash
   docker-compose ps redis
   ```

2. Test Redis connection:
   ```bash
   make docker-shell-redis
   # Then run: PING
   ```

### Application won't start
1. Check application logs:
   ```bash
   make docker-logs-app
   ```

2. Verify environment variables are set correctly in `.env`

3. Rebuild the application:
   ```bash
   make docker-rebuild
   ```

## Data Persistence

Data is persisted in Docker volumes:
- `postgres_data` - PostgreSQL database files
- `redis_data` - Redis data files

To completely remove all data:
```bash
make docker-clean
```

## Development Workflow

1. **Make code changes** in your local files

2. **Rebuild and restart** the application:
   ```bash
   make docker-rebuild
   ```

3. **View logs** to debug:
   ```bash
   make docker-logs-app
   ```

## Production Deployment

1. Update `.env` with production values:
   - Set strong, unique secrets for `ACCESS_TOKEN_SECRET` and `REFRESH_TOKEN_SECRET`
   - Set `ENVIRONMENT=production`
   - Use strong database credentials
   - Consider using external managed PostgreSQL and Redis services

2. Build and start:
   ```bash
   ENVIRONMENT=production docker-compose up -d
   ```

3. Monitor logs:
   ```bash
   make docker-logs
   ```

## Stopping Services

To stop all services:
```bash
make docker-down
```

To stop and remove all data:
```bash
make docker-clean
```
