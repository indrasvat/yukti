# Google OAuth Setup Guide

Yukti requires your own Google OAuth credentials. This is a one-time setup that takes about 5 minutes.

> **Why can't Yukti use shared credentials?**
> Google blocks third-party apps from using another application's OAuth credentials. Each user must create their own.

---

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

> **Note:** While in "Testing" mode, only test users can authenticate. This is fine for personal use.

## Step 4: Create OAuth Credentials

1. In the left sidebar, click **Clients**
2. Click **+ Create Client**
3. For "Application type", select **Desktop app**
4. Name it `Yukti CLI`
5. Click **Create**

## Step 5: Get Your Credentials

1. After creation, you'll see your new client in the list
2. Click the **download icon** (↓) to download the JSON file
3. Open the JSON file — you'll see:

```json
{
  "installed": {
    "client_id": "123456789-xxx.apps.googleusercontent.com",
    "client_secret": "GOCSPX-xxxxxxxxxxxxx",
    ...
  }
}
```

4. Keep this file open — you'll need the `client_id` and `client_secret` values

---

**Done!** Now return to the terminal and run:

```bash
yukti init
```

Paste your Client ID and Client Secret when prompted.

---

## Troubleshooting

### "This app is blocked"

You're using someone else's OAuth credentials. Create your own following this guide.

### "Access blocked: This app's request is invalid"

The OAuth consent screen isn't configured correctly:
- Go back to Step 3
- Ensure you've added yourself as a test user in the Audience section

### "Error 403: access_denied"

You haven't added yourself as a test user. Go to:
Google Cloud Console → Google Auth Platform → Audience → Add your email
