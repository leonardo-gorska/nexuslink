.PHONY: build test test-integration lint docker-up docker-down migrate loadtest seed

build:
	go build -o bin/api ./cmd/api
	go build -o bin/worker ./cmd/worker

test:
	go test -v -race -coverprofile=coverage.out -covermode=atomic ./...

test-integration:
	go test -v -tags=integration ./internal/adapter/...

lint:
	golangci-lint run

docker-up:
	docker compose -f deployments/docker-compose.yml up -d

docker-down:
	docker compose -f deployments/docker-compose.yml down

migrate:
	@echo "Migration script..."
	./scripts/migrate.sh

seed:
	@echo "Seeding data..."
	./scripts/seed.sh

loadtest:
	@echo "Running loadtest..."
	./scripts/loadtest.sh
