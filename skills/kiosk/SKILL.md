# kiosk skill

## Purpose

Use this skill to guide users through creating, publishing, and updating Kiosk apps with kiosk-cli.
It includes the full prompt context used by `kiosk init` and `kiosk publish`, plus command-by-command guidance.

## Kiosk registry overview

Kiosk is the app store for Claude Code apps. Publishing creates a public listing that points to a GitHub repo.

Important: **The project must be pushed to a public GitHub repository before publishing to Kiosk.**

## When to use this skill

- Creating or updating `KIOSK.md`
- Publishing a new app or updating an existing listing
- Updating metadata for an already published app (especially with `kiosk api update`)

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
2. **Ensure KIOSK.md exists and is current.** Use the init prompt below if you need to generate or update it.
3. **Generate metadata.** Use `kiosk api create` for new apps or `kiosk api update` for existing apps.
4. **Refresh KIOSK.md (optional).** Use `kiosk api refresh` after updates.

## Init prompt context (used by `kiosk init`)

# Create a KIOSK.md File

## What is KIOSK.md?

KIOSK.md is **complete installation documentation for an autonomous agent**. When published, a *different* agent with **zero context** about your project will use this file to install and run your app on a new machine.

Think of it as writing instructions for a highly capable but completely uninformed assistant. They can read code, run commands, and figure things out--but they don't know:
- What external services your app depends on
- What accounts need to be created
- What implicit setup steps you've done on your machine
- Why certain config values are set the way they are
- The "tribal knowledge" you've accumulated while building this

**Your job**: Capture everything that cannot be derived from the code alone, or point to exactly where in the code it can be found.

## Principles

1. **Zero context assumption**: The installing agent starts with nothing but the cloned repo and KIOSK.md. No README context, no conversation history, no understanding of your intentions.

2. **Minimal user intervention**: Installation should require the user to provide only what truly requires human action (creating accounts, obtaining API keys). Everything else should be automated or have sensible defaults.

3. **End-to-end verification**: Include steps to verify the app is fully running, not just installed.

4. **Concise but complete**: Be efficient with words, but don't omit critical information. Better to over-document than leave gaps.

5. **Pointers over prose**: Instead of explaining how something works, point to the file where the agent can read it: "See `lib/config.ts` for all configuration options."

## General Instructions

- **Use multiple choice questions**: Whenever you need the user's input, prefer using a tool to present multiple choice options. This greatly reduces the burden on the user.
- **Auto-detect first**: Always try to detect information automatically before asking the user. Only ask for confirmation or clarification when truly needed.
- **Be concise**: Keep your messages short and actionable.

---

## Step 1: Deep Project Analysis

Go beyond surface-level detection. You need to understand everything required to run this app.

### 1.1 Project Structure

- Scan the project directory to determine the language/ecosystem
- Check for README.md to understand what the app does
- Check if KIOSK.md already exists (also `Kiosk.md` or `kiosk.md`). If it exists, ask the user if they want to update it or start fresh.

### 1.2 Dependency Layers

**Runtime dependencies** (things that must be running):
- Databases (Postgres, MySQL, MongoDB, SQLite file location)
- Caches (Redis, Memcached)
- Message queues (RabbitMQ, Kafka)
- External services (S3, email providers, etc.)

**System dependencies** (things that must be installed):
- Language runtime and version
- System packages (ffmpeg, imagemagick, etc.)
- CLI tools (aws-cli, docker, etc.)

**Package dependencies**:
- Read package.json/requirements.txt/etc.
- Note any packages that require additional setup (native modules, etc.)

### 1.3 Configuration Archaeology

Dig through the code to find ALL configuration:

```bash
# Find environment variable usage
grep -r "process.env\\|os.environ\\|env\\(" --include="*.ts" --include="*.js" --include="*.py"

# Find config files
find . -name "*.config.*" -o -name ".env*" -o -name "config.*"
```

For each config value found:
- Is there a default? What is it?
- Is it required or optional?
- What is it used for?
- Where should the user get this value?

### 1.4 External Service Dependencies

Search for integrations that require accounts or setup:
- API clients (Stripe, Twilio, OpenAI, etc.)
- OAuth providers (Google, GitHub, etc.)
- Cloud services (AWS, GCP, Vercel, etc.)

For each, document:
- What account/credentials are needed
- Where to create them
- What callback URLs or settings to configure on the provider's side

### 1.5 Database & State

- Are there migrations to run?
- Is seed data required?
- What's the expected initial state?
- Are there any manual database setup steps?

