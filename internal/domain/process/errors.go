package process

import "errors"

// Domain errors for process operations.
var (
	ErrNotFound     = errors.New("process not found")
	ErrAccessDenied = errors.New("access denied to process")
)
