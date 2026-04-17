.PHONY: up down dev test vet build clean migrate-up migrate-down

up:
	docker compose up --build -d

down:
	docker compose down

dev:
	docker compose up --build

test:
	docker compose up --abort-on-container-exit --remove-orphans test

vet:
	go vet ./...

build:
	go build -o bin/foo ./cmd/server/
	go build -o bin/migrate ./cmd/migrate/

migrate-up:
	docker compose run --rm app ./migrate up

migrate-down:
	docker compose run --rm app ./migrate down

clean:
	rm -rf bin/
	docker compose down -v
