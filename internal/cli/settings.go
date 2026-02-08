package cli

import (
	"fmt"
	"path/filepath"
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
	"TeammateIdle":        true, // multi-agent workflow event (v2.1.33+)
	"TaskCompleted":       true, // multi-agent workflow event (v2.1.33+)
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
	"Bash":          true,
	"Read":          true,
	"Write":         true,
	"Edit":          true,
	"Glob":          true,
	"Grep":          true,
	"Task":          true,
	"Skill":         true,
	"WebSearch":     true,
	"WebFetch":      true,
	"TodoRead":      true,
	"TodoWrite":     true,
	"TaskOutput":    true,
	"AskUser":       true,
	"mcp__":         true, // MCP tool prefix (matched specially)
	"computer":      true,
	"text_editor":   true,
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

// validatePermissions validates the permissions section of settings.json.
// Expected structure: {"allow": ["Bash(npm*)", ...], "deny": ["Bash(rm*)", ...]}
func validatePermissions(perms any, filePath string) []cue.ValidationError {
	var errors []cue.ValidationError

	permsMap, ok := perms.(map[string]any)
	if !ok {
		errors = append(errors, cue.ValidationError{
			File:     filePath,
			Message:  "permissions must be an object with optional 'allow' and 'deny' arrays",
			Severity: "error",
			Source:   cue.SourceAnthropicDocs,
		})
		return errors
	}

	for key, val := range permsMap {
		if key != "allow" && key != "deny" {
			errors = append(errors, cue.ValidationError{
				File:     filePath,
				Message:  fmt.Sprintf("permissions: unknown key '%s'. Only 'allow' and 'deny' are supported", key),
				Severity: "error",
				Source:   cue.SourceAnthropicDocs,
			})
			continue
		}

		errors = append(errors, validatePermissionEntries(val, key, filePath)...)
	}

	return errors
}

// validatePermissionEntries validates a single permission list (allow or deny).
func validatePermissionEntries(entries any, listName string, filePath string) []cue.ValidationError {
	var errors []cue.ValidationError

	arr, ok := entries.([]any)
	if !ok {
		errors = append(errors, cue.ValidationError{
			File:     filePath,
			Message:  fmt.Sprintf("permissions.%s must be an array of tool permission strings", listName),
			Severity: "error",
			Source:   cue.SourceAnthropicDocs,
		})
		return errors
	}

	for i, entry := range arr {
		str, ok := entry.(string)
		if !ok || str == "" {
			errors = append(errors, cue.ValidationError{
				File:     filePath,
				Message:  fmt.Sprintf("permissions.%s[%d]: each entry must be a non-empty string", listName, i),
				Severity: "error",
				Source:   cue.SourceAnthropicDocs,
			})
			continue
		}

		// Extract tool name from patterns like "Bash(npm*)" or plain "Read"
		toolName := extractToolName(str)
		if !isKnownTool(toolName) {
			errors = append(errors, cue.ValidationError{
				File:     filePath,
				Message:  fmt.Sprintf("permissions.%s[%d]: unrecognized tool name '%s' in '%s'", listName, i, toolName, str),
				Severity: "suggestion",
				Source:   cue.SourceCClintObserve,
			})
		}
	}

	return errors
}

// extractToolName returns the tool name portion from a permission entry.
// "Bash(npm*)" -> "Bash", "Read" -> "Read", "mcp__foo" -> "mcp__"
func extractToolName(entry string) string {
	// Handle parenthesized patterns like "Bash(npm*)"
	if idx := strings.Index(entry, "("); idx > 0 {
		return entry[:idx]
	}
	// Handle MCP tool prefix like "mcp__server_tool"
	if strings.HasPrefix(entry, "mcp__") {
		return "mcp__"
	}
	return entry
}

// isKnownTool checks whether a tool name is in the known tools set.
func isKnownTool(name string) bool {
	return knownToolNames[name]
}

// validateMCPServers validates the mcpServers configuration map.
// Each entry maps a server name to an object with command, args, env, and cwd fields.
func validateMCPServers(mcpServers any, filePath string) []cue.ValidationError {
	var errors []cue.ValidationError

	serversMap, ok := mcpServers.(map[string]any)
	if !ok {
		errors = append(errors, cue.ValidationError{
			File:     filePath,
			Message:  "mcpServers must be an object mapping server names to configurations",
			Severity: "error",
			Source:   cue.SourceAnthropicDocs,
		})
		return errors
	}

	for serverName, serverConfig := range serversMap {
		if serverName == "" {
			errors = append(errors, cue.ValidationError{
				File:     filePath,
				Message:  "mcpServers: server name must not be empty",
				Severity: "error",
				Source:   cue.SourceAnthropicDocs,
			})
			continue
		}

		serverMap, ok := serverConfig.(map[string]any)
		if !ok {
			errors = append(errors, cue.ValidationError{
				File:     filePath,
				Message:  fmt.Sprintf("mcpServers '%s': server configuration must be an object", serverName),
				Severity: "error",
				Source:   cue.SourceAnthropicDocs,
			})
			continue
		}

		errors = append(errors, validateMCPServerEntry(serverName, serverMap, filePath)...)
	}

	return errors
}

// validateMCPServerEntry validates a single MCP server configuration entry.
func validateMCPServerEntry(serverName string, serverMap map[string]any, filePath string) []cue.ValidationError {
	var errors []cue.ValidationError

	// Validate command field (required, non-empty string)
	cmdVal, cmdExists := serverMap["command"]
	if !cmdExists {
		errors = append(errors, cue.ValidationError{
			File:     filePath,
			Message:  fmt.Sprintf("mcpServers '%s': missing required field 'command'", serverName),
			Severity: "error",
			Source:   cue.SourceAnthropicDocs,
		})
	} else if cmdStr, ok := cmdVal.(string); !ok {
		errors = append(errors, cue.ValidationError{
			File:     filePath,
			Message:  fmt.Sprintf("mcpServers '%s': 'command' must be a string", serverName),
			Severity: "error",
			Source:   cue.SourceAnthropicDocs,
		})
	} else if cmdStr == "" {
		errors = append(errors, cue.ValidationError{
			File:     filePath,
			Message:  fmt.Sprintf("mcpServers '%s': 'command' must not be empty", serverName),
			Severity: "error",
			Source:   cue.SourceAnthropicDocs,
		})
	}

	// Validate args field (optional, must be array of strings)
	if argsVal, argsExists := serverMap["args"]; argsExists {
		errors = append(errors, validateMCPServerArgs(serverName, argsVal, filePath)...)
	}

	// Validate env field (optional, must be object with string values)
	if envVal, envExists := serverMap["env"]; envExists {
		errors = append(errors, validateMCPServerEnv(serverName, envVal, filePath)...)
	}

	// Validate cwd field (optional, must be a string)
	if cwdVal, cwdExists := serverMap["cwd"]; cwdExists {
		if _, ok := cwdVal.(string); !ok {
			errors = append(errors, cue.ValidationError{
				File:     filePath,
				Message:  fmt.Sprintf("mcpServers '%s': 'cwd' must be a string", serverName),
				Severity: "error",
				Source:   cue.SourceAnthropicDocs,
			})
		}
	}

	return errors
}

// validateMCPServerArgs validates the args field of an MCP server entry.
func validateMCPServerArgs(serverName string, argsVal any, filePath string) []cue.ValidationError {
	argsArray, ok := argsVal.([]any)
	if !ok {
		return []cue.ValidationError{{
			File:     filePath,
			Message:  fmt.Sprintf("mcpServers '%s': 'args' must be an array of strings", serverName),
			Severity: "error",
			Source:   cue.SourceAnthropicDocs,
		}}
	}
	var errors []cue.ValidationError
	for i, arg := range argsArray {
		if _, isStr := arg.(string); !isStr {
			errors = append(errors, cue.ValidationError{
				File:     filePath,
				Message:  fmt.Sprintf("mcpServers '%s': args[%d] must be a string", serverName, i),
				Severity: "error",
				Source:   cue.SourceAnthropicDocs,
			})
		}
	}
	return errors
}

// validateMCPServerEnv validates the env field of an MCP server entry.
func validateMCPServerEnv(serverName string, envVal any, filePath string) []cue.ValidationError {
	envMap, ok := envVal.(map[string]any)
	if !ok {
		return []cue.ValidationError{{
			File:     filePath,
			Message:  fmt.Sprintf("mcpServers '%s': 'env' must be an object with string values", serverName),
			Severity: "error",
			Source:   cue.SourceAnthropicDocs,
		}}
	}
	var errors []cue.ValidationError
	for envKey, envValue := range envMap {
		if _, isStr := envValue.(string); !isStr {
			errors = append(errors, cue.ValidationError{
				File:     filePath,
				Message:  fmt.Sprintf("mcpServers '%s': env '%s' value must be a string", serverName, envKey),
				Severity: "error",
				Source:   cue.SourceAnthropicDocs,
			})
		}
	}
	return errors
}

// validateRules validates the rules array in settings.json.
// Each entry must be a non-empty string containing a valid glob pattern.
// Warns on suspicious patterns like absolute paths.
func validateRules(rules any, filePath string) []cue.ValidationError {
	var errors []cue.ValidationError

	rulesArray, ok := rules.([]any)
	if !ok {
		errors = append(errors, cue.ValidationError{
			File:     filePath,
			Message:  "rules must be an array of glob pattern strings",
			Severity: "error",
			Source:   cue.SourceAnthropicDocs,
		})
		return errors
	}

	for i, entry := range rulesArray {
		str, ok := entry.(string)
		if !ok || str == "" {
			errors = append(errors, cue.ValidationError{
				File:     filePath,
				Message:  fmt.Sprintf("rules[%d]: each entry must be a non-empty string", i),
				Severity: "error",
				Source:   cue.SourceAnthropicDocs,
			})
			continue
		}

		// Validate glob syntax using the shared helper from rules.go
		if err := validateGlobPattern(str); err != nil {
			errors = append(errors, cue.ValidationError{
				File:     filePath,
				Message:  fmt.Sprintf("rules[%d]: invalid glob pattern %q: %v", i, str, err),
				Severity: "error",
				Source:   cue.SourceAnthropicDocs,
			})
			continue
		}

		// Warn on absolute paths (not portable across machines)
		if filepath.IsAbs(str) || strings.HasPrefix(str, "/") {
			errors = append(errors, cue.ValidationError{
				File:     filePath,
				Message:  fmt.Sprintf("rules[%d]: absolute path %q is not portable; use relative glob patterns", i, str),
				Severity: "warning",
				Source:   cue.SourceCClintObserve,
			})
		}
	}

	return errors
}

// validateMatcherToolName validates a toolName pattern from a hook matcher.
// Patterns look like "Bash(npm*)", "Edit", "mcp__server_tool", etc.
// Returns errors if the base tool name is unrecognized or the glob portion is invalid.
func validateMatcherToolName(toolNamePattern string, location string, filePath string) []cue.ValidationError {
	var errors []cue.ValidationError

	if toolNamePattern == "" {
		return errors
	}

	// Extract the base tool name
	baseTool := extractToolName(toolNamePattern)
	if !isKnownTool(baseTool) {
		errors = append(errors, cue.ValidationError{
			File:     filePath,
			Message:  fmt.Sprintf("%s: unrecognized tool name '%s' in toolName pattern '%s'", location, baseTool, toolNamePattern),
			Severity: "suggestion",
			Source:   cue.SourceCClintObserve,
		})
	}

	// Validate the parenthetical glob pattern if present
	errors = append(errors, validateToolNameGlob(toolNamePattern, location, filePath)...)

	return errors
}

// validateToolNameGlob validates the glob portion inside parentheses of a toolName pattern.
func validateToolNameGlob(toolNamePattern, location, filePath string) []cue.ValidationError {
	openIdx := strings.Index(toolNamePattern, "(")
	if openIdx <= 0 {
		return nil
	}

	closeIdx := strings.LastIndex(toolNamePattern, ")")
	if closeIdx <= openIdx {
		return []cue.ValidationError{{
			File:     filePath,
			Message:  fmt.Sprintf("%s: unclosed parenthesis in toolName pattern '%s'", location, toolNamePattern),
			Severity: "error",
			Source:   cue.SourceAnthropicDocs,
		}}
	}

	globPart := toolNamePattern[openIdx+1 : closeIdx]
	if globPart == "" {
		return []cue.ValidationError{{
			File:     filePath,
			Message:  fmt.Sprintf("%s: empty glob pattern in parentheses for toolName '%s'", location, toolNamePattern),
			Severity: "warning",
			Source:   cue.SourceCClintObserve,
		}}
	}

	if err := validateGlobPattern(globPart); err != nil {
		return []cue.ValidationError{{
			File:     filePath,
			Message:  fmt.Sprintf("%s: invalid glob in toolName '%s': %v", location, toolNamePattern, err),
			Severity: "error",
			Source:   cue.SourceAnthropicDocs,
		}}
	}

	return nil
}

// validateHooks validates hooks for settings (full event set)
func validateHooks(hooks any, filePath string) []cue.ValidationError {
	return validateHooksWithEvents(hooks, filePath, validHookEvents, "PreToolUse, PostToolUse, PostToolUseFailure, PermissionRequest, Notification, UserPromptSubmit, Stop, Setup, SubagentStart, SubagentStop, PreCompact, SessionStart, SessionEnd, TeammateIdle, TaskCompleted")
}

// ValidateComponentHooks validates hooks for agents and skills (scoped event set)
func ValidateComponentHooks(hooks any, filePath string) []cue.ValidationError {
	return validateHooksWithEvents(hooks, filePath, validComponentHookEvents, "PreToolUse, PostToolUse, Stop")
}

// validateHooksWithEvents validates the hooks section with specified allowed events
func validateHooksWithEvents(hooks any, filePath string, allowedEvents map[string]bool, eventLabel string) []cue.ValidationError {
	var errors []cue.ValidationError

	hooksMap, ok := hooks.(map[string]any)
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
		hookArray, ok := eventConfig.([]any)
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
			errors = append(errors, validateHookMatcher(hookMatcher, eventName, i, filePath)...)
		}
	}

	return errors
}

