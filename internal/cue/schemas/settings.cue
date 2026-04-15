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

	// Optional: conditional filter using permission rule syntax (v2.1.85+)
	// e.g., "Bash(git *)" to only run when Bash is called with git commands
	"if"?: string

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
	source: "github" | "git" | "git-subdir" | "url" | "npm" | "file" | "directory" | "hostPattern" | "settings"
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

	// Auto-memory storage directory override (v2.1.74+)
	autoMemoryDirectory?: string

	// Worktree settings (v2.1.76+)
	worktree?: {
		// Sparse checkout paths for large monorepos
		sparsePaths?: [...string]
		...
	}

	// Feedback survey sample rate for enterprise admins (v2.1.76+)
	feedbackSurveyRate?: number

	// Sandbox settings (v2.1.83+)
	sandbox?: {
		failIfUnavailable?: bool
		...
	}

	// Deep link registration control (v2.1.83+)
	disableDeepLinkRegistration?: bool

	// Tool result file cleanup period in days (v2.1.83+)
	// Must be >= 1; 0 silently disables transcript persistence (v2.1.89+)
	cleanupPeriodDays?: int & >=1

	// Allowed channel plugins for enterprise admins (v2.1.84+)
	// Map of plugin patterns to allowed/disallowed
	allowedChannelPlugins?: {
		[string]: bool
	}

	// Thinking summaries opt-in (v2.1.88+)
	// Set true to show thinking summaries in interactive sessions (default: off)
	showThinkingSummaries?: bool

	// Disable skill shell execution (v2.1.91+)
	// Prevents inline shell execution in skills, custom slash commands, and plugin commands
	disableSkillShellExecution?: bool

	// Force remote settings refresh (v2.1.92+)
	// Blocks startup until remote managed settings are freshly fetched; exits if fetch fails (fail-closed)
	forceRemoteSettingsRefresh?: bool

	// Status line refresh interval in seconds (v2.1.97+)
	// Re-runs the status line command every N seconds
	refreshInterval?: int

	// Allow only managed hooks (v2.1.101+)
	// When set, only hooks from managed/enterprise settings and force-enabled plugins run
	allowManagedHooksOnly?: bool

	// TUI rendering mode (v2.1.110+)
	// Controls rendering mode for the CLI; "/tui fullscreen" switches to flicker-free fullscreen rendering.
	tui?: string

	// Conversation auto-scroll toggle (v2.1.110+)
	// Disable auto-scroll in fullscreen mode (default: true)
	autoScrollEnabled?: bool

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
