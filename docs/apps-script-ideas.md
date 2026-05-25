# Google Apps Script Ideas & Testing Reference

This document contains creative Apps Script ideas for testing Yukti's push/pull, deployment, and version management features. All scripts have been verified to work with the Google APIs.

## Available Google Services

### Built-in Services (No Setup Required)
| Service | Object | Description |
|---------|--------|-------------|
| Gmail | `GmailApp` | Read, send, and organize emails |
| Google Sheets | `SpreadsheetApp` | Create, read, and modify spreadsheets |
| Google Docs | `DocumentApp` | Create and edit documents |
| Google Drive | `DriveApp` | Manage files and folders |
| Google Calendar | `CalendarApp` | Manage events and calendars |
| Google Slides | `SlidesApp` | Create and edit presentations |
| Google Forms | `FormApp` | Create and manage forms |
| Google Maps | `Maps` | Geocoding, directions, elevation |
| Translate | `LanguageApp` | Translate text between languages |
| URL Fetch | `UrlFetchApp` | Make HTTP requests to external APIs |
| Utilities | `Utilities` | Encoding, hashing, formatting |
| Cache | `CacheService` | Temporary data storage |
| Properties | `PropertiesService` | Persistent key-value storage |

### Advanced Services (Require Enabling)
- YouTube Data API
- Google Analytics
- BigQuery
- Admin SDK (Directory, Reports)
- Cloud Natural Language API
- Vertex AI / Gemini

---

## Script Categories

### 1. Gmail Automation

#### Inbox Statistics
```javascript
/**
 * Get comprehensive inbox statistics
 * Useful for productivity dashboards
 */
function getInboxStats() {
  const inbox = GmailApp.getInboxThreads(0, 100);
  const unread = GmailApp.getInboxUnreadCount();
  const labels = GmailApp.getUserLabels();

  return {
    totalThreads: inbox.length,
    unreadCount: unread,
    labelCount: labels.length,
    labels: labels.map(l => l.getName())
  };
}
```

#### Smart Email Search
```javascript
/**
 * Search Gmail with rich metadata
 * @param {string} query - Gmail search query (e.g., "from:boss is:unread")
 */
function searchGmail(query) {
  const threads = GmailApp.search(query, 0, 10);
  return threads.map(thread => ({
    subject: thread.getFirstMessageSubject(),
    date: thread.getLastMessageDate(),
    messageCount: thread.getMessageCount(),
    isUnread: thread.isUnread(),
    labels: thread.getLabels().map(l => l.getName())
  }));
}
```

#### Auto-Label by Sender Domain
```javascript
/**
 * Automatically label emails by sender's domain
 * Run as a time-driven trigger
 */
function autoLabelByDomain() {
  const threads = GmailApp.search('is:unread -has:userlabels', 0, 50);

  threads.forEach(thread => {
    const messages = thread.getMessages();
    const from = messages[0].getFrom();
    const domain = from.match(/@([^>]+)/)?.[1];

    if (domain) {
      const labelName = 'Domains/' + domain.split('.').slice(-2).join('.');
      let label = GmailApp.getUserLabelByName(labelName);
      if (!label) {
        label = GmailApp.createLabel(labelName);
      }
      thread.addLabel(label);
    }
  });
}
```

#### Mail Merge with Personalization
```javascript
/**
 * Send personalized emails from a spreadsheet
 * Sheet columns: Email, Name, Company, CustomField
 */
function mailMerge() {
  const sheet = SpreadsheetApp.getActiveSpreadsheet().getActiveSheet();
  const data = sheet.getDataRange().getValues();
  const headers = data.shift();

  const template = `
    Hi {{Name}},

    Thank you for your interest in {{Company}}.
    {{CustomField}}

    Best regards,
    The Team
  `;

  data.forEach(row => {
    const recipient = row[0];
    let body = template;

    headers.forEach((header, i) => {
      body = body.replace(new RegExp('{{' + header + '}}', 'g'), row[i]);
    });

    GmailApp.sendEmail(recipient, 'Personalized Message', body);
  });
}
```

---

### 2. Google Sheets Automation

