.PHONY: build tests vet lint verify

build:
	mkdir -p build
	go build -o build/powerkit-cli ./cmd/powerkit-cli

tests:
	go test ./...

vet:
	go vet ./...

lint:
	golangci-lint run

verify: tests vet
