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
| 046 | [Skill exceeds 500 lines](#rule-046-skill-exceeds-500-lines) | suggestion |
| 047 | [Missing Examples section](#rule-047-missing-examples-section) | suggestion |
| 048 | [Name cannot start/end with hyphen](#rule-048-name-cannot-startend-with-hyphen) | error |
| 049 | [Consecutive hyphens in name](#rule-049-consecutive-hyphens-in-name) | error |
| 050 | [Name must match directory](#rule-050-name-must-match-directory) | warning |
| 052 | [Invalid allowed-tools format](#rule-052-invalid-allowed-tools-format) | warning |
| 053 | [Empty license field](#rule-053-empty-license-field) | suggestion |
| 054 | [Compatibility field too long](#rule-054-compatibility-field-too-long) | warning |
| 055 | [Invalid metadata structure](#rule-055-invalid-metadata-structure) | suggestion |
| 056 | [Script missing shebang](#rule-056-script-missing-shebang) | suggestion |
| 057 | [Script not executable](#rule-057-script-not-executable) | suggestion |
| 058 | [Reference file too large](#rule-058-reference-file-too-large) | suggestion |
| 059 | [Absolute path in markdown link](#rule-059-absolute-path-in-markdown-link) | warning |
| 060 | [Reference chain too deep](#rule-060-reference-chain-too-deep) | suggestion |

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

### Rule 048: Name cannot start/end with hyphen

**Severity:** error
**Component:** skill
**Category:** structural

**Description:**
Per agentskills.io specification, skill names must not start or end with a hyphen. This ensures compatibility with directory naming conventions and agent discovery systems.

**Pass Criteria:**
- `name` field does not start with `-`
- `name` field does not end with `-`

**Fail Message:**
`Skill name '{name}' cannot start or end with a hyphen`

**Example violations:**
- `-my-skill` (starts with hyphen)
- `my-skill-` (ends with hyphen)

**Source:** [agentskills.io specification](https://agentskills.io/specification) - Name validation rules

---

### Rule 049: Consecutive hyphens in name

**Severity:** error
**Component:** skill
**Category:** structural

**Description:**
Per agentskills.io specification, skill names must not contain consecutive hyphens (`--`). This ensures clean naming conventions and prevents parsing ambiguities.

**Pass Criteria:**
- `name` field does not contain `--`

**Fail Message:**
`Skill name '{name}' contains consecutive hyphens (--) which are not allowed`

**Example violations:**
- `my--skill`
- `data--analysis--tool`

**Source:** [agentskills.io specification](https://agentskills.io/specification) - Name validation rules

---

### Rule 050: Name must match directory

**Severity:** warning
**Component:** skill
**Category:** structural

**Description:**
Per agentskills.io specification, the skill name in frontmatter should match the parent directory name. This is important for skill discovery and prevents confusion.

**Pass Criteria:**
- `name` field equals parent directory name (case-sensitive)
- Skipped for root-level or special directories (`.`, `skills`, `.claude`)

**Fail Message:**
`Skill name '{name}' should match parent directory name '{directory}'`

**Example:**
```
# Directory: pdf-processing/SKILL.md
# Frontmatter: name: pdf-tools
❌ Mismatch: "pdf-tools" vs "pdf-processing"
```

**Source:** [agentskills.io specification](https://agentskills.io/specification) - Directory matching requirement

---

### Rule 052: Invalid allowed-tools format

**Severity:** warning
**Component:** skill
**Category:** best-practice

**Description:**
Per agentskills.io specification, the `allowed-tools` field should use space-delimited tool names with optional scope syntax (e.g., `Bash(git:*) Read Write`).

**Pass Criteria:**
- `allowed-tools` is `"*"` (wildcard), OR
- Each token matches pattern: starts with uppercase, optionally followed by scope in parentheses

**Fail Message:**
`allowed-tools format should be space-delimited tool names (e.g., 'Bash(git:*) Read Write')`

**Valid examples:**
```yaml
allowed-tools: "*"
allowed-tools: "Read Write Edit"
allowed-tools: "Bash(git:*) Bash(jq:*) Read"
```

**Source:** [agentskills.io specification](https://agentskills.io/specification) - Tool restrictions format

---

### Rule 053: Empty license field

**Severity:** suggestion
**Component:** skill
**Category:** documentation

**Description:**
Per agentskills.io specification, if a `license` field is provided, it should contain a valid SPDX identifier or reference to a bundled license file.

**Pass Criteria:**
- If `license` field present: non-empty string

**Fail Message:**
`license field is empty - provide SPDX identifier (e.g., 'MIT', 'Apache-2.0') or license file reference`

**Valid examples:**
```yaml
license: MIT
license: Apache-2.0
license: See LICENSE.txt
```

**Source:** [agentskills.io specification](https://agentskills.io/specification) - License field

---

### Rule 054: Compatibility field too long

**Severity:** warning
**Component:** skill
**Category:** documentation

**Description:**
Per agentskills.io specification, the optional `compatibility` field must not exceed 500 characters. This field describes environment requirements.

**Pass Criteria:**
- If `compatibility` field present: ≤500 characters

**Fail Message:**
`compatibility field is {length} chars (max 500 per agentskills.io spec)`

**Valid example:**
```yaml
compatibility: Designed for Claude Code. Requires git and internet access.
```

**Source:** [agentskills.io specification](https://agentskills.io/specification) - Compatibility field constraints

---

### Rule 055: Invalid metadata structure

**Severity:** suggestion
**Component:** skill
**Category:** documentation

**Description:**
Per agentskills.io specification, the optional `metadata` field should be a key-value mapping with primitive values (string, number, boolean).

**Pass Criteria:**
- If `metadata` field present: must be a mapping
- All values should be primitives

**Fail Message:**
`metadata field should be key-value mapping` or `metadata.{key} should be primitive value`

**Valid example:**
```yaml
metadata:
  author: example-org
  version: "1.0"
  category: data-processing
```

**Source:** [agentskills.io specification](https://agentskills.io/specification) - Metadata field structure

---

### Rule 056: Script missing shebang

**Severity:** suggestion
**Component:** skill
**Category:** best-practice

**Description:**
Per agentskills.io specification, scripts in the `scripts/` directory should include a shebang line for portability. This ensures scripts can be executed directly without specifying the interpreter.

**Pass Criteria:**
- Script files (.sh, .py, .js, etc.) start with `#!`

**Fail Message:**
`Script 'scripts/{filename}' missing shebang (e.g., #!/usr/bin/env python3)`

**Valid examples:**
```bash
#!/usr/bin/env bash
#!/usr/bin/env python3
#!/usr/bin/env node
```

**Source:** [agentskills.io specification](https://agentskills.io/specification) - Script best practices

---

### Rule 057: Script not executable

**Severity:** suggestion
**Component:** skill
**Category:** best-practice

**Description:**
Per agentskills.io specification, scripts should have executable permissions set. This allows agents to run scripts directly.

**Pass Criteria:**
- Script files have executable bit set (mode & 0111)

**Fail Message:**
`Script 'scripts/{filename}' is not executable (chmod +x)`

**Source:** [agentskills.io specification](https://agentskills.io/specification) - Script best practices

---

### Rule 058: Absolute path in markdown link

**Severity:** warning
**Component:** skill
**Category:** best-practice

**Description:**
Per agentskills.io specification, file references should use relative paths from the skill root directory. Absolute paths break portability.

**Pass Criteria:**
- Markdown links don't start with `/` or `~`

**Fail Message:**
`Use relative path instead of absolute: '{path}'`

**Valid example:**
```markdown
See [reference guide](references/REFERENCE.md) for details.
```

**Invalid example:**
```markdown
See [reference guide](/Users/foo/.claude/skills/my-skill/references/REFERENCE.md)
```

**Source:** [agentskills.io specification](https://agentskills.io/specification) - File references

---

### Rule 059: Reference chain too deep

**Severity:** suggestion
**Component:** skill
**Category:** best-practice

**Description:**
Per agentskills.io specification, file references should be kept one level deep from SKILL.md. Nested reference chains require multiple Read() calls and waste context.

**Pass Criteria:**
- Referenced files don't contain links to other reference files

**Fail Message:**
`Reference chain detected: SKILL.md → {ref1} → {ref2} (keep references 1 level deep)`

**Good example:**
```
SKILL.md → references/REFERENCE.md  ✅
SKILL.md → references/advanced/TOPIC.md  ✅
```

**Bad example:**
```
SKILL.md → references/REFERENCE.md → references/DEEP.md  ❌
```

**Source:** [agentskills.io specification](https://agentskills.io/specification) - Progressive disclosure architecture

---

## New Frontmatter Fields

### Claude Code Fields (v2.1.0+)

| Field | Type | Description |
|-------|------|-------------|
| `context` | `"fork"` | Run skill in forked sub-agent context |
| `agent` | string | Agent type for execution (e.g., "refactor-specialist") |
| `user-invocable` | bool | Show in slash command menu (default: true for /skills/) |
| `hooks` | object | Lifecycle hooks scoped to skill execution |

### agentskills.io Fields

Per [agentskills.io specification](https://agentskills.io/specification):

| Field | Type | Description |
|-------|------|-------------|
| `license` | string | SPDX identifier (e.g., "MIT", "Apache-2.0") or license file reference |
| `compatibility` | string | Environment requirements (max 500 chars) |
| `metadata` | object | Arbitrary key-value mapping for additional properties |

### Example with All Fields

```yaml
---
name: my-skill
description: Example skill demonstrating all optional fields. Use when working with PDF documents.
# Claude Code fields
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
# agentskills.io fields
license: MIT
compatibility: Designed for Claude Code. Requires git.
metadata:
  author: example-org
  version: "1.0"
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

Per [agentskills.io specification](https://agentskills.io/specification):

| Component | Lines | Action |
|-----------|-------|--------|
| SKILL.md | <500 | Core methodology only |
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
