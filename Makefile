.PHONY: build tests vet lint verify

build:
	mkdir -p build
	go build -o build/powerkit-cli ./cmd/powerkit-cli

tests:
	go test ./...

vet:
	go vet ./...

lint:
	@if ! command -v golangci-lint >/dev/null 2>&1; then \
	  echo "error: golangci-lint not found in PATH. Install golangci-lint to run lint checks."; \
	  exit 1; \
	fi
	golangci-lint run

verify: tests vet lint
