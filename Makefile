# Simple Makefile for a Go project

VERSION    ?= 1.0.0
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME := $(shell date -u +%Y-%m-%dT%H:%M:%SZ 2>/dev/null || echo "unknown")
LDFLAGS    := -X maya-canteen/internal/version.Version=$(VERSION) \
              -X maya-canteen/internal/version.GitCommit=$(GIT_COMMIT) \
              -X maya-canteen/internal/version.BuildTime=$(BUILD_TIME)

# Build the application
all: build test

build:
	@echo "Building..."
	@CGO_ENABLED=1 GOOS=linux go build -ldflags "$(LDFLAGS)" -o main cmd/api/main.go

build-windows:
	@echo "Generating Windows resources..."
	@cd cmd/api && go-winres make --product-version $(VERSION).0 --file-version $(VERSION).0
	@echo "Building Windows executable..."
	@CGO_ENABLED=1 GOOS=windows go build -ldflags "$(LDFLAGS)" -o maya-canteen.exe cmd/api/main.go

# Run the application
run:
	@go run cmd/api/main.go &
	@cd frontend && pnpm install --prefer-offline
	@cd frontend && pnpm run dev
# Create DB container
docker-run:
	@if docker compose up --build 2>/dev/null; then \
		: ; \
	else \
		echo "Falling back to Docker Compose V1"; \
		docker-compose up --build; \
	fi

# Shutdown DB container
docker-down:
	@if docker compose down 2>/dev/null; then \
		: ; \
	else \
		echo "Falling back to Docker Compose V1"; \
		docker-compose down; \
	fi

# Test the application
test:
	@echo "Testing..."
	@go test ./... -v

# Clean the binary
clean:
	@echo "Cleaning..."
	@rm -f main maya-canteen.exe
	@find cmd/api -name '*.syso' -delete

# Live Reload
watch:
	@if command -v air > /dev/null; then \
            air; \
            echo "Watching...";\
        else \
            read -p "Go's 'air' is not installed on your machine. Do you want to install it? [Y/n] " choice; \
            if [ "$$choice" != "n" ] && [ "$$choice" != "N" ]; then \
                go install github.com/air-verse/air@latest; \
                air; \
                echo "Watching...";\
            else \
                echo "You chose not to install air. Exiting..."; \
                exit 1; \
            fi; \
        fi

.PHONY: all build build-windows run test clean watch