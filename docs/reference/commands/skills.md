# cclint skills

Lint and validate skill definition files.

---

## Usage

```bash
cclint skills [flags]
```

---

## Description

The `skills` subcommand scans for and validates skill definition files in
your Claude Code project. Skills contain methodology, patterns, and reference
material for specialist domains.

### Supported File Patterns

- `.claude/skills/**/SKILL.md`
- `skills/**/SKILL.md`

### What Gets Validated

- Filename must be `SKILL.md`
- Frontmatter structure and required fields
- Name format (agentskills.io compliance)
- Line count limit (500 lines ±10% tolerance)
- Description quality (third-person, trigger phrases)
- Required sections (Quick Reference, Anti-Patterns)
- Directory structure validation

---

## Line Limit Rule

Skills must stay under **~550 lines** (500 lines ±10% tolerance).

### Rationale

Skills should contain core methodology only. Heavy documentation, schemas, or
examples should be moved to a `references/` subdirectory and loaded on demand.

### Example Violation Message

```text
my-skill/SKILL.md:600
suggestion: Skill is 600 lines. Best practice: keep skills under ~550
lines (500±10%) - move heavy docs to references/ subdirectory.
```

### What To Do Instead

Extract heavy content to reference files:

```markdown
# Core methodology in SKILL.md (<500 lines)

## Quick Reference

| User Question | Action |
|---------------|--------|
| Need detailed schema? | See [Schema Reference](references/schema.md) |
```

---

## SKILL.md Naming Rule

Skill files must be named `SKILL.md` (case-sensitive).

### Rationale

This naming convention is required by Anthropic's documentation standards for
skill discovery and loading.

### Example Violation Message

```text
my-skill/skill.md:1
error: Skill file must be named SKILL.md
```

---

## Quick Reference Requirement

Skills should include a Quick Reference section for semantic routing.

### Rationale

The Quick Reference table acts as a semantic routing guide, helping Claude
understand when to use the skill based on user questions or task context.

### Validation

- Checks for `## Quick Reference` heading (case-insensitive)
- Table format recommended: `| User Question | Action |`
- Weighted 8-10 points in quality scoring

### Example Format

```markdown
## Quick Reference

| User Question | Action |
|---------------|--------|
| Need to parse JSON? | Use for JSON parsing tasks |
| Implementing API client? | Use when building HTTP clients |
| Debugging async code? | Use proactively for async bugs |
```

---

## Frontmatter Fields

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `name` | string | Skill identifier (lowercase, hyphens) |
| `description` | string | What skill does (50+ chars, critical for discovery) |

### Optional Claude Code Fields

| Field | Type | Description |
|-------|------|-------------|
| `allowed-tools` | string | Tool access permissions (e.g., `"Bash Read Write"`) |
| `model` | string | Model to use when skill is active |
| `context` | string | `"fork"` for sub-agent context |
| `agent` | string | Agent type for execution |
| `user-invocable` | bool | Show in slash command menu |
| `hooks` | object | Skill-level lifecycle hooks |

### Optional agentskills.io Fields

| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| `license` | string | SPDX identifier or file ref | License information |
| `compatibility` | string | max 500 chars | Environment requirements |
| `metadata` | object | primitive values only | Arbitrary key-value pairs |

---

## Example Output

### Passing Skill

```text
✓ my-skill/SKILL.md (425 lines)
  Score: 88/100 (A)
```

### Failing Skill

```text
✗ my-skill/skill.md
  errors:
    - Line 1: Skill file must be named SKILL.md

  warnings:
    - Line 5: Skill name 'my-skill' should match parent directory name 'my-tool'

  suggestions:
    - Line 1: Add YAML frontmatter with name and description
    - Line 480: Consider adding '## Anti-Patterns' section
    - Line 600: Skill is 600 lines. Best practice: keep skills under
      ~550 lines (500±10%)
```

---

## See Also

- [Skills Lint Rules](../../rules/skills.md)
- [Scoring Reference](../../scoring/README.md)
- [Writing Skills Guide](../../guides/writing-skills.md)
- [Common Tasks](../../common-tasks.md)
