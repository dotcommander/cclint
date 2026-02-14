# CUE Schema Reference

## Purpose of CUE Validation

CCLint uses CUE (Configure, Unite, Execute) schemas to provide structural validation of component frontmatter. CUE is a configuration language that excels at data validation, constraint enforcement, and type checking.

The validation pipeline:

1. **Frontmatter Extraction** - YAML frontmatter is parsed from markdown files
2. **CUE Schema Validation** - Frontmatter data is validated against type-safe CUE schemas
3. **Go-Based Best Practice Checks** - Additional semantic checks (line limits, patterns, etc.)
4. **Scoring** - Quality scores (0-100) are calculated based on validation results

CUE schemas provide:
- **Type Safety**: Ensures fields match expected types (string, int, bool, arrays)
- **Format Validation**: Regex patterns enforce naming conventions (e.g., lowercase names)
- **Constraint Enforcement**: Length limits, required fields, value constraints
- **Extensibility**: Open schemas (`...`) allow additional fields without breaking validation

## Schema Locations

Schemas are embedded in the cclint binary via `//go:embed`:

```
internal/cue/schemas/
├── agent.cue      # Agent frontmatter schema
├── command.cue    # Command frontmatter schema
├── skill.cue      # Skill frontmatter schema
└── settings.cue   # Settings.json schema
```

## Agent Schema

**File**: `internal/cue/schemas/agent.cue`

