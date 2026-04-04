BINARY_NAME=llmvlt
VERSION?=0.1.0
LDFLAGS=-ldflags "-X main.version=$(VERSION)"

.PHONY: build install clean test

build:
	go build $(LDFLAGS) -o bin/$(BINARY_NAME) .

install:
	go install $(LDFLAGS) .

clean:
	rm -rf bin/

test:
	go test ./...

# Cross-compile for releases
release:
	GOOS=linux   GOARCH=amd64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-linux-amd64 .
	GOOS=linux   GOARCH=arm64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-linux-arm64 .
	GOOS=darwin  GOARCH=arm64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-macos-arm64 .
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-windows-amd64.exe .
