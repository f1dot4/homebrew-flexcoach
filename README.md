# FlexCLI

FlexCLI is a Go-based command-line interface for the FlexCoach AI fitness platform. It allows users and developers to interact with the backend API to manage profiles, training plans, goals, and system status.

## Features

- **Profile Management**: View and update user profile, body vitals, and preferences.
- **Training Plans**: Retrieve, update, or skip daily training plans.
- **Goal & Constraint Tracking**: Manage structured training goals and user constraints.
- **Activity Management**: Download individual activities or perform bulk exports (FIT, GPX, TCX, CSV, KML).
- **Device Connections**: Monitor and sync Garmin, Withings, and Rouvy data.
- **System Status**: Check backend health and service connectivity.

## Installation

### Homebrew (macOS)

```bash
brew install f1dot4/flexcli/flexcli
```

### apt (Debian / Ubuntu)

```bash
# Add the GPG key
curl -fsSL https://f1dot4.github.io/homebrew-flexcli/flexcli.gpg \
  | sudo gpg --dearmor -o /etc/apt/keyrings/flexcli.gpg

# Add the repository
echo "deb [arch=amd64 signed-by=/etc/apt/keyrings/flexcli.gpg] https://f1dot4.github.io/homebrew-flexcli stable main" \
  | sudo tee /etc/apt/sources.list.d/flexcli.list

# Install
sudo apt update && sudo apt install flexcli
```

### Docker

```bash
docker run --rm \
  -e FLEXCLI_SERVER=https://flexcoach.example.com \
  -e FLEXCLI_API_KEY=YOUR_API_KEY \
  ghcr.io/f1dot4/flexcli:latest \
  flexcli profile data activity list
```

#### Dawarich integration

Run as a sidecar to automatically import activities into [Dawarich](https://dawarich.app) every 2 hours. Files are placed in the Dawarich watched directory under your user's email subdirectory:

```yaml
flexcli:
  image: ghcr.io/f1dot4/flexcli:latest
  volumes:
    - /opt/dawarich/watched:/watched
  environment:
    FLEXCLI_API_KEY: ${FLEXCLI_API_KEY}
    FLEXCLI_SERVER: https://flexcoach.example.com
    FLEXCLI_FORMAT: gpx
    FLEXCLI_DAWARICH_USER: your@email.com
    FLEXCLI_IMPORT_SCHEDULE: "0 */2 * * *"
  restart: always
```

#### Bulk historical import

To import all activities for a given year into Dawarich:

```bash
docker exec -e YEAR=2025 flexcli sh -c '
  flexcli \
    --server "$FLEXCLI_SERVER" \
    --key "$FLEXCLI_API_KEY" \
    profile data activity download-bulk \
    --format gpx \
    --year "$YEAR" \
    --output /tmp/$YEAR-all.zip \
  && unzip -o /tmp/$YEAR-all.zip -d /watched/your@email.com/ \
  && rm /tmp/$YEAR-all.zip
'
```

Then trigger the Dawarich watcher to pick up the files:

```bash
docker exec dawarich_app bin/rails runner 'Imports::Watcher.new.call'
```

### From Source

Ensure you have Go 1.22+ installed.

```bash
cd flexcli
make build
# Binaries will be in bin/flexcli-mac and bin/flexcli-linux
```

## Quick Start

Connect directly to a server without saving a configuration:

```bash
flexcli --server https://flexcoach.example.com --key YOUR_API_KEY profile body vitals get
```

Or configure a persistent context:

```bash
flexcli config --server https://flexcoach.example.com --key YOUR_API_KEY --name production
flexcli context use production
flexcli status

# Download your latest activity as GPX
flexcli profile data activity download --format gpx

# Bulk download activities for a specific month
flexcli profile data activity download-bulk --year 2026 --month 4
```

## Global Flags

- `--server`: Override the FlexCoach server URL.
- `--key`: Override the API Key (can also use `FLEXCLI_API_KEY` environment variable).
- `--context`: Use a specific context from the configuration file.
- `--config`: Specify a custom configuration file (default: `~/.flexcli.json`).

## Development

### Building
```bash
make build
```

### One-time setup (apt signing key)

Before your first release, generate a GPG key for signing the apt repository:

```bash
make setup-apt
# Note the long key ID printed (e.g. ABCD1234EFGH5678)
```

### Releasing

```bash
GPG_KEY_ID=<your-key-id> make release v=0.1.7
```

Requires [`nfpm`](https://nfpm.goreleaser.com/) (`brew install goreleaser/tap/nfpm`) and `gpg`.

**What this does:**
1. Runs all Go tests.
2. Updates version strings in `main.go` and the Homebrew formula.
3. Cross-compiles binaries for macOS and Linux.
4. Re-generates the full CLI reference documentation.
5. Builds a signed `.deb` package via `nfpm`.
6. Commits, tags, and pushes to GitHub.
7. Fetches the real SHA256 of the GitHub-generated tarball and updates the formula.
8. Publishes the `.deb` to the `gh-pages` apt repository (signed with your GPG key).

Once complete, both `brew upgrade flexcli` and `apt upgrade flexcli` work immediately.
