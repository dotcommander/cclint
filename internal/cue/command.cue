package schemas

import (
	"regexp"
)

// ============================================================================
// Command Schema
// ============================================================================

// Valid model options for Claude Code
#Model :: "sonnet" | "opus" | "haiku" | "sonnet[1m]" | "opusplan" | "inherit"

// Known Claude Code tools for validation
#KnownTool :: "Read" | "Write" | "Edit" | "MultiEdit" | "Bash" |
              "Grep" | "Glob" | "LS" | "Task" | "NotebookEdit" |
              "WebFetch" | "WebSearch" | "TodoWrite" | "BashOutput" |
              "KillBash" | "ExitPlanMode" | "AskUserQuestion" |
              "LSP" | "Skill"

// Command frontmatter schema (all fields optional per Claude Code spec)
#Command :: {
	// Optional Claude Code fields
	name?: string & =~("^[a-z0-9-]+$")  // lowercase, numbers, hyphens only
	description?: string                // command description
	"allowed-tools"?: string | [...#KnownTool]
	"argument-hint"?: string             // argument hint for help
	model?: #Model
}

// Validation entry point for commands
validate: {
	input: #Command
	result: true
}