#### Create Formatted Report
```javascript
/**
 * Create a professionally formatted spreadsheet
 */
function createFormattedReport() {
  const ss = SpreadsheetApp.create('Weekly Report - ' + new Date().toLocaleDateString());
  const sheet = ss.getActiveSheet();

  // Headers
  const headers = ['Date', 'Category', 'Amount', 'Status', 'Notes'];
  sheet.getRange(1, 1, 1, headers.length).setValues([headers])
    .setFontWeight('bold')
    .setBackground('#4285f4')
    .setFontColor('white');

  // Sample data
  const data = [
    [new Date(), 'Sales', 1500, 'Completed', 'Q1 target met'],
    [new Date(), 'Marketing', 800, 'Pending', 'Campaign running'],
    [new Date(), 'Operations', 2200, 'Completed', 'Infrastructure upgrade']
  ];
  sheet.getRange(2, 1, data.length, data[0].length).setValues(data);

  // Formatting
  sheet.setFrozenRows(1);
  sheet.autoResizeColumns(1, headers.length);

  // Conditional formatting for Status
  const statusRange = sheet.getRange('D:D');
  const rules = [
    SpreadsheetApp.newConditionalFormatRule()
      .whenTextEqualTo('Completed')
      .setBackground('#d4edda')
      .setRanges([statusRange])
      .build(),
    SpreadsheetApp.newConditionalFormatRule()
      .whenTextEqualTo('Pending')
      .setBackground('#fff3cd')
      .setRanges([statusRange])
      .build()
  ];
  sheet.setConditionalFormatRules(rules);

  return ss.getUrl();
}
```

#### Data Validation Dashboard
```javascript
/**
 * Add data validation and dropdowns to a sheet
 */
function setupDataValidation() {
  const sheet = SpreadsheetApp.getActiveSpreadsheet().getActiveSheet();

  // Status dropdown
  const statusRule = SpreadsheetApp.newDataValidation()
    .requireValueInList(['Pending', 'In Progress', 'Completed', 'Cancelled'])
    .setAllowInvalid(false)
    .build();
  sheet.getRange('D2:D100').setDataValidation(statusRule);

  // Priority dropdown with colors
  const priorityRule = SpreadsheetApp.newDataValidation()
    .requireValueInList(['High', 'Medium', 'Low'])
    .setAllowInvalid(false)
    .build();
  sheet.getRange('E2:E100').setDataValidation(priorityRule);

  // Date validation
  const dateRule = SpreadsheetApp.newDataValidation()
    .requireDate()
    .setAllowInvalid(false)
    .build();
  sheet.getRange('A2:A100').setDataValidation(dateRule);
}
```

#### Cross-Sheet Data Aggregation
```javascript
/**
 * Aggregate data from multiple sheets into a summary
 */
function aggregateSheetData() {
  const ss = SpreadsheetApp.getActiveSpreadsheet();
  const sheets = ss.getSheets();

  let summary = SpreadsheetApp.getActiveSpreadsheet().getSheetByName('Summary');
  if (!summary) {
    summary = ss.insertSheet('Summary');
  }

  const results = [];
  sheets.forEach(sheet => {
    if (sheet.getName() === 'Summary') return;

    const data = sheet.getDataRange().getValues();
    const total = data.slice(1).reduce((sum, row) => {
      const value = parseFloat(row[2]) || 0; // Assuming amount in column C
      return sum + value;
    }, 0);

    results.push([sheet.getName(), data.length - 1, total]);
  });

  summary.clear();
  summary.getRange(1, 1, 1, 3).setValues([['Sheet', 'Rows', 'Total']]);
  if (results.length > 0) {
    summary.getRange(2, 1, results.length, 3).setValues(results);
  }
}
```

---

### 3. Google Calendar Integration

#### Upcoming Events Summary
```javascript
/**
 * Get upcoming events with rich details
 * @param {number} daysAhead - Number of days to look ahead
 */
function getUpcomingEvents(daysAhead) {
  const calendar = CalendarApp.getDefaultCalendar();
  const now = new Date();
  const endDate = new Date(now.getTime() + daysAhead * 24 * 60 * 60 * 1000);

  const events = calendar.getEvents(now, endDate);
  return events.map(event => ({
    title: event.getTitle(),
    start: event.getStartTime(),
    end: event.getEndTime(),
    location: event.getLocation(),
    description: event.getDescription(),
    guests: event.getGuestList().map(g => g.getEmail()),
    isAllDay: event.isAllDayEvent()
  }));
}
```

