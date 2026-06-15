package schemas

// ============================================================================
// Settings Schema (.claude/settings.json)
// ============================================================================

// Hook matcher pattern
#Matcher: string

// Hook type - command, prompt, agent, http, or mcp_tool (v2.1.118+ added mcp_tool)
#HookType: "command" | "prompt" | "agent" | "http" | "mcp_tool"

// Hook handler definition
// PreToolUse hooks can return additionalContext to the model (v2.1.9+)
#HookCommand: {
	type: #HookType

	// Required for type: "command"
	command?: string

	// Exec form (v2.1.139+) — alternative to `command`, type: "command" only.
	// Spawns the command directly without a shell, so path placeholders never need quoting.
	args?: [...string]

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

	// Async rewake controls (v2.1.156, type: "command" only).
	// When an async hook exits with code 2 it can wake the model back up and
	// block the turn; asyncRewake opts the async hook into that behavior.
	asyncRewake?:   bool   // wake the model on async exit code 2
	rewakeMessage?: string // custom prefix for the system-reminder injected on rewake
	rewakeSummary?: string // short summary shown to the user when the async hook rewakes

	timeout?: *30 | int  // default 30 seconds
	once?:    bool       // run only once per session (v2.1.0+)

	// PostToolUse only (v2.1.139+): feed the hook's rejection reason back to
	// Claude and continue the turn instead of failing.
	continueOnBlock?: bool

	// Optional: conditional filter using permission rule syntax (v2.1.85+)
	// e.g., "Bash(git *)" to only run when Bash is called with git commands
	"if"?: string

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
	// Type-specific fields (repo, url, path, package, hostPattern, ref, headers, skipLfs)
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

	// Strict marketplace allowlist (managed/enterprise settings)
	// Array of allowed marketplace sources. When set in managed settings,
	// ONLY these sources can be added as marketplaces. The check runs BEFORE
	// downloading, so blocked sources never touch the filesystem.
	// Each entry is an open struct: {source: "<variant>", <variant-specific fields>}.
	// Example: {source: "github", repo: "acme/approved-plugins"}
	strictKnownMarketplaces?: [...#MarketplaceSource]

	// Blocked marketplace sources (managed/enterprise settings)
	// Array of marketplace sources that are explicitly denied. Same element
	// shape as strictKnownMarketplaces. Also checked BEFORE downloading.
	blockedMarketplaces?: [...#MarketplaceSource]

	// Disable all hooks (v2.1.49+)
	// Non-managed settings cannot disable managed hooks set by enterprise policy
	disableAllHooks?: bool

	// Auto-memory storage directory override (v2.1.74+)
	autoMemoryDirectory?: string

	// Auto-memory toggle (v2.1.156) — enable auto-memory for this project
	autoMemoryEnabled?: bool

	// Auto-dream toggle (v2.1.156) — enable background memory consolidation
	autoDreamEnabled?: bool

	// Worktree settings (v2.1.76+)
	worktree?: {
		// Sparse checkout paths for large monorepos
		sparsePaths?: [...string]

		// Base ref for new worktrees (v2.1.133+)
		// "fresh" → branch from origin/<default>; "head" → branch from local HEAD
		// Default is "fresh" as of v2.1.133 (was effectively "head" since v2.1.128).
		baseRef?: "fresh" | "head"

		// Background-session isolation mode (v2.1.143+)
		// "none" → background sessions edit the working copy directly without
		// EnterWorktree (for repos where worktrees are impractical).
		bgIsolation?: "none"
		...
	}

	// Feedback survey sample rate for enterprise admins (v2.1.76+)
	feedbackSurveyRate?: number

	// Sandbox settings (v2.1.83+)
	sandbox?: {
		failIfUnavailable?: bool

		// Managed-settings paths to custom bubblewrap / socat binaries (v2.1.133+, Linux/WSL only)
		bwrapPath?: string
		socatPath?: string

		// Network-level domain controls (allowedDomains: v2.1.83+; deniedDomains: v2.1.113+)
		// deniedDomains takes precedence over allowedDomains wildcards
		network?: {
			allowedDomains?: [...string]
			deniedDomains?: [...string]
			...
		}
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

	// Allow only managed allowed-domains (v2.1.126+)
	// When set, sandbox.network.allowedDomains entries from non-managed settings
	// sources are ignored — only the managed/enterprise allowedDomains list is honored.
	allowManagedDomainsOnly?: bool

	// Allow only managed allowed-read-paths (v2.1.126+)
	// When set, sandbox.allowedReadPaths entries from non-managed settings sources
	// are ignored — only the managed/enterprise allowedReadPaths list is honored.
	allowManagedReadPathsOnly?: bool

	// Parent-tier settings merge behavior (v2.1.133+, admin tier)
	// "first-wins" (default) → SDK managedSettings does not merge into the policy chain.
	// "merge" → admins opt SDK managedSettings (parent tier) into the merged policy.
	parentSettingsBehavior?: "first-wins" | "merge"

	// Channels enablement for console (API key) auth (v2.1.128+)
	// Console orgs with managed settings must set this true to enable the
	// --channels flow under API-key authentication.
	channelsEnabled?: bool

	// Per-skill visibility overrides (v2.1.129+)
	// Map of skill name to visibility level. Lets users hide or collapse a
	// skill without editing its SKILL.md. Does not apply to plugin skills —
	// those are managed through `/plugin`.
	skillOverrides?: {
		[string]: "on" | "name-only" | "user-invocable-only" | "off"
	}

	// Auto mode classifier rules (v2.1.136+)
	// Managed-settings permission-rule pattern lists steering the auto-mode classifier.
	// hard_deny blocks unconditionally regardless of user intent or allow exceptions.
	// allow / soft_deny / deny tune the classifier's default disposition; environment
	// carries free-form context about the user's setup to help the classifier decide.
	autoMode?: {
		allow?:       [...string]
		soft_deny?:   [...string]
		hard_deny?:   [...string]
		deny?:        [...string]
		environment?: [...string]
		...
	}

	// WSL inheritance of Windows-side managed settings (v2.1.118+)
	// Policy/managed settings key — allows WSL on Windows to inherit Windows managed settings
	wslInheritsWindowsSettings?: bool

	// PR URL template (v2.1.119+)
	// Override github.com with a custom code-review URL for the footer PR badge
	prUrlTemplate?: string

	// Allow all claude.ai MCP connectors (v2.1.149+)
	// Enterprise managed setting that loads claude.ai cloud MCP connectors alongside managed-mcp.json
	allowAllClaudeAiMcps?: bool

	// Plugin suggestion marketplace allowlist (v2.1.152+)
	// Managed setting allowlisting org marketplaces whose plugins may be suggested via context-aware tips
	pluginSuggestionMarketplaces?: [...#MarketplaceSource]

	// Managed-settings MCP server allow/deny policies (referenced through v2.1.154)
	// Entry shape inferred as server name/pattern strings (single invalid entry mentioned in changelog)
	allowedMcpServers?: [...string]
	deniedMcpServers?: [...string]

	// TUI rendering mode (v2.1.110+)
	// Controls rendering mode for the CLI; "/tui fullscreen" switches to flicker-free fullscreen rendering.
	tui?: string

	// Conversation auto-scroll toggle (v2.1.110+)
	// Disable auto-scroll in fullscreen mode (default: true)
	autoScrollEnabled?: bool

	// Default subagent for dispatched sessions (v2.1.157+)
	// --agent <name> overrides this at invocation time.
	agent?: string

	// Fallback models tried in order when the primary model is overloaded
	// or unavailable (v2.1.166+). Each element is a model name or alias;
	// "default" expands to the default model. Effective list capped at 3.
	// CLI --fallback-model takes precedence.
	fallbackModel?: [...string]

	// Minimum Claude Code version required to start (v2.1.163+, semver).
	// Only enforced from managed (policy) settings.
	requiredMinimumVersion?: string

	// Maximum Claude Code version allowed to start (v2.1.163+, semver).
	// Only enforced from managed (policy) settings.
	requiredMaximumVersion?: string

	// Hide bundled skills, workflows, and built-in slash commands (v2.1.169+).
	// Env equivalent: CLAUDE_CODE_DISABLE_BUNDLED_SKILLS.
	disableBundledSkills?: bool

	// Ramp mouse-wheel scroll speed during fast scrolls, fullscreen mode only (v2.1.174+).
	wheelScrollAccelerationEnabled?: bool

	// When true and availableModels is non-empty, the Default model selection is
	// also constrained to the allowlist (v2.1.175+). Managed (policy) setting.
	enforceAvailableModels?: bool

	// Extra clickable footer badges shown when a regex matches turn output (v2.1.176+).
	// Array of {type, pattern, url, label?}; configurable via user or managed settings.
	footerLinksRegexes?: [...{...}]

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
