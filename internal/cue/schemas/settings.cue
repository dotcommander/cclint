package schemas

// ============================================================================
// Settings Schema (.claude/settings.json)
// ============================================================================

// Hook matcher pattern
#Matcher: string

// Hook type - command, prompt, agent, or http (v2.1.0+ added prompt and agent for plugins)
#HookType: "command" | "prompt" | "agent" | "http"

// Hook handler definition
// PreToolUse hooks can return additionalContext to the model (v2.1.9+)
#HookCommand: {
	type: #HookType

	// Required for type: "command"
	command?: string

	// Required for type: "prompt"
	prompt?: string

	// Required for type: "http"
	url?: string

	// Optional: HTTP headers (type: "http" only)
	headers?: {[string]: string}

	// Optional: environment variables to forward (type: "http" only)
	allowedEnvVars?: [...string]

	// Optional: status message displayed while hook runs
	statusMessage?: string

	// Optional: run command asynchronously (type: "command" only)
	async?: bool

	timeout?: *30 | int  // default 30 seconds
	once?:    bool       // run only once per session (v2.1.0+)

	// Enforce command field when type is "command"
	if type == "command" {
		command: string
	}

	// Enforce prompt field when type is "prompt"
	if type == "prompt" {
		prompt: string
	}

	// Enforce url field when type is "http"
	if type == "http" {
		url: string
	}
}

// Hook definition (can be nested under event arrays)
#Hook: {
	matcher: #Matcher
	hooks: [...#HookCommand]
}

// Marketplace source configuration (v2.1.45+)
#MarketplaceSource: {
	source: "github" | "git" | "url" | "npm" | "file" | "directory" | "hostPattern"
	// Type-specific fields (repo, url, path, package, hostPattern, ref, headers)
	...
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

	// Spinner tips customization (v2.1.45+)
	// Configure tips with an array of custom tip strings.
	// Set excludeDefault: true to show only custom tips instead of built-in ones.
	spinnerTipsOverride?: {
		tips: [...string] & [_, ...]
		excludeDefault?: *false | bool
	}

	// Plugin enablement (v2.1.45+)
	// Map of "plugin-name@marketplace-name" to enabled/disabled
	enabledPlugins?: {
		[string]: bool
	}

	// Additional plugin marketplaces (v2.1.45+)
	// Map of marketplace name to source configuration
	extraKnownMarketplaces?: {
		[string]: {
			source: #MarketplaceSource
		}
	}

	// Disable all hooks (v2.1.49+)
	// Non-managed settings cannot disable managed hooks set by enterprise policy
	disableAllHooks?: bool

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