// validateHookMatcher validates a single hook matcher entry within an event.
func validateHookMatcher(hookMatcher any, eventName string, idx int, filePath string) []cue.ValidationError {
	var errors []cue.ValidationError

	hookMatcherMap, ok := hookMatcher.(map[string]any)
	if !ok {
		return []cue.ValidationError{{
			File:     filePath,
			Message:  fmt.Sprintf("Event '%s' hook %d: must be an object with 'matcher' and 'hooks' fields", eventName, idx),
			Severity: "error",
			Source:   cue.SourceAnthropicDocs,
		}}
	}

	// Check for required 'matcher' field and validate toolName if present
	errors = append(errors, validateHookMatcherField(hookMatcherMap, eventName, idx, filePath)...)

	// Check for required 'hooks' field
	innerHooks, exists := hookMatcherMap["hooks"]
	if !exists {
		return append(errors, cue.ValidationError{
			File:     filePath,
			Message:  fmt.Sprintf("Event '%s' hook %d: missing required field 'hooks'", eventName, idx),
			Severity: "error",
			Source:   cue.SourceAnthropicDocs,
		})
	}

	innerHooksArray, ok := innerHooks.([]any)
	if !ok {
		return append(errors, cue.ValidationError{
			File:     filePath,
			Message:  fmt.Sprintf("Event '%s' hook %d: 'hooks' field must be an array", eventName, idx),
			Severity: "error",
			Source:   cue.SourceAnthropicDocs,
		})
	}

	for j, innerHook := range innerHooksArray {
		errors = append(errors, validateInnerHook(innerHook, eventName, idx, j, filePath)...)
	}

	return errors
}

