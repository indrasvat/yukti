# Google OAuth Setup Guide

Yukti requires your own Google OAuth credentials. This is a one-time setup that takes about 5 minutes.

> **Why can't Yukti use shared credentials?**
> Google blocks third-party apps from using another application's OAuth credentials. Each user must create their own.

## Step 1: Create a Google Cloud Project

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Click the project dropdown at the top of the page
3. Click **New Project**
4. Enter a name (e.g., "Yukti") and click **Create**
5. Wait for the project to be created, then select it

## Step 2: Enable the Apps Script API

1. In the search bar at the top, type **Apps Script API**
2. Click on "Apps Script API" in the results
3. Click the **Enable** button
4. Wait for the API to be enabled

## Step 3: Configure OAuth Consent Screen

1. In the left sidebar, click **Google Auth Platform**
   - If you don't see it, search for "OAuth consent screen" in the search bar
2. Click **Get started** or **Configure**
3. Fill in the required fields:
   - **App name:** `Yukti`
   - **User support email:** Your email address
   - **Audience:** Select `External`
4. Click **Save and continue** through the remaining steps
5. Go to **Audience** in the sidebar
6. Under "Test users", click **Add users**
7. Add your Google email address and click **Save**

> **Note:** While your app is in "Testing" mode, only test users can authenticate. This is fine for personal use.

## Step 4: Create OAuth Credentials

1. In the left sidebar, click **Clients**
2. Click **+ Create Client**
3. For "Application type", select **Desktop app**
4. Name it `Yukti CLI`
5. Click **Create**

## Step 5: Download Credentials

1. After creation, you'll see your new client in the list
2. Click the **download icon** (↓) to download the JSON file
3. Open the downloaded JSON file in a text editor
4. You'll see something like:

```json
{
  "installed": {
    "client_id": "123456789-abcdefg.apps.googleusercontent.com",
    "client_secret": "GOCSPX-xxxxxxxxxxxxx",
    ...
  }
}
```

5. Note the `client_id` and `client_secret` values — you'll need these next

## Step 6: Configure Yukti

Run the setup wizard:

```bash
yukti init
```

When prompted:
1. Press Enter to open Google Cloud Console (or 's' to skip if already there)
2. Paste your **Client ID** from the JSON file
3. Paste your **Client Secret** from the JSON file

Your credentials are now saved locally at:
- **macOS:** `~/Library/Application Support/yukti/config.json`
- **Linux:** `~/.config/yukti/config.json`

## Step 7: Authenticate

```bash
yukti login
```

This opens your browser. Sign in with your Google account and authorize Yukti.

## Step 8: Verify

```bash
yukti status
```

You should see green indicators showing you're logged in.

---

## Troubleshooting

### "This app is blocked" error

You're trying to use someone else's OAuth credentials. You must create your own following this guide.

### "client_secret is missing" error

Your config file is missing the client secret. Re-run `yukti init` and make sure to enter both the Client ID AND Client Secret from the downloaded JSON file.

### "Access blocked: This app's request is invalid"

The OAuth consent screen isn't configured correctly. Go back to Step 3 and ensure:
- You've set up the consent screen
- You've added yourself as a test user

### "Error 403: access_denied"

You haven't added yourself as a test user. Go to Google Cloud Console → Google Auth Platform → Audience → Add your email as a test user.

### Token expires frequently

This is normal. Google OAuth tokens expire after 1 hour. Yukti automatically refreshes them using the stored refresh token. If refresh fails, run `yukti login` again.
