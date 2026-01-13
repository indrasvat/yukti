// Package process provides the application service for script execution.
package process

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"yukti/internal/domain/process"
	"yukti/internal/infrastructure/google"
	"yukti/internal/infrastructure/logger"
)

// Service manages script execution and tracks execution history.
type Service struct {
	runner         *google.ScriptRunner
	loggingService *google.CloudLoggingService
	gcpProjectNum  string // GCP project number for Cloud Logging API

	// In-memory execution history (per script)
	mu         sync.RWMutex
	executions map[string][]ExecutionEntry // scriptID -> entries
	maxEntries int                         // Max entries per script
}

// ExecutionEntry represents a tracked execution with its result.
type ExecutionEntry struct {
	ID           string
	FunctionName string
	Status       process.Status
	StartTime    time.Time
	Duration     time.Duration
	Result       any    // Return value or nil
	Error        string // Error message if failed
	ScriptID     string

	// Console log entries (lazy loaded via Cloud Logging API)
	Logs         []google.LogEntry
	LogsLoaded   bool   // Whether logs have been fetched
	LogsError    string // Error message if log fetching failed
	LogsExpanded bool   // Whether logs are expanded in UI
}

// NewService creates a new process service.
func NewService(runner *google.ScriptRunner, loggingService *google.CloudLoggingService, gcpProjectNum string) *Service {
	return &Service{
		runner:         runner,
		loggingService: loggingService,
		gcpProjectNum:  gcpProjectNum,
		executions:     make(map[string][]ExecutionEntry),
		maxEntries:     50,
	}
}

// RunFunction executes a function and returns the result.
// The execution is tracked in the in-memory history.
func (s *Service) RunFunction(ctx context.Context, scriptID, functionName string) (*ExecutionEntry, error) {
	logger.Info("Running function %s in script %s", functionName, scriptID)

	// Create the entry
	entry := &ExecutionEntry{
		ID:           fmt.Sprintf("%d", time.Now().UnixNano()),
		FunctionName: functionName,
		Status:       process.StatusRunning,
		StartTime:    time.Now(),
		ScriptID:     scriptID,
	}

	// Add to history as running
	s.addEntry(scriptID, *entry)

	// Execute the function
	result, err := s.runner.Run(ctx, scriptID, functionName, nil)
	entry.Duration = time.Since(entry.StartTime)

	if err != nil {
		entry.Status = process.StatusFailed
		entry.Error = err.Error()
		logger.Error("Function %s failed (API error): %s", functionName, err.Error())
		s.updateEntry(scriptID, entry.ID, *entry)
		return entry, nil //nolint:nilerr // Intentionally return entry with nil error so UI can display failure
	}

	// Check for script execution errors
	if result.Error != nil {
		entry.Status = process.StatusFailed
		entry.Error = formatScriptError(result.Error)
		logger.Error("Function %s failed (script error): %s", functionName, entry.Error)
		s.updateEntry(scriptID, entry.ID, *entry)
		return entry, nil
	}

	// Success
	entry.Status = process.StatusCompleted
	if result.Response != nil {
		entry.Result = result.Response.Result
	}
	logger.Info("Function %s completed in %v", functionName, entry.Duration)
	s.updateEntry(scriptID, entry.ID, *entry)

	return entry, nil
}

// GetRecentExecutions returns recent execution entries for a script.
func (s *Service) GetRecentExecutions(scriptID string, limit int) []ExecutionEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entries := s.executions[scriptID]
	if len(entries) == 0 {
		return nil
	}

	if limit <= 0 || limit > len(entries) {
		limit = len(entries)
	}

	// Return most recent first (entries are stored oldest first)
	result := make([]ExecutionEntry, limit)
	for i := 0; i < limit; i++ {
		result[i] = entries[len(entries)-1-i]
	}

	return result
}

// FetchLogsForEntry fetches console logs for an execution entry from Cloud Logging API.
// It updates the entry in-place with the fetched logs.
func (s *Service) FetchLogsForEntry(ctx context.Context, entry *ExecutionEntry, limit int) error {
	if s.loggingService == nil {
		entry.LogsError = "Cloud Logging not configured"
		return nil
	}

	if s.gcpProjectNum == "" {
		entry.LogsError = "GCP project not configured"
		return nil
	}

	// Determine time range for log query
	// Fetch logs from (startTime - 1s) to (startTime + duration + 1s)
	startTime := entry.StartTime.Add(-1 * time.Second)
	endTime := entry.StartTime.Add(entry.Duration + 1*time.Second)

	// If still running, use current time
	if entry.Status == process.StatusRunning {
		endTime = time.Now().Add(1 * time.Second)
	}

	pageSize := limit
	if pageSize <= 0 {
		pageSize = 100
	}

	req := google.FetchLogsRequest{
		GCPProjectNumber: s.gcpProjectNum,
		FunctionName:     entry.FunctionName,
		StartTime:        startTime,
		EndTime:          endTime,
		PageSize:         pageSize,
	}

	resp, err := s.loggingService.FetchLogs(ctx, req)
	if err != nil {
		entry.LogsError = err.Error()
		entry.LogsLoaded = true
		logger.Error("Failed to fetch logs for %s: %s", entry.FunctionName, err.Error())
		return nil // Don't fail the whole operation, just mark error
	}

	entry.Logs = resp.Entries
	entry.LogsLoaded = true
	entry.LogsError = ""

	logger.Info("Fetched %d logs for %s", len(resp.Entries), entry.FunctionName)
	return nil
}

