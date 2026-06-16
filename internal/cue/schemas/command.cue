package schemas

import "strings"

// ============================================================================
// Command Schema
// ============================================================================

// Known Claude Code tools for validation (#KnownTool is generated from
// textutil.KnownTools at load time; see validator.go).

// ============================================================================
// Command Hook Definitions
// ============================================================================

// Hook command definition
#CommandHookCommand: {
	type:     "command"
	command?: string             // shell form
	args?:    [...string]        // exec form (v2.1.139+), alternative to command
	timeout?: int                // timeout in seconds
	once?:    bool               // run only once per session
	continueOnBlock?: bool       // PostToolUse only (v2.1.139+)
	"if"?: string                // conditional filter using permission rule syntax (v2.1.85+)
}

// Command hook entry
#CommandHook: {
	matcher?: string                      // optional tool matcher (e.g., "Bash", "Write")
	hooks: [...#CommandHookCommand]       // array of commands to run
}

// Command hooks configuration
#CommandHooks: {
	[string]: [...#CommandHook]
}

// ============================================================================
// Command Schema
// ============================================================================

// Command frontmatter schema (all fields optional per Claude Code spec)
// Source: https://code.claude.com/docs/en/slash-commands
#Command: {
	// Optional Claude Code fields
	name?: string & =~("^[a-z0-9-]+$") & strings.MaxRunes(64)  // lowercase, numbers, hyphens only, max 64 chars
	description?: string & strings.MaxRunes(1024)              // command description, max 1024 chars
	"allowed-tools"?: "*" | string | [...#KnownTool | =~"^mcp__" | =~"^[A-Za-z]+\\(.*\\)$"]           // commands use 'allowed-tools:', NOT 'tools:'
	"disallowed-tools"?: "*" | string | [...#KnownTool | =~"^mcp__" | =~"^[A-Za-z]+\\(.*\\)$"]        // remove tools from model while command is active (v2.1.152+)
	"argument-hint"?: string                                   // argument hint for help
	model?: #Model
	effort?: string                                            // reasoning effort level (v2.1.80+)
	"disable-model-invocation"?: bool                          // prevent SlashCommand tool from calling this
	hooks?: #CommandHooks                                      // command-level hooks (PreToolUse, PostToolUse, Stop)

	// Allow additional fields
	...
}

// Validation entry point for commands
validate: {
	input: #Command
	result: true
}
