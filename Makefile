
ifneq ($(wildcard .env),)
    include .env
    export
endif

MIGRATIONS_PATH = db/migration
DB_URL = postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSLMODE)


.PHONY: run dev build migrate-create migrate-up migrate-down migrate-down-all sqlc test docker-up docker-down docs help


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