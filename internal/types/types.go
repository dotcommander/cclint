// Package types provides shared types used across the cclint codebase.
// This package is at the bottom of the dependency graph and should not import
// any other internal packages to avoid circular dependencies.
package types

// ValidationError represents a validation error or warning.
type ValidationError struct {
	File     string
	Message  string
	Severity string // error, warning, suggestion, info
	Source   string // anthropic-docs, cclint-observation, agentskills-io
	Line     int
	Column   int
}

// Rule source constants.
const (
	SourceAnthropicDocs = "anthropic-docs"     // Official Anthropic documentation
	SourceCClintObserve = "cclint-observation" // Our best practice observations
	SourceAgentSkillsIO = "agentskills-io"     // agentskills.io specification
)

// Severity level constants.
const (
	SeverityError      = "error"
	SeverityWarning    = "warning"
	SeveritySuggestion = "suggestion"
	SeverityInfo       = "info"
)

// Component type constants.
const (
	TypeAgent   = "agent"
	TypeCommand = "command"
	TypeSkill   = "skill"
	TypeRule    = "rule"
)

// Hook type constants.
const (
	TypeHTTP = "http"
)
