.PHONY := build

VERSION:=0.6.1

build:
	go build ./...

test:
	go test ./...

release: scicommander-${VERSION}-linux-amd64.tar.gz \
	scicommander-${VERSION}-windows-amd64.zip \
	scicommander-${VERSION}-mac-amd64.tar.gz \
	scicommander-${VERSION}-mac-arm64.tar.gz

scicommander-${VERSION}-linux-amd64.tar.gz: linux/sci
	tar -zcvf $@ $<

scicommander-${VERSION}-windows-amd64.zip: windows/sci.exe
	zip $@ $<

scicommander-${VERSION}-mac-amd64.tar.gz: mac-intel/sci
	tar -zcvf $@ $<

scicommander-${VERSION}-mac-arm64.tar.gz: mac-m/sci
	tar -zcvf $@ $<

linux/sci:
	mkdir -p linux
	GOOS=linux GOARCH=amd64 go build -o linux/sci ./...

windows/sci.exe:
	mkdir -p windows
	GOOS=windows GOARCH=amd64 go build -o windows/sci.exe ./...

mac-intel/sci:
	mkdir -p mac-intel
	GOOS=darwin GOARCH=amd64 go build -o mac-intel/sci ./...

mac-m/sci:
	mkdir -p mac-m
	GOOS=darwin GOARCH=arm64 go build -o mac-m/sci ./...


clean:
	rm -rf linux windows mac-intel mac-m
