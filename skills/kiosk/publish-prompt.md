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
