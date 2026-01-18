# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build Commands

```bash
# Build the binary
go build -o kiosk .

# Build with version injection
go build -ldflags "-X github.com/reflective-technologies/kiosk-cli/cmd.Version=1.0.0" -o kiosk .

# Run tests
go test ./...

# Run a single test
go test -run TestName ./path/to/package
```

## Architecture

This is a CLI application using the Cobra framework. The binary is named `kiosk`.

- `main.go` - Entry point, calls `cmd.Execute()`
- `cmd/` - All CLI commands live here
  - `root.go` - Root command and `Execute()` function
  - Each subcommand is a separate file that registers itself via `init()`

### Adding New Commands

Create a new file in `cmd/` following this pattern:

```go
var myCmd = &cobra.Command{
    Use:   "mycommand",
    Short: "Description",
    Run: func(cmd *cobra.Command, args []string) {
        // implementation
    },
}

func init() {
    rootCmd.AddCommand(myCmd)
}
```
