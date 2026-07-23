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

scicommander-${VERSION}-linux-amd64.tar.gz: linux-v${VERSION}/sci
	tar -zcvf $@ $<

scicommander-${VERSION}-windows-amd64.zip: windows-v${VERSION}/sci.exe
	zip $@ $<

scicommander-${VERSION}-mac-amd64.tar.gz: mac-intel-v${VERSION}/sci
	tar -zcvf $@ $<

scicommander-${VERSION}-mac-arm64.tar.gz: mac-m-v${VERSION}/sci
	tar -zcvf $@ $<

linux-v${VERSION}/sci:
	mkdir -p $$(dirname $@)
	GOOS=linux GOARCH=amd64 go build -o $@ ./...

windows-v${VERSION}/sci.exe:
	mkdir -p $$(dirname $@)
	GOOS=windows GOARCH=amd64 go build -o $@ ./...

mac-intel-v${VERSION}/sci:
	mkdir -p $$(dirname $@)
	GOOS=darwin GOARCH=amd64 go build -o $@ ./...

mac-m-v${VERSION}/sci:
	mkdir -p $$(dirname $@)
	GOOS=darwin GOARCH=arm64 go build -o $@ ./...
