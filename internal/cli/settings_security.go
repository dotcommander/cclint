package cli

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/dotcommander/cclint/internal/cue"
)

// validateHookCommandSecurity checks for security issues in hook commands.
// Delegates to specific check functions for each security concern.
func validateHookCommandSecurity(cmd string, ctx hookContext) []cue.ValidationError {
	location := fmt.Sprintf("Event '%s' hook %d inner hook %d", ctx.EventName, ctx.HookIdx, ctx.InnerIdx)

	var warnings []cue.ValidationError
	warnings = append(warnings, checkUnquotedVariables(cmd, location, ctx.FilePath)...)
	warnings = append(warnings, checkPathTraversal(cmd, location, ctx.FilePath)...)
	warnings = append(warnings, checkHardcodedPaths(cmd, location, ctx.FilePath)...)
	warnings = append(warnings, checkSensitiveFileAccess(cmd, location, ctx.FilePath)...)
	warnings = append(warnings, checkDangerousPatterns(cmd, location, ctx.FilePath)...)

	return warnings
}

// checkUnquotedVariables detects unquoted variable expansion.
func checkUnquotedVariables(cmd, location, filePath string) []cue.ValidationError {
	// Matches $VAR or ${VAR} not preceded by quote and not followed by quote
	unquotedVarPattern := regexp.MustCompile(`[^"']\$\{?[A-Za-z_][A-Za-z0-9_]*\}?[^"']|^\$\{?[A-Za-z_][A-Za-z0-9_]*\}?[^"']`)
	if !unquotedVarPattern.MatchString(cmd) {
		return nil
	}

	// Check if it's truly unquoted (not a false positive)
	// Common false positive: $CLAUDE_PROJECT_DIR/path (this is often safe)
	if strings.Contains(cmd, `"$`) || strings.Contains(cmd, `'$`) {
		return nil
	}

	return []cue.ValidationError{{
		File:     filePath,
		Message:  fmt.Sprintf("%s: Unquoted variable expansion detected. Use \"$VAR\" to prevent word splitting", location),
		Severity: "warning",
		Source:   cue.SourceCClintObserve,
	}}
}

// checkPathTraversal detects path traversal attempts.
func checkPathTraversal(cmd, location, filePath string) []cue.ValidationError {
	if !strings.Contains(cmd, "..") {
		return nil
	}
	return []cue.ValidationError{{
		File:     filePath,
		Message:  fmt.Sprintf("%s: Path traversal '..' detected in hook command - potential security risk", location),
		Severity: "warning",
		Source:   cue.SourceCClintObserve,
	}}
}

// checkHardcodedPaths detects hardcoded absolute paths without $CLAUDE_PROJECT_DIR.
func checkHardcodedPaths(cmd, location, filePath string) []cue.ValidationError {
	absolutePathPattern := regexp.MustCompile(`["']/(?:Users|home|var|tmp|etc)/[^\s"']+`)
	if !absolutePathPattern.MatchString(cmd) || strings.Contains(cmd, "$CLAUDE_PROJECT_DIR") {
		return nil
	}
	return []cue.ValidationError{{
		File:     filePath,
		Message:  fmt.Sprintf("%s: Hardcoded absolute path detected. Consider using $CLAUDE_PROJECT_DIR for portability", location),
		Severity: "warning",
		Source:   cue.SourceCClintObserve,
	}}
}

// checkSensitiveFileAccess detects access to sensitive files.
func checkSensitiveFileAccess(cmd, location, filePath string) []cue.ValidationError {
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

	var warnings []cue.ValidationError
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
	return warnings
}

// checkDangerousPatterns detects command injection risks.
func checkDangerousPatterns(cmd, location, filePath string) []cue.ValidationError {
	dangerousPatterns := []struct {
		pattern string
		message string
	}{
		{`\beval\b`, "eval command detected - potential command injection risk"},
		{`\$\(.*\)`, "Command substitution detected - ensure input is sanitized"},
		{"`[^`]+`", "Backtick command substitution detected - ensure input is sanitized"},
		{`>\s*/dev/`, "Redirecting to /dev/ - verify this is intentional"},
	}

	var warnings []cue.ValidationError
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
