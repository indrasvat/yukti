// Package process provides the application service for script execution.
package process

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"yukti/internal/domain/process"
	"yukti/internal/infrastructure/google"
)

// Service manages script execution and tracks execution history.
type Service struct {
	runner *google.ScriptRunner

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
}

// NewService creates a new process service.
func NewService(runner *google.ScriptRunner) *Service {
	return &Service{
		runner:     runner,
		executions: make(map[string][]ExecutionEntry),
		maxEntries: 50,
	}
}

// RunFunction executes a function and returns the result.
// The execution is tracked in the in-memory history.
func (s *Service) RunFunction(ctx context.Context, scriptID, functionName string) (*ExecutionEntry, error) {
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
		s.updateEntry(scriptID, entry.ID, *entry)
		return entry, nil //nolint:nilerr // Intentionally return entry with nil error so UI can display failure
	}

	// Check for script execution errors
	if result.Error != nil {
		entry.Status = process.StatusFailed
		entry.Error = formatScriptError(result.Error)
		s.updateEntry(scriptID, entry.ID, *entry)
		return entry, nil
	}

	// Success
	entry.Status = process.StatusCompleted
	if result.Response != nil {
		entry.Result = result.Response.Result
	}
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
func formatScriptError(err *google.ExecutionError) string {
	if err == nil {
		return ""
	}

	// Try to get detailed error message from details
	for _, detail := range err.Details {
		if detail.ErrorMessage != "" {
			msg := detail.ErrorMessage
			if detail.ErrorType != "" {
				msg = detail.ErrorType + ": " + msg
			}
			return msg
		}
	}

	return err.Message
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
