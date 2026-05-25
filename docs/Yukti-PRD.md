# Yukti (युक्ति) — Product Requirements Document

> **Version:** 1.0.0  
> **Last Updated:** January 2026  
> **Status:** Draft  

---

## Executive Summary

**Yukti** (Sanskrit: युक्ति — "skillful means" or "clever device") is a blazingly fast, beautiful, and secure Terminal User Interface (TUI) for managing Google Apps Script projects. Built with Go using the Charmbracelet ecosystem (BubbleTea + LipGloss), Yukti aims to dramatically exceed the capabilities and user experience of Google's barebones web interface while providing a first-class terminal-native development workflow.

### Why Yukti?

Google's Apps Script web editor is functional but limited:
- No offline capabilities
- Slow navigation between projects
- Limited keyboard shortcuts
- No bulk operations
- Poor integration with developer workflows
- No syntax highlighting customization
- No local file management

Yukti addresses all these pain points while adding powerful features that make Apps Script development a joy.

---

## Goals & Non-Goals

### Goals

1. **Superior DX** — Make managing Apps Script projects faster, more intuitive, and more enjoyable than the web interface
2. **Offline-First** — Work with cached project data when disconnected, sync when online
3. **Power-User Optimized** — Full keyboard navigation, vim-like bindings, command palette
4. **Beautiful UI** — Thoughtfully designed TUI that's pleasant to use for extended periods
5. **Secure by Default** — Proper OAuth handling, no credential storage in plaintext, secure token refresh
6. **Extensible** — Plugin architecture for future feature additions
7. **Observable** — Comprehensive logging, metrics, and debugging capabilities

### Non-Goals

