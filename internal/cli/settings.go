package cli

import (
	"fmt"
	"regexp"
	"strings"

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
	"PreToolUse":          true,
	"PermissionRequest":   true,
	"PostToolUse":         true,
	"PostToolUseFailure":  true,
	"Notification":        true,
	"UserPromptSubmit":    true,
	"Stop":                true,
	"SubagentStart":       true,
	"SubagentStop":        true,
	"PreCompact":          true,
	"Setup":               true, // matcher values: "init", "maintenance" (not tool names)
	"SessionStart":        true,
	"SessionEnd":          true,
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
}

// validateSettingsSpecific implements settings-specific validation rules
func validateSettingsSpecific(data map[string]interface{}, filePath string) []cue.ValidationError {
	var errors []cue.ValidationError

	// Check hooks structure if present
	if hooks, ok := data["hooks"]; ok {
		errors = append(errors, validateHooks(hooks, filePath)...)
	}

	return errors
}

// validateHooks validates hooks for settings (full event set)
func validateHooks(hooks interface{}, filePath string) []cue.ValidationError {
	return validateHooksWithEvents(hooks, filePath, validHookEvents, "PreToolUse, PostToolUse, PostToolUseFailure, PermissionRequest, Notification, UserPromptSubmit, Stop, Setup, SubagentStart, SubagentStop, PreCompact, SessionStart, SessionEnd")
}

// ValidateComponentHooks validates hooks for agents and skills (scoped event set)
func ValidateComponentHooks(hooks interface{}, filePath string) []cue.ValidationError {
	return validateHooksWithEvents(hooks, filePath, validComponentHookEvents, "PreToolUse, PostToolUse, Stop")
}

