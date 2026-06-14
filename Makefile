.PHONY: run build tidy vet test sqlc migrate-up migrate-down tools

# --- Phase 0 ---

run: ## Run the API server
	go run ./cmd/api

build: ## Build the API binary to bin/api
	go build -o bin/api ./cmd/api

hashpw: ## Generate a bcrypt hash for ADMIN_PASSWORD_HASH (make hashpw PW=secret)
	go run ./cmd/hashpw "$(PW)"

tidy: ## Sync go.mod/go.sum
	go mod tidy

vet: ## Static analysis
	go vet ./...

test: ## Run tests
	go test ./...

# --- Phase 1+ (require tools below) ---

sqlc: ## Regenerate type-safe DB code from SQL
	sqlc generate

migrate-up: ## Apply all pending migrations (needs DB_PATH and goose)
	goose -dir migrations sqlite3 "$(DB_PATH)" up

migrate-down: ## Roll back the last migration
	goose -dir migrations sqlite3 "$(DB_PATH)" down

tools: ## Install dev tooling (goose, sqlc)
	go install github.com/pressly/goose/v3/cmd/goose@latest
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
