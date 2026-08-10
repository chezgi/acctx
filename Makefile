BINARY := bin/acctx
VERSION ?= dev
COMMIT ?= unknown
BUILD_DATE ?= unknown
LDFLAGS := -X acctx/internal/buildinfo.Version=$(VERSION) -X acctx/internal/buildinfo.Commit=$(COMMIT) -X acctx/internal/buildinfo.BuiltAt=$(BUILD_DATE)
.PHONY: build test test-race vet verify clean
build:
	mkdir -p bin
	go build -trimpath -ldflags "$(LDFLAGS)" -o $(BINARY) ./cmd/acctx
test:
	go test ./...
test-race:
	go test -race ./...
vet:
	go vet ./...
verify: vet test test-race build
clean:
	rm -rf bin
