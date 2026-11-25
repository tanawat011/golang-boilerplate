.PHONY: help run build test test-integration cover lint vuln fmt vet tidy docker-up docker-down docker-logs db-up clean

APP_NAME  := golang-boilerplate
MAIN_PATH := ./cmd/api
# Admin connection used by the integration tests (each test creates and drops
# its own database). Matches the PostgreSQL started by `make db-up`.
TEST_DATABASE_URL ?= postgres://postgres:postgres@127.0.0.1:15432/postgres?sslmode=disable

## help: Show this help
help:
	@sed -n 's/^## //p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/  /'

## run: Run the API on the host (needs .env and a database: make db-up)
run:
	go run $(MAIN_PATH)

## build: Build the binary into bin/
build:
	CGO_ENABLED=0 go build -trimpath -o bin/$(APP_NAME) $(MAIN_PATH)

## test: Run unit tests (no database needed)
test:
	go test -race ./...

## test-integration: Run integration tests against PostgreSQL (make db-up first)
test-integration:
	TEST_DATABASE_URL="$(TEST_DATABASE_URL)" go test -race -tags integration ./...

## cover: Unit test coverage report (coverage.html)
cover:
	go test -coverpkg=./... -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

## lint: Run golangci-lint
lint:
	golangci-lint run ./...

## vuln: Check dependencies for known vulnerabilities
vuln:
	go run golang.org/x/vuln/cmd/govulncheck@latest ./...

## fmt: Format the code
fmt:
	gofmt -s -w .

## vet: Run go vet
vet:
	go vet -tags integration ./...

## tidy: Tidy go.mod / go.sum
tidy:
	go mod tidy

## db-up: Start only PostgreSQL (127.0.0.1:15432)
db-up:
	docker compose up -d --wait postgres

## docker-up: Build and start the API and PostgreSQL
docker-up:
	docker compose up -d --build --wait

## docker-down: Stop the stack (add -v to the command to drop the data volume)
docker-down:
	docker compose down

## docker-logs: Follow the logs
docker-logs:
	docker compose logs -f

## clean: Remove build artefacts
clean:
	rm -rf bin/ coverage.out coverage.html
