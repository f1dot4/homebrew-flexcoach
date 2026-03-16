VERSION ?= $(shell git describe --tags --always --dirty)
LDFLAGS := -X main.Version=$(VERSION)

.PHONY: build clean release

build:
	mkdir -p bin
	GOOS=darwin GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o bin/flexcli-mac ./cmd/flexcli/
	GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o bin/flexcli-linux ./cmd/flexcli/

clean:
	rm -rf bin/

release:
	@if [ -z "$(v)" ]; then echo "Usage: make release v=0.1.x"; exit 1; fi
	@echo "Releasing v$(v)..."
	@# Update version in main.go
	@sed -i '' 's/Version     = ".*"/Version     = "$(v)"/' cmd/flexcli/main.go
	@# Update version in Formula
	@sed -i '' 's|url ".*tags/v.*.tar.gz"|url "https://github.com/f1dot4/homebrew-flexcli/archive/refs/tags/v$(v).tar.gz"|' Formula/flexcli.rb
	@# Rebuild to ensure binary matches
	@$(MAKE) build
	@# Create temporary tarball to calculate SHA256 (local approximation for now)
	@git archive --prefix=homebrew-flexcli-$(v)/ --format=tar.gz HEAD -o /tmp/flexcli-v$(v).tar.gz
	@SHA=$$(shasum -a 256 /tmp/flexcli-v$(v).tar.gz | cut -d' ' -f1); \
	sed -i '' "s/sha256 \".*\"/sha256 \"$$SHA\"/" Formula/flexcli.rb
	@rm /tmp/flexcli-v$(v).tar.gz
	@git add -f cmd/flexcli/main.go Formula/flexcli.rb bin/
	@git commit -m "chore: release v$(v)"
	@git tag v$(v)
	@echo "Release v$(v) committed and tagged locally."
	@echo "CRITICAL: GitHub-generated archives may have different SHAs than local git archive."
	@echo "Run: git push origin main && git push origin v$(v)"
	@echo "Then verify/update SHA in Formula/flexcli.rb if brew upgrade fails."
