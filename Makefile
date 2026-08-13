.PHONY: build test lint fmt mocks swagger web-build web-dev help

# build compiles every Go package under src/.
build:
	go build ./src/...

# test runs the unit-test packages; the $(...) form propagates a failing exit code.
test:
	go test $$(go list ./tests/... | grep '_unit_tests')

# lint requires golangci-lint v2 on PATH (go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest).
lint:
	golangci-lint run ./...

# fmt runs golines via golangci-lint's built-in formatter; no standalone golines needed.
fmt:
	golangci-lint fmt

# mocks requires the deepsearch-mockgen CLI (github.com/FournyP/deepsearch-mockgen-cli).
mocks:
	deepsearch-mockgen -S ./src -O ./tests -A -P

# swagger regenerates each host's API docs (needs the go toolchain; runs swag via go run).
swagger:
	./src/gateway_host/docs/generate_docs.sh
	./src/runtime_host/docs/generate_docs.sh

web-build:
	cd web && npm ci && npm run build

web-dev:
	cd web && npm run dev

help:
	@echo "build       - go build ./src/..."
	@echo "test        - go test unit-test packages"
	@echo "lint        - golangci-lint run ./... (needs golangci-lint v2)"
	@echo "fmt         - golangci-lint fmt, 120-char cap"
	@echo "mocks       - regenerate mocks (needs deepsearch-mockgen)"
	@echo "swagger     - regenerate host API docs (needs go toolchain)"
	@echo "web-build   - build the web console (cd web && npm ci && npm run build)"
	@echo "web-dev     - run the web console dev server (cd web && npm run dev)"
