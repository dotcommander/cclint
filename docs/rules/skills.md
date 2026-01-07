# Skill Lint Rules

Validation rules enforced by cclint for skill components.

---

## Overview

Skills must be named `SKILL.md` and follow Anthropic's documentation standards for skill discovery and composition. The linter enforces structural requirements, best practices, and documentation standards to ensure skills are discoverable and maintainable.

## Rule Index

| Rule | Title | Severity |
|------|-------|----------|
| 035 | [Skill file must be named SKILL.md](#rule-035-skill-file-must-be-named-skillmd) | error |
| 036 | [Empty skill file](#rule-036-empty-skill-file) | error |
| 037 | [Missing frontmatter](#rule-037-missing-frontmatter) | suggestion |
| 038 | [Reserved word in name](#rule-038-reserved-word-in-name) | error |
| 039 | [XML tags in description](#rule-039-xml-tags-in-description) | error |
| 040 | [First-person description](#rule-040-first-person-description) | suggestion |
| 041 | [Addressing user in description](#rule-041-addressing-user-in-description) | suggestion |
| 042 | [Description too short](#rule-042-description-too-short) | suggestion |
| 043 | [Missing trigger phrases](#rule-043-missing-trigger-phrases) | suggestion |
| 044 | [Invalid semver version format](#rule-044-invalid-semver-version-format) | warning |
| 045 | [Missing Anti-Patterns section](#rule-045-missing-anti-patterns-section) | suggestion |
| 046 | [Skill exceeds 550 lines](#rule-046-skill-exceeds-550-lines) | suggestion |
| 047 | [Missing Examples section](#rule-047-missing-examples-section) | suggestion |

---

## Rules

### Rule 035: Skill file must be named SKILL.md

**Severity:** error
**Component:** skill
**Category:** structural

**Description:**
Skill files must be named `SKILL.md` (case-sensitive) according to Anthropic's documentation standards. This naming convention is critical for skill discovery and loading.

**Pass Criteria:**
- File path ends with `/SKILL.md` (Unix) or `\SKILL.md` (Windows)

**Fail Message:**
`Skill file must be named SKILL.md`

**Source:** [Anthropic Docs - Skills](https://code.claude.com/docs/en/skills) - SKILL.md naming convention

---

### Rule 036: Empty skill file

**Severity:** error
**Component:** skill
**Category:** structural

**Description:**
Skill files must contain content. Empty files provide no value and cannot be loaded by the skill system.

**Pass Criteria:**
- File contains non-whitespace content

**Fail Message:**
`Skill file is empty`

**Source:** [Anthropic Docs - Skills](https://code.claude.com/docs/en/skills) - Skills must contain content

---

### Rule 037: Missing frontmatter

**Severity:** suggestion
**Component:** skill
**Category:** documentation

**Description:**
Skills should include YAML frontmatter with `name` and `description` fields. The description is critical for skill discovery, allowing Claude to find and load appropriate skills based on task requirements.

**Pass Criteria:**
- File starts with `---` (YAML frontmatter delimiter)

**Fail Message:**
`Add YAML frontmatter with name and description (description is critical for skill discovery)`

Or when skill name can be extracted:
`Add frontmatter: ---\nname: {skill-name}\ndescription: Brief summary of what this skill does\n--- (description critical for discovery)`

**Source:** [Anthropic Docs - Skills](https://code.claude.com/docs/en/skills) - Frontmatter with name/description for discovery

---

### Rule 038: Reserved word in name

**Severity:** error
**Component:** skill
**Category:** best-practice

**Description:**
Skill names cannot use reserved words like "anthropic" or "claude". These are reserved for official components.

**Pass Criteria:**
- `name` field does not equal "anthropic" or "claude" (case-insensitive)

**Fail Message:**
`Name '{name}' is a reserved word and cannot be used`

**Source:** [Anthropic Docs - Skills](https://code.claude.com/docs/en/skills) - Reserved words blocked

---

### Rule 039: XML tags in description

**Severity:** error
**Component:** skill
**Category:** best-practice

**Description:**
Skill descriptions must not contain XML-like tags (e.g., `<tag>`). These interfere with skill discovery and processing.

**Pass Criteria:**
- `description` field does not match pattern `<[a-zA-Z][^>]*>`

**Fail Message:**
`Description contains XML-like tags which are not allowed`

**Source:** [Anthropic Docs - Skills](https://code.claude.com/docs/en/skills) - No XML tags in text fields

---

### Rule 040: First-person description

**Severity:** suggestion
**Component:** skill
**Category:** best-practice

**Description:**
Skill descriptions should use third-person perspective (e.g., "Analyzes..." not "I analyze..."). This maintains consistency with Anthropic's documentation standards and improves skill discovery.

**Pass Criteria:**
- `description` does not start with: "I ", "I'm ", "I'll ", "I've ", "My ", "We ", "We're "

**Fail Message:**
`Skill description should use third person (e.g., 'Analyzes...' not 'I analyze...')`

**Source:** [Anthropic Docs - Skills](https://code.claude.com/docs/en/skills) - Third-person descriptions

---

### Rule 041: Addressing user in description

**Severity:** suggestion
**Component:** skill
**Category:** best-practice

**Description:**
Skill descriptions should describe what the skill does, not address the user directly (e.g., "You can..."). This improves clarity and skill discovery.

**Pass Criteria:**
- `description` does not start with "You "

**Fail Message:**
`Skill description should describe what it does, not address the user`

**Source:** [Anthropic Docs - Skills](https://code.claude.com/docs/en/skills) - Description style guidelines

---

### Rule 042: Description too short

**Severity:** suggestion
**Component:** skill
**Category:** documentation

**Description:**
Skill descriptions should be at least 50 characters to provide sufficient detail for skill discovery. Shorter descriptions may not provide enough context for Claude to determine when to use the skill.

**Pass Criteria:**
- `description` field is 50+ characters

**Fail Message:**
`Description is only {length} chars. Aim for 50+ to help with skill discovery.`

**Source:** [Anthropic Docs - Skills](https://code.claude.com/docs/en/skills) - Description critical for discovery

---

### Rule 043: Missing trigger phrases

**Severity:** suggestion
**Component:** skill
**Category:** documentation

**Description:**
Skill descriptions should include trigger phrases like "Use when...", "Use for...", or "Use proactively..." to help Claude understand when to invoke the skill.

**Pass Criteria:**
- `description` contains at least one of: "use when", "use for", "use proactively" (case-insensitive)

**Fail Message:**
`Consider adding trigger phrases like 'Use when...' or 'Use for...' to help skill discovery`

**Source:** [Anthropic Docs - Skills](https://code.claude.com/docs/en/skills) - Trigger phrase patterns

---

### Rule 044: Invalid semver version format

**Severity:** warning
**Component:** skill
**Category:** best-practice

**Description:**
If a `version` field is provided in frontmatter, it should follow semantic versioning format (e.g., "1.0.0"). This enables version tracking and compatibility checking.

**Pass Criteria:**
- `version` field matches pattern: `^\d+\.\d+\.\d+(-[a-zA-Z0-9.]+)?(\+[a-zA-Z0-9.]+)?$`
- Examples: "1.0.0", "2.1.3-beta", "1.0.0+build.123"

**Fail Message:**
`Version '{version}' should follow semver format (e.g., '1.0.0')`

**Source:** cclint observation - Semantic versioning for skill tracking

---

### Rule 045: Missing Anti-Patterns section

**Severity:** suggestion
**Component:** skill
**Category:** documentation

**Description:**
Skills should include an "Anti-Patterns" section documenting common mistakes, incorrect usage patterns, or approaches to avoid. This helps users understand not just how to use the skill correctly, but what to avoid.

**Pass Criteria:**
Document contains at least one of:
- `## Anti-Pattern` or `## Anti-Patterns`
- `### Anti-Pattern`
- `## Best Practices` with `### don't` subsection
- Table with `| Anti-Pattern` column

**Fail Message:**
`Consider adding '## Anti-Patterns' section to document common mistakes.`

**Source:** cclint observation - Best practice for methodology documentation

---

### Rule 046: Skill exceeds 550 lines

**Severity:** suggestion
**Component:** skill
**Category:** structural

**Description:**
Skills should be kept under ~550 lines (500 lines with 10% tolerance). Larger skills become difficult to maintain and slow to load. Heavy documentation, schemas, or examples should be moved to a `references/` subdirectory and loaded on demand.

**Pass Criteria:**
- File has 550 or fewer lines

**Fail Message:**
`Skill is {lines} lines. Best practice: keep skills under ~550 lines (500±10%) - move heavy docs to references/ subdirectory.`

**Source:** cclint observation - Skills should use references/ for heavy content

---

### Rule 047: Missing Examples section

**Severity:** suggestion
**Component:** skill
**Category:** documentation

**Description:**
Skills should include examples demonstrating usage. This helps users understand how to apply the skill in practice.

**Pass Criteria:**
Document contains at least one of:
- `## Example` or `## Examples`
- `## Expected Output`
- `## Usage`
- `### Example`

**Fail Message:**
`Consider adding '## Examples' section to illustrate skill usage.`

**Source:** cclint observation - Documentation best practice

---

## New Frontmatter Fields (v2.1.0+)

Claude Code 2.1.0 introduced several new optional frontmatter fields for skills:

| Field | Type | Description |
|-------|------|-------------|
| `context` | `"fork"` | Run skill in forked sub-agent context |
| `agent` | string | Agent type for execution (e.g., "refactor-specialist") |
| `user-invocable` | bool | Show in slash command menu (default: true for /skills/) |
| `hooks` | object | Lifecycle hooks scoped to skill execution |

### Example with New Fields

```yaml
---
name: my-skill
description: Example skill with v2.1.0 fields
context: fork
agent: refactor-specialist
user-invocable: true
hooks:
  PreToolUse:
    - matcher: "Bash"
      hooks:
        - type: command
          command: echo "Before bash"
          once: true
---
```

---

## Best Practices

### Skill Structure

A well-structured skill should include:

1. **YAML frontmatter** with `name` and detailed `description`
2. **Quick Reference** table (semantic routing)
3. **Workflow** section with step-by-step methodology
4. **Anti-Patterns** section documenting what to avoid
5. **Examples** section demonstrating usage
6. **References** subdirectory for heavy documentation (if needed)

### Size Guidelines

| Component | Lines | Action |
|-----------|-------|--------|
| SKILL.md | <550 | Core methodology only |
| references/ | unlimited | Heavy docs, schemas, examples |

### Description Guidelines

Good descriptions:
- ✅ Use third-person: "Analyzes API patterns..."
- ✅ Include trigger phrases: "Use when implementing OpenAI APIs..."
- ✅ 50+ characters for context
- ✅ Describe what, not who

Bad descriptions:
- ❌ First-person: "I analyze API patterns..."
- ❌ Addressing user: "You can use this to..."
- ❌ Too short: "API helper"
- ❌ Contains XML: "Use <openai> patterns"

---

## Related Documentation

- [Agent Lint Rules](agents.md)
- [Command Lint Rules](commands.md)
- [Quality Scoring](../scoring.md)
- [Cross-File Validation](../cross-file-validation.md)
