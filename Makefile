BINARY_NAME := sci

.PHONY := build

build:
	go build -o $(BINARY_NAME)

build-tiny:
	tinygo build --no-debug -o $(BINARY_NAME).tiny
