package project

import "errors"

// Domain errors for project operations.
var (
	ErrNotFound      = errors.New("project not found")
	ErrAccessDenied  = errors.New("access denied to project")
	ErrAlreadyExists = errors.New("project already exists")
	ErrInvalidTitle  = errors.New("invalid project title")
	ErrQuotaExceeded = errors.New("API quota exceeded")
)
