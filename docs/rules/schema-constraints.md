# Schema Constraints

CUE schema validation rules for Claude Code components.

## Overview

All Claude Code components (agents, commands, settings, CLAUDE.md) are validated against CUE schemas that enforce structural constraints, naming patterns, and field requirements. This document catalogs all schema-level constraints (Rules 105-124).

---

## Agent Schema Constraints (Rules 105-112)

### Rule 105: Agent Name Pattern

**Severity:** error
**Component:** agent
**Category:** schema

**Description:**
Agent names must contain only lowercase letters, numbers, and hyphens. No uppercase letters, underscores, or special characters are permitted.

**Constraint:**
`name: string & =~("^[a-z0-9-]+$")`

**Valid Values:**
- `memory-specialist`
- `error-chain-1`
- `web3-agent`

**Invalid Values:**
- `Memory-Specialist` (uppercase)
- `error_chain` (underscore)
- `web3.agent` (period)
- `agent@service` (special char)

**Source:** [Anthropic Docs - Sub-agents](https://code.claude.com/docs/en/sub-agents) - name field format

---

### Rule 106: Agent Name Max Length

**Severity:** error
**Component:** agent
**Category:** schema

**Description:**
Agent names are limited to 64 characters to ensure readability and compatibility with filesystem paths.

**Constraint:**
`name: string & strings.MaxRunes(64)`

**Valid Values:**
Maximum 64 Unicode characters (runes).

**Source:** [Anthropic Docs - Skills](https://code.claude.com/docs/en/skills) - 64 character name limit

---

### Rule 107: Agent Description Required

**Severity:** error
**Component:** agent
**Category:** schema

**Description:**
Every agent must have a non-empty description field in its frontmatter. This description is used for documentation and agent discovery.

**Constraint:**
`description: string & !=""`

**Valid Values:**
Any non-empty string.

**Source:** [Anthropic Docs - Sub-agents](https://code.claude.com/docs/en/sub-agents) - description required for discovery

---

### Rule 108: Agent Description Max Length

**Severity:** error
**Component:** agent
**Category:** schema

**Description:**
Agent descriptions are limited to 1024 characters to prevent excessive frontmatter bloat.

**Constraint:**
`description: string & strings.MaxRunes(1024)`

**Valid Values:**
Maximum 1024 Unicode characters (runes).

**Source:** [Anthropic Docs - Skills](https://code.claude.com/docs/en/skills) - 1024 character description limit

---

### Rule 109: Agent Model Enum Values

**Severity:** error
**Component:** agent
**Category:** schema

**Description:**
The model field must be one of the predefined Claude Code model identifiers. Custom model names are not permitted.

**Constraint:**
`model?: "sonnet" | "opus" | "haiku" | "sonnet[1m]" | "opusplan" | "inherit"`

**Valid Values:**
- `sonnet` - Claude Sonnet 3.5
- `opus` - Claude Opus 3.5
- `haiku` - Claude Haiku 3.5
- `sonnet[1m]` - Sonnet with 1 million token context
- `opusplan` - Opus with planning mode
- `inherit` - Inherit from parent/settings

**Source:** [Anthropic Docs - Sub-agents](https://code.claude.com/docs/en/sub-agents) - model field enum values

---

### Rule 110: Agent Color Enum Values

**Severity:** error
**Component:** agent
**Category:** schema

**Description:**
Agent color must be one of the 8 standard Claude Code colors for visual identification.

**Constraint:**
`color?: "red" | "blue" | "green" | "yellow" | "purple" | "orange" | "pink" | "cyan"`

**Valid Values:**
- `red`
- `blue`
- `green`
- `yellow`
- `purple`
- `orange`
- `pink`
- `cyan`

**Source:** cclint observation - Visual identification palette for agents

---

### Rule 111: Agent Tools Field Variants

**Severity:** error
**Component:** agent
**Category:** schema

**Description:**
The tools (or allowed-tools) field can be specified in three formats: wildcard for all tools, comma-separated string, or array of specific tool names.

**Constraint:**
`tools?: "*" | string | [...#KnownTool]`
`"allowed-tools"?: "*" | string | [...#KnownTool]`

**Valid Values:**
- `"*"` - All tools allowed
- `"Read,Write,Edit"` - Comma-separated string
- `["Read", "Write", "Edit"]` - Array of tool names

**Source:** [Anthropic Docs - Sub-agents](https://code.claude.com/docs/en/sub-agents) - tools/allowed-tools field

---

### Rule 112: Known Tools Whitelist

**Severity:** error
**Component:** agent
**Category:** schema

**Description:**
When tools are specified as an array, each tool name must be from the known Claude Code tools whitelist.

**Constraint:**
```
#KnownTool: "Read" | "Write" | "Edit" | "MultiEdit" | "Bash" |
	"Grep" | "Glob" | "LS" | "Task" | "NotebookEdit" |
	"WebFetch" | "WebSearch" | "TodoWrite" | "BashOutput" |
	"KillBash" | "ExitPlanMode" | "AskUserQuestion" |
	"LSP" | "Skill" | "DBClient"
```

**Valid Values:**
Read, Write, Edit, MultiEdit, Bash, Grep, Glob, LS, Task, NotebookEdit, WebFetch, WebSearch, TodoWrite, BashOutput, KillBash, ExitPlanMode, AskUserQuestion, LSP, Skill

**Source:** cclint observation - Derived from Claude Code tool set documented in system prompts

---

## Command Schema Constraints (Rules 113-117)

### Rule 113: Command Name Pattern

**Severity:** error
**Component:** command
**Category:** schema

**Description:**
Command names must contain only lowercase letters, numbers, and hyphens. Same pattern as agent names.

**Constraint:**
`name?: string & =~("^[a-z0-9-]+$")`

**Valid Values:**
- `improve-code`
- `test-runner`
- `db-migrate`

**Invalid Values:**
- `Refactor` (uppercase)
- `test_runner` (underscore)
- `db.migrate` (period)

**Source:** [Anthropic Docs - Slash Commands](https://code.claude.com/docs/en/slash-commands) - name pattern consistency

---

### Rule 114: Command Name Max Length

**Severity:** error
**Component:** command
**Category:** schema

**Description:**
Command names are limited to 64 characters for consistency with agent names.

**Constraint:**
`name?: string & strings.MaxRunes(64)`

**Valid Values:**
Maximum 64 Unicode characters (runes).

**Source:** cclint observation - Consistency with agent/skill 64-char limit

---

### Rule 115: Command Description Max Length

**Severity:** error
**Component:** command
**Category:** schema

**Description:**
Command descriptions are limited to 1024 characters. Unlike agents, command descriptions are optional.

**Constraint:**
`description?: string & strings.MaxRunes(1024)`

**Valid Values:**
Maximum 1024 Unicode characters (runes), or omitted entirely.

**Source:** cclint observation - Consistency with agent/skill 1024-char limit

---

### Rule 116: Command Allowed-Tools Variants

**Severity:** error
**Component:** command
**Category:** schema

**Description:**
The allowed-tools field can be specified as wildcard, comma-separated string, or array of specific tool names.

**Constraint:**
`"allowed-tools"?: "*" | string | [...#KnownTool]`

**Valid Values:**
- `"*"` - All tools allowed
- `"Task,TodoWrite"` - Comma-separated string
- `["Task", "TodoWrite"]` - Array of tool names

**Source:** [Anthropic Docs - Slash Commands](https://code.claude.com/docs/en/slash-commands) - allowed-tools field

---

### Rule 117: Command Model Enum

**Severity:** error
**Component:** command
**Category:** schema

**Description:**
Command model field must be one of the predefined Claude Code model identifiers. Same values as agent model.

**Constraint:**
`model?: "sonnet" | "opus" | "haiku" | "sonnet[1m]" | "opusplan" | "inherit"`

**Valid Values:**
- `sonnet`
- `opus`
- `haiku`
- `sonnet[1m]`
- `opusplan`
- `inherit`

**Source:** [Anthropic Docs - Slash Commands](https://code.claude.com/docs/en/slash-commands) - model field

---

### Rule 117a: Command Hooks Field (v2.1.0+)

**Severity:** info
**Component:** command
**Category:** schema

**Description:**
Commands can define lifecycle hooks (PreToolUse, PostToolUse, Stop) scoped to the command's execution.

**Constraint:**
```
hooks?: {
	[string]: [...#CommandHook]
}
```

**Valid Values:**
Object with event names mapping to arrays of hook definitions.

**Source:** Claude Code 2.1.0 changelog - hooks support for commands

---

## Skill Schema Constraints (v2.1.0+)

### Rule 117b: Skill Context Field

**Severity:** info
**Component:** skill
**Category:** schema

**Description:**
Skills can run in a forked sub-agent context using `context: fork`.

**Constraint:**
```
context?: "fork"
```

**Valid Values:**
- `fork` - Run skill in forked sub-agent context
- Omitted - Run in main context (default)

**Source:** Claude Code 2.1.0 changelog - context: fork for skills

---

### Rule 117c: Skill Agent Field

**Severity:** info
**Component:** skill
**Category:** schema

**Description:**
Skills can specify an agent type for execution.

**Constraint:**
```
agent?: string
```

**Valid Values:**
Any valid agent type name (e.g., "code-quality-specialist", "go-specialist").

**Source:** Claude Code 2.1.0 changelog - agent field for skills

---

### Rule 117d: Skill User-Invocable Field

**Severity:** info
**Component:** skill
**Category:** schema

**Description:**
Skills in `/skills/` directories are visible in the slash command menu by default. Use `user-invocable: false` to opt out.

**Constraint:**
```
"user-invocable"?: bool
```

**Valid Values:**
- `true` or omitted - Skill appears in slash command menu (default for /skills/ dirs)
- `false` - Skill hidden from slash command menu

**Source:** Claude Code 2.1.0 changelog - user-invocable field for skills

---

### Rule 117e: Skill Hooks Field

**Severity:** info
**Component:** skill
**Category:** schema

**Description:**
Skills can define lifecycle hooks scoped to the skill's execution.

**Constraint:**
```
hooks?: {
	[string]: [...#SkillHook]
}
```

**Valid Values:**
Object with event names (PreToolUse, PostToolUse, Stop) mapping to hook arrays.

**Source:** Claude Code 2.1.0 changelog - hooks support for skills

---

## Settings Schema Constraints (Rules 118-121)

### Rule 118: Hook Event Structure

**Severity:** error
**Component:** settings
**Category:** schema

**Description:**
The hooks configuration must be structured as a map of event names to arrays of hook definitions. Each event can have multiple hooks.

**Constraint:**
```
hooks?: {
	[string]: [...#Hook]
}
```

**Valid Values:**
```json
{
  "hooks": {
    "PreToolUse": [
      { "matcher": "...", "hooks": [...] }
    ],
    "PostToolUse": [
      { "matcher": "...", "hooks": [...] }
    ]
  }
}
```

**Source:** [Anthropic Docs - Hooks](https://code.claude.com/docs/en/hooks) - hooks configuration structure

---

### Rule 119: Hook Matcher Required

**Severity:** error
**Component:** settings
**Category:** schema

**Description:**
Every hook definition must include a matcher field that determines when the hook triggers.

**Constraint:**
```
#Hook: {
	matcher: #Matcher
	hooks: [...#HookCommand]
}
```

**Valid Values:**
Any string pattern that matches hook trigger conditions (glob patterns, regex, etc.).

**Source:** [Anthropic Docs - Hooks](https://code.claude.com/docs/en/hooks) - matcher field required

---

### Rule 120: Hook Type Enum

**Severity:** error
**Component:** settings
**Category:** schema

**Description:**
Hook commands must specify one of the valid hook types. As of Claude Code 2.1.0, three types are supported.

**Constraint:**
```
#HookType: "command" | "prompt" | "agent"

#HookCommand: {
	type: #HookType
	...
}
```

**Valid Values:**
- `command` - Execute a shell command
- `prompt` - Modify Claude's prompt (plugins, v2.1.0+)
- `agent` - Invoke an agent (plugins, v2.1.0+)

**Source:** [Anthropic Docs - Hooks](https://code.claude.com/docs/en/hooks) - hook type values

---

### Rule 121: Hook Command Field Required

**Severity:** error
**Component:** settings
**Category:** schema

**Description:**
Hook command definitions must include a command field specifying what to execute.

**Constraint:**
```
#HookCommand: {
	type:    "command"
	command: string
	...
}
```

**Valid Values:**
Any non-empty string representing a shell command.

**Source:** [Anthropic Docs - Hooks](https://code.claude.com/docs/en/hooks) - command field required for type:command

---

### Rule 121a: Hook Once Field (v2.1.0+)

**Severity:** error
**Component:** settings
**Category:** schema

**Description:**
Hook commands can specify `once: true` to run only once per session.

**Constraint:**
```
#HookCommand: {
	type:     #HookType
	command:  string
	timeout?: int
	once?:    bool
}
```

**Valid Values:**
- `true` - Run hook only once per session
- `false` or omitted - Run hook every time it triggers

**Source:** Claude Code 2.1.0 changelog - once field for hooks

---

### Rule 121b: Settings Language Field (v2.1.0+)

**Severity:** info
**Component:** settings
**Category:** schema

**Description:**
Configure Claude's response language.

**Constraint:**
```
language?: string
```

**Valid Values:**
Any language name string (e.g., "japanese", "spanish", "french").

**Source:** Claude Code 2.1.0 changelog - language setting

---

### Rule 121c: Settings respectGitignore Field (v2.1.0+)

**Severity:** info
**Component:** settings
**Category:** schema

**Description:**
Per-project control over @-mention file picker behavior.

**Constraint:**
```
respectGitignore?: bool
```

**Valid Values:**
- `true` - Hide gitignored files from @-mention picker
- `false` - Show all files in @-mention picker

**Source:** Claude Code 2.1.0 changelog - respectGitignore setting

---

## CLAUDE.md Schema Constraints (Rules 122-124)

### Rule 122: Section Structure

**Severity:** error
**Component:** claude_md
**Category:** schema

**Description:**
CLAUDE.md sections must include heading and content fields. References are optional.

**Constraint:**
```
#Section: {
	heading: string
	content: string
	references?: [...#RefLink]
}
```

**Valid Values:**
- `heading` - Section heading text (required)
- `content` - Section content/body (required)
- `references` - Array of reference links (optional)

**Source:** cclint observation - CLAUDE.md section structure convention

---

### Rule 123: Rule Name/Description Required

**Severity:** error
**Component:** claude_md
**Category:** schema

**Description:**
CLAUDE.md rule definitions must include both name and description fields.

**Constraint:**
```
#Rule: {
	name:        string
	description: string
}
```

**Valid Values:**
Both `name` and `description` must be non-empty strings.

**Source:** cclint observation - CLAUDE.md rule definition structure

---

### Rule 124: Reference Path Required

**Severity:** error
**Component:** claude_md
**Category:** schema

**Description:**
Reference links in CLAUDE.md sections must include a path field. The title is optional.

**Constraint:**
```
#RefLink: {
	path:   string
	title?: string
}
```

**Valid Values:**
- `path` - File path or URL (required)
- `title` - Display title (optional)

**Source:** cclint observation - CLAUDE.md reference link structure

---

## Schema Validation Process

CUE schemas are embedded in the linter binary using `//go:embed` directives. Validation occurs in this order:

1. **Frontmatter extraction** - YAML parsing from `---` delimited blocks
2. **CUE validation** - Frontmatter validated against component-specific schema
3. **Best practice checks** - Go-based validation for patterns not expressible in CUE
4. **Scoring** - Quality scoring based on structural and compositional metrics

Schema violations result in **error** severity findings that block component usage in Claude Code.

## Related Rules

- Rules 1-40: Agent best practices
- Rules 41-60: Command best practices
- Rules 61-80: Skill best practices
- Rules 81-104: Settings/Hook best practices
