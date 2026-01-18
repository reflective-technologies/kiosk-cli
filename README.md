# kiosk-cli

The app store for Claude Code apps.

## Installation

### Quick install (recommended)

```bash
curl -fsSL https://raw.githubusercontent.com/reflective-technologies/kiosk-cli/main/install.sh | sh
```

This installs to `~/.local/bin`. Set `KIOSK_INSTALL_DIR` to customize:

```bash
curl -fsSL https://raw.githubusercontent.com/reflective-technologies/kiosk-cli/main/install.sh | KIOSK_INSTALL_DIR=/usr/local/bin sh
```

### From source

```bash
go install github.com/reflective-technologies/kiosk-cli@latest
```

## Usage

```bash
kiosk --help
```

## Development

```bash
# Build
go build -o kiosk .

# Run tests
go test ./...
```
