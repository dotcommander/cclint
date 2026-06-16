package schemas

import "strings"

// ============================================================================
// Skill Schema
// Source: https://code.claude.com/docs/en/skills
// String substitutions available in skills:
//   ${CLAUDE_SESSION_ID} - current session ID (v2.1.9+)
//   ${CLAUDE_SKILL_DIR} - skill's own directory path (v2.1.69+)
// ============================================================================

// Known Claude Code tools for validation (#KnownTool is generated from
// textutil.KnownTools at load time; see validator.go).

// ============================================================================
// Skill Hook Definitions
// ============================================================================

// Hook command definition
#SkillHookCommand: {
	type:     "command"
	command?: string             // shell form
	args?:    [...string]        // exec form (v2.1.139+), alternative to command
	timeout?: int                // timeout in seconds
	once?:    bool               // run only once per session
	continueOnBlock?: bool       // PostToolUse only (v2.1.139+)
	"if"?: string                // conditional filter using permission rule syntax (v2.1.85+)
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
// Extended: https://agentskills.io/specification
#Skill: {
	// Required fields
	// Name: lowercase, numbers, hyphens only. No leading/trailing/consecutive hyphens.
	name: string & =~("^[a-z0-9]+(-[a-z0-9]+)*$") & strings.MaxRunes(64)
	description: string & !="" & strings.MaxRunes(1536)       // non-empty description, max 1536 chars (v2.1.105 raised from 250)

	// Optional Claude Code fields
	"argument-hint"?: string                                  // hint shown during autocomplete (e.g., "[issue-number]")
	"disable-model-invocation"?: bool                         // prevent Claude from auto-loading this skill
	"user-invocable"?: bool                                   // show in slash command menu (default true)
	"allowed-tools"?: "*" | string | [...#KnownTool | =~"^mcp__" | =~"^[A-Za-z]+\\(.*\\)$"]          // skills use 'allowed-tools:', NOT 'tools:'
	"disallowed-tools"?: "*" | string | [...#KnownTool | =~"^mcp__" | =~"^[A-Za-z]+\\(.*\\)$"]       // remove tools from model while skill is active (v2.1.152+)
	model?: #Model                                            // model to use when skill is active
	effort?: string                                           // reasoning effort level (v2.1.80+)
	context?: "fork"                                          // run skill in forked sub-agent context
	agent?: string                                            // agent type for execution
	hooks?: #SkillHooks                                       // skill-level hooks (PreToolUse, PostToolUse, Stop)

	// Optional agentskills.io fields
	license?: string                                          // SPDX identifier or license file reference
	compatibility?: string & strings.MaxRunes(500)            // environment requirements (max 500 chars)
	metadata?: {[string]: string | number | bool}             // arbitrary key-value mapping

	// Allow additional fields
	...
}

// Validation entry point for skills
validate: {
	input: #Skill
	result: true
}
