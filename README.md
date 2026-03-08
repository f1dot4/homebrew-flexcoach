# FlexCLI

FlexCLI is the command-line interface for the FlexCoach platform.

## Features
- Context management
- Goal tracking
- Constraint configuration
- Training plan access
- Statistics viewing

## Installation

You can install FlexCLI via Homebrew:

```bash
brew install f1dot4/flexcli/flexcli
```

### Building from Source

To build the `flexcli` binary from the source code:

```bash
# From the root of the flexcli repository
go build -o flexcli ./cmd/flexcli
```

## Development
```bash
go build ./cmd/flexcli
```
