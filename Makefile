BINARY_NAME := sci

.PHONY := build

build:
	go build -o $(BINARY_NAME)
