
ifneq ($(wildcard .env),)
    include .env
    export
endif

MIGRATIONS_PATH =migrations
DB_URL = postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@$(POSTGRES_HOST):$(POSTGRES_PORT)/$(POSTGRES_DB)?sslmode=$(POSTGRES_SSL_MODE)


.PHONY: run dev build migrate-create migrate-up migrate-down migrate-down-all sqlc test docker-up docker-down docs help gen-mocks test coverage gen-keys


run:
	go run cmd/api/main.go


build:
	go build -o bin/api cmd/api/main.go

migrate-create:
	migrate create -ext sql -dir $(MIGRATIONS_PATH) -seq $(name)

migrate-up:
	migrate -path $(MIGRATIONS_PATH) -database "$(DB_URL)" up


migrate-down:
	migrate -path $(MIGRATIONS_PATH) -database "$(DB_URL)" down 1

migrate-down-all:
	migrate -path $(MIGRATIONS_PATH) -database "$(DB_URL)" down -all


sqlc:
	sqlc generate

docs:
	swag init -g cmd/api/main.go -o docs

test:
	go test -v -cover ./...

docker-up:
	docker compose up -d

docker-down:
	docker compose down

gen-mocks:
	mockery

coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

integration-test:
	go test -v -tags=integration ./internal/integration/...

gen-keys:
	@mkdir -p certs
	@openssl genrsa -out certs/jwt_private.pem 2048
	@openssl rsa -in certs/jwt_private.pem -pubout -out certs/jwt_public.pem
	@echo "RSA Key pairs generated in ./certs/"