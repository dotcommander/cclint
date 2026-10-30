# Configuration Guide

cclint can be configured through configuration files or environment variables. This guide covers all supported configuration options.

## Configuration File Formats

cclint reads the first valid configuration file in this precedence order at the lookup directory:

1. `.cclintrc.json` - JSON format
2. `.cclintrc.yaml` - YAML format (recommended)
3. `.cclintrc.yml` - YAML format

With `--root`, cclint looks in that directory only. Otherwise it starts in the working directory and walks up to the nearest ancestor with a readable config. An unreadable or invalid file is skipped during lookup; correct syntax and verify the loaded settings if a file appears ignored.

## Example Configuration

### YAML Format (Recommended)

```yaml
# .cclintrc.yaml
root: /path/to/project
exclude:
  - "**/vendor/**"
  - "**/node_modules/**"
followSymlinks: false

# Output settings
format: console
output: ""
failOn: error

# Display options
quiet: false
verbose: false
showScores: false
showImprovements: false

# Processing options
concurrency: 10
parallel: true

# Rule settings
rules:
  strict: true

# Schema validation
schemas:
  enabled: true
  extensions: {}
```

### JSON Format

```json
{
  "root": "/path/to/project",
  "exclude": ["**/vendor/**", "**/node_modules/**"],
  "followSymlinks": false,
  "format": "console",
  "failOn": "error",
  "quiet": false,
  "verbose": false,
  "showScores": false,
  "showImprovements": false,
  "concurrency": 10,
  "parallel": true,
  "rules": {
    "strict": true
  },
  "schemas": {
    "enabled": true,
    "extensions": {}
  }
}
```

## Environment Variables

The loader binds these `CCLINT_` variables. For compound keys it accepts both the historical spelling shown below and a snake-case alias (for example, `CCLINT_FAIL_ON`). Shells do not expand `~` in quoted YAML/JSON values; use an absolute path for `root`.

| Configuration | Environment Variable | Example |
|---------------|---------------------|---------|
| `root` | `CCLINT_ROOT` | `export CCLINT_ROOT=/custom/path` |
| `exclude` | `CCLINT_EXCLUDE` | `export CCLINT_EXCLUDE="**/vendor/**,**/test/**"` (comma-separated patterns) |
| `followSymlinks` | `CCLINT_FOLLOWSYMLINKS` | `export CCLINT_FOLLOWSYMLINKS=true` |
| `format` | `CCLINT_FORMAT` | `export CCLINT_FORMAT=json` |
| `output` | `CCLINT_OUTPUT` | `export CCLINT_OUTPUT=report.json` |
| `failOn` | `CCLINT_FAILON` | `export CCLINT_FAILON=warning` |
| `quiet` | `CCLINT_QUIET` | `export CCLINT_QUIET=true` |
| `verbose` | `CCLINT_VERBOSE` | `export CCLINT_VERBOSE=true` |
| `showScores` | `CCLINT_SHOWSCORES` | `export CCLINT_SHOWSCORES=true` |
| `showImprovements` | `CCLINT_SHOWIMPROVEMENTS` | `export CCLINT_SHOWIMPROVEMENTS=true` |
| `no-cycle-check` | `CCLINT_NO_CYCLE_CHECK` | `export CCLINT_NO_CYCLE_CHECK=true` |
| `concurrency` | `CCLINT_CONCURRENCY` | `export CCLINT_CONCURRENCY=20` |
| `parallel` | `CCLINT_PARALLEL` | `export CCLINT_PARALLEL=false` |
| `rules.strict` | `CCLINT_RULES_STRICT` | `export CCLINT_RULES_STRICT=false` |
| `schemas.enabled` | `CCLINT_SCHEMAS_ENABLED` | `export CCLINT_SCHEMAS_ENABLED=false` |

### Priority Order

Configuration values are applied in the following order (later sources override earlier ones):

1. Default values
2. Configuration file (`.cclintrc.*`)
3. Environment variables (`CCLINT_*`)
4. Command-line flags (highest priority)

## Configuration Options

### `root`

**Type:** `string`
**Default:** nearest ancestor of the working directory containing `.claude/` or `.claude-plugin/plugin.json`, falling back to `~/.claude`.

The root directory containing Claude Code components to lint.

### `exclude`

**Type:** `array of strings`
**Default:** `[]`

Glob patterns for files/directories to exclude from linting. Supports doublestar patterns (`**`).

### `followSymlinks`

**Type:** `boolean`
**Default:** `false`

Whether to follow symbolic links during file discovery.

### `format`

**Type:** `string`
**Default:** `console`
**Valid values:** `console`, `json`, `markdown`

Output format for lint results.

### `output`

**Type:** `string`
**Default:** `""` (stdout)

File path to write output. If omitted, all formats write to stdout.

### `failOn`

**Type:** `string`
**Default:** `error`
**Valid values:** `error`, `warning`, `suggestion`

Minimum severity level that causes a non-zero exit code.

### `quiet`

**Type:** `boolean`
**Default:** `false`

Suppress all report output; communicate the result through exit status only.

### `verbose`

**Type:** `boolean`
**Default:** `false`

Enable detailed processing information.

### `showScores`

**Type:** `boolean`
**Default:** `false`

Display quality scores (0-100) for each component.

### `showImprovements`

**Type:** `boolean`
**Default:** `false`

Display improvement recommendations.

### `no-cycle-check`

**Type:** `boolean`
**Default:** `false`

Disable cyclic dependency detection.

### `concurrency`

**Type:** `integer`
**Default:** `10`
**Minimum:** `1`

Number of parallel workers for file processing.

### `parallel`

**Type:** `boolean`
**Default:** `true`

Enable parallel processing of files.

### `rules.strict`

**Type:** `boolean`
**Default:** `true`

Enable strict rule enforcement.

### `schemas.enabled`

**Type:** `boolean`
**Default:** `true`

Enable CUE schema validation for component frontmatter.

### `schemas.extensions`

**Type:** `object`
**Default:** `{}`

Stored in the configuration object; this guide does not imply that custom CUE schemas are loaded automatically.
