BINARY  := sage
VERSION := 0.1.0
SAGE_FREE_LLM_API_KEY ?=
SAGE_FREE_LLM_API_BASE_URL ?=
LDFLAGS := -ldflags="-s -w -X main.Version=$(VERSION) -X main.BuiltinFreeLLMAPIKey=$(SAGE_FREE_LLM_API_KEY) -X main.BuiltinFreeLLMAPIBaseURL=$(SAGE_FREE_LLM_API_BASE_URL)"

# ── development ───────────────────────────────────────────────────────────────

build:
	@go build $(LDFLAGS) -o $(BINARY) .

run:
	go run . $(ARGS)

install:
	@go install $(LDFLAGS) .

tidy:
	go mod tidy

# ── cross-platform release builds ─────────────────────────────────────────────

release: clean
	mkdir -p dist
	@GOOS=linux   GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY)-linux-amd64   .
	@GOOS=linux   GOARCH=arm64 go build $(LDFLAGS) -o dist/$(BINARY)-linux-arm64   .
	@GOOS=darwin  GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY)-darwin-amd64  .
	@GOOS=darwin  GOARCH=arm64 go build $(LDFLAGS) -o dist/$(BINARY)-darwin-arm64  .
	@GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY)-windows-amd64.exe .
	@echo "✅ Binaries in ./dist/"

clean:
	rm -f $(BINARY)
	rm -rf dist/

.PHONY: build run install tidy release clean
