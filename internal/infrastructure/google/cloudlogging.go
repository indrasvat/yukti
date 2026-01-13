// Package google provides the Cloud Logging API client for fetching console.log output.
package google

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"golang.org/x/oauth2"
)

// CloudLoggingBaseURL is the Cloud Logging API base URL.
const CloudLoggingBaseURL = "https://logging.googleapis.com/v2"

// Common Cloud Logging errors.
var (
	ErrNoLogsFound       = errors.New("no logs found")
	ErrScopeInsufficient = errors.New("logging.read scope not authorized - re-authenticate with: yukti logout && yukti login")
	ErrProjectNotLinked  = errors.New("script not linked to GCP project")
)

// LogSeverity represents the severity level of a log entry.
type LogSeverity string

const (
	SeverityDefault   LogSeverity = "DEFAULT"
	SeverityDebug     LogSeverity = "DEBUG"
	SeverityInfo      LogSeverity = "INFO"
	SeverityNotice    LogSeverity = "NOTICE"
	SeverityWarning   LogSeverity = "WARNING"
	SeverityError     LogSeverity = "ERROR"
	SeverityCritical  LogSeverity = "CRITICAL"
	SeverityAlert     LogSeverity = "ALERT"
	SeverityEmergency LogSeverity = "EMERGENCY"
)

// LogEntry represents a parsed Cloud Logging entry.
type LogEntry struct {
	Timestamp    time.Time
	Severity     LogSeverity
	FunctionName string
	Message      string
	StackTrace   string // For ERROR entries with stack traces
}

// FetchLogsRequest contains parameters for fetching logs.
type FetchLogsRequest struct {
	GCPProjectNumber string    // GCP project number (from OAuth client ID)
	FunctionName     string    // Optional: filter by function name
	StartTime        time.Time // Start of time range
	EndTime          time.Time // End of time range
	PageSize         int       // Max entries per request (default 100)
	PageToken        string    // For pagination
}

// FetchLogsResponse contains the fetched logs and pagination info.
type FetchLogsResponse struct {
	Entries       []LogEntry
	NextPageToken string
	TotalEntries  int
}

// CloudLoggingService provides methods to fetch logs from Cloud Logging API.
type CloudLoggingService struct {
	httpClient *http.Client
	logger     *slog.Logger
}

// NewCloudLoggingService creates a new Cloud Logging service client.
func NewCloudLoggingService(ctx context.Context, tokenSource oauth2.TokenSource, logger *slog.Logger) *CloudLoggingService {
	return &CloudLoggingService{
		httpClient: oauth2.NewClient(ctx, tokenSource),
		logger:     logger,
	}
}

// FetchLogs retrieves log entries from Cloud Logging API.
func (s *CloudLoggingService) FetchLogs(ctx context.Context, req FetchLogsRequest) (*FetchLogsResponse, error) {
	if req.GCPProjectNumber == "" {
		return nil, ErrProjectNotLinked
	}

	if req.PageSize <= 0 {
		req.PageSize = 100
	}

	// Build the filter query
	// Apps Script logs use resource.type = "app_script_function"
	// Note: Try without function name filter first to verify logs exist
	filter := `resource.type="app_script_function"`

	// Use wider time range to account for Cloud Logging ingestion delay
	if !req.StartTime.IsZero() {
		// Extend start time by 5 minutes to catch delayed logs
		extendedStart := req.StartTime.Add(-5 * time.Minute)
		filter += ` AND timestamp>="` + extendedStart.Format(time.RFC3339) + `"`
	}
	if !req.EndTime.IsZero() {
		// Extend end time by 1 minute
		extendedEnd := req.EndTime.Add(1 * time.Minute)
		filter += ` AND timestamp<="` + extendedEnd.Format(time.RFC3339) + `"`
	}
	// Skip function name filter for now - will filter in code if needed
	// Many Apps Script logs don't have labels.function_name set
	_ = req.FunctionName // Acknowledge unused for now

	// Build the request body
	requestBody := map[string]any{
		"resourceNames": []string{
			fmt.Sprintf("projects/%s", req.GCPProjectNumber),
		},
		"filter":   filter,
		"orderBy":  "timestamp asc",
		"pageSize": req.PageSize,
	}

	if req.PageToken != "" {
		requestBody["pageToken"] = req.PageToken
	}

	s.logger.Debug("Fetching logs",
		slog.String("project", req.GCPProjectNumber),
		slog.String("filter", filter),
		slog.Int("pageSize", req.PageSize),
	)

	// Make the API request
	body, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("marshaling request: %w", err)
	}

	url := CloudLoggingBaseURL + "/entries:list"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// Handle errors
	if resp.StatusCode >= 400 {
		return nil, s.handleError(resp)
	}

	// Parse response
	var apiResp cloudLoggingAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	// Convert API entries to our LogEntry type
	entries := make([]LogEntry, 0, len(apiResp.Entries))
	for _, e := range apiResp.Entries {
		entry := parseLogEntry(e)
		entries = append(entries, entry)
	}

	s.logger.Debug("Fetched logs",
		slog.Int("count", len(entries)),
		slog.String("nextPageToken", apiResp.NextPageToken),
	)

	return &FetchLogsResponse{
		Entries:       entries,
		NextPageToken: apiResp.NextPageToken,
		TotalEntries:  len(entries),
	}, nil
}

