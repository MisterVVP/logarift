.PHONY: test build run docker-up docker-down format

test:
	cd backend && go test ./...

build:
	cd backend && go build -o ../bin/logarift-api ./cmd/api

run:
	cd backend && go run ./cmd/api

docker-up:
	docker compose up --build

docker-down:
	docker compose down

format:
	gofmt -w backend
