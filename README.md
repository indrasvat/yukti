# Yukti

A beautiful TUI for managing Google Apps Script projects.

## Installation

### macOS (Apple Silicon)
```bash
curl -L https://github.com/robinsharma/yukti/releases/latest/download/yukti-darwin-arm64 -o yukti
chmod +x yukti
sudo mv yukti /usr/local/bin/
```

### macOS (Intel)
```bash
curl -L https://github.com/robinsharma/yukti/releases/latest/download/yukti-darwin-amd64 -o yukti
chmod +x yukti
sudo mv yukti /usr/local/bin/
```

### Linux
```bash
curl -L https://github.com/robinsharma/yukti/releases/latest/download/yukti-linux-amd64 -o yukti
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

Enter your Client ID and Client Secret when prompted. The wizard will also ask about token storage (file-based is recommended to avoid keychain prompts).

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

## Troubleshooting

| Problem | Solution |
|---------|----------|
| "This app is blocked" | You need your own OAuth credentials. [Setup guide →](docs/google-oauth-setup.md) |
| "client_secret is missing" | Re-run `yukti init` and enter both Client ID and Secret |
| "unverified developer" (macOS) | Run `xattr -d com.apple.quarantine yukti` |
| Keychain keeps asking for password | Re-run `yukti init` and choose file-based token storage |
| Token expired | Run `yukti login` again |

## Building from Source

```bash
git clone https://github.com/robinsharma/yukti.git
cd yukti
make build
./bin/yukti
```

## License

MIT
