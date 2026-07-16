.PHONY: help deps-up deps-down backend-run backend-build backend-test \
        frontend-install frontend-dev frontend-build \
        build build-backend build-frontend up down logs

DOCKERHUB_BACKEND_IMAGE  ?= abdullahprasetio/waphafiz-backend
DOCKERHUB_FRONTEND_IMAGE ?= abdullahprasetio/waphafiz-frontend
TAG                      ?= local
NUXT_PUBLIC_API_BASE     ?=

help: ## tampilkan daftar target
	@grep -E '^[a-zA-Z_-]+:.*## ' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*## "}; {printf "  %-20s %s\n", $$1, $$2}'

## ── Dev lokal (tanpa Docker image) ─────────────────────────

deps-up: ## start PostgreSQL + Redis untuk dev backend
	cd backend && make docker-up

deps-down: ## stop PostgreSQL + Redis
	cd backend && make docker-down

backend-run: ## jalankan backend lokal (go run)
	cd backend && make run

backend-build: ## build binary backend ke backend/bin/api
	cd backend && make build

backend-test: ## jalankan unit + integration test backend
	cd backend && make test-race

frontend-install: ## install dependency frontend
	cd frontend && npm ci

frontend-dev: ## jalankan dev server frontend (port 3000)
	cd frontend && npm run dev

frontend-build: ## build production frontend (.output/)
	cd frontend && npm run build

## ── Docker image (build lokal, platform native) ────────────

build-backend: ## build docker image backend: waphafiz-backend:local
	docker build -t $(DOCKERHUB_BACKEND_IMAGE):$(TAG) ./backend

build-frontend: ## build docker image frontend: waphafiz-frontend:local
	docker build -t $(DOCKERHUB_FRONTEND_IMAGE):$(TAG) \
		$(if $(NUXT_PUBLIC_API_BASE),--build-arg NUXT_PUBLIC_API_BASE=$(NUXT_PUBLIC_API_BASE),) \
		./frontend

build: build-backend build-frontend ## build kedua image sekaligus

## ── Full stack via docker-compose (butuh .env.prod) ────────

up: ## start full stack (postgres, redis, backend, frontend)
	docker compose --env-file .env.prod up -d --build

down: ## stop full stack
	docker compose --env-file .env.prod down

logs: ## tail log semua service
	docker compose --env-file .env.prod logs -f
