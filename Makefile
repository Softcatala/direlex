.PHONY: help build-assets build generate start lint fix clean

## help: Show this help message
help:
	@grep -E '^## ' $(MAKEFILE_LIST)

## build-assets: Build CSS and JavaScript assets
build-assets:
	go run ./cmd/build-assets

## build: Build the server binary
build: build-assets
	go build -buildvcs=false -ldflags="-s -w" -o direlex ./cmd/server

## generate: Generate static site
generate: build-assets
	go run ./cmd/generate

## start: Build and run the server
start: build
	./direlex

## lint: Run all Go linters
lint:
	go vet ./...
	gofmt -l .

## fix: Format Go code
fix:
	go fmt ./...

## clean: Remove built binaries and build artifacts
clean:
	rm -f direlex
	rm -rf build/
