# Command Lint Rules

Documentation of all validation rules enforced by cclint for command components.

---

## Rules 022-034

### Rule 022: Invalid Name Format

**Severity:** error
**Component:** command
**Category:** structural

**Description:**
Validates that the `name` field in command frontmatter (if present) follows the required format: lowercase alphanumeric characters with hyphens only.

**Pass Criteria:**
- Name contains only characters: `a-z`, `0-9`, `-`
- No uppercase letters, underscores, or special characters
- Name field is optional (derived from filename per Anthropic docs)

**Fail Message:**
`Name must be lowercase alphanumeric with hyphens only`

**Source:** [Anthropic Docs - Slash Commands](https://code.claude.com/docs/en/slash-commands) - Name derived from filename

---

### Rule 023: XML Tags in Description

**Severity:** error
**Component:** command
**Category:** structural

**Description:**
Detects XML-like tags in the `description` field of command frontmatter. XML tags are not allowed in command descriptions per Anthropic documentation.

**Pass Criteria:**
- Description field contains no XML-like tags (e.g., `<tag>`, `<div>`, `<example>`)
- Plain text descriptions only

**Fail Message:**
`Description contains XML-like tags which are not allowed`

**Source:** [Anthropic Docs - Skills](https://code.claude.com/docs/en/skills) - No XML tags in text fields

---

### Rule 024: Command Exceeds Line Limit

**Severity:** suggestion
**Component:** command
**Category:** best-practice

**Description:**
Checks that command files stay within the recommended size limit of ~55 lines (50 lines ±10% tolerance). Commands exceeding this threshold should delegate to specialist agents instead of implementing logic directly.

**Pass Criteria:**
- Total line count ≤ 55 lines
- Commands follow thin delegation pattern

**Fail Message:**
`Command is [N] lines. Best practice: keep commands under ~55 lines (50±10%) - delegate to specialist agents instead of implementing logic directly.`

**Source:**
cclint observation (thin/fat architecture pattern)

---

### Rule 025: Direct Implementation Patterns Detected

**Severity:** suggestion
**Component:** command
**Category:** best-practice

**Description:**
Detects implementation sections (`## Implementation`, `### Steps`) in command files. Commands should delegate to specialist agents rather than containing step-by-step implementation details.

**Pass Criteria:**
- No `## Implementation` section
- No `### Steps` section
- Command delegates to agents instead of implementing logic

**Fail Message:**
`Command contains implementation steps. Consider delegating to a specialist agent instead.`

**Source:**
cclint observation (thin command pattern)

---

### Rule 026: Missing allowed-tools with Task() Delegation

**Severity:** suggestion
**Component:** command
**Category:** security

**Description:**
Checks that commands using the `Task()` tool in their body include the `allowed-tools` permission in frontmatter. This ensures explicit permission for agent delegation.

**Pass Criteria:**
- If `Task(` appears in command body, `allowed-tools:` must be present in frontmatter
- Typically `allowed-tools: Task` or specific agent permissions

**Fail Message:**
`Command uses Task() but lacks 'allowed-tools' permission. Add 'allowed-tools: Task' to frontmatter.`

**Source:**
cclint observation (security best practice)

---

### Rule 027: Bloat Section "Quick Reference"

**Severity:** suggestion
**Component:** command
**Category:** best-practice

**Description:**
Detects `## Quick Reference` section in thin commands (those delegating via `Task()`). Quick reference tables belong in skills, not commands.

**Pass Criteria:**
- Thin commands (with `Task()` delegation) do not contain `## Quick Reference`
- Reference content resides in skill files

**Fail Message:**
`Thin command has '## Quick Reference' - belongs in skill, not command`

**Source:**
cclint observation (thin/fat architecture pattern)

---

### Rule 028: Bloat Section "Usage"

**Severity:** suggestion
**Component:** command
**Category:** best-practice

**Description:**
Detects `## Usage` section in thin commands (those delegating via `Task()`). Since thin commands delegate to agents that have full context, usage sections are redundant.

**Pass Criteria:**
- Thin commands (with `Task()` delegation) do not contain `## Usage`
- Usage documentation resides in agent or skill files

**Fail Message:**
`Thin command has '## Usage' - agent has full context, remove`

**Source:**
cclint observation (thin/fat architecture pattern)

---

### Rule 029: Bloat Section "Workflow"

**Severity:** suggestion
**Component:** command
**Category:** best-practice

**Description:**
Detects `## Workflow` section in thin commands (those delegating via `Task()`). Workflow details duplicate agent content and should not appear in thin commands.

**Pass Criteria:**
- Thin commands (with `Task()` delegation) do not contain `## Workflow`
- Workflow documentation resides in agent files

**Fail Message:**
`Thin command has '## Workflow' - duplicates agent content, remove`

**Source:**
cclint observation (thin/fat architecture pattern)

---

### Rule 030: Bloat Section "When to use"

**Severity:** suggestion
**Component:** command
**Category:** best-practice

**Description:**
Detects `## When to use` section in thin commands. This information should be captured in the command's `description` field instead.

**Pass Criteria:**
- Thin commands (with `Task()` delegation) do not contain `## When to use`
- Usage context documented in description field

**Fail Message:**
`Thin command has '## When to use' - belongs in description, remove`

**Source:**
cclint observation (thin/fat architecture pattern)

---

### Rule 031: Bloat Section "What it does"

**Severity:** suggestion
**Component:** command
**Category:** best-practice

**Description:**
Detects `## What it does` section in thin commands. This information should be captured in the command's `description` field instead.

**Pass Criteria:**
- Thin commands (with `Task()` delegation) do not contain `## What it does`
- Functionality described in description field

**Fail Message:**
`Thin command has '## What it does' - belongs in description, remove`

**Source:**
cclint observation (thin/fat architecture pattern)

---

### Rule 032: Excessive Code Examples

**Severity:** suggestion
**Component:** command
**Category:** best-practice

**Description:**
Checks that commands contain no more than 2 code examples. Excessive examples indicate the command may be too complex and should delegate to an agent.

**Pass Criteria:**
- Maximum of 2 code blocks with ```bash or ```shell fences
- Concise example usage only

**Fail Message:**
`Command has [N] code examples. Best practice: max 2 examples.`

**Source:**
cclint observation (documentation best practice)

---

### Rule 033: Success Criteria Not in Checkbox Format

**Severity:** suggestion
**Component:** command
**Category:** documentation

**Description:**
Checks that success criteria sections use checkbox format (`- [ ]`) instead of prose. This makes success conditions explicit and testable.

**Pass Criteria:**
- If `## Success` or `Success criteria:` section exists, it must contain `- [ ]` checkboxes
- Success criteria are actionable and verifiable

**Fail Message:**
`Success criteria should use checkbox format '- [ ]' not prose`

**Source:**
cclint observation (documentation best practice)

---

### Rule 034: Missing Usage Section in Fat Command

**Severity:** suggestion
**Component:** command
**Category:** documentation

**Description:**
Detects fat commands (>40 lines without `Task()` delegation) that lack a `## Usage` or `## Workflow` section. Fat commands implementing logic directly should document their usage, or better yet, delegate to a specialist agent.

**Pass Criteria:**
- Fat commands (>40 lines, no `Task()`) contain `## Usage` or `## Workflow` section
- OR command is refactored to delegate to an agent (preferred)

**Fail Message:**
`Fat command without Task delegation lacks '## Usage' section. Consider delegating to a specialist agent.`

**Source:**
cclint observation (documentation best practice)

---

## New Frontmatter Fields (v2.1.0+)

Claude Code 2.1.0 introduced the `hooks` field for commands:

| Field | Type | Description |
|-------|------|-------------|
| `hooks` | object | Lifecycle hooks scoped to command execution |

### Example with Hooks

```yaml
---
name: my-command
description: Example command with hooks
hooks:
  PreToolUse:
    - matcher: "Write"
      hooks:
        - type: command
          command: echo "Before write"
          once: true
  Stop:
    - hooks:
        - type: command
          command: echo "Command finished"
---
```

---

## Summary

| Rule Range | Category | Count |
|------------|----------|-------|
| 022-023 | Structural (format, schema) | 2 |
| 024-026 | Best Practice (architecture) | 3 |
| 027-031 | Best Practice (bloat detection) | 5 |
| 032-034 | Documentation | 3 |

**Total:** 13 command-specific rules (022-034)

## Related Documentation

- [Agent Rules](agents.md)
- [Skill Rules](skills.md)
- [Quality Scoring](../scoring.md)
