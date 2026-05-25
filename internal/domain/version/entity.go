// Package version defines the version domain entity and related types.
package version

import (
	"fmt"
	"time"
)

// Version represents an immutable snapshot of a project's code.
type Version struct {
	VersionNumber int
	Description   string
	CreateTime    time.Time
	ScriptID      string
}

// String returns a human-readable version string.
func (v *Version) String() string {
	if v.VersionNumber == 0 {
		return "HEAD"
	}
	return fmt.Sprintf("v%d", v.VersionNumber)
}

// IsHead returns true if this is the HEAD (development) version.
func (v *Version) IsHead() bool {
	return v.VersionNumber == 0
}
