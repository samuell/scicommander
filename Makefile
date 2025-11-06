.PHONY := build

build:
	go build ./...

test:
	go test ./...

build-windows: go.exe

go.exe:
	GOOS=windows GOARCH=amd64 go build -o win/sci.exe ./...

build-mac-intel:
	GOOS=darwin GOARCH=amd64 go build -o mac-intel/sci ./...

build-mac-m:
	GOOS=darwin GOARCH=arm64 go build -o mac-m/sci ./...

