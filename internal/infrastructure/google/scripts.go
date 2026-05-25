package google

import (
	"context"
	"fmt"
	"net/url"
)

// ScriptRunner executes Apps Script functions via the Execution API.
type ScriptRunner struct {
	client *Client
}

// NewScriptRunner creates a new script runner.
func NewScriptRunner(client *Client) *ScriptRunner {
	return &ScriptRunner{client: client}
}

// ExecutionRequest is the request body for running a script function.
type ExecutionRequest struct {
	Function   string `json:"function"`
	Parameters []any  `json:"parameters,omitempty"`
	DevMode    bool   `json:"devMode,omitempty"`
}

// ExecutionResult is the response from running a script function.
type ExecutionResult struct {
	Done     bool               `json:"done"`
	Response *ExecutionResponse `json:"response,omitempty"`
	Error    *ExecutionError    `json:"error,omitempty"`
}

// ExecutionResponse contains the return value from a successful execution.
type ExecutionResponse struct {
	Type   string `json:"@type,omitempty"`
	Result any    `json:"result"`
}

// ExecutionError contains details about a failed execution.
type ExecutionError struct {
	Code    int                    `json:"code"`
	Message string                 `json:"message"`
	Status  string                 `json:"status,omitempty"`
	Details []ExecutionErrorDetail `json:"details,omitempty"`
}

// ExecutionErrorDetail contains additional error information.
type ExecutionErrorDetail struct {
	Type             string             `json:"@type,omitempty"`
	ErrorMessage     string             `json:"errorMessage,omitempty"`
	ErrorType        string             `json:"errorType,omitempty"`
	ScriptStacktrace []ScriptStackFrame `json:"scriptStackTraceElements,omitempty"`
}

// ScriptStackFrame represents a frame in the script's stack trace.
type ScriptStackFrame struct {
	Function   string `json:"function,omitempty"`
	LineNumber int    `json:"lineNumber,omitempty"`
}

// Run executes a function in the given script.
// POST /v1/scripts/{scriptId}:run
//
// The function must be exported (not private) and the script must be
// deployed for API execution. DevMode=true runs the most recent saved version
// instead of the deployed version.
func (r *ScriptRunner) Run(ctx context.Context, scriptID, functionName string, params []any) (*ExecutionResult, error) {
	path := fmt.Sprintf("/scripts/%s:run", url.PathEscape(scriptID))

	req := ExecutionRequest{
		Function:   functionName,
		Parameters: params,
		DevMode:    true, // Use saved (development) version for TUI execution
	}

	var result ExecutionResult
	if err := r.client.Post(ctx, path, req, &result); err != nil {
		return nil, fmt.Errorf("running script: %w", err)
	}

	return &result, nil
}

// RunWithOptions executes a function with additional options.
func (r *ScriptRunner) RunWithOptions(ctx context.Context, scriptID string, req ExecutionRequest) (*ExecutionResult, error) {
	path := fmt.Sprintf("/scripts/%s:run", url.PathEscape(scriptID))

	var result ExecutionResult
	if err := r.client.Post(ctx, path, req, &result); err != nil {
		return nil, fmt.Errorf("running script: %w", err)
	}

	return &result, nil
}
