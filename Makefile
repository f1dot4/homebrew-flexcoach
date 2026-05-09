VERSION ?= $(shell git describe --tags --always --dirty)
LDFLAGS := -X main.Version=$(VERSION)
GPG_KEY_ID ?= 3005BD255C306C4E
DOCKER_IMAGE ?= ghcr.io/f1dot4/flexcli

.PHONY: build clean release deb publish-apt docker-build docker-push docs test setup-apt

test:
	go test ./...

build: test
	mkdir -p bin
	GOOS=darwin GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o bin/flexcli-mac ./cmd/flexcli/
	GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o bin/flexcli-linux ./cmd/flexcli/
	cp bin/flexcli-mac flexcli

deb:
	@if [ -z "$(v)" ]; then echo "Usage: make deb v=0.1.x"; exit 1; fi
	@which nfpm > /dev/null || (echo "nfpm not found — install with: brew install goreleaser/tap/nfpm"; exit 1)
	VERSION=$(v) nfpm package --config nfpm.yaml --packager deb --target bin/

publish-apt:
	@if [ -z "$(v)" ]; then echo "Usage: make publish-apt v=0.1.x"; exit 1; fi
	@GPG_KEY_ID=$(GPG_KEY_ID) bash scripts/publish-apt.sh bin/flexcli_$(v)_amd64.deb $(v)

docker-build:
	@if [ -z "$(v)" ]; then echo "Usage: make docker-build v=0.1.x"; exit 1; fi
	docker build --platform linux/amd64 -t $(DOCKER_IMAGE):$(v) -t $(DOCKER_IMAGE):latest .

docker-push:
	@if [ -z "$(v)" ]; then echo "Usage: make docker-push v=0.1.x"; exit 1; fi
	docker push $(DOCKER_IMAGE):$(v)
	docker push $(DOCKER_IMAGE):latest

docs: build
	python3 ../scripts/generate_cli_docs.py

clean:
	rm -rf bin/
	rm -f flexcli

# One-time setup: generate a GPG key for signing the apt repo.
# After running, note the key ID printed and set GPG_KEY_ID=<id> when releasing.
setup-apt:
	@echo "Generating a GPG key for apt repo signing..."
	@printf '%%no-protection\nKey-Type: RSA\nKey-Length: 4096\nName-Real: FlexCLI Releases\nName-Email: lukas.leidinger@willhaben.at\nExpire-Date: 0\n%%commit\n' | gpg --batch --gen-key -
	@echo ""
	@echo "Key generated. Your key IDs:"
	@gpg --list-secret-keys --keyid-format LONG lukas.leidinger@willhaben.at
	@echo ""
	@echo "Set the key ID when releasing: GPG_KEY_ID=<LONG-ID> make release v=x.y.z"

release: test docs
	@if [ -z "$(v)" ]; then echo "Usage: make release v=0.1.x"; exit 1; fi
	@if [ -z "$(GPG_KEY_ID)" ]; then echo "Error: GPG_KEY_ID is required.\nRun: GPG_KEY_ID=<your-key-id> make release v=$(v)\nTo generate a key run: make setup-apt"; exit 1; fi
	@which nfpm > /dev/null || (echo "nfpm not found — install with: brew install goreleaser/tap/nfpm"; exit 1)
	@echo "Releasing v$(v)..."
	@# Update version in main.go
	@sed -i '' 's/Version     = ".*"/Version     = "$(v)"/' cmd/flexcli/main.go
	@# Update version in Formula (placeholder SHA for now)
	@sed -i '' 's|url ".*tags/v.*.tar.gz"|url "https://github.com/f1dot4/homebrew-flexcli/archive/refs/tags/v$(v).tar.gz"|' Formula/flexcli.rb
	@sed -i '' 's/sha256 ".*"/sha256 "PLACEHOLDER"/' Formula/flexcli.rb
	@# Rebuild with explicit version so LDFLAGS reflect the release tag, not git-describe
	@$(MAKE) build VERSION=v$(v)
	@# Build .deb package
	@VERSION=$(v) nfpm package --config nfpm.yaml --packager deb --target bin/
	@git add -f cmd/flexcli/main.go Formula/flexcli.rb bin/ docs/CLI_REFERENCE.md
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
	@# Publish apt repo to gh-pages
	@GPG_KEY_ID=$(GPG_KEY_ID) bash scripts/publish-apt.sh bin/flexcli_$(v)_amd64.deb $(v)
	@# Build and push Docker image
	@$(MAKE) docker-build v=$(v)
	@$(MAKE) docker-push v=$(v)
	@echo "Release v$(v) complete. brew upgrade flexcli, apt upgrade flexcli, and docker pull will all work."
