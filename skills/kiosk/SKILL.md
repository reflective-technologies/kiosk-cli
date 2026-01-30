# kiosk skill

## Purpose

Use this skill to guide users through creating, publishing, and updating Kiosk apps with kiosk-cli.
It summarizes the kiosk-cli workflow and points to supplementary prompt files to keep this skill lightweight.

## Kiosk registry overview

Kiosk is the app store for Claude Code apps. Publishing creates a public listing that points to a GitHub repo.

Important: **The project must be pushed to a public GitHub repository before publishing to Kiosk.**

## When to use this skill

- Creating or updating `KIOSK.md`
- Publishing a new app or updating an existing listing
- Updating metadata for an already published app (especially with `kiosk api update`)

## Supplementary files

- `init-prompt.md`: Full prompt context for creating or updating `KIOSK.md` (used by `kiosk init`).
- `publish-prompt.md`: Full prompt context for publishing and metadata updates (used by `kiosk publish`).

When you need the full instructions, open the relevant file and follow it step by step.

## Command reference (what each command does)

### General
- `kiosk` (no args): Open the interactive TUI to browse and run apps.
- `kiosk tui`: Launch the interactive TUI explicitly.

### Workspace projects
- `kiosk new [project-name]`: Create a new workspace project and start requirements gathering.
- `kiosk edit [project-name]`: Resume a workspace project session.

### Run and manage apps
- `kiosk run <app>`: Install (if needed) and run an app.
- `kiosk install <app>`: Alias for `kiosk run`.
- `kiosk ls`: List installed apps (interactive list).
- `kiosk list`: Alias for `kiosk ls`.
- `kiosk rm <app>`: Remove an installed app.

### Auth
- `kiosk login`: Authenticate with GitHub for publishing.
- `kiosk whoami`: Show the current authenticated user.
- `kiosk logout`: Remove stored credentials.

### KIOSK.md creation
- `kiosk init`: Fetches the KIOSK.md creation prompt and runs Claude to generate the file.
- `kiosk api init-prompt`: Print the same prompt to stdout (useful for agents).

### Publish
- `kiosk publish`: Fetches the publish prompt and runs Claude in the current repo.
- `kiosk api publish-prompt`: Print the publish prompt to stdout.
- `kiosk audit`: Run a security audit prompt to scan for secrets before publishing.

### Direct API access (preferred for agent workflows)
- `kiosk api list`: List all published apps.
- `kiosk api get <app-id>`: Get details for an app.
- `kiosk api create -f <file>`: Publish a new app.
- `kiosk api update <app-id> -f <file>`: **Update metadata for a published app. Use this for changes after initial publish.**
- `kiosk api delete <app-id>`: Delete an app.
- `kiosk api refresh <app-id>`: Refresh KIOSK.md from the repo.

### Config and maintenance
- `kiosk config list/get/set`: View or change CLI config values (including the API URL).
- `kiosk update`: Update kiosk-cli to the latest version.
- `kiosk version`: Print version.
- `kiosk completion <shell>`: Generate shell completions.

## Publishing workflow (preferred: use `kiosk api` commands)

1. **Confirm GitHub repo is public and pushed.** Kiosk requires a publicly accessible GitHub repo.
2. **Ensure KIOSK.md exists and is current.** See `init-prompt.md` for the full workflow.
3. **Generate metadata.** Use `kiosk api create` for new apps or `kiosk api update` for existing apps.
4. **Refresh KIOSK.md (optional).** Use `kiosk api refresh` after updates.

For the detailed publish flow, follow `publish-prompt.md`.