**Source**: [https://code.claude.com/docs/en/sub-agents](https://code.claude.com/docs/en/sub-agents)

### Required Fields

| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| `name` | string | `^[a-z0-9-]+$`, max 64 chars | Lowercase, numbers, hyphens only |
| `description` | string | non-empty, max 1024 chars | Agent description |

### Optional Fields

| Field | Type | Values | Description |
|-------|------|--------|-------------|
| `model` | string | `sonnet`, `opus`, `haiku`, `sonnet[1m]`, `opusplan`, `inherit` | Model preference |
| `color` | string | `red`, `blue`, `green`, `yellow`, `purple`, `orange`, `pink`, `cyan` | Terminal color |
| `tools` | `*`, string, array | `*` for all, or tool names | Tool access allowlist (agents use `tools:` not `allowed-tools:`) |
| `disallowedTools` | `*`, string, array | Tool names | Tool access denylist |
| `permissionMode` | string | `default`, `acceptEdits`, `dontAsk`, `bypassPermissions`, `plan` | Permission handling |
| `skills` | string | comma-separated skill names | Skills to preload into agent context |
| `hooks` | object | Event → hooks mapping | Agent-level hooks (PreToolUse, PostToolUse, Stop) |

### Known Tools

```
Read, Write, Edit, MultiEdit, Bash, Grep, Glob, LS, Task, NotebookEdit,
WebFetch, WebSearch, TodoWrite, BashOutput, KillBash, ExitPlanMode,
AskUserQuestion, LSP, Skill, DBClient
```

### Hooks Format

```yaml
hooks:
  PreToolUse:
    - matcher: "Bash"
      hooks:
        - type: command
          command: /path/to/script.sh
          timeout: 30
          once: false
  Stop:
    - hooks:
        - type: command
          command: /cleanup.sh
```

**Hook Events**: `PreToolUse`, `PostToolUse`, `Stop`, `Start`

**Hook Fields**:
- `matcher` (optional): Tool name pattern to match
- `hooks`: Array of hook commands
  - `type`: Must be `"command"`
  - `command`: Shell command to execute
  - `timeout`: Timeout in seconds (optional)
  - `once`: Run only once per session (optional, v2.1.0+)

### Example

```yaml
---
name: code-poet-agent
description: Deep code-improvement specialist for architectural pattern detection
model: opus
color: purple
tools:
  - Read
  - Write
  - Edit
  - Grep
permissionMode: default
skills: claude-code-skills,component-optimization-patterns
hooks:
  PreToolUse:
    - matcher: Write
      hooks:
        - type: command
          command: .claude/hooks/pre-write-check
---
```

## Command Schema

**File**: `internal/cue/schemas/command.cue`

**Source**: [https://code.claude.com/docs/en/slash-commands](https://code.claude.com/docs/en/slash-commands)

### Required Fields

**None** - All fields are optional per Claude Code specification.

### Optional Fields

| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| `name` | string | `^[a-z0-9-]+$`, max 64 chars | Command name (defaults to filename) |
| `description` | string | max 1024 chars | Command description for help |
| `allowed-tools` | `*`, string, array | `*` for all, or tool names | Tools available to command (commands use `allowed-tools:` not `tools:`) |
| `argument-hint` | string | - | Argument hint text for help |
| `model` | string | `sonnet`, `opus`, `haiku`, `sonnet[1m]`, `opusplan`, `inherit` | Model preference |
| `disable-model-invocation` | bool | - | Prevent SlashCommand tool from calling |
| `hooks` | object | Event → hooks mapping | Command-level hooks |

### Hooks Format

Same format as Agent hooks (PreToolUse, PostToolUse, Stop).

### Example

```yaml
---
name: commit
description: Repository hygiene + commit organization
model: sonnet
allowed-tools:
  - Bash
  - Grep
  - Glob
argument-hint: "[options]"
hooks:
  PreToolUse:
    - matcher: Bash
      hooks:
        - type: command
          command: .claude/hooks/pre-commit-check
---
```

## Skill Schema

**File**: `internal/cue/schemas/skill.cue`

**Sources**:
- [https://code.claude.com/docs/en/skills](https://code.claude.com/docs/en/skills)
- [https://agentskills.io/specification](https://agentskills.io/specification)

### Required Fields

| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| `name` | string | `^[a-z0-9]+(-[a-z0-9]+)*$`, max 64 chars | No leading/trailing/consecutive hyphens |
| `description` | string | non-empty, max 1024 chars | Skill description |

### Optional Claude Code Fields

| Field | Type | Values | Description |
|-------|------|--------|-------------|
| `argument-hint` | string | e.g., `[issue-number]` | Hint shown during autocomplete |
| `disable-model-invocation` | bool | `false` (default) | Prevent Claude from auto-loading this skill |
| `user-invocable` | bool | `true` (default) | Include in slash command menu |
| `allowed-tools` | `*`, string, array | `*` for all, or tool names | Tools available (skills use `allowed-tools:` not `tools:`) |
| `model` | string | `sonnet`, `opus`, `haiku`, `sonnet[1m]`, `opusplan`, `inherit`, or full model name | Model preference |
| `context` | string | `fork` | Run in forked sub-agent context |
| `agent` | string | Agent name | Agent type for execution |
| `hooks` | object | Event → hooks mapping | Skill-level hooks (PreToolUse, PostToolUse, Stop) |

### Optional AgentSkills.io Fields

| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| `license` | string | SPDX identifier | License identifier or file reference |
| `compatibility` | string | max 500 chars | Environment requirements |
| `metadata` | object | String/number/bool values | Arbitrary key-value mapping |

### String Substitutions (v2.1.9+)

Available in skill hook commands:
- `${CLAUDE_SESSION_ID}` - Current session ID

### Hooks Format

Same format as Agent hooks (PreToolUse, PostToolUse, Stop).

### Example

```yaml
---
name: react-patterns
description: Implements core React 19+ patterns including hooks, state management, and accessibility
model: sonnet
allowed-tools: "*"
context: fork
user-invocable: true
license: MIT
compatibility: React 19+, Node.js 20+
metadata:
  version: "1.0.0"
  author: Example Author
hooks:
  PreToolUse:
    - matcher: Write
      hooks:
        - type: command
          command: .claude/hooks/react-pre-write
---
```

## Settings Schema

**File**: `internal/cue/schemas/settings.cue`

**Target**: `.claude/settings.json`

### Optional Fields

| Field | Type | Description |
|-------|------|-------------|
| `hooks` | object | Event → hooks mapping |
| `language` | string | Response language (e.g., `"japanese"`, `"spanish"`) - v2.1.0+ |
| `respectGitignore` | bool | Git integration - per-project @-mention picker - v2.1.0+ |
| `plansDirectory` | string | Custom plans directory path (default: `.claude/plans`) - v2.1.9+ |
| `model` | string | Default model |
| `permissions` | object | Permission settings |
| `mcp` | object | MCP server settings (supports `auto:N` syntax for tool search threshold) - v2.1.9+ |

### Hooks Format

Hook types include `command`, `prompt`, and `agent` (v2.1.0+):

```json
{
  "hooks": {
    "SessionStart": [
      {
        "matcher": ".*",
        "hooks": [
          {
            "type": "command",
            "command": "/path/to/startup-script.sh",
            "timeout": 30
          }
        ]
      }
    ]
  },
  "language": "spanish",
  "respectGitignore": true
}
```

**Hook Events**: `SessionStart`, `PreToolUse`, `PostToolUse`, `Stop`, etc.

**Hook Fields**:
- `matcher`: Pattern to match (optional for some events)
- `hooks`: Array of hook commands
  - `type`: `command`, `prompt`, or `agent` (v2.1.0+)
  - `command`: Shell command (for type=command)
  - `timeout`: Timeout in seconds, default 30
  - `once`: Run only once per session (v2.1.0+)

## Validation Rules Summary

| Component | Required Fields | Naming Convention | Max Length |
|-----------|-----------------|-------------------|------------|
| Agent | `name`, `description` | `^[a-z0-9-]+$` | name: 64, description: 1024 |
| Command | None | `^[a-z0-9-]+$` (if provided) | name: 64, description: 1024 |
| Skill | `name`, `description` | `^[a-z0-9]+(-[a-z0-9]+)*$` | name: 64, description: 1024 |
| Settings | None | - | - |

## Common Patterns

### Tools vs Allowed-Tools

- **Agents**: Use `tools:` field
- **Commands**: Use `allowed-tools:` field
- **Skills**: Use `allowed-tools:` field

### Model Values

Standard model names:
- `sonnet` - Default model for most tasks
- `opus` - Complex reasoning, architecture
- `haiku` - Quick tasks, lookups
- `sonnet[1m]` - Extended context
- `opusplan` - Planning mode
- `inherit` - Use parent model

Full model names are also accepted (e.g., `claude-sonnet-4-20250514`).

### Color Values (Agents Only)

Only 8 standard colors supported:
`red`, `blue`, `green`, `yellow`, `purple`, `orange`, `pink`, `cyan`
