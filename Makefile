.PHONY: build tests

build:
	mkdir -p build
	go build -o build/powerkit-cli ./cmd/powerkit-cli

tests:
	go test ./...
