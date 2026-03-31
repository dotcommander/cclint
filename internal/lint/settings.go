package lint

import (
	"maps"
	"slices"
	"strings"

	"github.com/dotcommander/cclint/internal/cue"
)

// LintSettings runs linting on settings files using the generic linter.
func LintSettings(rootPath string, quiet bool, verbose bool, noCycleCheck bool, exclude []string) (*LintSummary, error) {
	ctx, err := NewLinterContext(rootPath, quiet, verbose, noCycleCheck, exclude)
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
	"SessionStart":       true,
	"SessionEnd":         true,
	"TeammateIdle":       true, // multi-agent workflow event (v2.1.33+)
	"TaskCompleted":      true, // multi-agent workflow event (v2.1.33+)
	"ConfigChange":       true, // config file change event (v2.1.49+)
	"WorktreeCreate":     true, // worktree lifecycle event (v2.1.50+)
	"WorktreeRemove":     true, // worktree lifecycle event (v2.1.50+)
	"InstructionsLoaded": true, // instructions loaded event (v2.1.69+)
	"PostCompact":        true, // fires after compaction completes (v2.1.76+)
	"Elicitation":        true, // MCP elicitation request event (v2.1.76+)
	"ElicitationResult":  true, // MCP elicitation result event (v2.1.76+)
	"StopFailure":        true, // API error stop event (v2.1.78+)
	"CwdChanged":         true, // working directory change event (v2.1.83+)
	"FileChanged":        true, // file change event (v2.1.83+)
	"TaskCreated":        true, // task creation event (v2.1.84+)
	"PermissionDenied":   true, // auto mode denial event (v2.1.88+)
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
	"http":    true,
}

// eventLabel builds a sorted, comma-separated label from a hook event map.
func eventLabel(events map[string]bool) string {
	keys := slices.Sorted(maps.Keys(events))
	return strings.Join(keys, ", ")
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
