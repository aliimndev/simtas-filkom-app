# SIMTAS FILKOM — Backend Makefile (Jobs 22, 23, 26)
# Convenience targets for local development and CI.

.PHONY: build run test test-unit test-integration test-coverage coverage-check vet lint docs swagger

build:
	go build ./...

run:
	go run ./cmd/server

vet:
	go vet ./...

# ── Tests (Job 23) ─────────────────────────────────────────────────────
test:
	go test ./... -count=1

test-unit:
	go test ./pkg/... ./internal/usecase/... -count=1 -v

# Integration tests use the `integration` build tag (skipped by default).
test-integration:
	go test ./internal/handler/... -count=1 -v -tags integration

test-coverage:
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html

# Coverage gate (Job 23): the usecase layer must stay ≥ 80%. CI fails the
# pipeline when the measured coverage drops below the target. Uses awk for the
# float comparison so it runs on stock ubuntu runners (no bc dependency).
coverage-check:
	go test ./internal/usecase/... -coverprofile=coverage.out -count=1 > /dev/null
	@coverage=$$(go tool cover -func=coverage.out | grep '^total:' | awk '{print $$NF}' | tr -d '%'); \
	echo "usecase coverage: $$coverage%"; \
	awk -v c="$$coverage" 'BEGIN { exit !(c >= 80) }' || { echo "FAIL: usecase coverage below 80%"; exit 1; }

# ── API docs (Job 22) ──────────────────────────────────────────────────
# Regenerates OpenAPI docs from swagger annotations into ./docs so the
# generated package (backend/docs) can be imported by cmd/server/main.go.
# Requires:
#   go install github.com/swaggo/swag/cmd/swag@latest
docs:
	swag init -g cmd/server/main.go -o docs

swagger: docs
