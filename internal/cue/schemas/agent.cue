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

// Valid model options for Claude Code
#Model: "sonnet" | "opus" | "haiku" | "sonnet[1m]" | "opusplan" | "inherit"

// Known Claude Code tools for validation
#KnownTool: "Read" | "Write" | "Edit" | "MultiEdit" | "Bash" |
	"Grep" | "Glob" | "LS" | "Task" | "NotebookEdit" |
	"WebFetch" | "WebSearch" | "TodoWrite" | "BashOutput" |
	"KillBash" | "ExitPlanMode" | "AskUserQuestion" |
	"LSP" | "Skill" | "DBClient"

// Task(agent_type) restricts which sub-agents can be spawned (v2.1.33+)
#TaskAgentTool: =~"^Task\\([a-z0-9][a-z0-9-]*\\)$"

// Tools specification - can be "*" for all, comma-separated string, or array of specific tools/Task(agent) refs
#Tools: "*" | string | [...(#KnownTool | #TaskAgentTool)]

// ============================================================================
// Agent Hook Definitions
// ============================================================================

// Hook command definition (same as settings.cue)
#AgentHookCommand: {
	type:     "command"
	command:  string
	timeout?: int   // timeout in seconds
	once?:    bool  // run only once per session (v2.1.0+)
}

// Agent hook entry - matcher is optional (e.g., Stop hook doesn't need a matcher)
#AgentHook: {
	matcher?: string                    // optional tool matcher (e.g., "Bash", "Write")
	hooks: [...#AgentHookCommand]       // array of commands to run
}

// Agent hooks configuration - maps event names to hook arrays
// Known events: PreToolUse, PostToolUse, Stop, Start
// Allow any string key for future event types
#AgentHooks: {
	[string]: [...#AgentHook]
}

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
	permissionMode?: "default" | "acceptEdits" | "delegate" | "dontAsk" | "bypassPermissions" | "plan"
	maxTurns?: int & >0                                       // max conversation turns (positive integer)
	skills?: string                                           // skills to preload into agent context
	hooks?: #AgentHooks                                       // agent-level hooks (PreToolUse, PostToolUse, Stop)
	memory?: #MemoryScope                                     // persistent memory scope (v2.1.33+)
	mcpServers?: [...string]                                   // MCP server names available to agent
	isolation?: "worktree"                                    // subagent isolation mode (v2.1.49+)
	background?: bool                                         // always run as background task (v2.1.49+)

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
	valid: strings.Contains("sonnet,opus,haiku,sonnet[1m],opusplan,inherit", model)
}

// Check if name format is valid (lowercase, numbers, hyphens)
#isValidName: {
	name: string
	valid: regexp.Match("^([a-z0-9-]+)$", name)
}
