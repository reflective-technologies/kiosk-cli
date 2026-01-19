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

### Browse and run apps

```bash
# Run an app (installs if needed)
kiosk run <app-name>

# Run with sandbox mode (no file writes outside project)
kiosk run --sandbox <app-name>

# List installed apps
kiosk ls

# Remove an installed app
kiosk rm <app-name>
```

### Authentication

```bash
# Log in with GitHub (required for publishing)
kiosk login

# Show current authenticated user
kiosk whoami

# Log out
kiosk logout
```

### Publish your own app

```bash
# Initialize a new kiosk app project (creates Kiosk.md)
kiosk new

# Publish the current repo to kiosk.app (requires login)
kiosk publish
```

### Configuration

```bash
# List all config values
kiosk config list

# Get a specific config value
kiosk config get <key>

# Set a config value
kiosk config set <key> <value>
```

### Direct API access

For scripting and automation:

```bash
# List all published apps (JSON output)
kiosk api list

# Get app details
kiosk api get <app-id>

# Publish a new app
kiosk api create -f app.json

# Update an existing app
kiosk api update <app-id> -f app.json

# Delete an app
kiosk api delete <app-id>

# Refresh app's Kiosk.md from repository
kiosk api refresh <app-id>
```

### Other commands

```bash
# Update kiosk to the latest version
kiosk update

# Print version
kiosk version

# Generate shell completions
kiosk completion <bash|zsh|fish|powershell>
```

## Development

```bash
# Build
go build -o kiosk .

# Run tests
go test ./...
```
