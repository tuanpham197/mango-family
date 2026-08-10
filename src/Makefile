# Household Finance — Go + Vue monorepo (re-platform 2026-07)
# Design: ../docs/superpowers/specs/2026-07-09-go-vue-replatform-design.md
# Layout: api/ (Go · Gin · GORM) · web/ (Vue 3 + Vite) · db/migrations/ (goose SQL)
# Chạy `make help` để xem toàn bộ lệnh.

SHELL := /bin/bash

# ---- Config (override khi cần: make migrate-up DATABASE_URL=...) ----
POSTGRES_USER     ?= app
POSTGRES_PASSWORD ?= app_secret
POSTGRES_DB       ?= finance_dev
POSTGRES_PORT     ?= 5432
DATABASE_URL      ?= postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@localhost:$(POSTGRES_PORT)/$(POSTGRES_DB)?sslmode=disable
MIGRATIONS_DIR    := db/migrations
GOOSE             ?= goose

.PHONY: help tools up down ps logs clean-db \
        migrate-up migrate-down migrate-status migrate-create \
        seed seed-prod api dev-api web dev build \
        test-api test-api-integration test-web test-e2e test

help: ## Liệt kê các lệnh
	@grep -E '^[a-zA-Z0-9_-]+:.*?## ' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-22s\033[0m %s\n", $$1, $$2}'

tools: ## Cài công cụ dev (goose, air)
	go install github.com/pressly/goose/v3/cmd/goose@latest
	go install github.com/air-verse/air@latest

# ---- Docker / Postgres ----
up: ## Khởi động Postgres 16 (docker compose)
	docker compose up -d postgres
	@echo "Postgres sẵn sàng: $(DATABASE_URL)"

down: ## Dừng toàn bộ container
	docker compose down

ps: ## Trạng thái container
	docker compose ps

logs: ## Theo dõi log Postgres
	docker compose logs -f postgres

clean-db: ## XÓA volume Postgres (mất sạch dữ liệu dev) rồi khởi động lại + migrate
	docker compose down -v
	docker compose up -d postgres
	@sleep 3 && $(MAKE) migrate-up

# ---- Migrations (goose — SQL thuần, KHÔNG dùng GORM AutoMigrate) ----
migrate-up: ## Áp toàn bộ migration (00001_users … )
	$(GOOSE) -dir $(MIGRATIONS_DIR) postgres "$(DATABASE_URL)" up

migrate-down: ## Rollback 1 migration gần nhất
	$(GOOSE) -dir $(MIGRATIONS_DIR) postgres "$(DATABASE_URL)" down

migrate-status: ## Trạng thái migration
	$(GOOSE) -dir $(MIGRATIONS_DIR) postgres "$(DATABASE_URL)" status

migrate-create: ## Tạo migration mới: make migrate-create name=ten_migration
	@test -n "$(name)" || (echo "Usage: make migrate-create name=<ten_migration>"; exit 1)
	$(GOOSE) -dir $(MIGRATIONS_DIR) create $(name) sql

# ---- App (dev) ----
seed: ## Seed dev: Alice/Bob (hộ A), Carol (hộ B) + danh mục & tài khoản mặc định
	cd api && DATABASE_URL="$(DATABASE_URL)" go run ./cmd/seed

seed-prod: ## Seed tài khoản thật từ deploy/seed.env (cần DATABASE_URL=<Supabase>)
	@test -f deploy/seed.env || (echo "Thiếu deploy/seed.env — copy từ deploy/seed.env.example"; exit 1)
	cd api && set -a && . ../deploy/seed.env && set +a && DATABASE_URL="$(DATABASE_URL)" go run ./cmd/seed

api: ## Chạy Go API (Gin, cổng 8080)
	cd api && DATABASE_URL="$(DATABASE_URL)" go run .

dev-api: ## Chạy Go API hot-reload qua air
	cd api && DATABASE_URL="$(DATABASE_URL)" air

web: ## Chạy Vue dev server (Vite — proxy /api + /ws sang :8080)
	cd web && npm run dev

dev: up migrate-up ## Một phát: Postgres + migrations (lần đầu thêm `make seed`)
	@echo "→ Tiếp theo: 'make seed' (lần đầu), rồi 'make api' và 'make web' ở 2 terminal."

build: ## Build production: web/dist + Go binary (Go serve web/dist — một origin)
	cd web && npm run build
	cd api && go build -o bin/server .

# ---- Tests ----
test-api: ## Unit test Go (tầng biz)
	cd api && go test ./...

test-api-integration: ## Integration test storage (cần Postgres đang chạy)
	cd api && DATABASE_URL="$(DATABASE_URL)" go test ./... -tags=integration

test-web: ## Vitest (stores/components)
	cd web && npm run test:unit

test-e2e: ## Playwright e2e (18 kịch bản 001 + 23 kịch bản 002)
	cd web && npx playwright test

test: test-api test-web ## Toàn bộ unit test (Go + Vue)
