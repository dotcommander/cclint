package cli

import (
	"github.com/dotcommander/cclint/internal/cue"
)

// LintSettings runs linting on settings files using the generic linter.
func LintSettings(rootPath string, quiet bool, verbose bool, noCycleCheck bool) (*LintSummary, error) {
	ctx, err := NewLinterContext(rootPath, quiet, verbose, noCycleCheck)
	if err != nil {
		return nil, err
	}
	return lintBatch(ctx, NewSettingsLinter()), nil
}

// Valid hook events according to Anthropic documentation
var validHookEvents = map[string]bool{
	"PreToolUse":         true,
	"PermissionRequest":  true,
	"PostToolUse":        true,
	"PostToolUseFailure": true,
	"Notification":       true,
	"UserPromptSubmit":   true,
	"Stop":               true,
	"SubagentStart":      true,
	"SubagentStop":       true,
	"PreCompact":         true,
	"Setup":              true, // matcher values: "init", "maintenance" (not tool names)
	"SessionStart":       true,
	"SessionEnd":         true,
	"TeammateIdle":       true, // multi-agent workflow event (v2.1.33+)
	"TaskCompleted":      true, // multi-agent workflow event (v2.1.33+)
}

// validComponentHookEvents lists hook events valid for agents and skills.
// Components only support PreToolUse, PostToolUse, and Stop per Claude Code docs.
var validComponentHookEvents = map[string]bool{
	"PreToolUse":  true,
	"PostToolUse": true,
	"Stop":        true,
}

// Hook events that support prompt hooks
var promptHookEvents = map[string]bool{
	"Stop":              true,
	"SubagentStop":      true,
	"UserPromptSubmit":  true,
	"PreToolUse":        true,
	"PermissionRequest": true,
}

// Valid hook types
var validHookTypes = map[string]bool{
	"command": true,
	"prompt":  true,
	"agent":   true,
}

// knownToolNames lists the tool names recognized by Claude Code.
var knownToolNames = map[string]bool{
	"Bash":           true,
	"Read":           true,
	"Write":          true,
	"Edit":           true,
	"Glob":           true,
	"Grep":           true,
	"Task":           true,
	"Skill":          true,
	"WebSearch":      true,
	"WebFetch":       true,
	"TodoRead":       true,
	"TodoWrite":      true,
	"TaskOutput":     true,
	"AskUser":        true,
	"mcp__":          true, // MCP tool prefix (matched specially)
	"computer":       true,
	"text_editor":    true,
	"MultiModelTool": true,
}

// validateSettingsSpecific implements settings-specific validation rules
func validateSettingsSpecific(data map[string]any, filePath string) []cue.ValidationError {
	var errors []cue.ValidationError

	// Check hooks structure if present
	if hooks, ok := data["hooks"]; ok {
		errors = append(errors, validateHooks(hooks, filePath)...)
	}

	// Check permissions structure if present
	if perms, ok := data["permissions"]; ok {
		errors = append(errors, validatePermissions(perms, filePath)...)
	}

	// Check mcpServers structure if present
	if mcpServers, ok := data["mcpServers"]; ok {
		errors = append(errors, validateMCPServers(mcpServers, filePath)...)
	}

	// Check rules array if present
	if rules, ok := data["rules"]; ok {
		errors = append(errors, validateRules(rules, filePath)...)
	}

	return errors
}
