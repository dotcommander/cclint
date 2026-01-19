# cclint plugins

Lint and validate plugin manifest files.

---

## Usage

```bash
cclint plugins [flags]
```

---

## Description

The `plugins` subcommand scans for and validates plugin manifest files in your
Claude Code project. Plugins extend Claude Code with custom functionality and
are defined via JSON manifests.

### Supported File Patterns

- `**/.claude-plugin/plugin.json`

### What Gets Validated

- Required fields: `name`, `description`, `author.name`
- Version format: semver (major.minor.patch)
- Name format: lowercase alphanumeric with hyphens, max 64 characters
- Description quality: 50+ characters recommended
- Reserved words: `anthropic`, `claude` are prohibited
- File size limit: 5KB for optimal composition scoring

---

## File Size Limit

Plugin manifests should stay under **5KB** for optimal quality scoring.

### Scoring Tiers

| Size | Score | Grade |
|------|-------|-------|
| ≤1KB | 10/10 | Excellent |
| ≤2KB | 8/10 | Good |
| ≤5KB | 6/10 | OK |
| ≤10KB | 3/10 | Large (warning) |
| >10KB | 0/10 | Too large (error) |

### Rationale

Plugin manifests should be concise metadata files. Heavy documentation or
configuration should be externalized to separate files (README, config files,
etc.).

### Example Violation Message

```text
my-plugin/.claude-plugin/plugin.json:1
suggestion: Plugin manifest is 8KB. Best practice: keep plugin.json under
5KB for optimal loading - move documentation to README.md.
```

---

## Name Format Rules

Plugin names must follow these conventions:

- Lowercase alphanumeric characters with hyphens
- Maximum 64 characters
- Cannot use reserved words: `anthropic`, `claude`

### Example Valid Names

```json
{
  "name": "my-plugin",
  "name": "github-integration",
  "name": "dbtools-postgres"
}
```

### Example Violation Messages

```text
# Invalid: reserved word
plugin.json:3
error: Name 'claude-helper' is a reserved word and cannot be used

# Invalid: exceeds character limit
plugin.json:3
error: Name exceeds 64 character limit (72 chars)
```

---

## Required Fields

| Field | Type | Constraints | Description |
|-------|------|-------------|-------------|
| `name` | string | max 64 chars, lowercase alphanumeric with hyphens | Plugin identifier |
| `description` | string | max 1024 chars, 50+ recommended | What the plugin does |
| `author.name` | string | non-empty | Author or maintainer name |

---

## Recommended Fields

| Field | Type | Description |
|-------|------|-------------|
| `version` | string | Semver format (e.g., `1.0.0`) |
| `homepage` | string | Project URL |
| `repository` | string | Source code URL |
| `license` | string | SPDX identifier (e.g., `MIT`, `Apache-2.0`) |
| `keywords` | array | Discoverability tags |
| `readme` | string | Path to README file |

---

## Version Format

The `version` field should follow [semver](https://semver.org/) format:
`major.minor.patch`

### Examples

```json
{
  "version": "1.0.0",
  "version": "2.3.1",
  "version": "0.5.0-beta"
}
```

### Example Violation Message

```text
plugin.json:5
error: Invalid semver format 'v1.0' (expected major.minor.patch, e.g., 1.0.0)
```

---

## Example Manifest

```json
{
  "name": "example-plugin",
  "description": "A comprehensive example plugin that demonstrates common patterns for Claude Code extensions, including tool definitions and configuration management.",
  "version": "1.0.0",
  "author": {
    "name": "Your Name"
  },
  "homepage": "https://github.com/username/example-plugin",
  "repository": "https://github.com/username/example-plugin.git",
  "license": "MIT",
  "keywords": ["example", "demo", "tutorial"]
}
```

---

## Example Output

### Passing Plugin

```text
✓ my-plugin/.claude-plugin/plugin.json (2.3KB)
  Score: 85/100 (B)
```

### Failing Plugin

```text
✗ my-plugin/.claude-plugin/plugin.json
  errors:
    - Line 3: Required field 'name' is missing or empty
    - Line 8: Required field 'author.name' is missing or empty

  suggestions:
    - Line 3: Name 'claude-helper' is a reserved word and cannot be used
    - Line 5: Description is only 25 chars - consider expanding for clarity
    - Line 1: Consider adding 'version' field in semver format (e.g., 1.0.0)
    - Line 1: Consider adding 'homepage' field with project URL
    - Line 1: Consider adding 'repository' field with source code URL
    - Line 1: Consider adding 'license' field (e.g., MIT, Apache-2.0)
    - Line 1: Consider adding 'keywords' array for discoverability
```

---

## See Also

- [Agents Reference](agents.md)
- [Commands Reference](commands.md)
- [Skills Reference](skills.md)
- [CLI Reference](../cli.md)
- [Global Flags](../flags.md)
- [Quality Scoring](../scoring.md)