// validateHookMatcherField validates the matcher field of a hook matcher entry.
func validateHookMatcherField(hookMatcherMap map[string]any, eventName string, idx int, filePath string) []cue.ValidationError {
	matcherVal, matcherExists := hookMatcherMap["matcher"]
	if !matcherExists {
		return []cue.ValidationError{{
			File:     filePath,
			Message:  fmt.Sprintf("Event '%s' hook %d: missing required field 'matcher'", eventName, idx),
			Severity: "error",
			Source:   cue.SourceAnthropicDocs,
		}}
	}

	matcherMap, isMap := matcherVal.(map[string]any)
	if !isMap {
		return nil
	}

	toolNameVal, exists := matcherMap["toolName"]
	if !exists {
		return nil
	}

	toolNameStr, isStr := toolNameVal.(string)
	if !isStr || toolNameStr == "" {
		return nil
	}

	location := fmt.Sprintf("Event '%s' hook %d matcher", eventName, idx)
	return validateMatcherToolName(toolNameStr, location, filePath)
}

// validateInnerHook validates a single inner hook entry (type, command/prompt fields).
func validateInnerHook(innerHook any, eventName string, hookIdx, innerIdx int, filePath string) []cue.ValidationError {
	innerHookMap, ok := innerHook.(map[string]any)
	if !ok {
		return []cue.ValidationError{{
			File:     filePath,
			Message:  fmt.Sprintf("Event '%s' hook %d inner hook %d: must be an object", eventName, hookIdx, innerIdx),
			Severity: "error",
			Source:   cue.SourceAnthropicDocs,
		}}
	}

	hookType, typeExists := innerHookMap["type"]
	if !typeExists {
		return []cue.ValidationError{{
			File:     filePath,
			Message:  fmt.Sprintf("Event '%s' hook %d inner hook %d: missing required field 'type'", eventName, hookIdx, innerIdx),
			Severity: "error",
			Source:   cue.SourceAnthropicDocs,
		}}
	}

	hookTypeStr, ok := hookType.(string)
	if !ok {
		return []cue.ValidationError{{
			File:     filePath,
			Message:  fmt.Sprintf("Event '%s' hook %d inner hook %d: 'type' must be a string", eventName, hookIdx, innerIdx),
			Severity: "error",
			Source:   cue.SourceAnthropicDocs,
		}}
	}

	if !validHookTypes[hookTypeStr] {
		return []cue.ValidationError{{
			File:     filePath,
			Message:  fmt.Sprintf("Event '%s' hook %d inner hook %d: invalid type '%s'. Valid types: command, prompt, agent", eventName, hookIdx, innerIdx, hookTypeStr),
			Severity: "error",
			Source:   cue.SourceAnthropicDocs,
		}}
	}

	return validateInnerHookType(innerHookMap, hookTypeStr, eventName, hookIdx, innerIdx, filePath)
}

