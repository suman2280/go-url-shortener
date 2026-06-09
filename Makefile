.PHONY: dev build test clean lint swagger k6-up k6-redirect

APP_NAME = urlshortener

dev:
	docker compose up --build

build:
	go build -o bin/$(APP_NAME) ./cmd/api

run:
	go run ./cmd/api

test:
	go test -race -cover ./internal/... ./pkg/...

test-verbose:
	go test -race -cover -v ./internal/... ./pkg/...

lint:
	golangci-lint run ./...

clean:
	rm -rf bin/
	go clean

swagger:
	swag init -g cmd/api/main.go -o docs

migrate:
	go run ./cmd/api migrate

docker-build:
	docker compose build

docker-up:
	docker compose up -d

docker-down:
	docker compose down

docker-logs:
	docker compose logs -f app

docker-clean:
	docker compose down -v

k6-up:
	docker run --rm -i --network host grafana/k6 run - < k6/shorten.js

k6-redirect:
	docker run --rm -i --network host grafana/k6 run - < k6/redirect.js

.PHONY: dev build run test test-verbose lint clean swagger docker-build docker-up docker-down docker-logs docker-clean k6-up k6-redirect
