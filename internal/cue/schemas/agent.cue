package schemas

import (
	"regexp"
	"strings"
)

// ============================================================================
// Agent Schema
// ============================================================================

// Valid Claude Code colors - limited to 8 standard colors
#Color: "red" | "blue" | "green" | "yellow" | "purple" | "orange" | "pink" | "cyan"

// Valid model options for Claude Code
#Model: "sonnet" | "opus" | "haiku" | "sonnet[1m]" | "opusplan" | "inherit"

// Known Claude Code tools for validation
#KnownTool: "Read" | "Write" | "Edit" | "MultiEdit" | "Bash" |
	"Grep" | "Glob" | "LS" | "Task" | "NotebookEdit" |
	"WebFetch" | "WebSearch" | "TodoWrite" | "BashOutput" |
	"KillBash" | "ExitPlanMode" | "AskUserQuestion" |
	"LSP" | "Skill" | "DBClient"

// Tools specification - can be "*" for all, comma-separated string, or array of specific tools
#Tools: "*" | string | [...#KnownTool]

// Agent frontmatter schema
#Agent: {
	// Required fields
	name: string & =~("^[a-z0-9-]+$")  // lowercase, numbers, hyphens only
	description: string & !=""         // non-empty description

	// Optional Claude Code fields
	model?: #Model
	color?: #Color
	tools?: #Tools
	"allowed-tools"?: #Tools          // alternative name for tools field

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
	valid: strings.Contains("red,blue,green,yellow,purple,orange,pink,cyan", color)
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
