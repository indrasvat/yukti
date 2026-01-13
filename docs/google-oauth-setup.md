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

## Running Functions (Optional Setup)

To use Yukti's function execution feature (Ctrl+R), your Apps Script projects need additional setup.

### Step 6: Link Apps Script Project to GCP Project

Each Apps Script project you want to run functions from must be linked to your GCP project:

1. Open your Apps Script project at [script.google.com](https://script.google.com/)
2. Click the **⚙ gear icon** (Project Settings) in the left sidebar
3. Under "Google Cloud Platform (GCP) Project", click **Change project**
4. Enter your GCP **Project Number** (found in GCP Console → Project Info → Project number)
5. Click **Set project**

### Step 7: Deploy as API Executable

You must create an API Executable deployment at least once:

1. In your Apps Script project, click **Deploy** → **New deployment**
2. Click the **gear icon** next to "Select type" and choose **API Executable**
3. Optionally add a description (e.g., "API access for Yukti")
4. Click **Deploy**

> **Note:** With `devMode`, Yukti always runs the latest saved version of your script.
> The deployment is only needed once to enable API access.

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

### "resource not found: Requested entity was not found"

When running functions (Ctrl+R), this error means one of:

1. **Script not linked to GCP project**: Follow Steps 6-7 above
2. **No API Executable deployment**: Create a deployment as described in Step 7
3. **Wrong GCP project**: Ensure the Apps Script project is linked to the **same** GCP project where you created your OAuth credentials

### "access forbidden: Request had insufficient authentication scopes"

The OAuth token doesn't have permissions for services your script uses. You need to re-authenticate:

```bash
yukti logout && yukti login
```

This will request all necessary scopes including Gmail, Calendar, Drive, and Spreadsheets access.
