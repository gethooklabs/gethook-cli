.PHONY: build test lint clean install snapshot

BINARY := gethook
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -s -w -X main.version=$(VERSION)

build:
	go build -ldflags "$(LDFLAGS)" -o $(BINARY) .

test:
	go test ./... -count=1

lint:
	go vet ./...

clean:
	rm -f $(BINARY)

install: build
	mv $(BINARY) /usr/local/bin/$(BINARY)

snapshot:
	goreleaser release --snapshot --clean

# Run against local API (requires GETHOOK_API_BASE env var)
run-local:
	GETHOOK_API_BASE=http://localhost:8080 go run . $(ARGS)