#### Schedule from Spreadsheet
```javascript
/**
 * Create calendar events from spreadsheet data
 * Sheet columns: Title, Date, Start Time, Duration (hours), Location
 */
function createEventsFromSheet() {
  const sheet = SpreadsheetApp.getActiveSpreadsheet().getActiveSheet();
  const data = sheet.getDataRange().getValues();
  const calendar = CalendarApp.getDefaultCalendar();

  data.slice(1).forEach(row => {
    const [title, date, startTime, duration, location] = row;

    const startDateTime = new Date(date);
    startDateTime.setHours(startTime.getHours(), startTime.getMinutes());

    const endDateTime = new Date(startDateTime.getTime() + duration * 60 * 60 * 1000);

    calendar.createEvent(title, startDateTime, endDateTime, {
      location: location,
      description: 'Created by Apps Script'
    });
  });
}
```

#### Find Free Slots
```javascript
/**
 * Find free time slots in a date range
 * @param {Date} startDate - Start of search range
 * @param {Date} endDate - End of search range
 * @param {number} minDuration - Minimum slot duration in minutes
 */
function findFreeSlots(startDate, endDate, minDuration) {
  const calendar = CalendarApp.getDefaultCalendar();
  const events = calendar.getEvents(startDate, endDate);

  // Sort events by start time
  events.sort((a, b) => a.getStartTime() - b.getStartTime());

  const freeSlots = [];
  let lastEnd = startDate;

  events.forEach(event => {
    const eventStart = event.getStartTime();
    const gap = (eventStart - lastEnd) / (1000 * 60); // minutes

    if (gap >= minDuration) {
      freeSlots.push({
        start: new Date(lastEnd),
        end: new Date(eventStart),
        duration: gap
      });
    }

    const eventEnd = event.getEndTime();
    if (eventEnd > lastEnd) {
      lastEnd = eventEnd;
    }
  });

  // Check time after last event
  const finalGap = (endDate - lastEnd) / (1000 * 60);
  if (finalGap >= minDuration) {
    freeSlots.push({
      start: new Date(lastEnd),
      end: endDate,
      duration: finalGap
    });
  }

  return freeSlots;
}
```

---

### 4. Google Drive Management

#### File Organization
```javascript
/**
 * List files with detailed metadata
 * @param {number} limit - Maximum files to return
 */
function listDriveFiles(limit) {
  const files = DriveApp.getFiles();
  const result = [];

  while (files.hasNext() && result.length < limit) {
    const file = files.next();
    result.push({
      name: file.getName(),
      id: file.getId(),
      mimeType: file.getMimeType(),
      size: file.getSize(),
      lastUpdated: file.getLastUpdated(),
      url: file.getUrl(),
      sharingAccess: file.getSharingAccess().toString()
    });
  }

  return result;
}
```

#### Auto-Archive Old Files
```javascript
/**
 * Move files older than N days to an archive folder
 * @param {number} daysOld - Age threshold in days
 */
function archiveOldFiles(daysOld) {
  const cutoffDate = new Date(Date.now() - daysOld * 24 * 60 * 60 * 1000);

  // Get or create archive folder
  const archiveFolders = DriveApp.getFoldersByName('Archive');
  const archive = archiveFolders.hasNext()
    ? archiveFolders.next()
    : DriveApp.createFolder('Archive');

  const files = DriveApp.getFiles();
  const movedFiles = [];

  while (files.hasNext()) {
    const file = files.next();
    if (file.getLastUpdated() < cutoffDate) {
      file.moveTo(archive);
      movedFiles.push(file.getName());
    }
  }

  return {
    archiveFolder: archive.getUrl(),
    movedCount: movedFiles.length,
    files: movedFiles
  };
}
```

#### Duplicate Finder
```javascript
/**
 * Find potential duplicate files by name
 */
function findDuplicates() {
  const files = DriveApp.getFiles();
  const fileMap = {};
  const duplicates = [];

  while (files.hasNext()) {
    const file = files.next();
    const name = file.getName().toLowerCase();

    if (fileMap[name]) {
      fileMap[name].push({
        id: file.getId(),
        url: file.getUrl(),
        size: file.getSize(),
        lastUpdated: file.getLastUpdated()
      });
    } else {
      fileMap[name] = [{
        id: file.getId(),
        url: file.getUrl(),
        size: file.getSize(),
        lastUpdated: file.getLastUpdated()
      }];
    }
  }

  // Filter to only duplicates
  Object.keys(fileMap).forEach(name => {
    if (fileMap[name].length > 1) {
      duplicates.push({
        name: name,
        copies: fileMap[name]
      });
    }
  });

  return duplicates;
}
```

