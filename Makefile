.PHONY: test test-unit test-integration test-host tidy fmt vet

# Default: the containerized unit suite, mirroring rdt's `tests unit`.
test: test-unit

# Unit tests inside the app container (Go's -short skips anything that needs a
# live service, the idiomatic stand-in for rdt's `not integration_test` marker).
test-unit:
	docker compose build app
	docker compose run --rm app go test -short ./...

# Full suite including integration tests: bring up Postgres, wait for it to be
# healthy, then run without -short so service-backed tests run. (redis/minio
# join here as the modules that need them are ported.)
test-integration:
	docker compose up -d --wait postgres
	docker compose build app
	docker compose run --rm app go test ./...

# Host-side convenience (no container) — fast inner loop while developing.
test-host:
	go test ./...

tidy:
	go mod tidy

fmt:
	go fmt ./...

vet:
	go vet ./...
