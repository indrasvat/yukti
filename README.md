# Yukti

A beautiful TUI for managing Google Apps Script projects.

## Quick Start

### 1. Download

Download the latest release for your platform from [GitHub Releases](https://github.com/robinsharma/yukti/releases).

**macOS users:** If you downloaded via browser, remove the quarantine attribute:
```bash
xattr -d com.apple.quarantine yukti
```
> Downloads via `curl` or `wget` don't have this issue.

### 2. Install

Make the binary executable and move it to your PATH:
```bash
chmod +x yukti
sudo mv yukti /usr/local/bin/
```

Or keep it local:
```bash
chmod +x yukti
./yukti --help
```

### 3. Create Google OAuth Credentials

Yukti requires your own Google OAuth credentials (it cannot use shared credentials due to Google's security policies).

1. Go to [Google Cloud Console](https://console.cloud.google.com/)

2. **Create a new project** (or select an existing one)
   - Click the project dropdown at the top
   - Click "New Project"
   - Name it "Yukti" or similar

3. **Enable the Apps Script API**
   - Use the search bar at the top
   - Search for "Apps Script API"
   - Click "Enable"

4. **Configure OAuth consent screen**
   - Go to "Google Auth Platform" in the left sidebar
   - Click "Get started" or "Configure"
   - App name: `Yukti`
   - User support email: your email
   - Audience: `External`
   - Complete the wizard
   - Go to "Audience" and add yourself as a test user

5. **Create OAuth credentials**
   - Go to "Clients" in the left sidebar
   - Click "+ Create Client"
   - Application type: `Desktop app`
   - Name: `Yukti CLI`
   - Click "Create"
   - Click the download icon to download the JSON file
   - Open the JSON and note the `client_id` and `client_secret`

### 4. Initialize Yukti

Run the setup wizard:
```bash
yukti init
```

This will guide you through entering your OAuth credentials. They are stored locally at:
- macOS: `~/Library/Application Support/yukti/config.json`
- Linux: `~/.config/yukti/config.json`

### 5. Login

Authenticate with Google:
```bash
yukti login
```

This opens your browser for Google sign-in. After authorizing, your tokens are stored securely in your system keychain.

### 6. Launch the TUI

```bash
yukti
```

## Commands

| Command | Description |
|---------|-------------|
| `yukti` | Launch the TUI |
| `yukti init` | Set up OAuth credentials |
| `yukti login` | Authenticate with Google |
| `yukti logout` | Clear stored credentials |
| `yukti status` | Show auth and config status |
| `yukti version` | Show version info |

## Verify Setup

Check that everything is configured correctly:
```bash
yukti status
```

You should see green indicators for config and authentication:
```
  ━━ Configuration ━━

  ●  Config       ~/Library/Application Support/yukti/config.json
  ●  Client ID    57632406••••.com
  ●  Secret       Configured

  ━━ Authentication ━━

  ●  Status       Logged in
  ●  Expires in   ████████████████████ 59m
```

## Troubleshooting

### "This app is blocked" error during login

You're using someone else's OAuth credentials. Create your own following [Step 3](#3-create-google-oauth-credentials).

### "client_secret is missing" error

Your config is missing the client secret. Re-run `yukti init` and enter both the Client ID and Client Secret from the downloaded JSON file.

### Repeated keychain password prompts (macOS, development)

This happens during development when rebuilding the binary. For development, use:
```bash
export YUKTI_TOKEN_FILE=~/.config/yukti/dev-token.json
yukti login
yukti status
```

### "operation not permitted" or "unverified developer" (macOS)

Remove the quarantine attribute:
```bash
xattr -d com.apple.quarantine /path/to/yukti
```

Or allow in System Preferences → Security & Privacy → General.

## Building from Source

```bash
git clone https://github.com/robinsharma/yukti.git
cd yukti
make build
./bin/yukti --help
```

## License

MIT