// validateHooksWithEvents validates the hooks section with specified allowed events
func validateHooksWithEvents(hooks interface{}, filePath string, allowedEvents map[string]bool, eventLabel string) []cue.ValidationError {
	var errors []cue.ValidationError

	hooksMap, ok := hooks.(map[string]interface{})
	if !ok {
		errors = append(errors, cue.ValidationError{
			File:     filePath,
			Message:  "hooks must be an object mapping event names to hook configurations",
			Severity: "error",
			Source:   cue.SourceAnthropicDocs,
		})
		return errors
	}

	// Validate each event name and its hooks
	for eventName, eventConfig := range hooksMap {
		// Check if event name is valid
		if !allowedEvents[eventName] {
			errors = append(errors, cue.ValidationError{
				File:     filePath,
				Message:  fmt.Sprintf("Unknown hook event '%s'. Valid events: %s", eventName, eventLabel),
				Severity: "error",
				Source:   cue.SourceAnthropicDocs,
			})
			continue
		}

		// Validate the event's hook array
		hookArray, ok := eventConfig.([]interface{})
		if !ok {
			errors = append(errors, cue.ValidationError{
				File:     filePath,
				Message:  fmt.Sprintf("Event '%s': hook configuration must be an array", eventName),
				Severity: "error",
				Source:   cue.SourceAnthropicDocs,
			})
			continue
		}

		// Validate each hook matcher in the array
		for i, hookMatcher := range hookArray {
			hookMatcherMap, ok := hookMatcher.(map[string]interface{})
			if !ok {
				errors = append(errors, cue.ValidationError{
					File:     filePath,
					Message:  fmt.Sprintf("Event '%s' hook %d: must be an object with 'matcher' and 'hooks' fields", eventName, i),
					Severity: "error",
					Source:   cue.SourceAnthropicDocs,
				})
				continue
			}

			// Check for required 'matcher' field
			if _, exists := hookMatcherMap["matcher"]; !exists {
				errors = append(errors, cue.ValidationError{
					File:     filePath,
					Message:  fmt.Sprintf("Event '%s' hook %d: missing required field 'matcher'", eventName, i),
					Severity: "error",
					Source:   cue.SourceAnthropicDocs,
				})
			}

			// Check for required 'hooks' field
			innerHooks, exists := hookMatcherMap["hooks"]
			if !exists {
				errors = append(errors, cue.ValidationError{
					File:     filePath,
					Message:  fmt.Sprintf("Event '%s' hook %d: missing required field 'hooks'", eventName, i),
					Severity: "error",
					Source:   cue.SourceAnthropicDocs,
				})
				continue
			}

			// Validate the inner hooks array
			innerHooksArray, ok := innerHooks.([]interface{})
			if !ok {
				errors = append(errors, cue.ValidationError{
					File:     filePath,
					Message:  fmt.Sprintf("Event '%s' hook %d: 'hooks' field must be an array", eventName, i),
					Severity: "error",
					Source:   cue.SourceAnthropicDocs,
				})
				continue
			}

			// Validate each individual hook
			for j, innerHook := range innerHooksArray {
				innerHookMap, ok := innerHook.(map[string]interface{})
				if !ok {
					errors = append(errors, cue.ValidationError{
						File:     filePath,
						Message:  fmt.Sprintf("Event '%s' hook %d inner hook %d: must be an object", eventName, i, j),
						Severity: "error",
						Source:   cue.SourceAnthropicDocs,
					})
					continue
				}

				// Validate hook type
				hookType, typeExists := innerHookMap["type"]
				if !typeExists {
					errors = append(errors, cue.ValidationError{
						File:     filePath,
						Message:  fmt.Sprintf("Event '%s' hook %d inner hook %d: missing required field 'type'", eventName, i, j),
						Severity: "error",
						Source:   cue.SourceAnthropicDocs,
					})
					continue
				}

				hookTypeStr, ok := hookType.(string)
				if !ok {
					errors = append(errors, cue.ValidationError{
						File:     filePath,
						Message:  fmt.Sprintf("Event '%s' hook %d inner hook %d: 'type' must be a string", eventName, i, j),
						Severity: "error",
						Source:   cue.SourceAnthropicDocs,
					})
					continue
				}

				// Check if hook type is valid
				if !validHookTypes[hookTypeStr] {
					errors = append(errors, cue.ValidationError{
						File:     filePath,
						Message:  fmt.Sprintf("Event '%s' hook %d inner hook %d: invalid type '%s'. Valid types: command, prompt", eventName, i, j, hookTypeStr),
						Severity: "error",
						Source:   cue.SourceAnthropicDocs,
					})
					continue
				}

				// Validate type-specific requirements
				if hookTypeStr == "command" {
					if cmdVal, exists := innerHookMap["command"]; !exists {
						errors = append(errors, cue.ValidationError{
							File:     filePath,
							Message:  fmt.Sprintf("Event '%s' hook %d inner hook %d: type 'command' requires 'command' field", eventName, i, j),
							Severity: "error",
							Source:   cue.SourceAnthropicDocs,
						})
					} else if cmdStr, ok := cmdVal.(string); ok {
						// Validate hook command security
						securityWarnings := validateHookCommandSecurity(cmdStr, eventName, i, j, filePath)
						errors = append(errors, securityWarnings...)
					}
				} else if hookTypeStr == "prompt" {
					// Check if this event supports prompt hooks
					if !promptHookEvents[eventName] {
						errors = append(errors, cue.ValidationError{
							File:     filePath,
							Message:  fmt.Sprintf("Event '%s' hook %d inner hook %d: event '%s' does not support prompt hooks. Prompt hooks only supported for: Stop, SubagentStop, UserPromptSubmit, PreToolUse, PermissionRequest", eventName, i, j, eventName),
							Severity: "error",
							Source:   cue.SourceAnthropicDocs,
						})
					}

					if _, exists := innerHookMap["prompt"]; !exists {
						errors = append(errors, cue.ValidationError{
							File:     filePath,
							Message:  fmt.Sprintf("Event '%s' hook %d inner hook %d: type 'prompt' requires 'prompt' field", eventName, i, j),
							Severity: "error",
							Source:   cue.SourceAnthropicDocs,
						})
					}
				}
			}
		}
	}

	return errors
}

