# Yukti

The modern terminal interface for Google Apps Script.

## Features

- **Browse** - Navigate projects & files with fuzzy search (Ctrl+P)
- **Edit** - Syntax-aware code viewing with line numbers
- **Sync** - Create, clone, pull, diff, and push Apps Script workspaces from the terminal
- **Deploy** - Create immutable versions and versioned deployments from the terminal

## Installation

### macOS (Apple Silicon)
```bash
curl -L https://github.com/indrasvat/yukti/releases/latest/download/yukti-darwin-arm64 -o yukti
chmod +x yukti
sudo mv yukti /usr/local/bin/
```

### macOS (Intel)
```bash
curl -L https://github.com/indrasvat/yukti/releases/latest/download/yukti-darwin-amd64 -o yukti
chmod +x yukti
sudo mv yukti /usr/local/bin/
```

### Linux
```bash
curl -L https://github.com/indrasvat/yukti/releases/latest/download/yukti-linux-amd64 -o yukti
chmod +x yukti
sudo mv yukti /usr/local/bin/
```

> **Downloaded via browser?** Run `xattr -d com.apple.quarantine yukti` before moving.

## Setup

### Step 1: Enable Required Google APIs

In the [Google Cloud Console](https://console.cloud.google.com/), enable these APIs for your project:

| API | Purpose |
|-----|---------|
| **Apps Script API** | Access project content, deployments, and metrics |
| **Google Drive API** | List your Apps Script projects |

Navigate to **APIs & Services → Library** and enable both APIs.

### Step 2: Create Google OAuth Credentials

You need your own Google OAuth credentials (~5 minutes, one-time setup).

**[→ Follow the Google OAuth Setup Guide](docs/google-oauth-setup.md)**

You'll get a `client_id` and `client_secret` from a downloaded JSON file.

### Step 3: Run Setup Wizard

```bash
yukti init
```

Enter your Client ID and Client Secret when prompted. Tokens are stored locally with restricted file permissions.

### Step 4: Login

```bash
yukti login
```

Your browser opens → Sign in with Google → Authorize Yukti.

### Step 5: Verify & Launch

```bash
yukti status   # Check everything is configured
yukti          # Launch the TUI
```

You should see all green indicators:

```
  ━━ Configuration ━━
  ●  Config       ~/Library/Application Support/yukti/config.json
  ●  Client ID    57632406••••.com
  ●  Secret       Configured

  ━━ Authentication ━━
  ●  Status       Logged in
  ●  Expires in   ████████████████████ 59m
```

## Commands

| Command | Description |
|---------|-------------|
| `yukti` | Launch the TUI |
| `yukti init` | Set up OAuth credentials |
| `yukti login` | Authenticate with Google |
| `yukti logout` | Clear stored credentials |
| `yukti status` | Show configuration and auth status |
| `yukti version` | Show version info |
| `yukti new <title>` | Create a new Apps Script project and local workspace |
| `yukti clone <script-id>` | Clone a remote Apps Script project |
| `yukti pull` | Pull remote HEAD into the current workspace |
| `yukti diff` | Show local changes since the last pull or push |
| `yukti push` | Push local files to remote HEAD |
| `yukti release` | Push the current workspace, create a version, and deploy it |
| `yukti versions list` | List immutable versions for a script |
| `yukti versions create` | Create an immutable version from remote HEAD |
| `yukti deployments list` | List HEAD and versioned deployments |
| `yukti deployments create` | Create a deployment for a version |
| `yukti deployments update` | Move an existing deployment to a new version |
| `yukti deployments delete` | Delete a deployment |

## Workspace Sync

Yukti workspaces use a small `yukti.json` manifest to remember the Apps Script
project ID and the last remote snapshot Yukti saw. That lets `yukti push` catch
remote edits before replacing the project's HEAD files.

```bash
yukti new "Invoice Automation"
cd invoice-automation

# edit Code.gs / appsscript.json locally
yukti diff
yukti push
```

Supported file mapping:

| Local file | Apps Script file |
|------------|------------------|
| `Code.gs` | `Code` / `SERVER_JS` |
| `ui/Dialog.html` | `ui/Dialog` / `HTML` |
| `appsscript.json` | `appsscript` / `JSON` |

`yukti push` uploads the full project file set. If remote HEAD changed since
the last pull or push, Yukti stops and asks you to pull first. Use `--force`
only when you intentionally want to overwrite remote HEAD.

## Release & Deploy

`yukti release` is the fastest way to ship from a workspace:

```bash
yukti release --description "Production rollout"
```

It safely pushes local files, creates an immutable Apps Script version, creates
a versioned deployment, saves the deployment ID in `yukti.json`, and prints any
web app URL returned by Google. Later releases from the same workspace update
that saved deployment by default, so the URL stays stable.

To update an existing deployment without changing its URL:

```bash
yukti release --deployment-id AKfycbx... --description "Production rollout"
```

The TUI can inspect deployments and update selected versioned deployments for
any project you can access. Workspace URL stability is tracked through
`yukti.json`, so use `yukti release` from a workspace when you want future CLI
releases to keep updating the same saved deployment.

You can inspect release state at any time:

```bash
yukti versions list
yukti deployments list
```

For web apps, include `doGet(e)` or `doPost(e)` and configure `webapp` in
`appsscript.json` before releasing:

```json
{
  "webapp": {
    "access": "ANYONE_ANONYMOUS",
    "executeAs": "USER_DEPLOYING"
  }
}
```

## Troubleshooting

| Problem | Solution |
|---------|----------|
| "This app is blocked" | You need your own OAuth credentials. [Setup guide →](docs/google-oauth-setup.md) |
| "client_secret is missing" | Re-run `yukti init` and enter both Client ID and Secret |
| "unverified developer" (macOS) | Run `xattr -d com.apple.quarantine yukti` |
| Need a custom token location | Use `--token-file <path>` or set `YUKTI_TOKEN_FILE` |
| Token expired | Run `yukti login` again |

## Building from Source

```bash
git clone https://github.com/indrasvat/yukti.git
cd yukti
make build
./bin/yukti
```

## License

Apache 2.0 - see [LICENSE](LICENSE) for details.
