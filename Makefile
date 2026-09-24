BINARY := ghactions-pin-check
PKG    := ./cmd/ghactions-pin-check
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -s -w -X main.version=$(VERSION)

.PHONY: build test vet fmt tidy selfcheck clean

build:
	go build -ldflags "$(LDFLAGS)" -o bin/$(BINARY) $(PKG)

test:
	go test -race ./...

vet:
	go vet ./...

fmt:
	gofmt -w .

tidy:
	go mod tidy

selfcheck: build
	./bin/$(BINARY) -path .

clean:
	rm -rf bin