// validateHookCommandSecurity checks for security issues in hook commands
func validateHookCommandSecurity(cmd string, eventName string, hookIdx int, innerIdx int, filePath string) []cue.ValidationError {
	var warnings []cue.ValidationError
	location := fmt.Sprintf("Event '%s' hook %d inner hook %d", eventName, hookIdx, innerIdx)

	// Pattern 1: Unquoted variable expansion (potential word splitting/globbing)
	// Matches $VAR or ${VAR} not preceded by quote and not followed by quote
	unquotedVarPattern := regexp.MustCompile(`[^"']\$\{?[A-Za-z_][A-Za-z0-9_]*\}?[^"']|^\$\{?[A-Za-z_][A-Za-z0-9_]*\}?[^"']`)
	if unquotedVarPattern.MatchString(cmd) {
		// Check if it's truly unquoted (not a false positive)
		// Common false positive: $CLAUDE_PROJECT_DIR/path (this is often safe)
		if !strings.Contains(cmd, `"$`) && !strings.Contains(cmd, `'$`) {
			warnings = append(warnings, cue.ValidationError{
				File:     filePath,
				Message:  fmt.Sprintf("%s: Unquoted variable expansion detected. Use \"$VAR\" to prevent word splitting", location),
				Severity: "warning",
				Source:   cue.SourceCClintObserve,
			})
		}
	}

	// Pattern 2: Path traversal attempts
	if strings.Contains(cmd, "..") {
		warnings = append(warnings, cue.ValidationError{
			File:     filePath,
			Message:  fmt.Sprintf("%s: Path traversal '..' detected in hook command - potential security risk", location),
			Severity: "warning",
			Source:   cue.SourceCClintObserve,
		})
	}

	// Pattern 3: Hardcoded absolute paths without $CLAUDE_PROJECT_DIR
	absolutePathPattern := regexp.MustCompile(`["']/(?:Users|home|var|tmp|etc)/[^\s"']+`)
	if absolutePathPattern.MatchString(cmd) && !strings.Contains(cmd, "$CLAUDE_PROJECT_DIR") {
		warnings = append(warnings, cue.ValidationError{
			File:     filePath,
			Message:  fmt.Sprintf("%s: Hardcoded absolute path detected. Consider using $CLAUDE_PROJECT_DIR for portability", location),
			Severity: "warning",
			Source:   cue.SourceCClintObserve,
		})
	}

	// Pattern 4: Sensitive file access
	sensitivePatterns := []struct {
		pattern string
		message string
	}{
		{`\.env\b`, "Accessing .env file - ensure secrets are not logged"},
		{`\.git/`, "Accessing .git directory - potential security concern"},
		{`credentials`, "Accessing credentials file - ensure secure handling"},
		{`\.ssh/`, "Accessing .ssh directory - high security risk"},
		{`\.aws/`, "Accessing AWS config directory - ensure no secrets exposed"},
		{`id_rsa|id_ed25519|id_dsa`, "Accessing SSH private key - high security risk"},
	}

	for _, sp := range sensitivePatterns {
		matched, _ := regexp.MatchString(`(?i)`+sp.pattern, cmd)
		if matched {
			warnings = append(warnings, cue.ValidationError{
				File:     filePath,
				Message:  fmt.Sprintf("%s: %s", location, sp.message),
				Severity: "warning",
				Source:   cue.SourceCClintObserve,
			})
		}
	}

	// Pattern 5: Command injection risks (common dangerous patterns)
	dangerousPatterns := []struct {
		pattern string
		message string
	}{
		{`\beval\b`, "eval command detected - potential command injection risk"},
		{`\$\(.*\)`, "Command substitution detected - ensure input is sanitized"},
		{"`[^`]+`", "Backtick command substitution detected - ensure input is sanitized"},
		{`>\s*/dev/`, "Redirecting to /dev/ - verify this is intentional"},
	}

	for _, dp := range dangerousPatterns {
		matched, _ := regexp.MatchString(dp.pattern, cmd)
		if matched {
			warnings = append(warnings, cue.ValidationError{
				File:     filePath,
				Message:  fmt.Sprintf("%s: %s", location, dp.message),
				Severity: "warning",
				Source:   cue.SourceCClintObserve,
			})
		}
	}

	return warnings
}