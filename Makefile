BINARY  := sage
VERSION := 0.1.0
LDFLAGS := -ldflags="-s -w -X main.Version=$(VERSION)"

# ── development ───────────────────────────────────────────────────────────────

build:
	go build $(LDFLAGS) -o $(BINARY) .

run:
	go run . $(ARGS)

install:
	go install $(LDFLAGS) .

tidy:
	go mod tidy

# ── cross-platform release builds ─────────────────────────────────────────────

release: clean
	mkdir -p dist
	GOOS=linux   GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY)-linux-amd64   .
	GOOS=linux   GOARCH=arm64 go build $(LDFLAGS) -o dist/$(BINARY)-linux-arm64   .
	GOOS=darwin  GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY)-darwin-amd64  .
	GOOS=darwin  GOARCH=arm64 go build $(LDFLAGS) -o dist/$(BINARY)-darwin-arm64  .
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY)-windows-amd64.exe .
	@echo "✅ Binaries in ./dist/"

clean:
	rm -f $(BINARY)
	rm -rf dist/

.PHONY: build run install tidy release clean
