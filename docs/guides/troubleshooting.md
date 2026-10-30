# Troubleshooting Guide

Common issues and solutions when using cclint.

## Issue: File Not Found

**Symptom**: `error: no such file or directory` when running cclint

**Cause**:
- Incorrect path to component file
- Component not in expected directory structure
- Missing `--root` flag for non-standard project layouts

**Solution**:
```bash
# Verify the file exists
ls -la ./agents/my-agent.md

# Use full path if needed
cclint /absolute/path/to/agents/my-agent.md

# Specify project root for non-standard layouts
cclint --root /path/to/project agents
```

## Issue: Schema Validation Failed

**Symptom**: `schema validation failed` error for component frontmatter

**Cause**:
- Missing required frontmatter fields
- Invalid data types (e.g., string instead of array)
- Malformed YAML syntax

**Solution**:
```bash
# Check required fields for your component type
cclint --type agent ./my-file.md

# Inspect the YAML between the opening and closing --- lines.
# A Markdown file is not itself a YAML document; do not pipe the whole file to yq.

# Reference the schema documentation
# See: docs/reference/schemas.md
```

**Common Missing Fields**:

| Component | Required Fields |
|-----------|-----------------|
| Agent | `name`, `description` |
| Command | No required frontmatter fields (`name` is optional) |
| Skill | `name`, `description` |
| Plugin | `name`, `description`, `author.name` |

## Issue: Type Detection Failure

**Symptom**: `could not detect component type` warning

**Cause**:
- File doesn't match known naming patterns
- Ambiguous filename (e.g., `command.md` could be any type)

**Solution**:
```bash
# Explicitly specify type
cclint --type agent ./custom-name.md

# Available types: agent, command, skill, settings, context, plugin, rule, output-style
```

## Issue: Baseline Not Found

**Symptom**: `baseline file not found` error with `--baseline` flag

**Cause**:
- Baseline file doesn't exist
- Wrong baseline path specified

**Solution**:
```bash
# Create baseline first
cclint --baseline-create

# Specify custom path
cclint --baseline-path /custom/path/baseline.json --baseline agents
```

## Issue: CUE Import Errors

**Symptom**: `CUE validation failed` with import errors

**Cause**:
- CUE schemas not embedded in binary
- Corrupted build

**Solution**:
```bash
# Rebuild cclint to embed schemas
go build -o cclint .

# Run the freshly built binary (not an older cclint on PATH)
./cclint --verbose agents
```

## Issue: Permission Denied

**Symptom**: `permission denied` when accessing files

**Cause**:
- Insufficient file read permissions
- SELinux or similar security policies

**Solution**:
```bash
# Check file permissions
ls -la ./agents/

# If you own the files and intend to change their permissions, adjust only
# the affected file; otherwise ask the owner for read access.
```

## Issue: Orphan Skill Warning

**Symptom**: `info: orphan skill detected` message

**Cause**:
- Skill defined but never referenced by any component

**Solution**:
```bash
# This is informational - not an error
# Either:
# 1. Reference the skill from a component if it is intended for reuse.
# 2. If intentionally standalone, leave the informational suggestion as is.
# --quiet suppresses all report output, not just orphan suggestions.
```

## Issue: Scoring Inconsistency

**Symptom**: Quality scores differ between runs

**Cause**:
- Code changes between runs
- Different content, root, or scoring mode (`--scores` versus a plain lint run)
- Baseline filtering changes which findings are displayed, not the score calculation

**Solution**:
```bash
# Compare scores for the same files and root
cclint --root /path/to/project --scores agents

# Compare findings separately if baseline filtering is enabled
cclint --root /path/to/project --baseline agents
```

## Issue: Git Integration Not Working

**Symptom**: `--staged` or `--diff` flags show no files

**Cause**:
- Not in a git repository
- No relevant staged components (`--staged`) or relevant tracked/untracked changes (`--diff`)
- Selected files are outside the configured root or excluded; a component deletion may trigger full lint

**Solution**:
```bash
# Verify git status
git status

# Inspect the current selection without changing the index
git diff --cached --name-status
git status --short

# Git selection flags do not accept positional type filters
cclint --staged
cclint --diff
```

## Issue: Config File Not Loaded

**Symptom**: Custom settings in `.cclintrc.yaml`, `.cclintrc.yml`, or `.cclintrc.json` appear ignored

**Cause**:
- Config file in wrong directory
- Invalid YAML/JSON syntax
- Wrong filename

**Solution**:
```bash
# Valid filenames (in project root):
.cclintrc.json
.cclintrc.yaml
.cclintrc.yml

# Verify JSON syntax (if using the JSON variant)
jq . .cclintrc.json

# Use environment variables as alternative
export CCLINT_FORMAT=json
cclint agents
```

## Still Need Help?

- Check [Schema Reference](../reference/schemas.md) for frontmatter requirements
- Review [Configuration Guide](./configuration.md) for advanced options
- Review [Git Hooks Guide](./git-hooks.md) for CI/CD integration
- File an issue at: https://github.com/dotcommander/cclint/issues
