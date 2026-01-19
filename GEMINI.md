# kiosk-cli

The app store for Claude Code apps.

## Project Overview

`kiosk-cli` is a command-line interface tool designed to interact with the Kiosk app platform. It allows users to manage and interact with Claude Code applications.

**Tech Stack:**
*   **Language:** Go (1.24.6)
*   **CLI Framework:** Cobra (`github.com/spf13/cobra`)

## Architecture

The project follows a standard Go CLI architecture:

*   **`main.go`**: The entry point of the application. It calls `cmd.Execute()`.
*   **`cmd/`**: Contains the command definitions.
    *   `root.go`: Defines the root command (`kiosk`) and the `Execute` function.
    *   Subcommands (e.g., `ls.go`, `new.go`) are defined in separate files and register themselves via `init()`.
*   **`internal/`**: Contains internal application logic.
    *   `api/`: Client for interacting with the Kiosk backend.
    *   `appindex/`: Logic for managing the application index.
    *   `config/`: Configuration handling.

## Building and Running

### Prerequisites

*   Go 1.24+

### Build

To build the binary:

```bash
go build -o kiosk .
```

To build with version injection:

```bash
go build -ldflags "-X github.com/reflective-technologies/kiosk-cli/cmd.Version=1.0.0" -o kiosk .
```

### Run

You can run the built binary:

```bash
./kiosk --help
```

### Test

To run all tests:

```bash
go test ./...
```

To run a specific test:

```bash
go test -run TestName ./path/to/package
```

## Development Conventions

### Adding New Commands

To add a new command, create a new file in the `cmd/` directory (e.g., `cmd/mycommand.go`) and follow the Cobra pattern:

```go
package cmd

import "github.com/spf13/cobra"

var myCmd = &cobra.Command{
    Use:   "mycommand",
    Short: "Description of mycommand",
    Run: func(cmd *cobra.Command, args []string) {
        // Implementation
    },
}

func init() {
    rootCmd.AddCommand(myCmd)
}
```

### Issue Tracking (Beads)

This project uses **bd** (beads) for issue tracking.

*   `bd ready`: Find available work.
*   `bd show <id>`: View issue details.
*   `bd update <id> --status in_progress`: Claim work.
*   `bd close <id>`: Complete work.
*   `bd sync`: Sync with git.

Refer to `AGENTS.md` for the complete workflow, including the mandatory "Landing the Plane" steps (pushing code is required).
