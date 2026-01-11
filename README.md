# Yukti

A beautiful TUI for managing Google Apps Script projects.

## Quick Start

### 1. Download

Download the latest release for your platform from [GitHub Releases](https://github.com/robinsharma/yukti/releases).

**macOS users:** If you downloaded via browser, remove the quarantine attribute:
```bash
xattr -d com.apple.quarantine yukti
```

### 2. Install

```bash
chmod +x yukti
sudo mv yukti /usr/local/bin/
```

### 3. Set Up Google OAuth

Yukti requires your own Google OAuth credentials (one-time setup, ~5 minutes).

**[Follow the Google OAuth Setup Guide →](docs/google-oauth-setup.md)**

### 4. Initialize & Login

```bash
yukti init    # Enter your OAuth credentials
yukti login   # Authenticate with Google
```

### 5. Launch

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

```bash
yukti status
```

You should see green indicators:
```
  ━━ Configuration ━━
  ●  Config       ~/Library/Application Support/yukti/config.json
  ●  Client ID    57632406••••.com
  ●  Secret       Configured

  ━━ Authentication ━━
  ●  Status       Logged in
  ●  Expires in   ████████████████████ 59m
```

## File-Based Token Storage

By default, Yukti stores OAuth tokens in your system keychain. To use file-based storage instead (avoids keychain prompts):

**Option 1:** Add to your config file:
```json
{
  "oauth": { ... },
  "token_file": "default"
}
```

**Option 2:** Use the flag:
```bash
yukti --token-file default login
yukti --token-file default status
```

The `default` value stores tokens at `~/.config/yukti/token.json` (or `~/Library/Application Support/yukti/token.json` on macOS). You can also specify a custom path.

## Troubleshooting

| Error | Solution |
|-------|----------|
| "This app is blocked" | Create your own OAuth credentials. [Setup guide →](docs/google-oauth-setup.md) |
| "client_secret is missing" | Re-run `yukti init` with both Client ID and Secret from the JSON |
| macOS "unverified developer" | Run `xattr -d com.apple.quarantine yukti` |
| Repeated keychain prompts | Add `"token_file": "default"` to config or use `--token-file default` |

## Building from Source

```bash
git clone https://github.com/robinsharma/yukti.git
cd yukti
make build
./bin/yukti
```

## License

MIT