1. **Full IDE Replacement** — Yukti is for project management, not a complete IDE (use VS Code + clasp for heavy editing)
2. **Script Execution Environment** — No local execution of Apps Script code (use Apps Script API's `scripts.run`)
3. **Mobile Support** — Terminal-only; no mobile or tablet interfaces
4. **Multi-User Collaboration** — Single-user tool; no real-time collaboration features

---

## User Personas

### Primary: The Apps Script Power User

**Alex** is a Google Workspace administrator who manages 50+ Apps Script projects across the organization. They need to:
- Quickly navigate between projects
- Review and update deployments
- Monitor execution logs
- Perform bulk operations (rename, archive, share)
- Work efficiently with keyboard-only navigation

### Secondary: The Developer

**Sam** is a developer who builds Apps Script add-ons and automation tools. They need to:
- Rapidly iterate on code changes (push/pull)
- Manage multiple deployment versions
- View execution metrics and errors
- Integrate with their existing terminal workflow

### Tertiary: The Casual User

**Jordan** occasionally creates simple Apps Script automations. They need to:
- Easily find and open their projects
- View project contents without switching to browser
- Simple deployment management

---

## Feature Requirements

### P0 — Core Features (MVP)

These features are essential for the first usable release.

#### F1: Authentication & Authorization

**Description:** Secure OAuth2 authentication with Google APIs.

**Requirements:**
- OAuth2 flow with PKCE for enhanced security
- Secure token storage (OS keychain integration)
- Automatic token refresh
- Multiple account support
- Easy account switching

**Acceptance Criteria:**
- User can authenticate with `yukti login`
- Tokens stored securely in OS keychain
- Automatic re-authentication when tokens expire
- Support for 3+ Google accounts

```
┌─────────────────────────────────────────────────────────────────┐
│  Yukti Login                                                    │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│    🔐 Authentication Required                                   │
│                                                                 │
│    Press Enter to open browser for Google Sign-In...            │
│                                                                 │
│    Waiting for authentication callback...                       │
│    [████████████░░░░░░░░]                                       │
│                                                                 │
│    Account: user@example.com                                    │
│    Status:  ✓ Authenticated                                     │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

#### F2: Project List View

**Description:** Browse and manage all Apps Script projects.

**Requirements:**
- List all user's Apps Script projects
- Display project metadata (title, last modified, owner)
- Search/filter projects by name
- Sort by various criteria (name, date, owner)
- Pagination for large project lists
- Visual indicators for project type (standalone, bound)

**Acceptance Criteria:**
- Load and display 100+ projects efficiently
- Sub-second search/filter response
- Clear visual hierarchy
- Keyboard navigation (j/k, arrows, search)

```
┌─────────────────────────────────────────────────────────────────────────┐
│  ⚡ Yukti                                              user@example.com │
├─────────────────────────────────────────────────────────────────────────┤
│  🔍 Search: █                                      [?] Help  [q] Quit  │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  My Projects (47)                                    Sort: Modified ▼   │
│  ─────────────────────────────────────────────────────────────────────  │
│                                                                         │
│  ▸ 📄 Sample Scripts                    Me         Today, 4:57 PM      │
│    📄 DocMailer                         Me         Aug 16, 2021        │
│    📊 Sales Dashboard Automation        Team       Dec 3, 2025         │
│    📝 Form Response Handler             Me         Nov 28, 2025        │
│    📧 Email Scheduler                   Me         Nov 15, 2025        │
│    📁 Drive File Organizer              Shared     Oct 22, 2025        │
│                                                                         │
│                                                                         │
│  ─────────────────────────────────────────────────────────────────────  │
│  [Enter] Open  [n] New  [d] Delete  [r] Rename  [/] Search  [s] Star   │
└─────────────────────────────────────────────────────────────────────────┘
```

#### F3: Project Detail View

**Description:** View and manage individual project details.

**Requirements:**
- Display all project files with syntax highlighting
- Show project metadata and settings
- Navigate between files
- View file contents with line numbers
- Display project manifest (appsscript.json)

**Acceptance Criteria:**
- Render JavaScript with proper syntax highlighting
- Handle files up to 10,000 lines smoothly
- Support horizontal scrolling for long lines
- Display file tree for multi-file projects

```
┌─────────────────────────────────────────────────────────────────────────┐
│  ⚡ Yukti > Sample Scripts                                       [←]   │
├─────────────────────────────────────────────────────────────────────────┤
│  Files          │  Code.gs                                              │
│  ───────────────┼───────────────────────────────────────────────────────│
│                 │   1 │ function listMyFiles() {                        │
│  ▸ Code.gs      │   2 │   // "DriveApp" is the built-in wrapper for    │
│    Utils.gs     │   3 │   // the Drive API                              │
│    appsscript.  │   4 │   const files = DriveApp.getFiles();            │
│                 │   5 │   while (files.hasNext()) {                     │
│                 │   6 │     const file = files.next();                  │
│                 │   7 │     console.log(file.getName() + " (" +         │
│                 │   8 │                 file.getId() + ")");             │
│                 │   9 │   }                                              │
│                 │  10 │ }                                                │
│                 │                                                        │
│  ───────────────┼───────────────────────────────────────────────────────│
│  Info           │  Script ID: 1ABC...XYZ                                │
│  Created: Jan 5 │  Parent: None (Standalone)                            │
│  Modified: Today│  Timezone: America/Los_Angeles                        │
├─────────────────┴───────────────────────────────────────────────────────┤
│  [Tab] Switch pane  [e] Edit  [p] Push  [l] Pull  [d] Deploy  [←] Back │
└─────────────────────────────────────────────────────────────────────────┘
```

#### F4: File Push/Pull Operations

**Description:** Sync files between local filesystem and Apps Script.

**Requirements:**
- Pull project files to local directory
- Push local changes to Apps Script
- Detect and show file differences
- Conflict detection and resolution
- Batch operations for multiple files

**Acceptance Criteria:**
- Push/pull completes in < 3 seconds for typical projects
- Clear progress indication
- Rollback capability on failure
- Support for .claspignore-like patterns

#### F5: Deployment Management

**Description:** Create, view, and manage project deployments.

**Requirements:**
- List all deployments for a project
- Create new deployments with description
- Update existing deployments
- Delete deployments
- View deployment details (version, type, URL)

**Acceptance Criteria:**
- Display all deployment types (Web App, API Executable, Add-on)
- Show deployment URLs for easy access
- Quick copy-to-clipboard functionality

```
┌─────────────────────────────────────────────────────────────────────────┐
│  ⚡ Yukti > Sample Scripts > Deployments                          [←]  │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  Deployments (3)                                                        │
│  ─────────────────────────────────────────────────────────────────────  │
│                                                                         │
│  ▸ 🌐 Production Web App                                    v3         │
│       ID: AKfycbx...ABC                                                 │
│       URL: https://script.google.com/macros/s/AKfyc.../exec            │
│       Access: Anyone                                                    │
│       Created: Dec 15, 2025                                             │
│                                                                         │
│    🔌 API Executable                                        v2         │
│       ID: AKfycbx...DEF                                                 │
│       Access: Anyone with Google Account                                │
│       Created: Nov 30, 2025                                             │
│                                                                         │
│    📋 @HEAD (Development)                                   HEAD       │
│       ID: AKfycbx...GHI                                                 │
│       Auto-updates with each push                                       │
│                                                                         │
│  ─────────────────────────────────────────────────────────────────────  │
│  [Enter] Details  [n] New Deployment  [u] Update  [d] Delete  [c] Copy │
└─────────────────────────────────────────────────────────────────────────┘
```

#### F6: Version Management

**Description:** Create and manage project versions.

**Requirements:**
- List all versions with descriptions
- Create new versions
- View version metadata
- Compare versions (diff view)

**Acceptance Criteria:**
- Display version history with timestamps
- Clear version numbering
- Easy rollback to previous versions

### P1 — Enhanced Features

These features enhance usability significantly.

#### F7: Execution Log Viewer

**Description:** View script execution logs and errors.

**Requirements:**
- Real-time log streaming (tail -f style)
- Filter logs by function, status, time range
- Error highlighting and stack traces
- Export logs to file

**Acceptance Criteria:**
- Display logs within 2 seconds of execution
- Support for thousands of log entries
- Clear error formatting

```
┌─────────────────────────────────────────────────────────────────────────┐
│  ⚡ Yukti > Sample Scripts > Execution Logs                       [←]  │
├─────────────────────────────────────────────────────────────────────────┤
│  Filter: All Functions ▼    Status: All ▼    Time: Last 24h ▼          │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  ✓ listMyFiles          Completed    0.847s    Today 5:12 PM           │
│  ✓ listMyFiles          Completed    0.923s    Today 4:57 PM           │
│  ✗ sendEmails           Failed       2.103s    Today 4:45 PM           │
│    └─ Error: Exceeded maximum execution time                           │
│  ✓ processForm          Completed    0.234s    Today 4:30 PM           │
│  ✓ listMyFiles          Completed    0.891s    Today 4:15 PM           │
│  ⟳ cleanupFiles         Running...   12.4s     Today 4:14 PM           │
│                                                                         │
│  ─────────────────────────────────────────────────────────────────────  │
│  Total: 127 executions | Failed: 3 | Avg Duration: 1.2s                │
├─────────────────────────────────────────────────────────────────────────┤
│  [Enter] Details  [f] Filter  [r] Refresh  [t] Tail Mode  [e] Export   │
└─────────────────────────────────────────────────────────────────────────┘
```

#### F8: Trigger Management

**Description:** View and manage project triggers.

**Requirements:**
- List all triggers (time-based, event-based)
- Create new triggers
- Delete triggers
- View trigger execution history

**Acceptance Criteria:**
- Display trigger schedule clearly
- Show last/next execution times
- Error notifications for failed triggers

#### F9: Project Metrics Dashboard

**Description:** Visualize project usage metrics.

**Requirements:**
- Execution count over time
- Error rate trends
- User count (for add-ons)
- Quota usage

**Acceptance Criteria:**
- ASCII charts for metrics visualization
- Time range selection
- Export metrics data

```
┌─────────────────────────────────────────────────────────────────────────┐
│  ⚡ Yukti > Sample Scripts > Metrics                              [←]  │
├─────────────────────────────────────────────────────────────────────────┤
│  Time Range: Last 7 Days ▼                                              │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  Executions                                    Total: 1,247             │
│  250 │      ╭─╮                                                         │
│  200 │   ╭──╯ ╰──╮    ╭╮                                                │
│  150 │╭──╯       ╰────╯╰─╮                                              │
│  100 ││                  ╰──╮                                           │
│   50 │╯                     ╰──                                         │
│    0 └──────────────────────────                                        │
│       Mon  Tue  Wed  Thu  Fri  Sat  Sun                                 │
│                                                                         │
│  ─────────────────────────────────────────────────────────────────────  │
│  Error Rate: 2.4%      Avg Duration: 1.23s     Active Users: 47        │
│                                                                         │
├─────────────────────────────────────────────────────────────────────────┤
│  [←/→] Change range  [r] Refresh  [e] Export  [d] Details              │
└─────────────────────────────────────────────────────────────────────────┘
```

#### F10: Command Palette

**Description:** Quick-access command interface (VS Code style).

**Requirements:**
- Fuzzy search for commands
- Recent commands history
- Context-aware suggestions
- Keyboard shortcut hints

**Acceptance Criteria:**
- Open with Ctrl+P / Cmd+P
- < 50ms response time
- Support for 50+ commands

```
┌─────────────────────────────────────────────────────────────────────────┐
│                                                                         │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │  > deploy█                                                      │   │
│  ├─────────────────────────────────────────────────────────────────┤   │
│  │  ▸ Deploy Project                              Ctrl+Shift+D     │   │
│  │    Deploy: Create New Version                                   │   │
│  │    Deploy: Update Existing                                      │   │
│  │    Deploy: View All Deployments                Ctrl+D           │   │
│  │    Deploy: Delete Deployment                                    │   │
│  └─────────────────────────────────────────────────────────────────┘   │
│                                                                         │
│  [Sample Scripts project context]                                       │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

### P2 — Advanced Features

These features differentiate Yukti from alternatives.

#### F11: Code Editor with Syntax Highlighting

**Description:** In-TUI code editing with JavaScript syntax highlighting.

**Requirements:**
- JavaScript syntax highlighting (Chroma or similar)
- Line numbers
- Basic editing (insert, delete, copy, paste)
- Search and replace
- Go-to-line functionality
- Auto-indentation

**Acceptance Criteria:**
- Smooth editing for files up to 5,000 lines
- 60fps rendering during typing
- Undo/redo support

#### F12: GAS API Autocomplete (Stretch)

**Description:** Code completion for Google Apps Script APIs.

**Requirements:**
- Autocomplete for built-in services (DriveApp, SpreadsheetApp, etc.)
- Method signatures and documentation
- Parameter hints

**Acceptance Criteria:**
- Completion suggestions within 100ms
- Coverage for major GAS services
- Offline-capable (bundled definitions)

#### F13: Offline Mode

**Description:** Work with cached data when offline.

**Requirements:**
- Cache project list and metadata
- Cache project files locally
- Queue changes for sync when online
- Clear offline status indication

**Acceptance Criteria:**
- Seamless transition between online/offline
- No data loss during offline periods
- Conflict resolution on reconnect

#### F14: Bulk Operations

**Description:** Perform actions on multiple projects simultaneously.

**Requirements:**
- Multi-select projects
- Bulk delete
- Bulk archive/unarchive
- Bulk sharing changes
- Bulk tag/label management

**Acceptance Criteria:**
- Select up to 100 projects at once
- Progress indication for bulk operations
- Rollback on partial failure

### P3 — Nice-to-Have Features

#### F15: Project Templates

- Quick-start templates for common use cases
- Custom template creation
- Community template sharing

#### F16: Diff View

- Side-by-side file comparison
- Version diff
- Local vs remote diff

#### F17: Snippet Manager

- Save code snippets
- Quick insert into projects
- Snippet synchronization

#### F18: Theme System

- Multiple color themes
- Custom theme creation
- High-contrast accessibility theme

#### F19: Integration with External Editors

- Open in VS Code
- Open in vim/nvim
- External editor sync

---

## Technical Requirements

### Performance

| Metric | Target |
|--------|--------|
| Startup time | < 500ms |
| Project list load | < 1s for 100 projects |
| File render | < 100ms for 1000-line file |
| Search response | < 50ms |
| Push/pull operations | < 3s typical project |
| Memory usage | < 100MB typical |

### Security

- OAuth2 with PKCE flow
- Tokens stored in OS keychain (macOS Keychain, Windows Credential Manager, Linux Secret Service)
- No plaintext credential storage
- HTTPS-only API communication
- Token auto-refresh with secure storage

### Platform Support

| Platform | Support Level |
|----------|---------------|
| macOS (ARM64) | Full |
| macOS (AMD64) | Full |
| Linux (AMD64) | Full |
| Linux (ARM64) | Full |
| Windows (AMD64) | Full |
| Windows (ARM64) | Best effort |

### Terminal Requirements

- Minimum: 80x24 terminal
- Recommended: 120x40 terminal
- True color support (recommended)
- 256 color fallback
- 16 color fallback (basic functionality)

---

## User Experience Guidelines

### Keyboard-First Design

Every action should be accessible via keyboard:
- Arrow keys / hjkl for navigation
- Enter/Space for selection
- Escape for back/cancel
- Single-key shortcuts for common actions
- Command palette for everything else

### Progressive Disclosure

- Show essential information first
- Details available on demand
- Help text for complex features
- Contextual hints

### Error Handling

- Clear, actionable error messages
- Suggestions for resolution
- Log details for debugging
- Graceful degradation

### Feedback

- Immediate visual feedback for actions
- Progress indicators for long operations
- Success/failure confirmations
- Sound optional (terminal bell)

---

## Success Metrics

### Adoption

- 100+ GitHub stars within 3 months of release
- 50+ active users within 6 months
- Positive reception in developer communities (Reddit, HN)

### Quality

- < 5 critical bugs in first month
- < 2s average operation time
- 95%+ user satisfaction in feedback

### Performance

- All P0 features meet performance targets
- No memory leaks over extended use
- Consistent behavior across platforms

---

## Timeline & Milestones

### Phase 1: Foundation (Weeks 1-2)

- [ ] Project scaffolding and architecture
- [ ] OAuth2 implementation
- [ ] Basic project list view
- [ ] Navigation framework

### Phase 2: Core Features (Weeks 3-4)

- [ ] Project detail view
- [ ] File viewer with syntax highlighting
- [ ] Push/pull operations
- [ ] Deployment management

### Phase 3: Enhanced Features (Weeks 5-6)

- [ ] Version management
- [ ] Execution log viewer
- [ ] Command palette
- [ ] Trigger management

### Phase 4: Polish (Weeks 7-8)

- [ ] Metrics dashboard
- [ ] Offline mode
- [ ] Performance optimization
- [ ] Documentation and testing

### Phase 5: Release (Week 9+)

- [ ] Beta release
- [ ] Community feedback
- [ ] Bug fixes and improvements
- [ ] v1.0 release

---

## Appendix A: API Reference

### Apps Script API Endpoints Used

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/v1/projects` | POST | Create new project |
| `/v1/projects/{scriptId}` | GET | Get project metadata |
| `/v1/projects/{scriptId}/content` | GET | Get project files |
| `/v1/projects/{scriptId}/content` | PUT | Update project files |
| `/v1/projects/{scriptId}/metrics` | GET | Get project metrics |
| `/v1/projects/{scriptId}/deployments` | GET | List deployments |
| `/v1/projects/{scriptId}/deployments` | POST | Create deployment |
| `/v1/projects/{scriptId}/deployments/{deploymentId}` | PUT | Update deployment |
| `/v1/projects/{scriptId}/deployments/{deploymentId}` | DELETE | Delete deployment |
| `/v1/projects/{scriptId}/versions` | GET | List versions |
| `/v1/projects/{scriptId}/versions` | POST | Create version |
| `/v1/processes` | GET | List user processes |
| `/v1/projects/{scriptId}/processes` | GET | List script processes |

### OAuth2 Scopes Required

```
https://www.googleapis.com/auth/script.projects
https://www.googleapis.com/auth/script.projects.readonly
https://www.googleapis.com/auth/script.deployments
https://www.googleapis.com/auth/script.deployments.readonly
https://www.googleapis.com/auth/script.metrics
https://www.googleapis.com/auth/script.processes
```

---

## Appendix B: Competitive Analysis

### clasp (Google)

**Strengths:**
- Official Google tool
- TypeScript support (via bundler)
- Good CI/CD integration
- Established user base

**Weaknesses:**
- CLI-only (no TUI)
- Requires Node.js
- Limited project browsing
- No metrics visualization

### Yukti Differentiators

- Beautiful TUI interface
- No Node.js dependency
- Visual project management
- Metrics and log visualization
- Offline capabilities
- Single binary distribution

---

## Appendix C: Glossary

| Term | Definition |
|------|------------|
| Apps Script | Google's JavaScript-based scripting platform |
| Deployment | A published version of a script project |
| Version | An immutable snapshot of project code |
| Trigger | An automation rule that runs functions |
| Bound Script | A script attached to a Google Doc/Sheet/Form |
| Standalone Script | An independent script project |
| Manifest | The appsscript.json configuration file |

---

*Document maintained by the Yukti development team. For questions or suggestions, please open an issue on GitHub.*
