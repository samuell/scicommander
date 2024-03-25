BINARY_NAME := scicmd

.PHONY := build

build:
	go build -o $(BINARY_NAME)
