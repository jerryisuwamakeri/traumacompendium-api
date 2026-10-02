.PHONY: run build test test-unit test-integration lint fmt vet tidy migrate docker-up docker-down docker-logs clean

BINARY := bin/api
MAIN   := ./cmd/api

run: ## Run the API locally (reads .env via godotenv)
	go run $(MAIN)

build: ## Build the API binary into bin/
	go build -trimpath -o $(BINARY) $(MAIN)

test: ## Run the full test suite (integration tests skip without TEST_DATABASE_DSN/REDIS_ADDR)
	go test ./...

test-unit: ## Run only unit tests (no real MySQL/Redis required)
	go test $$(go list ./... | grep -v internal/repository/mysql)

test-integration: ## Run the full suite against a real MySQL/Redis (requires TEST_DATABASE_DSN)
	go test ./internal/repository/... ./internal/http/middleware/... -v

lint: ## Run golangci-lint
	golangci-lint run ./...

fmt: ## Format the whole tree
	gofmt -l -w .

vet: ## Run go vet
	go vet ./...

tidy: ## Tidy go.mod/go.sum
	go mod tidy

migrate: build ## Apply pending migrations by starting the API briefly (it migrates on boot)
	@echo "Migrations run automatically on API startup; 'make run' or 'make build && $(BINARY)' is sufficient."

docker-up: ## Start api + mysql + redis via docker compose
	docker compose up --build

docker-down: ## Stop and remove the docker compose stack
	docker compose down

docker-logs: ## Tail the api container's logs
	docker compose logs -f api

clean: ## Remove build artifacts
	rm -rf bin/
