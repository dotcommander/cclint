package schemas

// ============================================================================
// Settings Schema (.claude/settings.json)
// ============================================================================

// Hook matcher pattern
#Matcher: string

// Hook type - command, prompt, or agent (v2.1.0+ added prompt and agent for plugins)
#HookType: "command" | "prompt" | "agent"

// Hook command definition
// PreToolUse hooks can return additionalContext to the model (v2.1.9+)
#HookCommand: {
	type:     #HookType
	command:  string
	timeout?: *30 | int  // default 30 seconds
	once?:    bool       // run only once per session (v2.1.0+)
}

// Hook definition (can be nested under event arrays)
#Hook: {
	matcher: #Matcher
	hooks: [...#HookCommand]
}

// Claude Code settings.json schema
#Settings: {
	// Optional hooks configuration
	// Events map to arrays of hook configurations
	hooks?: {
		[string]: [...#Hook]
	}

	// Language setting (v2.1.0+)
	// Configure Claude's response language (e.g., "japanese", "spanish")
	language?: string

	// Git integration settings (v2.1.0+)
	// Per-project control over @-mention file picker behavior
	respectGitignore?: bool

	// Plans directory (v2.1.9+)
	// Customize where plan files are stored (default: .claude/plans)
	plansDirectory?: string

	// All other fields are allowed - settings.json is extensible
	// MCP settings can use auto:N syntax (v2.1.9+) for tool search auto-enable threshold
	// where N is the context window percentage (0-100)
	// Common fields include:
	// - model: string
	// - permissions: object
	// - mcp: object
	// - etc.
	...
}

// Validation entry point for settings
validate: {
	input: #Settings
	result: true
}