// validateInnerHookType validates type-specific requirements for a hook entry.
func validateInnerHookType(hookMap map[string]any, hookType, eventName string, hookIdx, innerIdx int, filePath string) []cue.ValidationError {
	var errors []cue.ValidationError

	switch hookType {
	case "command":
		cmdVal, exists := hookMap["command"]
		if !exists {
			errors = append(errors, cue.ValidationError{
				File:     filePath,
				Message:  fmt.Sprintf("Event '%s' hook %d inner hook %d: type 'command' requires 'command' field", eventName, hookIdx, innerIdx),
				Severity: "error",
				Source:   cue.SourceAnthropicDocs,
			})
		} else if cmdStr, ok := cmdVal.(string); ok {
			errors = append(errors, validateHookCommandSecurity(cmdStr, eventName, hookIdx, innerIdx, filePath)...)
		}
	case "prompt":
		if !promptHookEvents[eventName] {
			errors = append(errors, cue.ValidationError{
				File:     filePath,
				Message:  fmt.Sprintf("Event '%s' hook %d inner hook %d: event '%s' does not support prompt hooks. Prompt hooks only supported for: Stop, SubagentStop, UserPromptSubmit, PreToolUse, PermissionRequest", eventName, hookIdx, innerIdx, eventName),
				Severity: "error",
				Source:   cue.SourceAnthropicDocs,
			})
		}
		if _, exists := hookMap["prompt"]; !exists {
			errors = append(errors, cue.ValidationError{
				File:     filePath,
				Message:  fmt.Sprintf("Event '%s' hook %d inner hook %d: type 'prompt' requires 'prompt' field", eventName, hookIdx, innerIdx),
				Severity: "error",
				Source:   cue.SourceAnthropicDocs,
			})
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