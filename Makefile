.PHONY: run build test lint clean docker-up docker-down migrate help

# Variables
APP_NAME=golang-boilerplate
MAIN_PATH=./cmd/api

## help: Show this help message
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@sed -n 's/^## //p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/  /'

## run: Run the application
run:
	go run $(MAIN_PATH)

## build: Build the application binary
build:
	CGO_ENABLED=0 go build -o bin/$(APP_NAME) $(MAIN_PATH)

## test: Run all tests
test:
	go test ./... -v -cover

## test-coverage: Run tests with coverage report
test-coverage:
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html

## lint: Run linter
lint:
	golangci-lint run ./...

## clean: Clean build artifacts
clean:
	rm -rf bin/
	rm -f coverage.out coverage.html

## docker-up: Start all services with Docker Compose
docker-up:
	docker-compose up -d --build

## docker-down: Stop all Docker Compose services
docker-down:
	docker-compose down

## docker-logs: View Docker Compose logs
docker-logs:
	docker-compose logs -f

## migrate: Run database migrations
migrate:
	go run $(MAIN_PATH) migrate

## tidy: Tidy Go modules
tidy:
	go mod tidy

## fmt: Format Go code
fmt:
	gofmt -s -w .

## vet: Run go vet
vet:
	go vet ./...
