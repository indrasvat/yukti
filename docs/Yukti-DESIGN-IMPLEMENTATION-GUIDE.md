# Yukti (युक्ति) — Design & Implementation Guide

> **Version:** 1.0.0
> **Last Updated:** January 11, 2026
> **Status:** Living Document — Phase 1 Complete  

---

## Table of Contents

1. [Architecture Overview](#architecture-overview)
2. [Project Structure](#project-structure)
3. [Core Design Principles](#core-design-principles)
4. [Component Architecture](#component-architecture)
5. [API Client Design](#api-client-design)
6. [TUI Architecture](#tui-architecture)
7. [Plugin System](#plugin-system)
8. [Testing Strategy](#testing-strategy)
9. [Observability](#observability)
10. [Build & Development](#build--development)
11. [Implementation Phases](#implementation-phases)
12. [UI Mockups](#ui-mockups)

---

## Architecture Overview

Yukti follows a clean architecture with clear separation between:
- **Domain Layer** — Core business logic and entities
- **Application Layer** — Use cases and orchestration
- **Infrastructure Layer** — External services (Google APIs, filesystem, keychain)
- **Presentation Layer** — TUI components (BubbleTea models)

```
┌─────────────────────────────────────────────────────────────────────────┐
│                         Presentation Layer                              │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐   │
│  │ ProjectList │  │ProjectDetail│  │ Deployments │  │   Metrics   │   │
│  │    View     │  │    View     │  │    View     │  │    View     │   │
│  └──────┬──────┘  └──────┬──────┘  └──────┬──────┘  └──────┬──────┘   │
│         └────────────────┴────────────────┴────────────────┘           │
│                                   │                                     │
│                          ┌────────┴────────┐                           │
│                          │   App Router    │                           │
│                          └────────┬────────┘                           │
├──────────────────────────────────┼──────────────────────────────────────┤
│                         Application Layer                               │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐   │
│  │  Project    │  │ Deployment  │  │   Version   │  │   Process   │   │
│  │  Service    │  │  Service    │  │   Service   │  │   Service   │   │
│  └──────┬──────┘  └──────┬──────┘  └──────┬──────┘  └──────┬──────┘   │
│         └────────────────┴────────────────┴────────────────┘           │
│                                   │                                     │
├──────────────────────────────────┼──────────────────────────────────────┤
│                          Domain Layer                                   │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐   │
│  │   Project   │  │ Deployment  │  │   Version   │  │   Process   │   │
│  │   Entity    │  │   Entity    │  │   Entity    │  │   Entity    │   │
│  └─────────────┘  └─────────────┘  └─────────────┘  └─────────────┘   │
│                                                                         │
│  ┌───────────────────────────────────────────────────────────────────┐ │
│  │                    Repository Interfaces                          │ │
│  └───────────────────────────────────────────────────────────────────┘ │
├─────────────────────────────────────────────────────────────────────────┤
│                       Infrastructure Layer                              │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐   │
│  │   Google    │  │  Keychain   │  │ Filesystem  │  │   Config    │   │
│  │ API Client  │  │   Client    │  │   Client    │  │   Store     │   │
│  └─────────────┘  └─────────────┘  └─────────────┘  └─────────────┘   │
└─────────────────────────────────────────────────────────────────────────┘
```

---

## Project Structure

```
yukti/
├── cmd/
│   └── yukti/
│       └── main.go                 # Application entry point
├── internal/
│   ├── domain/                     # Domain layer
│   │   ├── project/
│   │   │   ├── entity.go           # Project entity
│   │   │   ├── repository.go       # Repository interface
│   │   │   └── errors.go           # Domain errors
│   │   ├── deployment/
│   │   │   ├── entity.go
│   │   │   ├── repository.go
│   │   │   └── errors.go
│   │   ├── version/
│   │   │   ├── entity.go
│   │   │   ├── repository.go
│   │   │   └── errors.go
│   │   └── process/
│   │       ├── entity.go
│   │       ├── repository.go
│   │       └── errors.go
│   ├── application/                # Application layer (use cases)
│   │   ├── project/
│   │   │   ├── service.go
│   │   │   ├── list.go
│   │   │   ├── get.go
│   │   │   ├── create.go
│   │   │   ├── update.go
│   │   │   └── delete.go
│   │   ├── deployment/
│   │   │   ├── service.go
│   │   │   ├── list.go
│   │   │   ├── create.go
│   │   │   ├── update.go
│   │   │   └── delete.go
│   │   ├── version/
│   │   │   └── service.go
│   │   └── process/
│   │       └── service.go
│   ├── infrastructure/             # Infrastructure layer
│   │   ├── google/
│   │   │   ├── client.go           # HTTP client wrapper
│   │   │   ├── auth.go             # OAuth2 implementation
│   │   │   ├── projects.go         # Projects API
│   │   │   ├── deployments.go      # Deployments API
│   │   │   ├── versions.go         # Versions API
│   │   │   └── processes.go        # Processes API
│   │   ├── keychain/
│   │   │   ├── store.go            # Keychain interface
│   │   │   ├── darwin.go           # macOS implementation
│   │   │   ├── linux.go            # Linux implementation
│   │   │   └── windows.go          # Windows implementation
│   │   ├── filesystem/
│   │   │   ├── local.go            # Local file operations
│   │   │   └── sync.go             # Sync operations
│   │   ├── cache/
│   │   │   └── store.go            # Local cache for offline
│   │   └── config/
│   │       └── config.go           # Configuration management
│   ├── tui/                        # Presentation layer
│   │   ├── app.go                  # Main TUI application
│   │   ├── router.go               # View routing
│   │   ├── keys.go                 # Keybindings
│   │   ├── styles/
│   │   │   ├── theme.go            # Theme definitions
│   │   │   └── colors.go           # Color palette
│   │   ├── components/             # Reusable components
│   │   │   ├── header.go
│   │   │   ├── footer.go
│   │   │   ├── list.go
│   │   │   ├── table.go
│   │   │   ├── spinner.go
│   │   │   ├── input.go
│   │   │   ├── modal.go
│   │   │   ├── toast.go
│   │   │   ├── code_viewer.go
│   │   │   ├── file_tree.go
│   │   │   └── command_palette.go
│   │   └── views/                  # Page views
│   │       ├── login.go
│   │       ├── project_list.go
│   │       ├── project_detail.go
│   │       ├── deployments.go
│   │       ├── versions.go
│   │       ├── processes.go
│   │       ├── metrics.go
│   │       └── settings.go
│   └── plugin/                     # Plugin system
│       ├── manager.go
│       ├── interface.go
│       └── loader.go
├── pkg/                            # Public packages
│   ├── syntax/                     # Syntax highlighting
│   │   └── javascript.go
│   └── ascii/                      # ASCII chart rendering
│       └── charts.go
├── plugins/                        # Built-in plugins
│   └── example/
│       └── plugin.go
├── docs/
│   ├── Yukti-PRD.md
│   └── Yukti-DESIGN-IMPLEMENTATION-GUIDE.md
├── scripts/
│   ├── install.sh
│   └── release.sh
├── testdata/                       # Test fixtures
├── .github/
│   └── workflows/
│       └── ci.yml
├── Makefile
├── go.mod
├── go.sum
├── CLAUDE.md                       # AI agent learnings
├── README.md
├── LICENSE
└── .goreleaser.yml
```

---

## Core Design Principles

### 1. SOLID Principles

#### Single Responsibility Principle (SRP)
Each component has one reason to change:

```go
// Good: ProjectService only handles project use cases
type ProjectService struct {
    repo    project.Repository
    logger  *slog.Logger
}

func (s *ProjectService) List(ctx context.Context) ([]project.Project, error)
func (s *ProjectService) Get(ctx context.Context, id string) (*project.Project, error)
func (s *ProjectService) Create(ctx context.Context, req CreateRequest) (*project.Project, error)

// Bad: God service doing everything
type GodService struct {
    // handles projects, deployments, versions, auth, cache...
}
```

#### Open/Closed Principle (OCP)
Open for extension, closed for modification via interfaces:

```go
// Repository interface allows multiple implementations
type Repository interface {
    List(ctx context.Context) ([]Project, error)
    Get(ctx context.Context, id string) (*Project, error)
    Create(ctx context.Context, p *Project) error
    Update(ctx context.Context, p *Project) error
    Delete(ctx context.Context, id string) error
}

// Google API implementation
type GoogleRepository struct {
    client *google.Client
}

// Cache implementation (for offline mode)
type CachedRepository struct {
    remote Repository
    cache  cache.Store
}

// Mock implementation (for testing)
type MockRepository struct {
    projects map[string]*Project
}
```

#### Liskov Substitution Principle (LSP)
Subtypes must be substitutable for their base types:

```go
// Any Repository implementation can be used interchangeably
func NewProjectService(repo project.Repository) *ProjectService {
    return &ProjectService{repo: repo}
}

// Works with any implementation
service := NewProjectService(googleRepo)
service := NewProjectService(cachedRepo)
service := NewProjectService(mockRepo)
```

#### Interface Segregation Principle (ISP)
Many specific interfaces over one general interface:

```go
// Good: Specific interfaces
type ProjectReader interface {
    List(ctx context.Context) ([]Project, error)
    Get(ctx context.Context, id string) (*Project, error)
}

type ProjectWriter interface {
    Create(ctx context.Context, p *Project) error
    Update(ctx context.Context, p *Project) error
    Delete(ctx context.Context, id string) error
}

// Components only depend on what they need
type ProjectListView struct {
    reader ProjectReader  // Only needs read access
}

type ProjectEditView struct {
    writer ProjectWriter  // Only needs write access
}
```

#### Dependency Inversion Principle (DIP)
High-level modules don't depend on low-level modules:

```go
// Domain layer defines interfaces
package project

type Repository interface {
    Get(ctx context.Context, id string) (*Project, error)
}

// Infrastructure layer implements interfaces
package google

type ProjectRepository struct {
    client *Client
}

func (r *ProjectRepository) Get(ctx context.Context, id string) (*project.Project, error) {
    // Implementation using Google API
}

// Application layer depends on abstractions
package application

type ProjectService struct {
    repo project.Repository  // Interface, not concrete type
}
```

### 2. Separation of Concerns

The TUI layer never directly calls Google APIs:

```go
// TUI View
type ProjectListModel struct {
    service *application.ProjectService  // Depends on application layer
    // NOT: client *google.Client         // Never depends on infrastructure
}

func (m ProjectListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        if msg.String() == "enter" {
            return m, m.loadProjects()
        }
    case projectsLoadedMsg:
        m.projects = msg.projects
    }
    return m, nil
}

func (m ProjectListModel) loadProjects() tea.Cmd {
    return func() tea.Msg {
        projects, err := m.service.List(context.Background())
        if err != nil {
            return errorMsg{err}
        }
        return projectsLoadedMsg{projects}
    }
}
```

### 3. Testability

Every component designed for easy testing:

```go
// Test with mock repository
func TestProjectService_List(t *testing.T) {
    mockRepo := &MockRepository{
        projects: map[string]*project.Project{
            "abc123": {ID: "abc123", Title: "Test Project"},
        },
    }
    
    service := application.NewProjectService(mockRepo)
    
    projects, err := service.List(context.Background())
    
    assert.NoError(t, err)
    assert.Len(t, projects, 1)
    assert.Equal(t, "Test Project", projects[0].Title)
}

// Test TUI with teatest
func TestProjectListView(t *testing.T) {
    mockService := &MockProjectService{
        ListFunc: func(ctx context.Context) ([]project.Project, error) {
            return []project.Project{{ID: "1", Title: "Test"}}, nil
        },
    }
    
    model := views.NewProjectListModel(mockService)
    tm := teatest.NewTestModel(t, model)
    
    tm.Send(tea.KeyMsg{Type: tea.KeyEnter})
    tm.WaitFinished(t, teatest.WithFinalModel(func(m tea.Model) {
        plm := m.(views.ProjectListModel)
        assert.Len(t, plm.Projects(), 1)
    }))
}
```

---

## Component Architecture

### Domain Entities

```go
// internal/domain/project/entity.go
package project

import "time"

type Project struct {
    ID           string
    Title        string
    ParentID     string  // For bound scripts
    CreateTime   time.Time
    UpdateTime   time.Time
    Creator      User
    LastModifier User
}

type User struct {
    Domain   string
    Email    string
    Name     string
    PhotoURL string
}

type File struct {
    Name         string
    Type         FileType
    Source       string
    LastModified time.Time
    FunctionSet  *FunctionSet  // Parsed functions
}

type FileType string

const (
    FileTypeServer FileType = "SERVER_JS"
    FileTypeHTML   FileType = "HTML"
    FileTypeJSON   FileType = "JSON"
)

type FunctionSet struct {
    Functions []Function
}

type Function struct {
    Name       string
    Parameters []string
}
```

```go
// internal/domain/deployment/entity.go
package deployment

import "time"

type Deployment struct {
    ID           string
    Version      *Version
    Config       DeploymentConfig
    UpdateTime   time.Time
    EntryPoints  []EntryPoint
}

type Version struct {
    VersionNumber int
    Description   string
    CreateTime    time.Time
}

type DeploymentConfig struct {
    ScriptID    string
    VersionID   int
    Description string
}

type EntryPoint struct {
    Type        EntryPointType
    WebApp      *WebAppConfig
    ExecutionAPI *ExecutionAPIConfig
    AddOn       *AddOnConfig
}

type EntryPointType string

const (
    EntryPointWebApp    EntryPointType = "WEB_APP"
    EntryPointExecAPI   EntryPointType = "EXECUTION_API"
    EntryPointAddOn     EntryPointType = "ADD_ON"
)

type WebAppConfig struct {
    Access      WebAppAccess
    URL         string
}

type WebAppAccess string

const (
    WebAppAccessMyself   WebAppAccess = "MYSELF"
    WebAppAccessDomain   WebAppAccess = "DOMAIN"
    WebAppAccessAnyone   WebAppAccess = "ANYONE"
    WebAppAccessAnyoneAnonymous WebAppAccess = "ANYONE_ANONYMOUS"
)
```

### Repository Interfaces

```go
// internal/domain/project/repository.go
package project

import "context"

type Repository interface {
    // Queries
    List(ctx context.Context, opts ListOptions) ([]Project, error)
    Get(ctx context.Context, id string) (*Project, error)
    GetContent(ctx context.Context, id string) (*Content, error)
    GetMetrics(ctx context.Context, id string, opts MetricsOptions) (*Metrics, error)
    
    // Commands
    Create(ctx context.Context, req CreateRequest) (*Project, error)
    UpdateContent(ctx context.Context, id string, content *Content) error
}

type ListOptions struct {
    PageSize  int
    PageToken string
}

type Content struct {
    ScriptID string
    Files    []File
}

type CreateRequest struct {
    Title    string
    ParentID string  // Optional, for bound scripts
}

type MetricsOptions struct {
    StartTime time.Time
    EndTime   time.Time
}

type Metrics struct {
    ActiveUsers   []MetricValue
    TotalUsers    []MetricValue
    Executions    []MetricValue
    FailedPercent []MetricValue
}

type MetricValue struct {
    Time  time.Time
    Value int64
}
```

---

## API Client Design

### HTTP Client Wrapper

```go
// internal/infrastructure/google/client.go
package google

import (
    "context"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "net/url"
    "time"
    
    "golang.org/x/oauth2"
)

const baseURL = "https://script.googleapis.com/v1"

type Client struct {
    httpClient *http.Client
    baseURL    string
    logger     *slog.Logger
}

func NewClient(ctx context.Context, tokenSource oauth2.TokenSource, logger *slog.Logger) *Client {
    return &Client{
        httpClient: oauth2.NewClient(ctx, tokenSource),
        baseURL:    baseURL,
        logger:     logger,
    }
}

func (c *Client) do(ctx context.Context, method, path string, body io.Reader, result any) error {
    url := c.baseURL + path
    
    c.logger.Debug("API request",
        slog.String("method", method),
        slog.String("url", url),
    )
    
    req, err := http.NewRequestWithContext(ctx, method, url, body)
    if err != nil {
        return fmt.Errorf("creating request: %w", err)
    }
    
    if body != nil {
        req.Header.Set("Content-Type", "application/json")
    }
    
    start := time.Now()
    resp, err := c.httpClient.Do(req)
    duration := time.Since(start)
    
    c.logger.Debug("API response",
        slog.String("method", method),
        slog.String("url", url),
        slog.Int("status", resp.StatusCode),
        slog.Duration("duration", duration),
    )
    
    if err != nil {
        return fmt.Errorf("executing request: %w", err)
    }
    defer resp.Body.Close()
    
    if resp.StatusCode >= 400 {
        return c.handleError(resp)
    }
    
    if result != nil {
        if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
            return fmt.Errorf("decoding response: %w", err)
        }
    }
    
    return nil
}

func (c *Client) Get(ctx context.Context, path string, result any) error {
    return c.do(ctx, http.MethodGet, path, nil, result)
}

func (c *Client) Post(ctx context.Context, path string, body, result any) error {
    b, err := json.Marshal(body)
    if err != nil {
        return fmt.Errorf("encoding request: %w", err)
    }
    return c.do(ctx, http.MethodPost, path, bytes.NewReader(b), result)
}

func (c *Client) Put(ctx context.Context, path string, body, result any) error {
    b, err := json.Marshal(body)
    if err != nil {
        return fmt.Errorf("encoding request: %w", err)
    }
    return c.do(ctx, http.MethodPut, path, bytes.NewReader(b), result)
}

func (c *Client) Delete(ctx context.Context, path string) error {
    return c.do(ctx, http.MethodDelete, path, nil, nil)
}
```

### OAuth2 Implementation

```go
// internal/infrastructure/google/auth.go
package google

import (
    "context"
    "crypto/rand"
    "encoding/base64"
    "fmt"
    "net"
    "net/http"
    
    "golang.org/x/oauth2"
    "golang.org/x/oauth2/google"
)

var scopes = []string{
    "https://www.googleapis.com/auth/script.projects",
    "https://www.googleapis.com/auth/script.deployments",
    "https://www.googleapis.com/auth/script.metrics",
    "https://www.googleapis.com/auth/script.processes",
}

type Authenticator struct {
    config       *oauth2.Config
    keychain     keychain.Store
    logger       *slog.Logger
}

func NewAuthenticator(clientID, clientSecret string, keychain keychain.Store, logger *slog.Logger) *Authenticator {
    return &Authenticator{
        config: &oauth2.Config{
            ClientID:     clientID,
            ClientSecret: clientSecret,
            Scopes:       scopes,
            Endpoint:     google.Endpoint,
            RedirectURL:  "http://localhost:0/callback",
        },
        keychain: keychain,
        logger:   logger,
    }
}

func (a *Authenticator) Login(ctx context.Context) (*oauth2.Token, error) {
    // Generate PKCE verifier
    verifier := generateVerifier()
    challenge := generateChallenge(verifier)
    
    // Start local callback server
    listener, err := net.Listen("tcp", "localhost:0")
    if err != nil {
        return nil, fmt.Errorf("starting callback server: %w", err)
    }
    defer listener.Close()
    
    port := listener.Addr().(*net.TCPAddr).Port
    a.config.RedirectURL = fmt.Sprintf("http://localhost:%d/callback", port)
    
    // Generate state for CSRF protection
    state := generateState()
    
    // Build auth URL with PKCE
    authURL := a.config.AuthCodeURL(state,
        oauth2.SetAuthURLParam("code_challenge", challenge),
        oauth2.SetAuthURLParam("code_challenge_method", "S256"),
    )
    
    // Channel for receiving auth code
    codeChan := make(chan string, 1)
    errChan := make(chan error, 1)
    
    // Start callback handler
    server := &http.Server{
        Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            if r.URL.Path != "/callback" {
                return
            }
            
            if r.URL.Query().Get("state") != state {
                errChan <- fmt.Errorf("invalid state")
                return
            }
            
            code := r.URL.Query().Get("code")
            if code == "" {
                errChan <- fmt.Errorf("no code in callback")
                return
            }
            
            w.Header().Set("Content-Type", "text/html")
            w.Write([]byte(successHTML))
            codeChan <- code
        }),
    }
    
    go server.Serve(listener)
    
    // Open browser for authentication
    if err := openBrowser(authURL); err != nil {
        return nil, fmt.Errorf("opening browser: %w", err)
    }
    
    a.logger.Info("Waiting for authentication...", slog.String("url", authURL))
    
    // Wait for callback or timeout
    select {
    case code := <-codeChan:
        // Exchange code for token with PKCE verifier
        token, err := a.config.Exchange(ctx, code,
            oauth2.SetAuthURLParam("code_verifier", verifier),
        )
        if err != nil {
            return nil, fmt.Errorf("exchanging code: %w", err)
        }
        
        // Store token in keychain
        if err := a.keychain.StoreToken(token); err != nil {
            a.logger.Warn("Failed to store token", slog.Any("error", err))
        }
        
        return token, nil
        
    case err := <-errChan:
        return nil, err
        
    case <-ctx.Done():
        return nil, ctx.Err()
    }
}

func (a *Authenticator) GetToken(ctx context.Context) (*oauth2.Token, error) {
    // Try to load from keychain
    token, err := a.keychain.LoadToken()
    if err != nil {
        return nil, fmt.Errorf("loading token: %w", err)
    }
    
    if token == nil {
        return nil, ErrNotAuthenticated
    }
    
    // Check if token needs refresh
    if token.Valid() {
        return token, nil
    }
    
    // Refresh token
    tokenSource := a.config.TokenSource(ctx, token)
    newToken, err := tokenSource.Token()
    if err != nil {
        return nil, fmt.Errorf("refreshing token: %w", err)
    }
    
    // Store refreshed token
    if err := a.keychain.StoreToken(newToken); err != nil {
        a.logger.Warn("Failed to store refreshed token", slog.Any("error", err))
    }
    
    return newToken, nil
}

func generateVerifier() string {
    b := make([]byte, 32)
    rand.Read(b)
    return base64.RawURLEncoding.EncodeToString(b)
}

func generateChallenge(verifier string) string {
    h := sha256.Sum256([]byte(verifier))
    return base64.RawURLEncoding.EncodeToString(h[:])
}

func generateState() string {
    b := make([]byte, 16)
    rand.Read(b)
    return base64.RawURLEncoding.EncodeToString(b)
}
```

### Projects Repository Implementation

```go
// internal/infrastructure/google/projects.go
package google

import (
    "context"
    "fmt"
    "net/url"
    
    "yukti/internal/domain/project"
)

type ProjectRepository struct {
    client *Client
}

func NewProjectRepository(client *Client) *ProjectRepository {
    return &ProjectRepository{client: client}
}

func (r *ProjectRepository) List(ctx context.Context, opts project.ListOptions) ([]project.Project, error) {
    path := "/projects"
    
    if opts.PageToken != "" {
        path += "?pageToken=" + url.QueryEscape(opts.PageToken)
    }
    
    var resp projectsListResponse
    if err := r.client.Get(ctx, path, &resp); err != nil {
        return nil, fmt.Errorf("listing projects: %w", err)
    }
    
    projects := make([]project.Project, len(resp.Projects))
    for i, p := range resp.Projects {
        projects[i] = mapProject(p)
    }
    
    return projects, nil
}

func (r *ProjectRepository) Get(ctx context.Context, id string) (*project.Project, error) {
    path := fmt.Sprintf("/projects/%s", id)
    
    var resp projectResponse
    if err := r.client.Get(ctx, path, &resp); err != nil {
        return nil, fmt.Errorf("getting project %s: %w", id, err)
    }
    
    p := mapProject(resp)
    return &p, nil
}

func (r *ProjectRepository) GetContent(ctx context.Context, id string) (*project.Content, error) {
    path := fmt.Sprintf("/projects/%s/content", id)
    
    var resp contentResponse
    if err := r.client.Get(ctx, path, &resp); err != nil {
        return nil, fmt.Errorf("getting content for %s: %w", id, err)
    }
    
    return mapContent(resp), nil
}

func (r *ProjectRepository) Create(ctx context.Context, req project.CreateRequest) (*project.Project, error) {
    body := createProjectRequest{
        Title:    req.Title,
        ParentID: req.ParentID,
    }
    
    var resp projectResponse
    if err := r.client.Post(ctx, "/projects", body, &resp); err != nil {
        return nil, fmt.Errorf("creating project: %w", err)
    }
    
    p := mapProject(resp)
    return &p, nil
}

func (r *ProjectRepository) UpdateContent(ctx context.Context, id string, content *project.Content) error {
    path := fmt.Sprintf("/projects/%s/content", id)
    
    body := updateContentRequest{
        ScriptID: content.ScriptID,
        Files:    mapFilesToAPI(content.Files),
    }
    
    if err := r.client.Put(ctx, path, body, nil); err != nil {
        return fmt.Errorf("updating content for %s: %w", id, err)
    }
    
    return nil
}

// API response/request types
type projectsListResponse struct {
    Projects []projectResponse `json:"projects"`
}

type projectResponse struct {
    ScriptID       string       `json:"scriptId"`
    Title          string       `json:"title"`
    ParentID       string       `json:"parentId,omitempty"`
    CreateTime     string       `json:"createTime"`
    UpdateTime     string       `json:"updateTime"`
    Creator        userResponse `json:"creator"`
    LastModifyUser userResponse `json:"lastModifyUser"`
}

type userResponse struct {
    Domain   string `json:"domain"`
    Email    string `json:"email"`
    Name     string `json:"name"`
    PhotoURL string `json:"photoUrl"`
}

type contentResponse struct {
    ScriptID string         `json:"scriptId"`
    Files    []fileResponse `json:"files"`
}

type fileResponse struct {
    Name         string       `json:"name"`
    Type         string       `json:"type"`
    Source       string       `json:"source"`
    LastModified string       `json:"lastModifyTime"`
    FunctionSet  *functionSet `json:"functionSet,omitempty"`
}

type functionSet struct {
    Values []functionValue `json:"values"`
}

type functionValue struct {
    Name string `json:"name"`
}

// Mapping functions
func mapProject(p projectResponse) project.Project {
    return project.Project{
        ID:           p.ScriptID,
        Title:        p.Title,
        ParentID:     p.ParentID,
        CreateTime:   parseTime(p.CreateTime),
        UpdateTime:   parseTime(p.UpdateTime),
        Creator:      mapUser(p.Creator),
        LastModifier: mapUser(p.LastModifyUser),
    }
}

func mapUser(u userResponse) project.User {
    return project.User{
        Domain:   u.Domain,
        Email:    u.Email,
        Name:     u.Name,
        PhotoURL: u.PhotoURL,
    }
}
```

---

## TUI Architecture

### Main Application Model

```go
// internal/tui/app.go
package tui

import (
    "context"
    
    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/lipgloss"
    
    "yukti/internal/application"
    "yukti/internal/tui/styles"
    "yukti/internal/tui/views"
)

type App struct {
    // Dependencies
    projectService    *application.ProjectService
    deploymentService *application.DeploymentService
    versionService    *application.VersionService
    processService    *application.ProcessService
    
    // State
    currentView  View
    viewStack    []View
    width        int
    height       int
    
    // Components
    header       Header
    footer       Footer
    cmdPalette   *CommandPalette
    toastManager *ToastManager
    
    // Context
    ctx context.Context
}

type View interface {
    tea.Model
    Title() string
    ShortHelp() []key.Binding
}

func NewApp(ctx context.Context, services *Services) *App {
    return &App{
        ctx:               ctx,
        projectService:    services.Project,
        deploymentService: services.Deployment,
        versionService:    services.Version,
        processService:    services.Process,
        currentView:       views.NewProjectListView(services.Project),
        viewStack:         make([]View, 0),
        header:            NewHeader(),
        footer:            NewFooter(),
        cmdPalette:        NewCommandPalette(),
        toastManager:      NewToastManager(),
    }
}

func (a *App) Init() tea.Cmd {
    return tea.Batch(
        a.currentView.Init(),
        a.header.Init(),
        a.footer.Init(),
    )
}

func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    var cmds []tea.Cmd
    
    switch msg := msg.(type) {
    case tea.WindowSizeMsg:
        a.width = msg.Width
        a.height = msg.Height
        // Propagate size to components
        cmds = append(cmds, a.propagateSize())
        
    case tea.KeyMsg:
        // Global keybindings
        switch {
        case key.Matches(msg, keys.Quit):
            return a, tea.Quit
        case key.Matches(msg, keys.CommandPalette):
            a.cmdPalette.Show()
        case key.Matches(msg, keys.Back):
            if len(a.viewStack) > 0 {
                a.currentView = a.viewStack[len(a.viewStack)-1]
                a.viewStack = a.viewStack[:len(a.viewStack)-1]
            }
        }
        
    case NavigateMsg:
        a.viewStack = append(a.viewStack, a.currentView)
        a.currentView = msg.View
        cmds = append(cmds, a.currentView.Init())
        
    case ToastMsg:
        cmds = append(cmds, a.toastManager.Show(msg))
    }
    
    // Update current view
    var cmd tea.Cmd
    a.currentView, cmd = a.currentView.Update(msg)
    cmds = append(cmds, cmd)
    
    // Update components
    a.header, cmd = a.header.Update(msg)
    cmds = append(cmds, cmd)
    
    a.footer, cmd = a.footer.Update(msg)
    cmds = append(cmds, cmd)
    
    if a.cmdPalette.Visible() {
        a.cmdPalette, cmd = a.cmdPalette.Update(msg)
        cmds = append(cmds, cmd)
    }
    
    a.toastManager, cmd = a.toastManager.Update(msg)
    cmds = append(cmds, cmd)
    
    return a, tea.Batch(cmds...)
}

func (a *App) View() string {
    // Build layout
    header := a.header.View()
    content := a.currentView.View()
    footer := a.footer.View()
    
    // Calculate content height
    contentHeight := a.height - lipgloss.Height(header) - lipgloss.Height(footer)
    
    content = lipgloss.NewStyle().
        Height(contentHeight).
        Width(a.width).
        Render(content)
    
    view := lipgloss.JoinVertical(lipgloss.Left,
        header,
        content,
        footer,
    )
    
    // Overlay command palette if visible
    if a.cmdPalette.Visible() {
        view = a.overlay(view, a.cmdPalette.View())
    }
    
    // Overlay toasts
    if a.toastManager.HasToasts() {
        view = a.overlayToasts(view, a.toastManager.View())
    }
    
    return view
}

func (a *App) overlay(base, overlay string) string {
    // Center overlay on screen
    overlayStyle := lipgloss.NewStyle().
        Width(a.width).
        Height(a.height).
        Align(lipgloss.Center, lipgloss.Center)
    
    return lipgloss.Place(
        a.width,
        a.height,
        lipgloss.Center,
        lipgloss.Center,
        overlay,
        lipgloss.WithWhitespaceForeground(styles.Overlay),
    )
}
```

### Theme & Styles

```go
// internal/tui/styles/theme.go
package styles

import "github.com/charmbracelet/lipgloss"

// Color palette
var (
    // Primary colors
    Primary    = lipgloss.Color("#7C3AED")  // Purple
    Secondary  = lipgloss.Color("#10B981")  // Green
    Accent     = lipgloss.Color("#F59E0B")  // Amber
    
    // Background colors
    Background     = lipgloss.Color("#1A1B26")
    Surface        = lipgloss.Color("#24283B")
    Overlay        = lipgloss.Color("#414868")
    
    // Text colors
    TextPrimary   = lipgloss.Color("#C0CAF5")
    TextSecondary = lipgloss.Color("#9AA5CE")
    TextMuted     = lipgloss.Color("#565F89")
    
    // Status colors
    Success = lipgloss.Color("#9ECE6A")
    Warning = lipgloss.Color("#E0AF68")
    Error   = lipgloss.Color("#F7768E")
    Info    = lipgloss.Color("#7AA2F7")
    
    // Border colors
    Border       = lipgloss.Color("#414868")
    BorderFocus  = lipgloss.Color("#7C3AED")
)

// Base styles
var (
    BaseStyle = lipgloss.NewStyle().
        Background(Background).
        Foreground(TextPrimary)
    
    TitleStyle = lipgloss.NewStyle().
        Bold(true).
        Foreground(Primary).
        MarginBottom(1)
    
    SubtitleStyle = lipgloss.NewStyle().
        Foreground(TextSecondary)
    
    MutedStyle = lipgloss.NewStyle().
        Foreground(TextMuted)
)

// Component styles
var (
    HeaderStyle = lipgloss.NewStyle().
        Background(Surface).
        Foreground(TextPrimary).
        Padding(0, 2).
        BorderStyle(lipgloss.NormalBorder()).
        BorderBottom(true).
        BorderForeground(Border)
    
    FooterStyle = lipgloss.NewStyle().
        Background(Surface).
        Foreground(TextMuted).
        Padding(0, 2).
        BorderStyle(lipgloss.NormalBorder()).
        BorderTop(true).
        BorderForeground(Border)
    
    PanelStyle = lipgloss.NewStyle().
        Background(Surface).
        BorderStyle(lipgloss.RoundedBorder()).
        BorderForeground(Border).
        Padding(1)
    
    FocusedPanelStyle = PanelStyle.
        BorderForeground(BorderFocus)
    
    ListItemStyle = lipgloss.NewStyle().
        PaddingLeft(2)
    
    SelectedItemStyle = ListItemStyle.
        Background(Overlay).
        Foreground(Primary)
    
    ButtonStyle = lipgloss.NewStyle().
        Foreground(TextPrimary).
        Background(Surface).
        Padding(0, 2).
        BorderStyle(lipgloss.RoundedBorder()).
        BorderForeground(Border)
    
    PrimaryButtonStyle = ButtonStyle.
        Background(Primary).
        Foreground(lipgloss.Color("#FFFFFF"))
    
    InputStyle = lipgloss.NewStyle().
        BorderStyle(lipgloss.RoundedBorder()).
        BorderForeground(Border).
        Padding(0, 1)
    
    FocusedInputStyle = InputStyle.
        BorderForeground(BorderFocus)
)

// Status badges
func SuccessBadge(text string) string {
    return lipgloss.NewStyle().
        Foreground(Success).
        Render("✓ " + text)
}

func ErrorBadge(text string) string {
    return lipgloss.NewStyle().
        Foreground(Error).
        Render("✗ " + text)
}

func WarningBadge(text string) string {
    return lipgloss.NewStyle().
        Foreground(Warning).
        Render("⚠ " + text)
}

func InfoBadge(text string) string {
    return lipgloss.NewStyle().
        Foreground(Info).
        Render("ℹ " + text)
}
```

### Project List View

```go
// internal/tui/views/project_list.go
package views

import (
    "context"
    "fmt"
    "strings"
    
    "github.com/charmbracelet/bubbles/key"
    "github.com/charmbracelet/bubbles/list"
    "github.com/charmbracelet/bubbles/textinput"
    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/lipgloss"
    
    "yukti/internal/application"
    "yukti/internal/domain/project"
    "yukti/internal/tui/styles"
)

type ProjectListModel struct {
    service  *application.ProjectService
    projects []project.Project
    filtered []project.Project
    
    list       list.Model
    search     textinput.Model
    searching  bool
    
    width  int
    height int
    
    loading bool
    err     error
}

func NewProjectListModel(service *application.ProjectService) *ProjectListModel {
    // Initialize list
    l := list.New([]list.Item{}, projectItemDelegate{}, 0, 0)
    l.Title = "My Projects"
    l.SetShowStatusBar(true)
    l.SetFilteringEnabled(false)  // We handle filtering ourselves
    l.Styles.Title = styles.TitleStyle
    l.Styles.TitleBar = styles.PanelStyle
    
    // Initialize search input
    search := textinput.New()
    search.Placeholder = "Search projects..."
    search.CharLimit = 100
    search.Width = 30
    
    return &ProjectListModel{
        service: service,
        list:    l,
        search:  search,
        loading: true,
    }
}

func (m *ProjectListModel) Init() tea.Cmd {
    return m.loadProjects()
}

func (m *ProjectListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    var cmds []tea.Cmd
    
    switch msg := msg.(type) {
    case tea.WindowSizeMsg:
        m.width = msg.Width
        m.height = msg.Height
        m.list.SetSize(msg.Width, msg.Height-4)  // Account for header/footer
        
    case tea.KeyMsg:
        if m.searching {
            switch {
            case key.Matches(msg, keys.Escape):
                m.searching = false
                m.search.Blur()
                return m, nil
            case key.Matches(msg, keys.Enter):
                m.filterProjects()
                m.searching = false
                m.search.Blur()
                return m, nil
            default:
                var cmd tea.Cmd
                m.search, cmd = m.search.Update(msg)
                m.filterProjects()
                return m, cmd
            }
        }
        
        switch {
        case key.Matches(msg, keys.Search):
            m.searching = true
            m.search.Focus()
            return m, textinput.Blink
            
        case key.Matches(msg, keys.Enter):
            if item, ok := m.list.SelectedItem().(projectItem); ok {
                return m, Navigate(NewProjectDetailModel(m.service, item.project.ID))
            }
            
        case key.Matches(msg, keys.Refresh):
            m.loading = true
            return m, m.loadProjects()
            
        case key.Matches(msg, keys.New):
            return m, m.showNewProjectModal()
            
        case key.Matches(msg, keys.Delete):
            if item, ok := m.list.SelectedItem().(projectItem); ok {
                return m, m.confirmDelete(item.project)
            }
        }
        
    case projectsLoadedMsg:
        m.loading = false
        m.projects = msg.projects
        m.filterProjects()
        
    case projectsErrorMsg:
        m.loading = false
        m.err = msg.err
    }
    
    // Update list
    var cmd tea.Cmd
    m.list, cmd = m.list.Update(msg)
    cmds = append(cmds, cmd)
    
    return m, tea.Batch(cmds...)
}

func (m *ProjectListModel) View() string {
    if m.loading {
        return m.renderLoading()
    }
    
    if m.err != nil {
        return m.renderError()
    }
    
    var b strings.Builder
    
    // Search bar
    if m.searching {
        searchStyle := styles.FocusedInputStyle
        b.WriteString(searchStyle.Render(m.search.View()))
        b.WriteString("\n\n")
    } else if m.search.Value() != "" {
        searchStyle := styles.InputStyle
        b.WriteString(searchStyle.Render(fmt.Sprintf("🔍 %s", m.search.Value())))
        b.WriteString("\n\n")
    }
    
    // Project list
    b.WriteString(m.list.View())
    
    return styles.PanelStyle.
        Width(m.width).
        Height(m.height).
        Render(b.String())
}

func (m *ProjectListModel) Title() string {
    return "Projects"
}

func (m *ProjectListModel) ShortHelp() []key.Binding {
    return []key.Binding{
        keys.Enter,
        keys.Search,
        keys.New,
        keys.Delete,
        keys.Refresh,
    }
}

// Helper methods

func (m *ProjectListModel) loadProjects() tea.Cmd {
    return func() tea.Msg {
        projects, err := m.service.List(context.Background())
        if err != nil {
            return projectsErrorMsg{err}
        }
        return projectsLoadedMsg{projects}
    }
}

func (m *ProjectListModel) filterProjects() {
    query := strings.ToLower(m.search.Value())
    if query == "" {
        m.filtered = m.projects
    } else {
        m.filtered = make([]project.Project, 0)
        for _, p := range m.projects {
            if strings.Contains(strings.ToLower(p.Title), query) {
                m.filtered = append(m.filtered, p)
            }
        }
    }
    
    // Update list items
    items := make([]list.Item, len(m.filtered))
    for i, p := range m.filtered {
        items[i] = projectItem{project: p}
    }
    m.list.SetItems(items)
}

func (m *ProjectListModel) renderLoading() string {
    return styles.PanelStyle.
        Width(m.width).
        Height(m.height).
        Align(lipgloss.Center, lipgloss.Center).
        Render("Loading projects...")
}

func (m *ProjectListModel) renderError() string {
    return styles.PanelStyle.
        Width(m.width).
        Height(m.height).
        Align(lipgloss.Center, lipgloss.Center).
        Render(styles.ErrorBadge(m.err.Error()))
}

// List item implementation

type projectItem struct {
    project project.Project
}

func (i projectItem) Title() string       { return i.project.Title }
func (i projectItem) Description() string { return i.project.LastModifier.Email }
func (i projectItem) FilterValue() string { return i.project.Title }

type projectItemDelegate struct{}

func (d projectItemDelegate) Height() int                             { return 2 }
func (d projectItemDelegate) Spacing() int                            { return 0 }
func (d projectItemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

func (d projectItemDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
    i, ok := item.(projectItem)
    if !ok {
        return
    }
    
    var style lipgloss.Style
    if index == m.Index() {
        style = styles.SelectedItemStyle
    } else {
        style = styles.ListItemStyle
    }
    
    // Project icon based on type
    icon := "📄"
    if i.project.ParentID != "" {
        icon = "📎" // Bound script
    }
    
    title := style.Render(fmt.Sprintf("%s %s", icon, i.project.Title))
    desc := styles.MutedStyle.Render(formatTime(i.project.UpdateTime))
    
    fmt.Fprintf(w, "%s\n%s\n", title, desc)
}

// Messages

type projectsLoadedMsg struct {
    projects []project.Project
}

type projectsErrorMsg struct {
    err error
}
```

### Code Viewer Component

```go
// internal/tui/components/code_viewer.go
package components

import (
    "fmt"
    "strings"
    
    "github.com/charmbracelet/bubbles/viewport"
    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/lipgloss"
    
    "yukti/internal/tui/styles"
    "yukti/pkg/syntax"
)

type CodeViewer struct {
    viewport viewport.Model
    content  string
    filename string
    
    lineNumbers bool
    syntax      bool
    
    width  int
    height int
}

func NewCodeViewer() *CodeViewer {
    vp := viewport.New(0, 0)
    vp.Style = lipgloss.NewStyle().
        BorderStyle(lipgloss.RoundedBorder()).
        BorderForeground(styles.Border)
    
    return &CodeViewer{
        viewport:    vp,
        lineNumbers: true,
        syntax:      true,
    }
}

func (c *CodeViewer) SetContent(filename, content string) {
    c.filename = filename
    c.content = content
    c.renderContent()
}

func (c *CodeViewer) SetSize(width, height int) {
    c.width = width
    c.height = height
    c.viewport.Width = width - 2   // Account for borders
    c.viewport.Height = height - 2
    c.renderContent()
}

func (c *CodeViewer) Update(msg tea.Msg) (*CodeViewer, tea.Cmd) {
    var cmd tea.Cmd
    c.viewport, cmd = c.viewport.Update(msg)
    return c, cmd
}

func (c *CodeViewer) View() string {
    // Header with filename
    header := styles.SubtitleStyle.
        Width(c.width).
        Render(c.filename)
    
    // Viewport content
    content := c.viewport.View()
    
    // Footer with position
    percent := float64(c.viewport.ScrollPercent()) * 100
    footer := styles.MutedStyle.
        Width(c.width).
        Align(lipgloss.Right).
        Render(fmt.Sprintf("%.0f%%", percent))
    
    return lipgloss.JoinVertical(lipgloss.Left,
        header,
        content,
        footer,
    )
}

func (c *CodeViewer) renderContent() {
    lines := strings.Split(c.content, "\n")
    
    // Calculate line number width
    lineNumWidth := len(fmt.Sprintf("%d", len(lines)))
    
    var b strings.Builder
    
    for i, line := range lines {
        // Line number
        if c.lineNumbers {
            lineNum := styles.MutedStyle.
                Width(lineNumWidth).
                Align(lipgloss.Right).
                Render(fmt.Sprintf("%d", i+1))
            b.WriteString(lineNum)
            b.WriteString(" │ ")
        }
        
        // Syntax highlighted line
        if c.syntax && isJavaScript(c.filename) {
            line = syntax.HighlightJavaScript(line)
        }
        
        b.WriteString(line)
        b.WriteString("\n")
    }
    
    c.viewport.SetContent(b.String())
}

func isJavaScript(filename string) bool {
    return strings.HasSuffix(filename, ".gs") ||
           strings.HasSuffix(filename, ".js")
}
```

---

## Plugin System

### Plugin Interface

```go
// internal/plugin/interface.go
package plugin

import (
    "context"
    
    tea "github.com/charmbracelet/bubbletea"
)

// Plugin represents a Yukti plugin
type Plugin interface {
    // Metadata
    Name() string
    Version() string
    Description() string
    
    // Lifecycle
    Init(ctx context.Context, api PluginAPI) error
    Shutdown() error
}

// ViewPlugin adds new views to Yukti
type ViewPlugin interface {
    Plugin
    
    // Views returns the views provided by this plugin
    Views() []View
}

// CommandPlugin adds commands to the command palette
type CommandPlugin interface {
    Plugin
    
    // Commands returns the commands provided by this plugin
    Commands() []Command
}

// HookPlugin allows hooking into Yukti events
type HookPlugin interface {
    Plugin
    
    // OnProjectLoad is called when a project is loaded
    OnProjectLoad(project *domain.Project) error
    
    // OnBeforePush is called before a push operation
    OnBeforePush(project *domain.Project) error
    
    // OnAfterPush is called after a successful push
    OnAfterPush(project *domain.Project) error
}

// PluginAPI provides services to plugins
type PluginAPI interface {
    // Services
    ProjectService() *application.ProjectService
    DeploymentService() *application.DeploymentService
    
    // TUI helpers
    ShowToast(message string, level ToastLevel)
    Navigate(view View)
    
    // Config
    GetConfig(key string) (string, bool)
    SetConfig(key, value string) error
}

// View represents a plugin-provided view
type View struct {
    ID       string
    Title    string
    Icon     string
    Model    tea.Model
    Priority int  // Higher = appears earlier in navigation
}

// Command represents a plugin-provided command
type Command struct {
    ID          string
    Title       string
    Description string
    Shortcut    string
    Execute     func(ctx context.Context) tea.Cmd
}

type ToastLevel string

const (
    ToastInfo    ToastLevel = "info"
    ToastSuccess ToastLevel = "success"
    ToastWarning ToastLevel = "warning"
    ToastError   ToastLevel = "error"
)
```

### Plugin Manager

```go
// internal/plugin/manager.go
package plugin

import (
    "context"
    "fmt"
    "log/slog"
    "plugin"
    "path/filepath"
)

type Manager struct {
    plugins  []Plugin
    api      PluginAPI
    logger   *slog.Logger
    pluginDir string
}

func NewManager(api PluginAPI, pluginDir string, logger *slog.Logger) *Manager {
    return &Manager{
        plugins:   make([]Plugin, 0),
        api:       api,
        logger:    logger,
        pluginDir: pluginDir,
    }
}

func (m *Manager) LoadPlugins(ctx context.Context) error {
    // Find plugin files
    matches, err := filepath.Glob(filepath.Join(m.pluginDir, "*.so"))
    if err != nil {
        return fmt.Errorf("finding plugins: %w", err)
    }
    
    for _, path := range matches {
        if err := m.loadPlugin(ctx, path); err != nil {
            m.logger.Warn("Failed to load plugin",
                slog.String("path", path),
                slog.Any("error", err),
            )
            continue
        }
    }
    
    m.logger.Info("Loaded plugins", slog.Int("count", len(m.plugins)))
    return nil
}

func (m *Manager) loadPlugin(ctx context.Context, path string) error {
    // Open plugin
    p, err := plugin.Open(path)
    if err != nil {
        return fmt.Errorf("opening plugin: %w", err)
    }
    
    // Look up the New function
    sym, err := p.Lookup("New")
    if err != nil {
        return fmt.Errorf("looking up New: %w", err)
    }
    
    // Create plugin instance
    newFunc, ok := sym.(func() Plugin)
    if !ok {
        return fmt.Errorf("invalid New function signature")
    }
    
    plug := newFunc()
    
    // Initialize plugin
    if err := plug.Init(ctx, m.api); err != nil {
        return fmt.Errorf("initializing plugin: %w", err)
    }
    
    m.plugins = append(m.plugins, plug)
    
    m.logger.Info("Loaded plugin",
        slog.String("name", plug.Name()),
        slog.String("version", plug.Version()),
    )
    
    return nil
}

func (m *Manager) Shutdown() {
    for _, p := range m.plugins {
        if err := p.Shutdown(); err != nil {
            m.logger.Warn("Error shutting down plugin",
                slog.String("name", p.Name()),
                slog.Any("error", err),
            )
        }
    }
}

func (m *Manager) Views() []View {
    var views []View
    for _, p := range m.plugins {
        if vp, ok := p.(ViewPlugin); ok {
            views = append(views, vp.Views()...)
        }
    }
    return views
}

func (m *Manager) Commands() []Command {
    var commands []Command
    for _, p := range m.plugins {
        if cp, ok := p.(CommandPlugin); ok {
            commands = append(commands, cp.Commands()...)
        }
    }
    return commands
}
```

---

## Testing Strategy

### Unit Tests

```go
// internal/application/project/service_test.go
package project_test

import (
    "context"
    "testing"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    
    "yukti/internal/application/project"
    domain "yukti/internal/domain/project"
)

type MockRepository struct {
    mock.Mock
}

func (m *MockRepository) List(ctx context.Context, opts domain.ListOptions) ([]domain.Project, error) {
    args := m.Called(ctx, opts)
    return args.Get(0).([]domain.Project), args.Error(1)
}

func (m *MockRepository) Get(ctx context.Context, id string) (*domain.Project, error) {
    args := m.Called(ctx, id)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*domain.Project), args.Error(1)
}

func TestProjectService_List(t *testing.T) {
    tests := []struct {
        name     string
        setup    func(*MockRepository)
        expected int
        wantErr  bool
    }{
        {
            name: "returns projects successfully",
            setup: func(m *MockRepository) {
                m.On("List", mock.Anything, mock.Anything).Return(
                    []domain.Project{
                        {ID: "1", Title: "Project 1"},
                        {ID: "2", Title: "Project 2"},
                    },
                    nil,
                )
            },
            expected: 2,
            wantErr:  false,
        },
        {
            name: "returns empty list",
            setup: func(m *MockRepository) {
                m.On("List", mock.Anything, mock.Anything).Return(
                    []domain.Project{},
                    nil,
                )
            },
            expected: 0,
            wantErr:  false,
        },
        {
            name: "handles error",
            setup: func(m *MockRepository) {
                m.On("List", mock.Anything, mock.Anything).Return(
                    []domain.Project(nil),
                    assert.AnError,
                )
            },
            expected: 0,
            wantErr:  true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            repo := new(MockRepository)
            tt.setup(repo)
            
            service := project.NewService(repo, nil)
            
            projects, err := service.List(context.Background())
            
            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
                assert.Len(t, projects, tt.expected)
            }
            
            repo.AssertExpectations(t)
        })
    }
}
```

### TUI Tests with teatest

```go
// internal/tui/views/project_list_test.go
package views_test

import (
    "testing"
    "time"
    
    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/x/exp/teatest"
    
    "yukti/internal/tui/views"
)

func TestProjectListView_Navigation(t *testing.T) {
    mockService := &MockProjectService{
        ListFunc: func(ctx context.Context) ([]project.Project, error) {
            return []project.Project{
                {ID: "1", Title: "Project 1"},
                {ID: "2", Title: "Project 2"},
                {ID: "3", Title: "Project 3"},
            }, nil
        },
    }
    
    model := views.NewProjectListModel(mockService)
    tm := teatest.NewTestModel(t, model)
    
    // Wait for initial load
    tm.WaitUpdate(t, tea.Msg(views.ProjectsLoadedMsg{}), teatest.WithDuration(time.Second))
    
    // Navigate down
    tm.Send(tea.KeyMsg{Type: tea.KeyDown})
    tm.Send(tea.KeyMsg{Type: tea.KeyDown})
    
    // Check selection
    tm.WaitFinished(t, teatest.WithFinalModel(func(m tea.Model) {
        plm := m.(*views.ProjectListModel)
        assert.Equal(t, 2, plm.SelectedIndex())
    }))
}

func TestProjectListView_Search(t *testing.T) {
    mockService := &MockProjectService{
        ListFunc: func(ctx context.Context) ([]project.Project, error) {
            return []project.Project{
                {ID: "1", Title: "Alpha Project"},
                {ID: "2", Title: "Beta Project"},
                {ID: "3", Title: "Gamma Project"},
            }, nil
        },
    }
    
    model := views.NewProjectListModel(mockService)
    tm := teatest.NewTestModel(t, model)
    
    // Wait for load
    tm.WaitUpdate(t, tea.Msg(views.ProjectsLoadedMsg{}), teatest.WithDuration(time.Second))
    
    // Start search
    tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
    
    // Type search query
    tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b', 'e', 't', 'a'}})
    
    // Check filtered results
    tm.WaitFinished(t, teatest.WithFinalModel(func(m tea.Model) {
        plm := m.(*views.ProjectListModel)
        assert.Len(t, plm.FilteredProjects(), 1)
        assert.Equal(t, "Beta Project", plm.FilteredProjects()[0].Title)
    }))
}
```

### Integration Tests with iTerm2 Driver

Use the `iterm2-driver` Claude Code skill to test-drive the TUI as it is built.

If it's not already installed on the machine, see instructions at https://github.com/indrasvat/claude-code-skills/blob/main/README.md.

---

## Observability

### Structured Logging

```go
// internal/infrastructure/logging/logger.go
package logging

import (
    "context"
    "io"
    "log/slog"
    "os"
    "path/filepath"
    "runtime"
)

type Config struct {
    Level      slog.Level
    Format     string  // "json" or "text"
    Output     string  // "stdout", "stderr", or file path
    AddSource  bool
}

func NewLogger(cfg Config) (*slog.Logger, error) {
    var writer io.Writer
    
    switch cfg.Output {
    case "stdout":
        writer = os.Stdout
    case "stderr":
        writer = os.Stderr
    default:
        // File output
        dir := filepath.Dir(cfg.Output)
        if err := os.MkdirAll(dir, 0755); err != nil {
            return nil, err
        }
        f, err := os.OpenFile(cfg.Output, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
        if err != nil {
            return nil, err
        }
        writer = f
    }
    
    opts := &slog.HandlerOptions{
        Level:     cfg.Level,
        AddSource: cfg.AddSource,
    }
    
    var handler slog.Handler
    if cfg.Format == "json" {
        handler = slog.NewJSONHandler(writer, opts)
    } else {
        handler = slog.NewTextHandler(writer, opts)
    }
    
    return slog.New(handler), nil
}

// WithContext adds context values to the logger
func WithContext(ctx context.Context, logger *slog.Logger) *slog.Logger {
    // Add request ID if present
    if reqID, ok := ctx.Value("request_id").(string); ok {
        logger = logger.With(slog.String("request_id", reqID))
    }
    
    // Add user email if present
    if email, ok := ctx.Value("user_email").(string); ok {
        logger = logger.With(slog.String("user", email))
    }
    
    return logger
}

// WithOperation adds operation context
func WithOperation(logger *slog.Logger, op string) *slog.Logger {
    return logger.With(slog.String("operation", op))
}
```

### Metrics Collection

```go
// internal/infrastructure/metrics/metrics.go
package metrics

import (
    "sync"
    "time"
)

type Metrics struct {
    mu sync.RWMutex
    
    // Counters
    apiCalls       map[string]int64
    apiErrors      map[string]int64
    
    // Gauges
    cacheSize      int64
    activeRequests int64
    
    // Histograms (simplified)
    apiLatencies   map[string][]time.Duration
}

func NewMetrics() *Metrics {
    return &Metrics{
        apiCalls:     make(map[string]int64),
        apiErrors:    make(map[string]int64),
        apiLatencies: make(map[string][]time.Duration),
    }
}

func (m *Metrics) RecordAPICall(endpoint string, duration time.Duration, err error) {
    m.mu.Lock()
    defer m.mu.Unlock()
    
    m.apiCalls[endpoint]++
    m.apiLatencies[endpoint] = append(m.apiLatencies[endpoint], duration)
    
    if err != nil {
        m.apiErrors[endpoint]++
    }
}

func (m *Metrics) SetCacheSize(size int64) {
    m.mu.Lock()
    defer m.mu.Unlock()
    m.cacheSize = size
}

func (m *Metrics) IncrementActiveRequests() {
    m.mu.Lock()
    defer m.mu.Unlock()
    m.activeRequests++
}

func (m *Metrics) DecrementActiveRequests() {
    m.mu.Lock()
    defer m.mu.Unlock()
    m.activeRequests--
}

func (m *Metrics) Snapshot() MetricsSnapshot {
    m.mu.RLock()
    defer m.mu.RUnlock()
    
    snapshot := MetricsSnapshot{
        APICallCounts:  make(map[string]int64),
        APIErrorCounts: make(map[string]int64),
        APIAvgLatency:  make(map[string]time.Duration),
        CacheSize:      m.cacheSize,
        ActiveRequests: m.activeRequests,
    }
    
    for k, v := range m.apiCalls {
        snapshot.APICallCounts[k] = v
    }
    
    for k, v := range m.apiErrors {
        snapshot.APIErrorCounts[k] = v
    }
    
    for k, latencies := range m.apiLatencies {
        if len(latencies) > 0 {
            var total time.Duration
            for _, l := range latencies {
                total += l
            }
            snapshot.APIAvgLatency[k] = total / time.Duration(len(latencies))
        }
    }
    
    return snapshot
}

type MetricsSnapshot struct {
    APICallCounts  map[string]int64
    APIErrorCounts map[string]int64
    APIAvgLatency  map[string]time.Duration
    CacheSize      int64
    ActiveRequests int64
}
```

---

## Build & Development

### Makefile

```makefile
# Makefile for Yukti

# Variables
BINARY_NAME=yukti
VERSION=$(shell git describe --tags --always --dirty)
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME)"

GO=go
GOTEST=$(GO) test
GOLINT=golangci-lint

# Directories
BUILD_DIR=./build
DOCS_DIR=./docs

# Default target
.DEFAULT_GOAL := help

##@ Development

.PHONY: dev
dev: ## Run in development mode with live reload
	@air -c .air.toml

.PHONY: run
run: ## Run the application
	@$(GO) run ./cmd/yukti

.PHONY: build
build: ## Build the binary
	@echo "Building $(BINARY_NAME)..."
	@$(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/yukti

.PHONY: build-all
build-all: ## Build for all platforms
	@echo "Building for all platforms..."
	@GOOS=darwin GOARCH=amd64 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 ./cmd/yukti
	@GOOS=darwin GOARCH=arm64 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 ./cmd/yukti
	@GOOS=linux GOARCH=amd64 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 ./cmd/yukti
	@GOOS=linux GOARCH=arm64 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 ./cmd/yukti
	@GOOS=windows GOARCH=amd64 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe ./cmd/yukti

.PHONY: install
install: build ## Install the binary
	@cp $(BUILD_DIR)/$(BINARY_NAME) $(GOPATH)/bin/

##@ Testing

.PHONY: test
test: ## Run unit tests
	@$(GOTEST) -v -race -coverprofile=coverage.out ./...

.PHONY: test-short
test-short: ## Run short tests only
	@$(GOTEST) -v -short ./...

.PHONY: test-integration
test-integration: ## Run integration tests
	@$(GOTEST) -v -tags=integration ./tests/integration/...

.PHONY: coverage
coverage: test ## Generate coverage report
	@$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

.PHONY: benchmark
benchmark: ## Run benchmarks
	@$(GOTEST) -bench=. -benchmem ./...

##@ Code Quality

.PHONY: lint
lint: ## Run linters
	@$(GOLINT) run ./...

.PHONY: fmt
fmt: ## Format code
	@$(GO) fmt ./...
	@goimports -w .

.PHONY: vet
vet: ## Run go vet
	@$(GO) vet ./...

.PHONY: tidy
tidy: ## Tidy go modules
	@$(GO) mod tidy

.PHONY: check
check: fmt vet lint test ## Run all checks

##@ CI

.PHONY: ci
ci: tidy check build ## Full CI pipeline
	@echo "CI passed!"

##@ Documentation

.PHONY: docs
docs: ## Generate documentation
	@$(GO) doc -all ./... > $(DOCS_DIR)/api.txt

##@ Release

.PHONY: release
release: ## Create a release with goreleaser
	@goreleaser release --clean

.PHONY: snapshot
snapshot: ## Create a snapshot release (no publish)
	@goreleaser release --snapshot --clean

##@ Utilities

.PHONY: clean
clean: ## Clean build artifacts
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html

.PHONY: deps
deps: ## Download dependencies
	@$(GO) mod download

.PHONY: update-deps
update-deps: ## Update dependencies
	@$(GO) get -u ./...
	@$(GO) mod tidy

.PHONY: help
help: ## Display this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)
```

### CI/CD Pipeline

```yaml
# .github/workflows/ci.yml
name: CI

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.25'
      
      - name: Cache Go modules
        uses: actions/cache@v4
        with:
          path: ~/go/pkg/mod
          key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}
          restore-keys: |
            ${{ runner.os }}-go-
      
      - name: Download dependencies
        run: make deps
      
      - name: Run tests
        run: make test
      
      - name: Upload coverage
        uses: codecov/codecov-action@v4
        with:
          files: coverage.out

  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.25'
      
      - name: golangci-lint
        uses: golangci/golangci-lint-action@v6
        with:
          version: latest

  build:
    runs-on: ubuntu-latest
    needs: [test, lint]
    strategy:
      matrix:
        goos: [linux, darwin, windows]
        goarch: [amd64, arm64]
        exclude:
          - goos: windows
            goarch: arm64
    steps:
      - uses: actions/checkout@v4
      
      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.25'
      
      - name: Build
        env:
          GOOS: ${{ matrix.goos }}
          GOARCH: ${{ matrix.goarch }}
        run: |
          make build
          mv build/yukti build/yukti-${{ matrix.goos }}-${{ matrix.goarch }}${{ matrix.goos == 'windows' && '.exe' || '' }}
      
      - name: Upload artifact
        uses: actions/upload-artifact@v4
        with:
          name: yukti-${{ matrix.goos }}-${{ matrix.goarch }}
          path: build/yukti-*
```

---

## Implementation Phases

### Phase 1: Foundation ✅ COMPLETED (January 11, 2026)

**Goals:**
- Project scaffolding
- OAuth2 authentication
- Basic API client
- Minimal TUI shell

**Tasks:**
1. ✅ Initialize Go module and project structure
2. ✅ Implement OAuth2 with PKCE
3. ✅ Create keychain integration (macOS first, Linux/Windows file-based)
4. ✅ Build basic HTTP client wrapper
5. ✅ Create minimal BubbleTea application shell
6. ✅ Implement login flow

**Deliverables:**
- ✅ Working authentication flow
- ✅ Basic TUI with welcome screen and login view

**Implementation Notes:**
- Go 1.25.5 with golangci-lint v2.1.6
- Catppuccin Mocha color theme (brighter, warmer than original Tokyo Night)
- Vim-style keybindings (hjkl navigation)
- Stack-based router for view navigation
- macOS Keychain via `github.com/keybase/go-keychain`
- File-based token storage alternative (avoids keychain prompts on rebuild)
- Config stored at `~/Library/Application Support/yukti/config.json` (macOS) or `~/.config/yukti/config.json`
- OAuth credentials required: `client_id` and `client_secret`
- CLI commands implemented: `init`, `login`, `logout`, `status`, `version`
- Setup wizard (`yukti init`) guides users through OAuth credential setup
- Token storage configurable via `--token-file` flag or `token_file` config option
- Welcome screen: Logo box with rounded border, feature list with icons, version info
- Status bar: Proper spacing with separators, styled key badges
- Header: Shows auth status indicator (green dot = logged in, gray = logged out)
- iTerm2 automation scripts for TUI testing (`.claude/automations/test_tui.py`)

### Phase 2: Core Views (Week 3-4)

**Goals:**
- Project list view
- Project detail view
- File viewer

**Tasks:**
1. Implement project repository
2. Build project list view with navigation
3. Create project detail view
4. Implement code viewer with syntax highlighting
5. Add keyboard navigation
6. Implement search/filter

**Deliverables:**
- Fully functional project browser
- Code viewing with syntax highlighting

**Implementation Notes (January 11, 2026):**
- **API Discovery:** Apps Script API lacks a `projects.list` endpoint. Uses Google Drive API with `mimeType='application/vnd.google-apps.script'` query to list standalone projects.
- **Required APIs:** Both Apps Script API and Google Drive API must be enabled in Google Cloud Console
- **OAuth Scopes:** Added `https://www.googleapis.com/auth/drive.readonly` for project listing
- **Project Repository:** `internal/infrastructure/google/project_repo.go` - Uses Drive API for List, Apps Script API for Get/GetContent
- **Views Implemented:**
  - `views/projects.go` - Project list with bubbles/list component
  - `views/project_detail.go` - File list with metadata (type, line count, functions)
  - `views/code_viewer.go` - Syntax highlighting via chroma library, vim-style navigation
- **Navigation Pattern:** Stack-based navigation using ViewFactory pattern to avoid circular imports
- **Known Limitation:** Container-bound scripts (attached to Sheets/Docs) not visible via Drive API - only standalone scripts appear

**Future Enhancements:**
- Split-pane layout (file tree left, code right) - per PRD F3 mockup
- Tab switching between panes
- Live file editing
- GAS API autocomplete (stretch goal F12)

### Phase 3: Operations (Week 5-6)

**Goals:**
- Push/pull operations
- Deployment management
- Version management

**Tasks:**
1. Implement content sync (push/pull)
2. Build deployment management views
3. Create version management
4. Add progress indicators
5. Implement error handling and rollback

**Deliverables:**
- Full CRUD for deployments and versions
- Reliable push/pull operations

### Phase 4: Advanced Features (Week 7-8)

**Goals:**
- Execution logs
- Metrics dashboard
- Command palette
- Offline mode

**Tasks:**
1. Implement process/log viewer
2. Build metrics dashboard with ASCII charts
3. Create command palette
4. Implement local caching for offline mode
5. Add sync queue for offline changes

**Deliverables:**
- Complete feature set
- Offline capability

### Phase 5: Polish & Release (Week 9+)

**Goals:**
- Performance optimization
- Testing completion
- Documentation
- Release

**Tasks:**
1. Performance profiling and optimization
2. Complete test coverage
3. Write user documentation
4. Create release pipeline
5. Beta testing
6. v1.0 release

**Deliverables:**
- Production-ready release
- Comprehensive documentation

---

## UI Mockups

### HTML Mockup: Project List

```html
<!DOCTYPE html>
<html>
<head>
    <style>
        body {
            background: #1A1B26;
            color: #C0CAF5;
            font-family: 'JetBrains Mono', monospace;
            padding: 20px;
            line-height: 1.4;
        }
        .container {
            max-width: 900px;
            margin: 0 auto;
            border: 1px solid #414868;
            border-radius: 8px;
            overflow: hidden;
        }
        .header {
            background: #24283B;
            padding: 8px 16px;
            border-bottom: 1px solid #414868;
            display: flex;
            justify-content: space-between;
        }
        .logo { color: #7C3AED; font-weight: bold; }
        .user { color: #9AA5CE; }
        .search-bar {
            padding: 12px 16px;
            border-bottom: 1px solid #414868;
        }
        .search-input {
            background: #24283B;
            border: 1px solid #414868;
            border-radius: 4px;
            padding: 8px 12px;
            color: #C0CAF5;
            width: 300px;
        }
        .content { padding: 16px; }
        .title-row {
            display: flex;
            justify-content: space-between;
            margin-bottom: 16px;
        }
        .section-title { font-weight: bold; }
        .sort { color: #9AA5CE; }
        .project-list { list-style: none; padding: 0; margin: 0; }
        .project-item {
            padding: 12px;
            border-radius: 4px;
            margin-bottom: 4px;
            cursor: pointer;
        }
        .project-item:hover { background: #24283B; }
        .project-item.selected { 
            background: #414868; 
            border-left: 3px solid #7C3AED;
        }
        .project-title { display: flex; align-items: center; gap: 8px; }
        .project-meta { 
            color: #565F89; 
            font-size: 0.85em;
            margin-top: 4px;
        }
        .footer {
            background: #24283B;
            padding: 8px 16px;
            border-top: 1px solid #414868;
            color: #565F89;
            font-size: 0.85em;
        }
        .key { 
            background: #414868; 
            padding: 2px 6px; 
            border-radius: 3px;
            margin-right: 4px;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <span class="logo">⚡ Yukti</span>
            <span class="user">user@example.com</span>
        </div>
        <div class="search-bar">
            <input type="text" class="search-input" placeholder="🔍 Search projects...">
        </div>
        <div class="content">
            <div class="title-row">
                <span class="section-title">My Projects (47)</span>
                <span class="sort">Sort: Modified ▼</span>
            </div>
            <ul class="project-list">
                <li class="project-item selected">
                    <div class="project-title">📄 Sample Scripts</div>
                    <div class="project-meta">Me • Today, 4:57 PM</div>
                </li>
                <li class="project-item">
                    <div class="project-title">📄 DocMailer</div>
                    <div class="project-meta">Me • Aug 16, 2021</div>
                </li>
                <li class="project-item">
                    <div class="project-title">📊 Sales Dashboard Automation</div>
                    <div class="project-meta">Team • Dec 3, 2025</div>
                </li>
                <li class="project-item">
                    <div class="project-title">📝 Form Response Handler</div>
                    <div class="project-meta">Me • Nov 28, 2025</div>
                </li>
                <li class="project-item">
                    <div class="project-title">📧 Email Scheduler</div>
                    <div class="project-meta">Me • Nov 15, 2025</div>
                </li>
            </ul>
        </div>
        <div class="footer">
            <span class="key">↵</span>Open
            <span class="key">n</span>New
            <span class="key">d</span>Delete
            <span class="key">/</span>Search
            <span class="key">?</span>Help
            <span class="key">q</span>Quit
        </div>
    </div>
</body>
</html>
```

---

## Appendix: CLAUDE.md Template

This file captures learnings during development:

```markdown
# CLAUDE.md — Yukti Development Learnings

## Overview
This document captures learnings, fixes, and patterns discovered during Yukti development.
Future AI sessions should read this file to avoid repeating mistakes.

## Git Conventions
- Always create frequent, atomic, relevant, one-liner conventional commits. Commit early, commit often.
- Never bulk-add files etc (`git add -A` etc). Always explicitly enumerate the files to be staged/committed.

## API Learnings

### Google Apps Script API

**Rate Limits:**
- Projects API: 5000 requests/day
- Deployments API: 1000 requests/day
- Content updates: 50/minute

**Quirks:**
- `getContent` returns files in arbitrary order
- Empty projects have one file: `appsscript.json`
- Bound scripts require `parentId` in creation request

## Code Patterns

### BubbleTea Best Practices

1. Always return `tea.Cmd` from Update, never block
2. Use channels for long-running operations
3. Handle `tea.WindowSizeMsg` early in Update
4. Propagate size changes to child components

### LipGloss Styling

1. Create styles once, not in View()
2. Use AdaptiveColor for light/dark themes
3. Calculate widths dynamically from terminal size

## Bug Fixes

### Issue: OAuth token not refreshing
**Symptom:** 401 errors after 1 hour
**Fix:** Use oauth2.TokenSource wrapper, not raw token
**Date:** 2026-01-XX

### Issue: Viewport not scrolling
**Symptom:** Code viewer stuck at top
**Fix:** Must call viewport.SetContent() after size change
**Date:** 2026-01-XX

## Performance Notes

- Project list: Pagination required for >100 projects
- Code viewer: 5000+ lines causes noticeable lag
- Syntax highlighting: Cache highlighted output

## Testing Notes

- `teatest` requires explicit `tea.Quit` to finish
- iTerm2 driver automation scripts MUST have proper cleanup code so iTerm2 terminal is properly cleaned up, running TUI apps are properly stopped etc.
- Mock repositories should implement full interface
```

---

*This is a living document. Update it as implementation progresses.*
