package deployment

import "errors"

// Domain errors for deployment operations.
var (
	ErrNotFound        = errors.New("deployment not found")
	ErrVersionNotFound = errors.New("version not found")
	ErrAccessDenied    = errors.New("access denied to deployment")
	ErrInvalidConfig   = errors.New("invalid deployment configuration")
)
