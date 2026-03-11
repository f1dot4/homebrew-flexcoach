# FlexCLI

FlexCLI is a Go-based command-line interface for the FlexCoach AI fitness platform. It allows users and developers to interact with the backend API to manage profiles, training plans, goals, and system status.

## Features

- **Profile Management**: View and update user profile, body vitals, and preferences.
- **Training Plans**: Retrieve, update, or skip daily training plans.
- **Goal & Constraint Tracking**: Manage structured training goals and user constraints.
- **Device Connections**: Monitor and sync Garmin, Withings, and Rouvy data.
- **System Status**: Check backend health and service connectivity.

## Installation

### Homebrew (Recommended)

```bash
brew install f1dot4/flexcli/flexcli
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

### Releasing
Automate version bumping, Homebrew formula update, tagging, and local commits:

```bash
make release v=0.1.4
```
