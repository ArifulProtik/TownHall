.DEFAULT_GOAL := help

# Colors
GREEN  := \033[0;32m
YELLOW := \033[0;33m
CYAN   := \033[0;36m
RESET  := \033[0m

BIN_DIR   := ./tmp
BIN       := $(BIN_DIR)/main
MAIN_PKG  := ./cmd/api
SCHEMA_DIR := ./ent/schema

.PHONY: help
help: ## Show this help with available commands
	@echo ""
	@echo "$(CYAN)TownHall commands$(RESET)"
	@echo ""
	@awk 'BEGIN {FS = ":.*##"} /^[a-zA-Z0-9_.-]+:.*##/ {printf "  $(GREEN)%-15s$(RESET) %s\n", $$1, $$2}' $(MAKEFILE_LIST)
	@echo ""

.PHONY: dev
dev: ## Run with live reload (air)
	@air

.PHONY: run
run: ## Run the API without reload (go run)
	@go run $(MAIN_PKG)

.PHONY: build
build: ## Build binary to ./tmp/main
	@mkdir -p $(BIN_DIR)
	@go build -o $(BIN) $(MAIN_PKG)
	@echo "$(GREEN)built $(BIN)$(RESET)"

.PHONY: test
test: ## Run all tests
	@go test ./...

.PHONY: test-verbose
test-verbose: ## Run all tests verbosely
	@go test -v ./...

.PHONY: vet
vet: ## Run go vet
	@go vet ./...

.PHONY: lint
lint: ## Run golangci-lint
	@golangci-lint run ./...

.PHONY: lint-fix
lint-fix: ## Run golangci-lint with auto-fix
	@golangci-lint run --fix ./...

.PHONY: tidy
tidy: ## Tidy go modules
	@go mod tidy

.PHONY: check
check: vet lint test ## Vet + lint + test (CI friendly)

.PHONY: ent-new
ent-new: ## Create new ent schema: make ent-new NAME=Group
ifndef NAME
	@echo "$(YELLOW)usage: make ent-new NAME=Group$(RESET)"
	@exit 1
endif
	@go run -mod=mod entgo.io/ent/cmd/ent new --target $(SCHEMA_DIR) $(NAME)

.PHONY: ent-generate
ent-generate: ## Regenerate ent code from schemas
	@go run -mod=mod entgo.io/ent/cmd/ent generate $(SCHEMA_DIR)

.PHONY: ent-describe
ent-describe: ## Print ent schema graph description
	@go run -mod=mod entgo.io/ent/cmd/ent describe $(SCHEMA_DIR)

.PHONY: setup
setup: ## Copy .env.example to .env (if missing) + tidy
	@if [ ! -f .env ]; then cp .env.example .env && echo "$(GREEN)created .env$(RESET)"; else echo "$(YELLOW).env already exists$(RESET)"; fi
	@go mod tidy

.PHONY: up
up: ## Start app + observability stack (Grafana :3000). Stop `make dev` first (:8080 clash)
	@docker compose up -d --build
	@echo "$(GREEN)Grafana: http://localhost:3000 (admin/admin)$(RESET)"

.PHONY: down
down: ## Stop the compose stack (keeps volumes)
	@docker compose down

.PHONY: logs
logs: ## Follow app container logs
	@docker compose logs -f app

.PHONY: clean
clean: ## Remove build artifacts (tmp binary, air log)
	@rm -f $(BIN) build-errors.log
	@echo "$(GREEN)cleaned$(RESET)"