### 1.6 Startup Sequence

Trace the actual startup:
- What command starts the app?
- What order do things need to happen? (migrations before server, etc.)
- Are there background workers or multiple processes?
- What ports are used? Any potential conflicts?

### 1.7 Hidden Assumptions

Look for things the creator might take for granted:
- Hardcoded localhost URLs that assume certain services
- File paths that assume certain directory structures
- Permissions or ownership requirements
- Network configurations

---

## Step 2: Synthesize Findings

Review everything you discovered in Step 1. You should now have a complete picture of:
- All dependencies (runtime, system, package)
- All configuration values and where they come from
- All external services and what credentials they need
- The startup sequence and verification steps

**Only ask the user if something cannot be determined from the code.** For example:
- A config value with no documentation or obvious purpose
- An API integration where you can't determine what account type is needed
- Ambiguous setup steps with multiple possible interpretations

If you do need to ask, use multiple choice questions and be specific about what's unclear. But the goal is to derive everything autonomously.

---

## Step 3: Write KIOSK.md

Structure the file for maximum clarity. Use this template as a guide:

```markdown
# Installing [App Name]

> **For the installing agent**: This file contains everything needed to set up this app. Follow steps in order. Use multiple choice questions for user input. Verify each major step before proceeding.

## Quick Summary
[One paragraph: what this app does and what you'll have running at the end]

## Prerequisites

### System Requirements
- [OS requirements if any]
- [Language runtime + version]
- [System packages: e.g., "ffmpeg for video processing"]

### External Services
- [Service]: [What it's for] - [How to get credentials]
  - Create account at [URL]
  - [Specific setup steps like "Create OAuth app with callback URL http://localhost:3000/auth/callback"]

### Required Credentials
| Variable | Description | How to obtain |
|----------|-------------|---------------|
| DATABASE_URL | PostgreSQL connection string | Create a Postgres database |
| OPENAI_API_KEY | For AI features | https://platform.openai.com/api-keys |

## Installation

### 1. Install Dependencies
```bash
[exact commands]
```

### 2. Database Setup
```bash
[migration commands]
[seed commands if needed]
```

### 3. Configuration
Create `.env` file:
```
KEY=value  # description
```

Hardcoded values to update:
- `path/to/file.ts:42` - Change `SOME_VALUE` to your value

### 4. Start the App
```bash
[start command]
```

## Verification

After installation, verify everything works:
- [ ] App starts without errors
- [ ] Can access http://localhost:PORT
- [ ] [Key functionality works]

## Troubleshooting

### [Common Issue 1]
[Solution]

## File Reference

Key files for the installing agent to understand:
- `lib/config.ts` - All configuration options
- `lib/db/schema.ts` - Database structure
- `.env.example` - Environment variable template
```

**Important context**: The installing agent will already be running from within the cloned repository. DO NOT include steps about cloning the repo. Focus on installation, configuration, and setup.

---

## Step 4: Validate Completeness

Before finishing, verify your KIOSK.md answers these questions:
- [ ] Can someone with just this file and the code get the app running?
- [ ] Are all external services documented with setup instructions?
- [ ] Are all environment variables listed with how to obtain them?
- [ ] Is the verification section specific enough to confirm success?
- [ ] Would YOU be able to install this from scratch using only this file?

---

## Step 5: Write and Confirm

Write KIOSK.md to the repository root.

Let the user know the file has been created. Remind them:
- Review the file to make sure it's accurate
- When ready, commit and push it to their repository
- To publish to Kiosk, use the publish prompt or run `kiosk publish`

## Publish prompt context (used by `kiosk publish`)

# Publish to Kiosk

## What is Kiosk?

Kiosk is the App Store for Claude Code. Publishing your app to Kiosk creates a permalink page where anyone can discover and install your app with a single command.

You are helping the user publish their app to Kiosk.

## General Instructions

- **Use multiple choice questions**: Whenever you need the user's input or opinion, prefer using a tool to present multiple choice options (if such a tool is available). This greatly reduces the burden on the user compared to asking open-ended questions.
- **Auto-detect first**: Always try to detect information automatically before asking the user. Only ask for confirmation or clarification when needed.
- **Be concise**: Keep your messages short and actionable. The user wants to publish quickly.

## Prerequisites

1. The app must be in a Git repository with a remote (GitHub, GitLab, etc.)
2. The code should be pushed to the remote
3. A `KIOSK.md` file is recommended (but not required) for best installation experience

---

## Step 1: Gather Project Information

### 1.1 Check Git Status

