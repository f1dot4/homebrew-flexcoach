VERSION ?= $(shell git describe --tags --always --dirty)
LDFLAGS := -X main.Version=$(VERSION)

.PHONY: build clean release docs test

test:
	go test ./...

build: test
	mkdir -p bin
	GOOS=darwin GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o bin/flexcli-mac ./cmd/flexcli/
	GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o bin/flexcli-linux ./cmd/flexcli/
	cp bin/flexcli-mac flexcli

docs: build
	python3 ../scripts/generate_cli_docs.py

clean:
	rm -rf bin/
	rm -f flexcli

release: test docs
	@if [ -z "$(v)" ]; then echo "Usage: make release v=0.1.x"; exit 1; fi
	@echo "Releasing v$(v)..."
	@# Update version in main.go
	@sed -i '' 's/Version     = ".*"/Version     = "$(v)"/' cmd/flexcli/main.go
	@# Update version in Formula (placeholder SHA for now)
	@sed -i '' 's|url ".*tags/v.*.tar.gz"|url "https://github.com/f1dot4/homebrew-flexcli/archive/refs/tags/v$(v).tar.gz"|' Formula/flexcli.rb
	@sed -i '' 's/sha256 ".*"/sha256 "PLACEHOLDER"/' Formula/flexcli.rb
	@# Rebuild with explicit version so LDFLAGS reflect the release tag, not git-describe
	@$(MAKE) build VERSION=v$(v)
	@git add -f cmd/flexcli/main.go Formula/flexcli.rb bin/
	@git commit -m "chore: release v$(v)"
	@git tag v$(v)
	@echo "Pushing tag to GitHub to generate archive..."
	@git push origin main
	@git push origin v$(v)
	@echo "Waiting for GitHub to generate the archive..."
	@sleep 5
	@echo "Downloading GitHub-generated tarball to get real SHA256..."
	@curl -sL https://github.com/f1dot4/homebrew-flexcli/archive/refs/tags/v$(v).tar.gz \
		-o /tmp/flexcli-v$(v)-github.tar.gz
	@SHA=$$(shasum -a 256 /tmp/flexcli-v$(v)-github.tar.gz | cut -d' ' -f1); \
	echo "Real SHA256: $$SHA"; \
	sed -i '' "s/sha256 \"PLACEHOLDER\"/sha256 \"$$SHA\"/" Formula/flexcli.rb
	@rm /tmp/flexcli-v$(v)-github.tar.gz
	@git add Formula/flexcli.rb
	@git commit -m "fix: correct sha256 for v$(v)"
	@git push origin main
	@echo "Release v$(v) complete. brew upgrade flexcli will work immediately."
