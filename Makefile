.PHONY: build test test-integration lint migrate-up run compose-up compose-down

build:
	go build ./...

test:
	go test ./...

# Requires a reachable Postgres via DATABASE_URL. `make compose-up` starts one locally.
test-integration:
	go test -tags=integration ./...

lint:
	golangci-lint run

migrate-up:
	go run ./cmd/migrate

run:
	go run ./cmd/tonberry

compose-up:
	docker compose up -d postgres

compose-down:
	docker compose down
