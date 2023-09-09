LDFLAGS = -s -w
GOFLAGS = -trimpath -ldflags="$(LDFLAGS)"

all: build test

.PHONY: release
release: test
	mkdir -p bin
	GOOS=darwin  GOARCH=arm64 go build $(GOFLAGS) -o bin/man-get-darwin-arm64 -v
	GOOS=linux   GOARCH=amd64 go build $(GOFLAGS) -o bin/man-get-linux-amd64 -v
	GOOS=windows GOARCH=amd64 go build $(GOFLAGS) -o bin/man-get-windows-amd64.exe -v

.PHONY: build
build:
	mkdir -p bin
	go build $(GOFLAGS) -o bin/man-get -v

.PHONY: test
test:
	go test -v ./...
