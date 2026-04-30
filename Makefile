APP_NAME=fintrack
DB_URL?=postgres://fintrack:fintrack_secret@localhost:5432/fintrack?sslmode=disable
MIGRATIONS_DIR=migrations/postgres

.PHONY: run build clean test lint migrate-new migrate-up migrate-down migrate-status seed docker-up docker-down

run:
	go run cmd/web/main.go

seed:
	go run cmd/seed/main.go

build:
	go build -o bin/$(APP_NAME) cmd/web/main.go

clean:
	rm -rf bin

test:
	go test ./... -v

lint:
	golangci-lint run ./...

migrate-new:
	goose -dir $(MIGRATIONS_DIR) create $(name) sql

migrate-up:
	goose -dir $(MIGRATIONS_DIR) postgres "$(DB_URL)" up

migrate-down:
	goose -dir $(MIGRATIONS_DIR) postgres "$(DB_URL)" down

migrate-status:
	goose -dir $(MIGRATIONS_DIR) postgres "$(DB_URL)" status

docker-up:
	docker compose up -d

docker-down:
	docker compose down
