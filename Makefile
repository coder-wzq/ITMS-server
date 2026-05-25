# ITMS Makefile

APP_NAME := itms-server
GO := go
GOFLAGS := -v

.PHONY: all build run test lint clean migrate-up migrate-down proto

# Build all services
build:
	$(GO) build $(GOFLAGS) -o bin/$(APP_NAME) ./cmd/gateway

# Run gateway
run:
	$(GO) run ./cmd/gateway -config deploy/configs/config.dev.yaml

# Run all tests
test:
	$(GO) test -race -cover ./...

# Run tests with coverage report
test-cover:
	$(GO) test -coverprofile=coverage.out ./... && $(GO) tool cover -html=coverage.out

# Lint code
lint:
	golangci-lint run ./...

# Format code
fmt:
	$(GO) fmt ./...

# Vet code
vet:
	$(GO) vet ./...

# Download dependencies
deps:
	$(GO) mod download
	$(GO) mod tidy

# Clean build artifacts
clean:
	rm -rf bin/ logs/

# Database migrations (golang-migrate)
migrate-up:
	go run ./cmd/migrate -action up

migrate-down:
	go run ./cmd/migrate -action down

migrate-version:
	go run ./cmd/migrate -action version

migrate-force:
	go run ./cmd/migrate -action force -version $(V)

# Generate proto files
proto:
	@for svc in services/*/; do \
		if [ -f "$$svc"*.proto ]; then \
			protoc --go_out=. --go-grpc_out=. $$svc*.proto; \
		fi \
	done

# Docker operations
docker-build:
	docker build -t itms-gateway:latest -f deploy/docker/Dockerfile .

docker-up:
	docker compose -f deploy/docker/docker-compose.yml up -d

docker-down:
	docker compose -f deploy/docker/docker-compose.yml down

docker-logs:
	docker compose -f deploy/docker/docker-compose.yml logs -f

# Help
help:
	@echo "ITMS Build System"
	@echo ""
	@echo "Usage:"
	@echo "  make build       Build all services"
	@echo "  make run         Run gateway"
	@echo "  make test        Run tests"
	@echo "  make lint        Lint code"
	@echo "  make fmt         Format code"
	@echo "  make vet         Vet code"
	@echo "  make clean       Clean artifacts"
	@echo "  make migrate-up  Apply database migrations"
	@echo "  make migrate-down Rollback migrations"
	@echo "  make docker-build Build Docker image"
	@echo "  make docker-up   Start Docker Compose"
	@echo "  make docker-down Stop Docker Compose"
