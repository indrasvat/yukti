// Package process defines the process (execution) domain entity and related types.
package process

import "time"

// Process represents a script execution record.
type Process struct {
	ID            string
	ScriptID      string
	FunctionName  string
	Status        Status
	StartTime     time.Time
	Duration      time.Duration
	ExecutingUser string
	ProcessType   Type
}

// Status represents the execution status of a process.
type Status string

const (
	StatusRunning   Status = "RUNNING"
	StatusCompleted Status = "COMPLETED"
	StatusFailed    Status = "FAILED"
	StatusTimedOut  Status = "TIMED_OUT"
	StatusCanceled  Status = "CANCELED"
	StatusUnknown   Status = "UNKNOWN"
)

// Type represents the type of process execution.
type Type string

const (
	TypeEditor    Type = "EDITOR"
	TypeAddOn     Type = "ADD_ON"
	TypeWebApp    Type = "WEB_APP"
	TypeAPI       Type = "EXECUTION_API"
	TypeTrigger   Type = "TRIGGER"
	TypeTimeBased Type = "TIME_DRIVEN"
)

// IsTerminal returns true if the process has finished execution.
func (p *Process) IsTerminal() bool {
	switch p.Status {
	case StatusCompleted, StatusFailed, StatusTimedOut, StatusCanceled:
		return true
	default:
		return false
	}
}

// IsSuccess returns true if the process completed successfully.
func (p *Process) IsSuccess() bool {
	return p.Status == StatusCompleted
}

// IsError returns true if the process failed.
func (p *Process) IsError() bool {
	return p.Status == StatusFailed || p.Status == StatusTimedOut
}
