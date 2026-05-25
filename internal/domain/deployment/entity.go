// Package deployment defines the deployment domain entity and related types.
package deployment

import "time"

// Deployment represents a deployed version of an Apps Script project.
type Deployment struct {
	ID          string
	Version     *Version
	Config      Config
	UpdateTime  time.Time
	EntryPoints []EntryPoint
}

// Version represents a specific version snapshot used in a deployment.
type Version struct {
	VersionNumber int
	Description   string
	CreateTime    time.Time
}

// Config contains deployment configuration.
type Config struct {
	ScriptID    string
	VersionID   int
	Description string
}

// EntryPoint represents an access point for a deployment.
type EntryPoint struct {
	Type         EntryPointType
	WebApp       *WebAppConfig
	ExecutionAPI *ExecutionAPIConfig
	AddOn        *AddOnConfig
}

// EntryPointType defines the type of entry point.
type EntryPointType string

const (
	EntryPointWebApp  EntryPointType = "WEB_APP"
	EntryPointExecAPI EntryPointType = "EXECUTION_API"
	EntryPointAddOn   EntryPointType = "ADD_ON"
)

// WebAppConfig contains web app deployment configuration.
type WebAppConfig struct {
	Access WebAppAccess
	URL    string
}

// WebAppAccess defines who can access a web app.
type WebAppAccess string

const (
	WebAppAccessMyself          WebAppAccess = "MYSELF"
	WebAppAccessDomain          WebAppAccess = "DOMAIN"
	WebAppAccessAnyone          WebAppAccess = "ANYONE"
	WebAppAccessAnyoneAnonymous WebAppAccess = "ANYONE_ANONYMOUS"
)

// ExecutionAPIConfig contains execution API configuration.
type ExecutionAPIConfig struct {
	Access ExecutionAPIAccess
}

// ExecutionAPIAccess defines who can execute the API.
type ExecutionAPIAccess string

const (
	ExecutionAPIAccessMyself WebAppAccess = "MYSELF"
	ExecutionAPIAccessDomain WebAppAccess = "DOMAIN"
	ExecutionAPIAccessAnyone WebAppAccess = "ANYONE"
)

// AddOnConfig contains add-on deployment configuration.
type AddOnConfig struct {
	AddOnType  string
	ReportName string
}

// IsWebApp returns true if this deployment has a web app entry point.
func (d *Deployment) IsWebApp() bool {
	for _, ep := range d.EntryPoints {
		if ep.Type == EntryPointWebApp {
			return true
		}
	}
	return false
}

// WebAppURL returns the web app URL if available.
func (d *Deployment) WebAppURL() string {
	for _, ep := range d.EntryPoints {
		if ep.Type == EntryPointWebApp && ep.WebApp != nil {
			return ep.WebApp.URL
		}
	}
	return ""
}
