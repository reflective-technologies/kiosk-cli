# Kiosk App Platform

Kiosk is an App Store platform for Claude Code applications. It enables developers to publish, share, and distribute Claude Code apps through a centralized marketplace.

## How It Works

### Publishing Workflow (Two-Step Process)

1. **Init Phase**: Run `/api/prompts/init` to create a `KIOSK.md` file in your repository
   - Analyzes project structure and dependencies
   - Identifies configuration requirements
   - Generates installation instructions
2. **Publish Phase**: Run `/api/prompts/publish` to publish to the registry
   - Gathers project info from Git
   - Verifies KIOSK.md exists
   - Creates or updates app entry via API

### Installation Workflow

1. Users browse apps on kiosk.app
2. Users copy the installation prompt (plain text)
3. Claude Code executes the prompt:
   - Clones the repository
   - Follows KIOSK.md instructions
   - Performs cleanup

## API Reference

Base URL: `https://kiosk.app`

### Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/kiosk` | List all published apps |
| POST | `/api/kiosk` | Publish a new app |
| GET | `/api/kiosk/[appId]` | Get app details by ID |
| PUT | `/api/kiosk/[appId]` | Update an existing app |
| DELETE | `/api/kiosk/[appId]` | Delete an app |
| POST | `/api/kiosk/[appId]/refresh` | Refresh cached KIOSK.md from repository |
| GET | `/api/kiosk/[appId]/install` | Get installation prompt (plain text) |
| GET | `/api/prompts/init` | Get KIOSK.md creation prompt |
| GET | `/api/prompts/publish` | Get publishing prompt |

### Response Formats

- **JSON**: All `/api/kiosk` endpoints return JSON
- **Plain Text**: Prompt endpoints (`/install`, `/prompts/*`) return `text/plain; charset=utf-8`

### App Schema

```typescript
interface KioskApp {
  id: string;              // URL-friendly slug (auto-generated from name)
  name: string;            // App display name
  description: string;     // Short description
  gitUrl: string;          // Repository URL
  branch?: string;         // Git branch (defaults to main/master)
  subdirectory?: string;   // Path within repo (for monorepos)
  screenshot?: string;     // Preview image URL
  instructions?: string;   // Additional setup notes
  kioskMd?: string;        // Cached installation instructions from repo
  createdAt: string;       // ISO timestamp
  updatedAt: string;       // ISO timestamp
}
```

### Create/Update Request

```typescript
interface CreateKioskRequest {
  name: string;            // Required
  description: string;     // Required
  gitUrl: string;          // Required - must be valid Git URL
  branch?: string;
  subdirectory?: string;
  screenshot?: string;
  instructions?: string;
}
```

### Git URL Validation

The API validates Git URLs using regex and supports HTTPS and SSH formats:

- GitHub: `https://github.com/owner/repo` or `git@github.com:owner/repo`
- GitLab: `https://gitlab.com/owner/repo` or `git@gitlab.com:owner/repo`
- Bitbucket: `https://bitbucket.org/owner/repo` or `git@bitbucket.org:owner/repo`

### Error Responses

| Status | Description |
|--------|-------------|
| 400 | Validation error (missing fields, invalid Git URL) |
| 404 | App not found |
| 500 | Server error |

All errors return: `{ "error": "message" }`

## KIOSK.md File

The `KIOSK.md` file is placed in the root of a repository (or subdirectory) and contains instructions that Claude Code follows to install the app. This is the core of app distribution.

### File Location

The API searches for these filenames (in order): `KIOSK.md`, `Kiosk.md`, `kiosk.md`

### Branch Resolution

If no branch is specified, the API tries: specified branch → `main` → `master`

### Raw File URLs

The API constructs raw file URLs based on provider:
- **GitHub**: `https://raw.githubusercontent.com/{owner}/{repo}/{branch}/{path}`
- **GitLab**: `https://gitlab.com/{owner}/{repo}/-/raw/{branch}/{path}`
- **Bitbucket**: `https://bitbucket.org/{owner}/{repo}/raw/{branch}/{path}`