---

### 5. Google Docs Automation

#### Document Generator
```javascript
/**
 * Create a formatted document with multiple sections
 */
function createFormattedDoc() {
  const doc = DocumentApp.create('Meeting Notes - ' + new Date().toLocaleDateString());
  const body = doc.getBody();

  // Title
  body.appendParagraph('Meeting Notes')
    .setHeading(DocumentApp.ParagraphHeading.TITLE);

  // Metadata
  body.appendParagraph('Date: ' + new Date().toLocaleDateString())
    .setAttributes({
      [DocumentApp.Attribute.ITALIC]: true,
      [DocumentApp.Attribute.FOREGROUND_COLOR]: '#666666'
    });

  body.appendHorizontalRule();

  // Sections
  const sections = ['Attendees', 'Agenda', 'Discussion', 'Action Items', 'Next Steps'];
  sections.forEach(section => {
    body.appendParagraph(section)
      .setHeading(DocumentApp.ParagraphHeading.HEADING1);
    body.appendParagraph('');
  });

  return doc.getUrl();
}
```

#### Template Fill-In
```javascript
/**
 * Replace placeholders in a document template
 * @param {string} templateId - ID of the template document
 * @param {Object} replacements - Key-value pairs for replacements
 */
function fillTemplate(templateId, replacements) {
  const template = DriveApp.getFileById(templateId);
  const copy = template.makeCopy('Filled - ' + template.getName());
  const doc = DocumentApp.openById(copy.getId());
  const body = doc.getBody();

  Object.keys(replacements).forEach(key => {
    body.replaceText('{{' + key + '}}', replacements[key]);
  });

  doc.saveAndClose();
  return copy.getUrl();
}
```

---

### 6. External API Integration

#### Slack Notification
```javascript
/**
 * Send a message to Slack via webhook
 * @param {string} webhookUrl - Slack incoming webhook URL
 * @param {string} message - Message to send
 * @param {string} channel - Optional channel override
 */
function sendSlackMessage(webhookUrl, message, channel) {
  const payload = {
    text: message,
    username: 'Apps Script Bot',
    icon_emoji: ':robot_face:'
  };

  if (channel) {
    payload.channel = channel;
  }

  const options = {
    method: 'post',
    contentType: 'application/json',
    payload: JSON.stringify(payload)
  };

  const response = UrlFetchApp.fetch(webhookUrl, options);
  return response.getResponseCode();
}
```

#### Discord Webhook
```javascript
/**
 * Send an embed message to Discord
 * @param {string} webhookUrl - Discord webhook URL
 * @param {Object} data - Message data
 */
function sendDiscordEmbed(webhookUrl, data) {
  const payload = {
    embeds: [{
      title: data.title,
      description: data.description,
      color: data.color || 5814783, // Blue
      fields: data.fields || [],
      timestamp: new Date().toISOString()
    }]
  };

  const options = {
    method: 'post',
    contentType: 'application/json',
    payload: JSON.stringify(payload)
  };

  const response = UrlFetchApp.fetch(webhookUrl, options);
  return response.getResponseCode();
}
```

#### Generic REST API Client
```javascript
/**
 * Make authenticated API requests
 * @param {string} url - API endpoint
 * @param {string} method - HTTP method
 * @param {Object} headers - Request headers
 * @param {Object} payload - Request body
 */
function apiRequest(url, method, headers, payload) {
  const options = {
    method: method || 'get',
    headers: headers || {},
    muteHttpExceptions: true
  };

  if (payload && (method === 'post' || method === 'put' || method === 'patch')) {
    options.contentType = 'application/json';
    options.payload = JSON.stringify(payload);
  }

  const response = UrlFetchApp.fetch(url, options);

  return {
    statusCode: response.getResponseCode(),
    headers: response.getHeaders(),
    body: JSON.parse(response.getContentText())
  };
}
```

---

### 7. AI Integration (Gemini)

