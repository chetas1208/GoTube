.PHONY: dev up down build migrate migrate-down seed test test-api test-worker test-web lint clean

# --- Docker Compose ---
up:
	docker compose -f infra/docker/docker-compose.yml --env-file .env up --build

down:
	docker compose -f infra/docker/docker-compose.yml --env-file .env down -v

dev:
	docker compose -f infra/docker/docker-compose.yml --env-file .env up --build --watch

logs:
	docker compose -f infra/docker/docker-compose.yml --env-file .env logs -f

# --- Database ---
migrate:
	cd backend/api && go run cmd/api/main.go migrate up

migrate-down:
	cd backend/api && go run cmd/api/main.go migrate down

seed:
	cd backend/api && go run cmd/api/main.go seed

# --- Build ---
build-api:
	cd backend/api && go build -o bin/api cmd/api/main.go

build-worker:
	cd backend/worker && go build -o bin/worker cmd/worker/main.go

build-web:
	cd frontend && npm run build

# --- Test ---
test: test-api test-worker test-web

test-api:
	cd backend/api && go test ./... -v -cover

test-worker:
	cd backend/worker && go test ./... -v -cover

test-web:
	cd frontend && npm test

# --- Lint ---
lint-api:
	cd backend/api && golangci-lint run ./...

lint-web:
	cd frontend && npm run lint

# --- Clean ---
clean:
	rm -rf backend/api/bin backend/worker/bin frontend/.next tmp/*
