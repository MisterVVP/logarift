.PHONY: deps-backend test test-backend test-math build build-backend build-math run run-backend run-math docker-up docker-down format frontend-install frontend-dev frontend-build

test: test-math test-backend

deps-backend:
	cd backend && GOTOOLCHAIN=local go mod download

test-backend: deps-backend
	cd backend && GOTOOLCHAIN=local go test ./...

test-math:
	$(MAKE) -C math-engine test

build: build-math build-backend

build-backend: deps-backend
	cd backend && GOTOOLCHAIN=local go build -o ../bin/logarift-api ./cmd/api

build-math:
	$(MAKE) -C math-engine OUT=../bin/logarift-math-engine

run: run-backend

run-backend: deps-backend
	cd backend && LOGARIFT_MATH_ENGINE_URL=http://localhost:8090 GOTOOLCHAIN=local go run ./cmd/api

run-math: build-math
	LOGARIFT_MATH_ENGINE_PORT=8090 ./bin/logarift-math-engine --serve

docker-up:
	docker compose up --build

docker-down:
	docker compose down

format:
	gofmt -w backend

frontend-install:
	cd frontend && npm install

frontend-dev:
	cd frontend && npm run dev

frontend-build:
	cd frontend && npm run build