// FetchAllLogsForEntry fetches all console logs with pagination.
func (s *Service) FetchAllLogsForEntry(ctx context.Context, entry *ExecutionEntry) error {
	if s.loggingService == nil {
		entry.LogsError = "Cloud Logging not configured"
		return nil
	}

	if s.gcpProjectNum == "" {
		entry.LogsError = "GCP project not configured"
		return nil
	}

	startTime := entry.StartTime.Add(-1 * time.Second)
	endTime := entry.StartTime.Add(entry.Duration + 1*time.Second)

	if entry.Status == process.StatusRunning {
		endTime = time.Now().Add(1 * time.Second)
	}

	req := google.FetchLogsRequest{
		GCPProjectNumber: s.gcpProjectNum,
		FunctionName:     entry.FunctionName,
		StartTime:        startTime,
		EndTime:          endTime,
	}

	logs, err := s.loggingService.FetchAllLogs(ctx, req)
	if err != nil {
		entry.LogsError = err.Error()
		entry.LogsLoaded = true
		return nil //nolint:nilerr // Intentionally return nil - error is stored in entry for UI display
	}

	entry.Logs = logs
	entry.LogsLoaded = true
	entry.LogsError = ""

	return nil
}

// ToggleLogsExpanded toggles the log expansion state for an entry.
func (s *Service) ToggleLogsExpanded(scriptID, entryID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	entries := s.executions[scriptID]
	for i := range entries {
		if entries[i].ID == entryID {
			entries[i].LogsExpanded = !entries[i].LogsExpanded
			s.executions[scriptID] = entries
			return
		}
	}
}

// GetEntryByID returns an execution entry by ID.
func (s *Service) GetEntryByID(scriptID, entryID string) *ExecutionEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entries := s.executions[scriptID]
	for i := range entries {
		if entries[i].ID == entryID {
			return &entries[i]
		}
	}
	return nil
}

// addEntry adds a new entry to the history.
func (s *Service) addEntry(scriptID string, entry ExecutionEntry) {
	s.mu.Lock()
	defer s.mu.Unlock()

	entries := s.executions[scriptID]

	// Trim to max entries
	if len(entries) >= s.maxEntries {
		entries = entries[1:]
	}

	entries = append(entries, entry)
	s.executions[scriptID] = entries
}

// updateEntry updates an existing entry by ID.
func (s *Service) updateEntry(scriptID, entryID string, updated ExecutionEntry) {
	s.mu.Lock()
	defer s.mu.Unlock()

	entries := s.executions[scriptID]
	for i := range entries {
		if entries[i].ID == entryID {
			entries[i] = updated
			s.executions[scriptID] = entries
			return
		}
	}
}

// formatScriptError formats a script execution error for display.
// It categorizes common GAS errors and provides user-friendly messages.
func formatScriptError(err *google.ExecutionError) string {
	if err == nil {
		return ""
	}

	// Check for common GAS error patterns by status/code
	if msg := formatErrorByStatusCode(err); msg != "" {
		return msg
	}

	// Check for specific error messages in the message field
	if msg := formatErrorByMessage(err.Message); msg != "" {
		return msg
	}

	// Try to get detailed error message from details
	if msg := formatErrorFromDetails(err.Details); msg != "" {
		return msg
	}

	return err.Message
}

// formatErrorByStatusCode returns a user-friendly message based on HTTP status codes.
func formatErrorByStatusCode(err *google.ExecutionError) string {
	switch {
	case err.Status == "PERMISSION_DENIED" || err.Code == 403:
		return "Script needs additional permissions - run in browser first"
	case err.Status == "UNAUTHENTICATED" || err.Code == 401:
		return "Re-authenticate: yukti logout && yukti login"
	case err.Status == "DEADLINE_EXCEEDED" || err.Code == 504:
		return "Execution timed out (6 min limit)"
	case err.Status == "NOT_FOUND" || err.Code == 404:
		return "Script not found - ensure script is linked to the same GCP project as your OAuth credentials"
	case err.Status == "RESOURCE_EXHAUSTED" || err.Code == 429:
		return "Rate limit exceeded - try again in a moment"
	default:
		return ""
	}
}

// formatErrorByMessage returns a user-friendly message based on error message content.
func formatErrorByMessage(message string) string {
	msg := strings.ToLower(message)
	switch {
	case strings.Contains(msg, "permission"):
		return "Script needs permissions - run in browser first to authorize"
	case strings.Contains(msg, "not found"):
		return "Script not found - check that script is deployed and linked to your GCP project"
	case strings.Contains(msg, "timeout") || strings.Contains(msg, "timed out"):
		return "Execution timed out (6 min limit)"
	case strings.Contains(msg, "authorization"):
		return "Authorization required - complete setup in browser first"
	default:
		return ""
	}
}

// formatErrorFromDetails extracts error message from execution error details.
func formatErrorFromDetails(details []google.ExecutionErrorDetail) string {
	for _, detail := range details {
		if detail.ErrorMessage == "" {
			continue
		}
		detailMsg := detail.ErrorMessage
		if detail.ErrorType != "" {
			detailMsg = detail.ErrorType + ": " + detailMsg
		}
		// Add hint for common script errors
		if strings.Contains(strings.ToLower(detailMsg), "you do not have permission") {
			return detailMsg + " (run in browser to grant permissions)"
		}
		return detailMsg
	}
	return ""
}

// FormatResult formats a result value for display.
func FormatResult(result any) string {
	if result == nil {
		return "undefined"
	}

	// Try to pretty-print JSON
	b, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Sprintf("%v", result)
	}

	return string(b)
}

// FormatResultCompact formats a result value for compact display.
func FormatResultCompact(result any) string {
	if result == nil {
		return "undefined"
	}

	b, err := json.Marshal(result)
	if err != nil {
		return fmt.Sprintf("%v", result)
	}

	// Truncate if too long
	str := string(b)
	if len(str) > 60 {
		str = str[:57] + "..."
	}

	return str
}
