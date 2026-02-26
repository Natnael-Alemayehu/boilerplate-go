.PHONY: build run test clean migrate-up migrate-down migrate-status sqlc-generate docker-up docker-down docker-logs deps lint fmt vet generate-keys

BINARY_NAME=api
MAIN_PATH=./cmd/api

build:
	go build -o bin/$(BINARY_NAME) $(MAIN_PATH)

run:
	go run $(MAIN_PATH)

test:
	go test -v -race -coverprofile=coverage.out ./...

test-coverage: test
	go tool cover -html=coverage.out

clean:
	rm -rf bin/
	rm -f coverage.out

deps:
	go mod download
	go mod tidy

lint:
	golangci-lint run ./...

fmt:
	go fmt ./...

vet:
	go vet ./...

migrate-up:
	goose -dir db/migrations postgres "$(DATABASE_URL)" up

migrate-down:
	goose -dir db/migrations postgres "$(DATABASE_URL)" down

migrate-status:
	goose -dir db/migrations postgres "$(DATABASE_URL)" status

migrate-create:
	@read -p "Enter migration name: " name; \
	goose -dir db/migrations create $$name sql

generate-keys:
	@./scripts/generate_keys.sh keys
	@echo "Generating refresh key pair..."
	@./scripts/generate_keys.sh keys/refresh
	@mv keys/refresh/private.pem keys/refresh_private.pem
	@mv keys/refresh/public.pem keys/refresh_public.pem
	@rmdir keys/refresh
	@mv keys/private.pem keys/access_private.pem
	@mv keys/public.pem keys/access_public.pem

sqlc-generate:
	sqlc generate

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

docker-logs:
	docker-compose logs -f

docker-build:
	docker-compose build

docker-restart:
	docker-compose restart

redis-cli:
	docker-compose exec redis redis-cli

psql:
	docker-compose exec postgres psql -U postgres -d app

dev: docker-up
	sleep 3
	go run $(MAIN_PATH)

all: deps sqlc-generate build