#### Using GeminiApp Library
```javascript
/**
 * Analyze text using Gemini API
 * Requires: GeminiApp library (Script ID: 1Kw_P7VGzNTr6_PabmXiHCxPNSGxSeD2RwHM3R8CRjHlOpYq7mpSxJJPQ)
 * @param {string} apiKey - Gemini API key
 * @param {string} text - Text to analyze
 */
function analyzeWithGemini(apiKey, text) {
  const genAI = new GeminiApp(apiKey);
  const model = genAI.getGenerativeModel({ model: 'gemini-1.5-flash' });

  const result = model.generateContent([
    'Analyze the following text and provide key insights:',
    text
  ]);

  return result.response.text();
}
```

#### Direct Gemini API Call
```javascript
/**
 * Call Gemini API directly without library
 * @param {string} apiKey - Gemini API key
 * @param {string} prompt - The prompt to send
 */
function callGeminiDirect(apiKey, prompt) {
  const url = 'https://generativelanguage.googleapis.com/v1beta/models/gemini-1.5-flash:generateContent?key=' + apiKey;

  const payload = {
    contents: [{
      parts: [{
        text: prompt
      }]
    }]
  };

  const options = {
    method: 'post',
    contentType: 'application/json',
    payload: JSON.stringify(payload),
    muteHttpExceptions: true
  };

  const response = UrlFetchApp.fetch(url, options);
  const result = JSON.parse(response.getContentText());

  if (result.candidates && result.candidates[0]) {
    return result.candidates[0].content.parts[0].text;
  }

  return result;
}
```

---

### 8. Triggers & Automation

#### Time-Driven Triggers
```javascript
/**
 * Set up automated triggers programmatically
 */
function setupTriggers() {
  // Delete existing triggers first
  ScriptApp.getProjectTriggers().forEach(trigger => {
    ScriptApp.deleteTrigger(trigger);
  });

  // Daily summary at 9 AM
  ScriptApp.newTrigger('dailySummary')
    .timeBased()
    .atHour(9)
    .everyDays(1)
    .create();

  // Hourly check
  ScriptApp.newTrigger('hourlyCheck')
    .timeBased()
    .everyHours(1)
    .create();

  // Weekly report on Monday at 8 AM
  ScriptApp.newTrigger('weeklyReport')
    .timeBased()
    .onWeekDay(ScriptApp.WeekDay.MONDAY)
    .atHour(8)
    .create();
}
```

#### Form Submit Trigger
```javascript
/**
 * Process form submissions automatically
 * Attach to Form as onFormSubmit trigger
 * @param {Object} e - Event object
 */
function onFormSubmit(e) {
  const responses = e.values;
  const timestamp = responses[0];
  const email = responses[1];
  const name = responses[2];

  // Send confirmation email
  GmailApp.sendEmail(email,
    'Thank you for your submission',
    'Hi ' + name + ',\n\nWe received your submission at ' + timestamp + '.\n\nBest regards'
  );

  // Log to spreadsheet
  const logSheet = SpreadsheetApp.openById('YOUR_LOG_SHEET_ID').getActiveSheet();
  logSheet.appendRow([timestamp, email, name, 'Processed']);

  // Notify via Slack (if configured)
  // sendSlackMessage(WEBHOOK_URL, 'New form submission from ' + name);
}
```

---

## Testing Checklist

Use these scripts to test Yukti's features:

### Push/Pull Testing
1. Create a new script with `createFormattedReport()`
2. Push changes to Google
3. Modify via Apps Script editor
4. Pull changes back
5. Verify conflict handling

### Deployment Testing
1. Deploy as web app
2. Test different version numbers
3. Verify deployment listing
4. Update deployments

### Version Testing
1. Create multiple versions with descriptions
2. List version history
3. Roll back to previous versions
4. Compare version differences

---

## Resources

- [Official Apps Script Samples](https://github.com/googleworkspace/apps-script-samples)
- [Apps Script Reference](https://developers.google.com/apps-script/reference)
- [Advanced Services Guide](https://developers.google.com/apps-script/guides/services/advanced)
- [GeminiApp Library](https://github.com/mhawksey/GeminiApp)
- [Apps Script Awesome List](https://github.com/oshliaer/google-apps-script-awesome-list)

---

## Project IDs for Testing

These are sample projects that can be used for Yukti testing:

| Project | Script ID | Description |
|---------|-----------|-------------|
| Yukti Test Scripts | `1XawIjT8_t7YrgT4uB8wmxXqnJfdHPfPwthmcoED7jc9Sr0rv7hV1Hq6D` | Comprehensive API test suite |

Add your own test projects here as you create them.
