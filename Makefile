VERSION ?= 0.1.0
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE    ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
BINARY  := rune

LDFLAGS := -s -w \
	-X github.com/rune-context/rune/internal/version.Version=$(VERSION) \
	-X github.com/rune-context/rune/internal/version.Commit=$(COMMIT) \
	-X github.com/rune-context/rune/internal/version.Date=$(DATE)

PLATFORMS := linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64 windows/arm64

.PHONY: build clean test release install

build:
	go build -ldflags "$(LDFLAGS)" -o $(BINARY) .

install: build
	install -d $(HOME)/.local/bin
	install -m 755 $(BINARY) $(HOME)/.local/bin/$(BINARY)
	@echo "✓ Installed to $(HOME)/.local/bin/$(BINARY)"
	@echo "  Make sure $(HOME)/.local/bin is in your PATH"

test:
	go test ./...

clean:
	rm -f $(BINARY)
	rm -rf release/

release: clean
	@mkdir -p release
	@for platform in $(PLATFORMS); do \
		os=$${platform%/*}; \
		arch=$${platform#*/}; \
		ext=""; \
		if [ "$$os" = "windows" ]; then ext=".exe"; fi; \
		echo "Building $$os/$$arch..."; \
		GOOS=$$os GOARCH=$$arch go build -ldflags "$(LDFLAGS)" -o release/$(BINARY)$$ext . ; \
		if [ "$$os" = "windows" ]; then \
			cd release && zip $(BINARY)-$$os-$$arch.zip $(BINARY)$$ext && rm $(BINARY)$$ext && cd ..; \
		else \
			cd release && tar -czf $(BINARY)-$$os-$$arch.tar.gz $(BINARY)$$ext && rm $(BINARY)$$ext && cd ..; \
		fi; \
	done
	@echo "✓ Release artifacts in release/"
	@ls -la release/
