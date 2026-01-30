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
grep -r "process.env\|os.environ\|env\(" --include="*.ts" --include="*.js" --include="*.py"

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
