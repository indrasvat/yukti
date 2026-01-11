// Package project defines the project domain entity and related types.
package project

import "time"

// Project represents a Google Apps Script project.
type Project struct {
	ID           string
	Title        string
	ParentID     string // For bound scripts (attached to Docs/Sheets/Forms)
	CreateTime   time.Time
	UpdateTime   time.Time
	Creator      User
	LastModifier User
}

// User represents a Google user.
type User struct {
	Domain   string
	Email    string
	Name     string
	PhotoURL string
}

// Content represents the content of a project (files).
type Content struct {
	ScriptID string
	Files    []File
}

// File represents a single file in a project.
type File struct {
	Name         string
	Type         FileType
	Source       string
	LastModified time.Time
	FunctionSet  *FunctionSet
}

// FileType represents the type of a script file.
type FileType string

const (
	FileTypeServer FileType = "SERVER_JS"
	FileTypeHTML   FileType = "HTML"
	FileTypeJSON   FileType = "JSON"
)

// FunctionSet contains parsed function information from a file.
type FunctionSet struct {
	Functions []Function
}

// Function represents a parsed function from a script file.
type Function struct {
	Name       string
	Parameters []string
}

// IsStandalone returns true if the project is a standalone script.
func (p *Project) IsStandalone() bool {
	return p.ParentID == ""
}

// IsBound returns true if the project is bound to a Google Doc/Sheet/Form.
func (p *Project) IsBound() bool {
	return p.ParentID != ""
}
