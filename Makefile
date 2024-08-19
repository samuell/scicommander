.PHONY := build

build:
	go build ./...

build-darwin-amd64:
	mkdir -p build/darwin-amd64
	go build -o build/darwin-amd64/sci ./...

build-darwin-arm64:
	mkdir -p build/darwin-arm64
	go build -o build/darwin-arm64/sci ./...

test:
	go test ./...
