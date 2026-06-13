BINARY=eks-node-recycler
VERSION?=0.1.0
LDFLAGS=-ldflags "-X main.version=$(VERSION)"

.PHONY: build lint test clean release

build:
	go build $(LDFLAGS) -o bin/$(BINARY) .

lint:
	golangci-lint run ./...

test:
	go test -v -race ./...

clean:
	rm -rf bin/

release:
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY)-linux-amd64 .
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o dist/$(BINARY)-darwin-arm64 .
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY)-windows-amd64.exe .
