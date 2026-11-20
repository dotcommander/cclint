# Linting by Component Type

cclint selects component types with positional arguments on the default lint
command — `cclint agents commands` is not a subcommand, it is the lint command
with two type filters. A run with no type filter lints every default type
discovered under the root.

Rule-level detail lives in [lint rules](../rules/README.md); frontmatter field
schemas live in the [schema reference](schemas.md).

## Usage

```bash
cclint                          # all default types
cclint agents commands          # selected types
cclint ./agents/reviewer.md     # path mode: specific files or directories
```

## Types at a Glance

| Type filter | Files discovered | Size guidance |
|------------|------------------|---------------|
| `agents` | `.claude/agents/**/*.md`, `agents/**/*.md`, plugin-shipped | ~200 lines |
| `commands` | `.claude/commands/**/*.md`, `commands/**/*.md`, plugin-shipped | ~55 lines (50 ±10%) |
| `skills` | `.claude/skills/**/SKILL.md`, `skills/**/SKILL.md`, plugin-shipped | ~550 lines (500 ±10%) |
| `plugins` | `.claude-plugin/plugin.json` (plugin repo root), `plugins/**/.claude-plugin/plugin.json`, `.claude/plugins/**/.claude-plugin/plugin.json` | ≤5,000 bytes |
| `settings` | `.claude/settings.json`, `.claude/settings.local.json`, layout-root `settings.json` (e.g. `~/.claude`) | — |
| `rules` | `.claude/rules/**/*.md`, `rules/**/*.md`, plugin-shipped | — |
| `output-styles` | `.claude/output-styles/**/*.md`, `output-styles/**/*.md`, plugin-shipped | — |
| `context` | `CLAUDE.md`, `CLAUDE.local.md`, `AGENTS.md`, `.claude/CLAUDE.md`, `.claude/AGENTS.md` | — |

Notes:

- **Plugin-shipped components** are anchored to the trees Claude Code loads
  them from — `plugins/marketplaces/<marketplace>/plugins/<plugin>/<type>/`
  and the legacy `plugins/cache/<a>/<b>/current/` trees. They are linted under
  their own component types; the `plugins` filter covers manifests only.
- **`context` is not part of a default run.** Pass `cclint context` to lint
  context files explicitly.

## Agents

Agents are specialist definitions with frontmatter-declared tools. Validated:
frontmatter structure (name format, model, enum fields), Foundation/Workflow
section structure, `Skill()` loading references, and body-tool consistency
(suppressed when 8+ tools are declared — a general-purpose agent). Bodies have
a 200-line guidance limit.

