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

## Setup (First Time Only)

### Step 1: Create Google OAuth Credentials

You need your own Google OAuth credentials (one-time setup, ~5 minutes).

**[→ Follow the Google OAuth Setup Guide](docs/google-oauth-setup.md)**

After completing the guide, you'll have a `client_id` and `client_secret`.

### Step 2: Configure Yukti

```bash
yukti init
```

This walks you through entering your credentials. They're saved to your config file.

### Step 3: Enable File-Based Token Storage (Recommended)

This avoids macOS keychain permission popups. Edit your config file:

```bash
# macOS
nano ~/Library/Application\ Support/yukti/config.json

# Linux
nano ~/.config/yukti/config.json
```

Add `"token_file": "default"`:

```json
{
  "oauth": {
    "client_id": "your-client-id.apps.googleusercontent.com",
    "client_secret": "your-client-secret"
  },
  "token_file": "default"
}
```

### Step 4: Login

```bash
yukti login
```

Your browser opens. Sign in with Google and authorize Yukti.

### Step 5: Verify

```bash
yukti status
```

You should see all green:

```
  ━━ Configuration ━━
  ●  Config       ~/Library/Application Support/yukti/config.json
  ●  Client ID    57632406••••.com
  ●  Secret       Configured

  ━━ Authentication ━━
  ●  Status       Logged in
  ●  Expires in   ████████████████████ 59m
```

### Step 6: Launch

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
| `yukti status` | Show configuration and auth status |
| `yukti version` | Show version info |

## Troubleshooting

| Problem | Solution |
|---------|----------|
| "This app is blocked" | You need your own OAuth credentials. [Setup guide →](docs/google-oauth-setup.md) |
| "client_secret is missing" | Re-run `yukti init` and enter both Client ID and Secret |
| "unverified developer" (macOS) | Run `xattr -d com.apple.quarantine yukti` |
| Keychain keeps asking for password | Add `"token_file": "default"` to your config |
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
