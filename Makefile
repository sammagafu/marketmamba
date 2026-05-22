.PHONY: help install build run test docker-build docker-up docker-down docker-logs clean fmt lint

help:
	@echo "Market Mamba - Available Commands"
	@echo ""
	@echo "Development:"
	@echo "  make install        - Download Go dependencies"
	@echo "  make build         - Build binary"
	@echo "  make run           - Run locally"
	@echo "  make test          - Run tests"
	@echo "  make fmt           - Format code"
	@echo ""
	@echo "Docker:"
	@echo "  make docker-build  - Build Docker image"
	@echo "  make docker-up     - Start containers"
	@echo "  make docker-down   - Stop containers"
	@echo "  make docker-logs   - View container logs"
	@echo ""
	@echo "Database:"
	@echo "  make db-migrate    - Run database migrations"
	@echo "  make db-clean      - Drop all tables"
	@echo ""
	@echo "Utilities:"
	@echo "  make clean         - Remove build artifacts"
	@echo "  make lint          - Run linter (requires golangci-lint)"

install:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy

web-build:
	@echo "Building Vue frontend..."
	cd web && npm install && npm run build
	rm -rf internal/api/dist && cp -r web/dist internal/api/dist

build: web-build
	@echo "Building binary..."
	go build -o forex-bot cmd/server/main.go

run: build
	@echo "Running application..."
	source .env 2>/dev/null || true && ./forex-bot

test:
	@echo "Running tests..."
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

fmt:
	@echo "Formatting code..."
	go fmt ./...

lint:
	@echo "Linting code..."
	golangci-lint run ./...

docker-build:
	@echo "Building Docker image..."
	docker-compose build

docker-up:
	@echo "Starting containers..."
	docker-compose up -d
	@echo "Waiting for services to be ready..."
	sleep 3
	@docker-compose exec app go run cmd/server/main.go &

docker-down:
	@echo "Stopping containers..."
	docker-compose down

docker-logs:
	docker-compose logs -f

db-migrate:
	@echo "Running database migrations..."
	psql -U forexbot -d forexbot -f migrations/001_init_schema.sql

db-clean:
	@echo "WARNING: This will delete all tables!"
	@read -p "Continue? [y/N] " -n 1 -r; \
	echo; \
	if [[ $$REPLY =~ ^[Yy]$$ ]]; then \
		docker-compose exec postgres psql -U forexbot -d forexbot -c "DROP TABLE IF EXISTS command_logs CASCADE; DROP TABLE IF EXISTS bot_states CASCADE; DROP TABLE IF EXISTS daily_stats CASCADE; DROP TABLE IF EXISTS risk_settings CASCADE; DROP TABLE IF EXISTS accounts CASCADE; DROP TABLE IF EXISTS positions CASCADE; DROP TABLE IF EXISTS trades CASCADE;"; \
	fi

clean:
	@echo "Cleaning build artifacts..."
	rm -f forex-bot
	rm -f coverage.out coverage.html
	go clean

.DEFAULT_GOAL := help
