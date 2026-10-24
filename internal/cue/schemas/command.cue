package schemas

import "strings"

// ============================================================================
// Command Schema
// ============================================================================

// Known Claude Code tools for validation (#KnownTool is generated from
// textutil.KnownTools at load time; see schema_loader.go).

// ============================================================================
// Command Hook Definitions
// ============================================================================

// Component hook definitions (#CommandHookCommand, #CommandHook, #CommandHooks)
// are aliases from component_hooks.cue, composed by schema_loader.go.

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