See [agent lint rules](../rules/agents.md), [writing agents](../guides/writing-agents.md),
and the [agent schema](schemas.md#agent-schema).

## Commands

Commands work best as thin routes from user invocations to specialist agents.
Validated: frontmatter structure and required fields, name format, line count,
delegation-pattern compliance, bloat-section detection, and load-blocking name
characters (hidden/bidi/blank-like, `<`, `>`, over 256 characters in file,
folder, or frontmatter names).

### Line Limit

Commands must stay under **~55 lines** (50 lines ±10% tolerance). Commands
follow the **thin delegation pattern**: they route to specialist agents rather
than implementing logic. A command past 55 lines signals methodology that
belongs in an agent or skill.

```text
my-command.md:60
suggestion: Command is 60 lines. Best practice: keep commands under ~55
lines (50±10%) - delegate to specialist agents instead of implementing
logic directly.
```

### Delegation Pattern

A thin command delegates via `Task()` or `Agent()` with a delegation-tool
allowlist:

```markdown
---
name: my-command
description: Example thin command
allowed-tools: Task
---

Use the specialist agent for this task.
```

A fat command contains implementation sections — `## Implementation` steps,
`### Steps` sections, inline code examples without delegation — and should
delegate instead:

- Rule 025 detects implementation sections
- Rule 026 suggests `allowed-tools: Task` when `Task()` is used without an allowlist
- Rules 027-031 detect bloat sections in thin commands

See [command lint rules](../rules/commands.md) for the full rule set.

## Skills

Skills contain methodology, patterns, and reference material. Validated: the
`SKILL.md` filename (case-sensitive; lowercase `skill.md` is discovered only
so its naming error can be reported), frontmatter structure, name format,
description quality, line count, suggested sections, and directory structure.
Frontmatter fields are documented in the [skill schema](schemas.md#skill-schema).

### Line Limit

Skills must stay under **~550 lines** (500 lines ±10% tolerance). SKILL.md is
the entry point; heavy documentation, schemas, and examples move to a
`references/` subdirectory loaded on demand.

```text
my-skill/SKILL.md:600
suggestion: Skill is 600 lines. Best practice: keep skills under ~550
lines (500±10%) - move heavy docs to references/ subdirectory.
```

See [writing skills](../guides/writing-skills.md) for the thin→fat methodology.

### Quick Reference Section

A `## Quick Reference` table acts as a semantic routing guide, helping Claude
choose the skill from user questions or task context. Absence reduces quality
scoring (8-10 points); it does not affect schema validity.

```markdown
## Quick Reference

| User Question | Action |
|---------------|--------|
| Need to parse JSON? | Use for JSON parsing tasks |
| Implementing API client? | Use when building HTTP clients |
```

## Plugins

Plugin manifests are JSON files with required metadata. Validated: required
fields, name format (lowercase alphanumeric with hyphens, max 64 characters,
reserved words `anthropic` and `claude` prohibited), semver version format,
and description quality (50+ characters recommended).

### Required and Recommended Fields

| Field | Type | Constraints |
|-------|------|-------------|
| `name` | string | max 64 chars, lowercase alphanumeric with hyphens |
| `description` | string | max 1024 chars, 50+ recommended |
| `author.name` | string | non-empty |

Recommended: `version` (semver), `homepage`, `repository`, `license` (SPDX
identifier), `keywords` (discoverability tags), `readme` (path to README).

### Manifest Size

Manifests earn full composition points at ≤1,000 bytes; the scorer uses bytes,
not 1,024-byte KiB:

| Size (bytes) | Composition points |
|------|-------|
| ≤1,000 | 10/10 |
| ≤2,000 | 8/10 |
| ≤5,000 | 6/10 |
| ≤10,000 | 3/10 |
| >10,000 | 0/10 |

```text
my-plugin/.claude-plugin/plugin.json:1
suggestion: Plugin manifest is 8KB. Best practice: keep plugin.json under 5KB
for optimal loading - move documentation to README.md.
```

See [plugin lint rules](../rules/plugins.md) for the full rule set.

## Settings

Claude Code reads exactly three discoverable settings files per root:
`.claude/settings.json` (project tier, checked in), `.claude/settings.local.json`
(project tier, not checked in), and `settings.json` at a layout root such as
`~/.claude` (user tier).

Validated: hook configuration (event names, matcher and hook fields, hook
types), permission rules and keys, MCP server configuration (`mcp`,
`managedMcpServers`), and security checks. Field-by-field detail lives in the
[settings schema](schemas.md#settings-schema) and
[settings lint rules](../rules/settings.md).

## Rules

Rule files are plain Markdown with optional `paths` frontmatter that scopes a
rule to matching files. Validated: `.md` extension, non-empty content, symlink
targets that exist (warning), and a `paths` value that is a string or list of
valid globs.

## Output Styles

Output styles customize the system prompt. Validated: required `name` (a
free-form display label — no kebab-case constraint) and `description`,
optional `keep-coding-instructions` boolean, a non-empty Markdown body, and
unknown frontmatter fields. Output styles participate in quality scoring.

## Context Files

Context files are `CLAUDE.md`, `CLAUDE.local.md`, `AGENTS.md`,
`.claude/CLAUDE.md`, and `.claude/AGENTS.md`. They are linted explicitly:

```bash
cclint context
```

Validated: section structure, and `@include` directives — includes referencing
binary files (images, archives, fonts) are flagged because Claude Code
silently skips them.

## Path Mode

Passing file or directory paths instead of type filters lints exactly those
files:

```bash
cclint ./agents/reviewer.md a.md b.md   # files (type auto-detected)
cclint ./commands/                      # a directory (singular names work too)
cclint --type agent ./custom/file.md    # override type detection
```

Type detection recognizes component directory names (`agents/`, `commands/`,
`skills/`, `output-styles/`), the `SKILL.md`, `settings.json`, and
`plugin.json` basenames, and `.claude-plugin` paths. Type filters and file
paths cannot be mixed in one invocation.

## Example Output

```text
✓ my-command.md (42 lines)
  Score: 85/100 (B)
```

```text
✗ complex-command.md
  errors:
    - Line 15: Name must be lowercase alphanumeric with hyphens only
    - Line 1: Required field 'description' is missing or empty

  suggestions:
    - Line 60: Command is 60 lines. Best practice: keep commands under
      ~55 lines (50±10%) - delegate to specialist agents instead of
      implementing logic directly.
```

## See Also

- [Common Tasks](../common-tasks.md)
- [Lint Rules](../rules/README.md)
- [Schema Reference](schemas.md)
- [Scoring Reference](../scoring/README.md)
- [Configuration Guide](../guides/configuration.md)
