# ──────────────────────────────────────────────────────────────────────────────
# Solid VPN — Makefile
#
# Prerequisites: Go 1.23+, Rust 1.82+, Docker, Docker Compose v2
#
# Quick start:
#   cp .env.example .env   # fill in secrets
#   make docker-up         # start all services
#   make migrate           # run database migrations
# ──────────────────────────────────────────────────────────────────────────────

.DEFAULT_GOAL := help
.PHONY: help dev build test lint format migrate \
        docker-up docker-down docker-build docker-logs \
        build-api build-vpn \
        test-api test-vpn \
        lint-api lint-vpn \
        format-api format-vpn \
        clean check-env

# ── Paths ─────────────────────────────────────────────────────────────────────
API_DIR        := apps/api
VPN_DIR        := apps/vpn-engine
AGENT_DIR      := apps/agent
MIGRATIONS_DIR := $(API_DIR)/migrations
BIN_DIR        := bin

# ── Tools ─────────────────────────────────────────────────────────────────────
GO             := go
CARGO          := cargo
DOCKER_COMPOSE := docker compose

# ── Build metadata ────────────────────────────────────────────────────────────
GIT_COMMIT  := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME  := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null || echo "unknown")
VERSION     := $(shell git describe --tags --always 2>/dev/null || echo "0.0.0-dev")
LDFLAGS     := -ldflags "-s -w \
  -X main.version=$(VERSION) \
  -X main.commit=$(GIT_COMMIT) \
  -X main.buildTime=$(BUILD_TIME)"

