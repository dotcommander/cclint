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

# Verify YAML syntax
cat ./my-file.md | yq eval '.'

# Reference the schema documentation
# See: docs/reference/schemas.md
```

**Common Missing Fields**:

| Component | Required Fields |
|-----------|-----------------|
| Agent | `name`, `description`, `model` |
| Command | `name`, `description` |
| Skill | `title`, `tooltip` |
| Plugin | `name`, `version`, `author.name` |

## Issue: Type Detection Failure

**Symptom**: `could not detect component type` warning

**Cause**:
- File doesn't match known naming patterns
- Ambiguous filename (e.g., `command.md` could be any type)

**Solution**:
```bash
# Explicitly specify type
cclint --type agent ./custom-name.md

# Available types: agent, command, skill, plugin
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

# Verify schemas are embedded
cclint --verbose agents
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

# Fix permissions if needed
chmod 644 ./agents/*.md
```

## Issue: Orphan Skill Warning

**Symptom**: `info: orphan skill detected` message

**Cause**:
- Skill defined but never referenced by any component

**Solution**:
```bash
# This is informational - not an error
# Either:
# 1. Reference the skill in a component: Skill(skill-name)
# 2. Remove the unused skill file
# 3. Suppress with: cclint --quiet skills
```

## Issue: Scoring Inconsistency

**Symptom**: Quality scores differ between runs

**Cause**:
- Code changes between runs
- Different baseline filtering active
- Cached scoring data

**Solution**:
```bash
# Run without baseline for true scores
cclint --scores agents

# Verify baseline state
cclint --baseline agents
```

## Issue: Git Integration Not Working

**Symptom**: `--staged` or `--diff` flags show no files

**Cause**:
- Not in a git repository
- No staged or modified files

**Solution**:
```bash
# Verify git status
git status

# Stage files first
git add ./agents/my-agent.md
cclint --staged agents

# Or lint all uncommitted changes
cclint --diff agents
```

## Issue: Config File Not Loaded

**Symptom**: Custom settings in `.cclintrc` ignored

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

# Verify syntax
cat .cclintrc.json | jq .

# Use environment variables as alternative
export CCLINT_FORMAT=json
cclint agents
```

## Still Need Help?

- Check [Schema Reference](../reference/schemas.md) for frontmatter requirements
- Review [Configuration Guide](./configuration.md) for advanced options
- Review [Git Hooks Guide](./git-hooks.md) for CI/CD integration
- File an issue at: https://github.com/dotcommander/cclint/issues
