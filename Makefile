.PHONY: deps-backend deps-llm-adapter test test-backend test-llm-adapter test-math build build-backend build-llm-adapter build-math run run-backend run-llm-adapter run-math docker-up docker-down format frontend-install frontend-dev frontend-build

test: test-math test-backend test-llm-adapter

deps-backend:
	cd backend && GOTOOLCHAIN=local go mod download

test-backend: deps-backend
	cd backend && GOTOOLCHAIN=local go test ./...

test-math:
	$(MAKE) -C math-engine test

deps-llm-adapter:
	cd llm-adapter && GOTOOLCHAIN=local go mod download

test-llm-adapter: deps-llm-adapter
	cd llm-adapter && GOTOOLCHAIN=local go test ./...

build: build-math build-backend build-llm-adapter

build-backend: deps-backend
	cd backend && GOTOOLCHAIN=local go build -o ../bin/logarift-api ./cmd/api

build-math:
	$(MAKE) -C math-engine OUT=../bin/logarift-math-engine

build-llm-adapter: deps-llm-adapter
	cd llm-adapter && GOTOOLCHAIN=local go build -o ../bin/logarift-llm-adapter ./cmd/llm-adapter

run: run-backend

run-backend: deps-backend
	cd backend && LOGARIFT_MATH_ENGINE_URL=http://localhost:8090 GOTOOLCHAIN=local go run ./cmd/api

run-llm-adapter: deps-llm-adapter
	cd llm-adapter && GOTOOLCHAIN=local go run ./cmd/llm-adapter

run-math: build-math
	LOGARIFT_MATH_ENGINE_PORT=8090 ./bin/logarift-math-engine --serve

docker-up:
	docker compose up --build

docker-down:
	docker compose down

format:
	gofmt -w backend llm-adapter

frontend-install:
	cd frontend && npm install

frontend-dev:
	cd frontend && npm run dev

frontend-build:
	cd frontend && npm run build
