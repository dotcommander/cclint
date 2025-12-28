package schemas

import "strings"

// ============================================================================
// Command Schema
// ============================================================================

// Valid model options for Claude Code
#Model: "sonnet" | "opus" | "haiku" | "sonnet[1m]" | "opusplan" | "inherit"

// Known Claude Code tools for validation
#KnownTool: "Read" | "Write" | "Edit" | "MultiEdit" | "Bash" |
	"Grep" | "Glob" | "LS" | "Task" | "NotebookEdit" |
	"WebFetch" | "WebSearch" | "TodoWrite" | "BashOutput" |
	"KillBash" | "ExitPlanMode" | "AskUserQuestion" |
	"LSP" | "Skill" | "DBClient"

// Command frontmatter schema (all fields optional per Claude Code spec)
// Source: https://code.claude.com/docs/en/slash-commands
#Command: {
	// Optional Claude Code fields
	name?: string & =~("^[a-z0-9-]+$") & strings.MaxRunes(64)  // lowercase, numbers, hyphens only, max 64 chars
	description?: string & strings.MaxRunes(1024)              // command description, max 1024 chars
	"allowed-tools"?: "*" | string | [...#KnownTool]           // commands use 'allowed-tools:', NOT 'tools:'
	"argument-hint"?: string                                   // argument hint for help
	model?: #Model
	"disable-model-invocation"?: bool                          // prevent SlashCommand tool from calling this

	// Allow additional fields
	...
}

// Validation entry point for commands
validate: {
	input: #Command
	result: true
}
