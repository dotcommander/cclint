package schemas

import "strings"

// ============================================================================
// Skill Schema
// Source: https://code.claude.com/docs/en/skills
// ============================================================================

// Valid model options for Claude Code
#Model: "sonnet" | "opus" | "haiku" | "sonnet[1m]" | "opusplan" | "inherit" |
	=~"^claude-[a-z0-9-]+$"  // allow full model names like claude-sonnet-4-20250514

// Known Claude Code tools for validation
#KnownTool: "Read" | "Write" | "Edit" | "MultiEdit" | "Bash" |
	"Grep" | "Glob" | "LS" | "Task" | "NotebookEdit" |
	"WebFetch" | "WebSearch" | "TodoWrite" | "BashOutput" |
	"KillBash" | "ExitPlanMode" | "AskUserQuestion" |
	"LSP" | "Skill" | "DBClient"

// ============================================================================
// Skill Hook Definitions
// ============================================================================

// Hook command definition
#SkillHookCommand: {
	type:     "command"
	command:  string
	timeout?: int   // timeout in seconds
	once?:    bool  // run only once per session
}

// Skill hook entry
#SkillHook: {
	matcher?: string                     // optional tool matcher (e.g., "Bash", "Write")
	hooks: [...#SkillHookCommand]        // array of commands to run
}

// Skill hooks configuration
#SkillHooks: {
	[string]: [...#SkillHook]
}

// ============================================================================
// Skill Schema
// ============================================================================

// Skill frontmatter schema
// Source: https://code.claude.com/docs/en/skills
#Skill: {
	// Required fields
	name: string & =~("^[a-z0-9-]+$") & strings.MaxRunes(64)  // lowercase, numbers, hyphens only, max 64 chars
	description: string & !="" & strings.MaxRunes(1024)       // non-empty description, max 1024 chars

	// Optional Claude Code fields
	"allowed-tools"?: "*" | string | [...#KnownTool]          // skills use 'allowed-tools:', NOT 'tools:'
	model?: #Model                                            // model to use when skill is active
	context?: "fork"                                          // run skill in forked sub-agent context
	agent?: string                                            // agent type for execution
	"user-invocable"?: bool                                   // opt-out of slash command menu (default true)
	hooks?: #SkillHooks                                       // skill-level hooks (PreToolUse, PostToolUse, Stop)

	// Allow additional fields
	...
}

// Validation entry point for skills
validate: {
	input: #Skill
	result: true
}
