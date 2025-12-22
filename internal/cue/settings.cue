package schemas

// ============================================================================
// Settings Schema (.claude/settings.json)
// ============================================================================

// Hook matcher pattern
#Matcher :: string

// Hook command definition
#HookCommand :: {
	type:     "command"
	command:  string
	timeout?: *30 | int  // default 30 seconds
}

// Hook definition
#Hook :: {
	matcher: #Matcher
	hooks: [...#HookCommand]
}

// Claude Code settings.json schema
#Settings :: {
	// Optional hooks configuration
	hooks?: {
		[string]: #Hook
	}

	// All other fields are allowed - settings.json is extensible
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
