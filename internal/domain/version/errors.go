package version

import "errors"

// Domain errors for version operations.
var (
	ErrNotFound     = errors.New("version not found")
	ErrAccessDenied = errors.New("access denied to version")
)
