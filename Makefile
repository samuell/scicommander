.PHONY := build

build:
	go build ./...

build-tiny:
	tinygo build --no-debug -o $(BINARY_NAME).tiny

test:
	go test ./...
