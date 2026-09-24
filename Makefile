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

.PHONY: dev-all
dev-all: ## Run backend (air) + frontend (bun) concurrently
	@trap 'kill 0' INT TERM; air & bun --cwd=ui run dev & wait

.PHONY: run
run: ## Run the API without reload (go run)
	@go run $(MAIN_PKG)

.PHONY: build
build: ## Build binary to ./tmp/main
	@mkdir -p $(BIN_DIR)
	@go build -o $(BIN) $(MAIN_PKG)
	@echo "$(GREEN)built $(BIN)$(RESET)"

# Go package list. Lazily evaluated and only for Go-dependent targets, so
# parsing the Makefile never invokes `go list` for unrelated targets
# (e.g. dev-docker). Fails fast when discovery errors instead of letting
# the grep pipe mask the failure with an empty list.
GO_TARGETS := $(filter test test-verbose vet check,$(MAKECMDGOALS))
ifeq ($(GO_TARGETS),)
GO_PKGS :=
else
_GO_PKGS_RAW := $(shell go list ./...)
ifeq ($(_GO_PKGS_RAW),)
$(error go list ./... failed: cannot determine Go packages)
endif
GO_PKGS = $(shell printf '%s\n' $(_GO_PKGS_RAW) | grep -v /ui)
endif

.PHONY: test
test: ## Run all tests
	@go test $(GO_PKGS)

.PHONY: test-verbose
test-verbose: ## Run all tests verbosely
	@go test -v $(GO_PKGS)

.PHONY: vet
vet: ## Run go vet
	@go vet $(GO_PKGS)

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

.PHONY: dev-docker
dev-docker: ## Start dev backend stack (hot-reload api :8080, pg, redis). Pair with `bun --cwd=ui run dev` for the frontend
	@docker compose -f docker-compose.dev.yml up -d --build
	@echo "$(GREEN)api: http://localhost:8080 | run frontend natively: bun --cwd=ui run dev$(RESET)"

.PHONY: dev-docker-full
dev-docker-full: ## Start dev stack WITH dockerized web (:5173). Heavier; native bun is the default
	@docker compose -f docker-compose.dev.yml --profile web up -d --build
	@echo "$(GREEN)api: http://localhost:8080 | web: http://localhost:5173$(RESET)"

.PHONY: dev-docker-down
dev-docker-down: ## Stop the dev stack (keeps volumes)
	@docker compose -f docker-compose.dev.yml --profile web down

.PHONY: dev-docker-logs
dev-docker-logs: ## Follow dev api logs
	@docker compose -f docker-compose.dev.yml logs -f api

.PHONY: up
up: ## Start prod app + observability stack (Grafana :3000). Stop `make dev` first (:8080 clash)
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