- **Remote URL**: `git remote get-url origin` -- this is the repository URL for Kiosk
- **Current branch**: `git branch --show-current`
- **Uncommitted changes**: `git status --porcelain` -- warn the user if they have unpushed changes

If there's no git remote, help the user set one up (offer to use `gh repo create` if GitHub CLI is available).

### 1.2 Detect Project Info

Look for project name and description in:
- `package.json`, `pyproject.toml`, `Cargo.toml`, `go.mod`, etc.
- `README.md`

### 1.3 Check for KIOSK.md

Check if KIOSK.md exists (also check `Kiosk.md`, `kiosk.md` for backwards compatibility).
If it doesn't exist, let the user know their app will still work but won't have detailed installation instructions. Offer to help create one using the init prompt.

### 1.4 Get Kiosk URL

Get the base URL for the Kiosk instance:
```bash
kiosk config get apiUrl
```

Use this URL (referred to as `<kiosk-url>` below) when constructing links to the app's page.

## Step 2: Set App Metadata (Iterative)

This step configures the public-facing information for your app on Kiosk--the name, description, and other metadata that users will see when browsing.

### Writing Great Metadata

**Description** -- Write like an App Store listing:
- 1-2 sentences, punchy and scannable
- Lead with the value proposition, not technical details
- Focus on benefits over features
- Use active voice, avoid jargon

**How It Works** -- A mini-README explaining what the app does:
- Start with an intro paragraph summarizing the app (no heading needed - parent section is already titled "How It Works")
- Use `###` for sub-headings to maintain proper heading hierarchy
- Structure with sections like: **Workflow** (numbered steps), **Features** (bullet list), **Requirements** (what the user needs)
- Be thorough -- this helps users understand the app before installing

### 2.1 Check if App Exists

Check if the app already exists on Kiosk:
```bash
kiosk api get <expected-app-id>
```

The app ID is typically a slugified version of the app name.

### 2.2 Create or Update

If the app doesn't exist, create it:
```bash
kiosk api create -f - <<EOF
{
  "name": "<app-name>",
  "description": "<app-description>",
  "gitUrl": "<github-url>",
  "branch": "<branch-if-not-main>",
  "subdirectory": "<path-if-not-root>",
  "howItWorks": "<numbered-steps>"
}
EOF
```

If the app already exists, update it:
```bash
kiosk api update <app-id> -f - <<EOF
{
  "name": "<app-name>",
  "description": "<app-description>",
  "howItWorks": "<numbered-steps>"
}
EOF
```

Only include optional fields (`branch`, `subdirectory`) when they apply.

### 2.3 Generate Description and "How It Works"

Analyze the project to understand what it does and craft compelling metadata:

1. **Read the README and KIOSK.md** to understand the app's purpose
2. **Identify the core value** -- what problem does it solve for users?
3. **Write the description** -- 1-2 sentences, benefit-focused, exciting
4. **Write howItWorks** -- a thorough markdown explanation of what the app does, its features, and requirements

### 2.4 Get Feedback

Present the app name, description, and "How It Works" steps. Use the question tool to ask:

**"How does this look for your app's public listing?"**
- Options: "Looks good!" / "I'd like to tweak it"

If the user wants changes, update the metadata and repeat until satisfied.

## Step 3: Celebrate!

Once published, present the result using the `<kiosk-url>` you retrieved earlier:

---

**Your app is live on Kiosk!**

**[App Name]** is now available for anyone to install with Claude Code.

**Your app's page:** `<kiosk-url>/kiosk/[app-id]`

**Share it:**
> I just published [App Name] to Kiosk! Install it in Claude Code: `<kiosk-url>/kiosk/[app-id]`

---

**Reminder:** If you have uncommitted or unpushed changes, make sure to push them so other users can access your latest code.

## CLI Reference

Use `kiosk api` for all API interactions. Run `kiosk api -h` to see available commands.

- `kiosk api create -f <file>` -- Publish a new app
- `kiosk api update <app-id> -f <file>` -- Update an existing app
- `kiosk api get <app-id>` -- Get app details
- `kiosk api list` -- List all published apps
- `kiosk api delete <app-id>` -- Delete an app

### JSON Fields

**Required:**
- `name`: Display name for the app
- `description`: What the app does
- `gitUrl`: HTTPS URL to the Git repository

**Optional:**
- `branch`: Git branch (only if not the default branch)
- `subdirectory`: Path within the repo (only if app is not at root)
- `instructions`: Additional setup instructions
- `howItWorks`: Markdown-formatted explanation of what the app does, features, and requirements
