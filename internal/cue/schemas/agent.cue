package schemas

import (
	"regexp"
	"strings"
)

// ============================================================================
// Agent Schema
// ============================================================================

// Valid Claude Code colors - limited to 11 standard colors
#Color: "red" | "blue" | "green" | "yellow" | "purple" | "orange" | "pink" | "cyan" | "gray" | "magenta" | "white"

// Valid memory scopes for persistent agent memory (v2.1.33+)
#MemoryScope: "user" | "project" | "local"

// Task(agent_type) restricts which sub-agents can be spawned (v2.1.33+)
#TaskAgentTool: =~"^Task\\([a-z0-9][a-z0-9-]*\\)$"

// Tools specification - can be "*" for all, comma-separated string, or array of specific tools/Task(agent) refs
#Tools: "*" | string | [...(#KnownTool | #TaskAgentTool | =~"^mcp__" | =~"^[A-Za-z]+\\(.*\\)$")]

// Component hook definitions (#AgentHookCommand, #AgentHook, #AgentHooks) are
// shared verbatim with command/skill: see component_hooks.cue, appended to
// every root by the schema loader (schema_loader.go).

// ============================================================================
// Agent Schema
// ============================================================================

// Agent frontmatter schema
// Source: https://code.claude.com/docs/en/sub-agents
#Agent: {
	// Required fields
	name: string & =~("^[a-z0-9-]+$") & strings.MaxRunes(64)  // lowercase, numbers, hyphens only, max 64 chars
	description: string & !="" & strings.MaxRunes(1024)       // non-empty description, max 1024 chars

	// Optional Claude Code fields
	model?: #Model
	color?: #Color                                            // UI display color (set via /agents wizard)
	tools?: #Tools                                            // tool access allowlist; agents use 'tools:', NOT 'allowed-tools:'
	disallowedTools?: #Tools                                  // tool access denylist
	permissionMode?: "default" | "acceptEdits" | "bypassPermissions" | "plan" | "dontAsk" | "auto"
	maxTurns?: int & >0                                       // max conversation turns (positive integer)
	autoCompactWindow?: int & >=100000 & <=1000000   // token count at which this agent compacts its own conversation when it runs as a subagent (v2.1.296+)
	effort?:   string                                         // reasoning effort level (v2.1.78+)
	skills?: string | [...string]                             // skills to preload into agent context
	hooks?: #AgentHooks                                       // agent-level hooks (PreToolUse, PostToolUse, Stop)
	memory?: #MemoryScope                                     // persistent memory scope (v2.1.33+)
	mcpServers?: [...string]                                   // MCP server names available to agent
	isolation?: "worktree"                                    // subagent isolation mode (v2.1.49+)
	background?: bool                                         // always run as background task (v2.1.49+)
	initialPrompt?: string                                    // auto-submit first turn (v2.1.83+)
	requiredMcpServers?: [...string]                          // agent only runs when these MCP servers are connected (v2.1.156)
	criticalSystemReminder_EXPERIMENTAL?: string             // experimental: reminder re-injected as a system message (v2.1.156)
	omitClaudeMd?: bool                                      // run without user, project and local CLAUDE.md files (v2.1.271)

	// Experimental agent features (v2.1.248+)
	experimental?: {
		cacheTtl?: "5m" | "1h" // per-agent prompt cache TTL when no subagent TTL setting is configured
		...
	}

	// Allow additional fields
	...
}

// Validation entry point for agents
validate: {
	input: #Agent
	result: true
}

// ============================================================================
// Additional Validation Functions
// ============================================================================

// Check if a color value is valid
#isValidColor: {
	color: string
	valid: strings.Contains("red,blue,green,yellow,purple,orange,pink,cyan,gray,magenta,white", color)
}

// Check if model value is valid
#isValidModel: {
	model: string
	valid: strings.Contains("sonnet,opus,haiku,fable,best,sonnet[1m],opus[1m],fable[1m],opusplan,inherit", model) || regexp.Match("^claude-[a-z0-9-]+(\\[[0-9a-z]+\\])?$", model)
}

// Check if name format is valid (lowercase, numbers, hyphens)
#isValidName: {
	name: string
	valid: regexp.Match("^([a-z0-9-]+)$", name)
}