# ──────────────────────────────────────────────────────────────────────────────
# help — list all targets with descriptions
# ──────────────────────────────────────────────────────────────────────────────
help: ## Show this help message
	@echo ""
	@echo "  Solid VPN — available make targets"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) \
	  | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-22s\033[0m %s\n", $$1, $$2}'
	@echo ""

# ──────────────────────────────────────────────────────────────────────────────
# dev — start the full local development stack
# ──────────────────────────────────────────────────────────────────────────────
dev: check-env docker-up ## Start the full local development stack (alias: docker-up)

# ──────────────────────────────────────────────────────────────────────────────
# build — compile all services
# ──────────────────────────────────────────────────────────────────────────────
build: build-api build-vpn ## Build all services (Go API + Rust VPN engine)

build-api: ## Build the Go control-plane API binary
	@echo "→ Building Go API..."
	@mkdir -p $(BIN_DIR)
	$(GO) build $(LDFLAGS) -o $(BIN_DIR)/vpn-api ./$(API_DIR)/cmd/api
	@echo "✓ $(BIN_DIR)/vpn-api"

build-vpn: ## Build the Rust VPN engine binary
	@echo "→ Building Rust VPN engine..."
	$(CARGO) build --release --manifest-path $(VPN_DIR)/Cargo.toml
	@echo "✓ $(VPN_DIR)/target/release/vpn-engine"

# ──────────────────────────────────────────────────────────────────────────────
# test — run all tests
# ──────────────────────────────────────────────────────────────────────────────
test: test-api test-vpn ## Run all tests (Go + Rust)

test-api: ## Run Go API tests
	@echo "→ Running Go tests..."
	$(GO) test -v -race -coverprofile=coverage.out ./$(API_DIR)/...
	$(GO) tool cover -func=coverage.out | tail -1
	@echo "✓ Go tests passed"

test-vpn: ## Run Rust VPN engine tests
	@echo "→ Running Rust tests..."
	$(CARGO) test --manifest-path $(VPN_DIR)/Cargo.toml
	@echo "✓ Rust tests passed"

# ──────────────────────────────────────────────────────────────────────────────
# lint — run linters
# ──────────────────────────────────────────────────────────────────────────────
lint: lint-api lint-vpn ## Run all linters (Go + Rust)

lint-api: ## Lint the Go API (go vet + staticcheck)
	@echo "→ Linting Go API..."
	$(GO) vet ./$(API_DIR)/...
	@if command -v staticcheck > /dev/null 2>&1; then \
	  staticcheck ./$(API_DIR)/...; \
	else \
	  echo "  staticcheck not found — skipping (install: go install honnef.co/go/tools/cmd/staticcheck@latest)"; \
	fi
	@echo "✓ Go lint passed"

lint-vpn: ## Lint the Rust VPN engine (clippy)
	@echo "→ Linting Rust VPN engine..."
	$(CARGO) clippy --manifest-path $(VPN_DIR)/Cargo.toml -- -D warnings
	@echo "✓ Rust lint passed"

# ──────────────────────────────────────────────────────────────────────────────
# format — auto-format code
# ──────────────────────────────────────────────────────────────────────────────
format: format-api format-vpn ## Auto-format all code (Go + Rust)

format-api: ## Format Go code (gofmt + goimports)
	@echo "→ Formatting Go..."
	$(GO) fmt ./$(API_DIR)/...
	@if command -v goimports > /dev/null 2>&1; then \
	  goimports -w $(API_DIR); \
	else \
	  echo "  goimports not found — skipping (install: go install golang.org/x/tools/cmd/goimports@latest)"; \
	fi
	@echo "✓ Go format done"

format-vpn: ## Format Rust code (rustfmt)
	@echo "→ Formatting Rust..."
	$(CARGO) fmt --manifest-path $(VPN_DIR)/Cargo.toml
	@echo "✓ Rust format done"

# ──────────────────────────────────────────────────────────────────────────────
# migrate — run database migrations
# ──────────────────────────────────────────────────────────────────────────────
migrate: check-env ## Run all pending database migrations (requires DATABASE_URL)
	@echo "→ Running migrations..."
	@if command -v migrate > /dev/null 2>&1; then \
	  migrate -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" up; \
	elif command -v goose > /dev/null 2>&1; then \
	  goose -dir $(MIGRATIONS_DIR) postgres "$(DATABASE_URL)" up; \
	else \
	  echo "  No migration tool found."; \
	  echo "  Install golang-migrate: https://github.com/golang-migrate/migrate"; \
	  echo "  Or goose: go install github.com/pressly/goose/v3/cmd/goose@latest"; \
	  exit 1; \
	fi
	@echo "✓ Migrations applied"

migrate-down: check-env ## Roll back the last migration
	@echo "→ Rolling back last migration..."
	@if command -v migrate > /dev/null 2>&1; then \
	  migrate -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" down 1; \
	elif command -v goose > /dev/null 2>&1; then \
	  goose -dir $(MIGRATIONS_DIR) postgres "$(DATABASE_URL)" down; \
	else \
	  echo "No migration tool found."; exit 1; \
	fi

migrate-status: check-env ## Show current migration status
	@if command -v goose > /dev/null 2>&1; then \
	  goose -dir $(MIGRATIONS_DIR) postgres "$(DATABASE_URL)" status; \
	else \
	  echo "goose not found — install: go install github.com/pressly/goose/v3/cmd/goose@latest"; \
	fi

# ──────────────────────────────────────────────────────────────────────────────
# docker-up / docker-down — manage the local dev environment
# ──────────────────────────────────────────────────────────────────────────────
docker-up: check-env ## Start all services via Docker Compose
	@echo "→ Starting Solid VPN development environment..."
	$(DOCKER_COMPOSE) up --build -d
	@echo ""
	@echo "  Services:"
	@echo "    API         → http://localhost:8080"
	@echo "    API health  → http://localhost:8080/health"
	@echo "    Grafana     → http://localhost:3000"
	@echo "    Prometheus  → http://localhost:9091"
	@echo "    PostgreSQL  → localhost:5432"
	@echo ""
	@echo "  Logs: make docker-logs"
	@echo "  Stop: make docker-down"

docker-down: ## Stop and remove all Docker Compose services
	@echo "→ Stopping Solid VPN development environment..."
	$(DOCKER_COMPOSE) down
	@echo "✓ Services stopped"

docker-down-volumes: ## Stop services AND remove data volumes (DESTROYS database data)
	@echo "→ Stopping services and removing volumes..."
	$(DOCKER_COMPOSE) down -v
	@echo "✓ Services and volumes removed"

docker-build: ## Rebuild all Docker images without cache
	@echo "→ Rebuilding Docker images..."
	$(DOCKER_COMPOSE) build --no-cache
	@echo "✓ Images rebuilt"

docker-logs: ## Follow logs for all services
	$(DOCKER_COMPOSE) logs -f

docker-logs-api: ## Follow logs for the API service only
	$(DOCKER_COMPOSE) logs -f api

docker-logs-vpn: ## Follow logs for the VPN engine service only
	$(DOCKER_COMPOSE) logs -f vpn-engine

docker-ps: ## Show running containers
	$(DOCKER_COMPOSE) ps

# ──────────────────────────────────────────────────────────────────────────────
# clean — remove build artifacts
# ──────────────────────────────────────────────────────────────────────────────
clean: ## Remove all build artifacts
	@echo "→ Cleaning build artifacts..."
	@rm -rf $(BIN_DIR)
	@rm -f coverage.out coverage.html
	$(CARGO) clean --manifest-path $(VPN_DIR)/Cargo.toml
	@echo "✓ Clean done"

# ──────────────────────────────────────────────────────────────────────────────
# check-env — verify .env exists before commands that need it
# ──────────────────────────────────────────────────────────────────────────────
check-env: ## Verify that .env exists (copy from .env.example if missing)
	@if [ ! -f .env ]; then \
	  echo ""; \
	  echo "  ERROR: .env file not found."; \
	  echo "  Run: cp .env.example .env  — then fill in the required secrets."; \
	  echo ""; \
	  exit 1; \
	fi

# ──────────────────────────────────────────────────────────────────────────────
# Utility targets
# ──────────────────────────────────────────────────────────────────────────────
tidy: ## Tidy Go module dependencies
	@echo "→ Tidying Go modules..."
	$(GO) mod tidy -C $(API_DIR)
	@echo "✓ Done"

vendor: ## Vendor Go dependencies
	$(GO) mod vendor -C $(API_DIR)

deps-update: ## Update all Go dependencies to latest minor/patch
	$(GO) get -u ./$(API_DIR)/...
	$(GO) mod tidy -C $(API_DIR)

health-check: ## Check the API health endpoint (requires running stack)
	@curl -sf http://localhost:8080/health

.PHONY: migrate-down migrate-status docker-down-volumes docker-build \
        docker-logs docker-logs-api docker-logs-vpn docker-ps \
        tidy vendor deps-update health-check