// FetchAllLogs fetches all logs with automatic pagination.
func (s *CloudLoggingService) FetchAllLogs(ctx context.Context, req FetchLogsRequest) ([]LogEntry, error) {
	var allEntries []LogEntry
	pageToken := ""

	for {
		req.PageToken = pageToken
		resp, err := s.FetchLogs(ctx, req)
		if err != nil {
			return nil, err
		}

		allEntries = append(allEntries, resp.Entries...)

		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken

		// Safety limit to prevent infinite loops
		if len(allEntries) > 10000 {
			s.logger.Warn("Reached safety limit of 10000 log entries")
			break
		}
	}

	return allEntries, nil
}

// handleError processes API error responses.
func (s *CloudLoggingService) handleError(resp *http.Response) error {
	body, _ := io.ReadAll(resp.Body)

	s.logger.Debug("Cloud Logging API error",
		slog.Int("status", resp.StatusCode),
		slog.String("body", string(body)),
	)

	// Check for scope issues
	if resp.StatusCode == http.StatusForbidden {
		var errResp struct {
			Error struct {
				Code    int    `json:"code"`
				Message string `json:"message"`
				Status  string `json:"status"`
			} `json:"error"`
		}
		if json.Unmarshal(body, &errResp) == nil {
			if errResp.Error.Status == "PERMISSION_DENIED" {
				return ErrScopeInsufficient
			}
		}
		return fmt.Errorf("%w: %s", ErrForbidden, string(body))
	}

	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("%w: %s", ErrUnauthorized, string(body))
	}

	return fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
}

// cloudLoggingAPIResponse represents the Cloud Logging API response.
type cloudLoggingAPIResponse struct {
	Entries       []cloudLoggingEntry `json:"entries"`
	NextPageToken string              `json:"nextPageToken"`
}

// cloudLoggingEntry represents a raw log entry from the API.
type cloudLoggingEntry struct {
	Timestamp   string            `json:"timestamp"`
	Severity    string            `json:"severity"`
	TextPayload string            `json:"textPayload"`
	JSONPayload map[string]any    `json:"jsonPayload"`
	Labels      map[string]string `json:"labels"`
	Resource    struct {
		Type   string            `json:"type"`
		Labels map[string]string `json:"labels"`
	} `json:"resource"`
}

// parseLogEntry converts a raw API entry to our LogEntry type.
func parseLogEntry(raw cloudLoggingEntry) LogEntry {
	entry := LogEntry{
		Severity: LogSeverity(raw.Severity),
	}

	// Parse timestamp
	if t, err := time.Parse(time.RFC3339Nano, raw.Timestamp); err == nil {
		entry.Timestamp = t
	} else if t, err := time.Parse(time.RFC3339, raw.Timestamp); err == nil {
		entry.Timestamp = t
	}

	// Get function name from labels
	if fn, ok := raw.Labels["function_name"]; ok {
		entry.FunctionName = fn
	}

	// Get message - prefer textPayload, fall back to jsonPayload
	if raw.TextPayload != "" {
		entry.Message = raw.TextPayload
	} else if raw.JSONPayload != nil {
		// Try common message fields
		if msg, ok := raw.JSONPayload["message"].(string); ok {
			entry.Message = msg
		} else if msg, ok := raw.JSONPayload["msg"].(string); ok {
			entry.Message = msg
		} else {
			// Serialize the whole payload
			if b, err := json.Marshal(raw.JSONPayload); err == nil {
				entry.Message = string(b)
			}
		}

		// Check for stack trace
		if stack, ok := raw.JSONPayload["stack_trace"].(string); ok {
			entry.StackTrace = stack
		} else if stack, ok := raw.JSONPayload["stackTrace"].(string); ok {
			entry.StackTrace = stack
		}
	}

	return entry
}

// SeverityIcon returns the display icon for a log severity.
func SeverityIcon(severity LogSeverity) string {
	switch severity {
	case SeverityDebug, SeverityInfo, SeverityDefault:
		return "ℹ️"
	case SeverityNotice:
		return "📋"
	case SeverityWarning:
		return "⚠️"
	case SeverityError, SeverityCritical, SeverityAlert, SeverityEmergency:
		return "❌"
	default:
		return "ℹ️"
	}
}

// SeverityColor returns the hex color code for a log severity (Catppuccin Mocha).
func SeverityColor(severity LogSeverity) string {
	switch severity {
	case SeverityDebug, SeverityInfo, SeverityDefault:
		return "#89B4FA" // Blue
	case SeverityNotice:
		return "#F9E2AF" // Yellow
	case SeverityWarning:
		return "#FAB387" // Orange/Peach
	case SeverityError, SeverityCritical, SeverityAlert, SeverityEmergency:
		return "#F38BA8" // Red
	default:
		return "#CDD6F4" // Default text
	}
}
