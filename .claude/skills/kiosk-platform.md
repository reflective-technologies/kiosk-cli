# Kiosk App Platform

Kiosk is an App Store platform for Claude Code applications. It enables developers to publish, share, and distribute Claude Code apps through a centralized marketplace.

## How It Works

1. Developers create a `Kiosk.md` file in their repository containing installation instructions
2. Developers publish their app to the platform via the API (name, description, git URL)
3. Users browse apps on kiosk.app and copy installation prompts
4. Claude Code executes the prompt to install the app locally

## API Reference

Base URL: `https://kiosk.app` (or local dev server)

### Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/kiosk` | List all published apps |
| POST | `/api/kiosk` | Publish a new app |
| GET | `/api/kiosk/[appId]` | Get app details by ID |
| PUT | `/api/kiosk/[appId]` | Update an existing app |
| DELETE | `/api/kiosk/[appId]` | Delete an app |
| POST | `/api/kiosk/[appId]/refresh` | Refresh cached Kiosk.md from repository |

### App Schema

```typescript
interface KioskApp {
  id: string;              // URL-friendly slug (auto-generated)
  name: string;            // App display name
  description: string;     // Short description
  gitUrl: string;          // Repository URL
  branch?: string;         // Git branch (defaults to main/master)
  subdirectory?: string;   // Path to Kiosk.md in repo
  screenshot?: string;     // Preview image URL
  instructions?: string;   // Additional setup notes
  kioskMd?: string;        // Cached installation instructions
  createdAt: string;       // ISO timestamp
  updatedAt: string;       // ISO timestamp
}
```

### Create/Update Request

```typescript
interface CreateAppRequest {
  name: string;            // Required
  description: string;     // Required
  gitUrl: string;          // Required - GitHub, GitLab, or Bitbucket URL
  branch?: string;
  subdirectory?: string;
  screenshot?: string;
  instructions?: string;
}
```

### Supported Git Providers

- GitHub: `https://github.com/owner/repo`
- GitLab: `https://gitlab.com/owner/repo`
- Bitbucket: `https://bitbucket.org/owner/repo`

## Kiosk.md File

The `Kiosk.md` file is placed in the root of a repository (or subdirectory) and contains instructions that Claude Code follows to install the app. This is the core of app distribution - it tells Claude Code exactly how to set up the application for a user.
