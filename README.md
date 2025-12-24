# cclint

> A linter for Claude Code components — agents, commands, skills, and settings.

cclint validates YAML frontmatter, checks structural patterns, and suggests improvements based on established conventions. It's designed to catch errors early and help maintain consistency across Claude Code projects.

## Features

- **Multi-format validation**: Checks agents, commands, skills, settings, and CLAUDE.md files
- **YAML frontmatter parsing**: Validates required fields and data types
- **CUE schema validation**: Embedded schemas for type-safe validation
- **Quality scoring**: 0-100 scores with tier grading (A-F)
- **Cross-file validation**: Detects missing skill references
- **Multiple output formats**: Console, JSON, Markdown
- **CI/CD ready**: Exit codes and JSON output for automation

## Installation

### Via Go install (recommended)

```bash
go install github.com/dotcommander/cclint@latest
```

This installs `cclint` to `~/go/bin/cclint` (or `$GOPATH/bin`). Ensure `~/go/bin` is in your `PATH`:

```bash
export PATH=$PATH:~/go/bin
```

### From source

```bash
git clone https://github.com/dotcommander/cclint.git && cd cclint
go build -o cclint .
ln -sf $(pwd)/cclint ~/go/bin/cclint
```

## Quick Start

```bash
# Lint everything in current directory
cclint

# Lint specific component types
cclint agents
cclint commands
cclint skills

# Verbose mode (shows suggestions)
cclint -v agents
```

## Usage

### Basic Commands

```bash
cclint                           # Lint all component types
cclint agents                    # Lint only agent files
cclint commands                  # Lint only command files
cclint skills                    # Lint only skill files
cclint settings                  # Lint only settings files
cclint context                   # Lint CLAUDE.md files
```

### Common Options

```bash
# Project root (auto-detected by default)
cclint --root ~/my-project agents

# Verbose output (shows suggestions)
cclint -v
cclint --verbose

# Quiet mode (errors only, no suggestions)
cclint -q
cclint --quiet

# Show quality scores (0-100 with tier grades)
cclint --scores agents

# Show improvement recommendations with point values
cclint --improvements agents

# Output formats
cclint --format json --output report.json
cclint --format markdown --output report.md
```

### Exit Codes

| Code | Meaning |
|------|---------|
| 0 | No errors (suggestions don't affect exit code) |
| 1 | Errors found |

## What It Checks

### Agents

| Type | Check |
|------|-------|
| **Error** | Missing `name` field |
| **Error** | Missing `description` field |
| **Error** | Invalid name format (must be lowercase letters, numbers, hyphens only) |
| **Error** | Name doesn't match filename |
| **Error** | Invalid `color` value |
| **Error** | Missing skill referenced in `Skill()` calls |
| **Suggestion** | Line count > 200 (move methodology to skills) |
| **Suggestion** | Missing `model` field |
| **Suggestion** | Contains "Quick Reference" section (belongs in skill) |
| **Suggestion** | Contains "When to Use" section (use description triggers) |
| **Suggestion** | Contains "What it does" section (belongs in description) |
| **Suggestion** | Contains "Usage" section (belongs in skill) |
| **Suggestion** | Inline scoring formula detected (move to skill) |
| **Suggestion** | Inline priority matrix detected (move to skill) |
| **Suggestion** | Description lacks "Use PROACTIVELY when..." pattern |
| **Suggestion** | No skill reference found with Foundation/Workflow sections |

### Commands

| Type | Check |
|------|-------|
| **Error** | Missing `allowed-tools` field |
| **Error** | Missing `description` field |
| **Suggestion** | Line count > 50 (delegate to specialist agent) |
| **Suggestion** | Contains implementation steps (should delegate) |
| **Suggestion** | Uses `Task()` but lacks `allowed-tools` permission |
| **Suggestion** | Contains "Quick Reference" (belongs in skill) |
| **Suggestion** | Contains "Usage/Workflow/When to use" (bloat in thin commands) |
| **Suggestion** | Has "What it does" section (belongs in description) |
| **Suggestion** | Redundant title header (filename identifies command) |
| **Suggestion** | Success criteria should use checkbox format `- [ ]` |
| **Suggestion** | More than 2 code examples (max 2 recommended) |

### Skills

| Type | Check |
|------|-------|
| **Error** | Filename must be `SKILL.md` (not skill name) |
| **Suggestion** | Line count > 500 (move heavy docs to `references/`) |
| **Suggestion** | Missing "Anti-Patterns" section |
| **Suggestion** | Missing "Examples" section |

## Output Examples

### Console Output (default)

```bash
$ cclint agents

✗ agents/broken-agent.md
    ✗ Required field 'name' is missing or empty
    ✗ Required field 'description' is missing or empty

✓ All other agents passed

2/3 passed, 2 errors (45ms)
```

### Verbose Mode (with suggestions)

```bash
$ cclint -v agents

💡 agents/fat-agent.md
    💡 Agent is 251 lines. Best practice: keep agents under 200 lines.
    💡 Missing 'model' specification
    💡 Agent has '## Quick Reference' - belongs in skill, not agent
    💡 Description lacks 'Use PROACTIVELY when...' pattern

✓ agents/lean-agent.md [B 72]

2/2 passed, 0 errors, 4 suggestions (52ms)
```

### With Quality Scores

```bash
$ cclint --scores agents

✓ agents/lean-agent.md [A 92]
✓ agents/decent-agent.md [B 78]
💡 agents/minimal-agent.md [C 55]

3/3 passed, 0 errors, 3 suggestions
```

### JSON Output

```bash
$ cclint --format json --output report.json
```

```json
{
  "header": {
    "tool": "cclint",
    "version": "1.0.0",
    "timestamp": "2025-12-23T20:30:00Z"
  },
  "summary": {
    "total_files": 15,
    "successful_files": 14,
    "failed_files": 1,
    "total_errors": 2,
    "total_warnings": 0,
    "duration": "125ms"
  },
  "results": [
    {
      "file": "agents/broken.md",
      "type": "agent",
      "success": false,
      "errors": [
        {
          "file": "agents/broken.md",
          "message": "Required field 'name' is missing or empty",
          "severity": "error",
          "line": 1
        }
      ]
    }
  ]
}
```

## Configuration

Create a `.cclintrc.json` (or `.yaml`/`.yml`) in your project root:

```json
{
  "root": ".",
  "format": "console",
  "output": "",
  "failOn": "error",
  "quiet": false,
  "verbose": false,
  "showScores": false,
  "showImprovements": false,
  "rules": {
    "strict": true
  },
  "schemas": {
    "enabled": true
  }
}
```

### Environment Variables

Prefix with `CCLINT_`:

```bash
export CCLINT_ROOT=~/my-project
export CCLINT_VERBOSE=true
export CCLINT_FORMAT=json
export CCLINT_OUTPUT=lint-report.json
```

### Configuration Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `root` | string | `~/.claude` | Project root directory |
| `exclude` | []string | - | Paths to exclude from linting |
| `followSymlinks` | bool | `false` | Follow symbolic links |
| `format` | string | `console` | Output format: `console`, `json`, `markdown` |
| `output` | string | - | Output file path (required for non-console formats) |
| `failOn` | string | `error` | Fail level: `error`, `warning`, `suggestion` |
| `quiet` | bool | `false` | Suppress non-error output |
| `verbose` | bool | `false` | Show suggestions |
| `showScores` | bool | `false` | Show quality scores |
| `showImprovements` | bool | `false` | Show improvement recommendations |
| `rules.strict` | bool | `true` | Enable strict validation |
| `schemas.enabled` | bool | `true` | Enable CUE schema validation |

## CI/CD Integration

### GitHub Actions

```yaml
name: Lint Claude Code Components

on: [push, pull_request]

jobs:
  cclint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Install cclint
        run: |
          go install github.com/dotcommander/cclint@latest

      - name: Run cclint
        run: |
          cclint --format json --output cclint-report.json

      - name: Upload results
        if: always()
        uses: actions/upload-artifact@v4
        with:
          name: cclint-report
          path: cclint-report.json
```

### GitLab CI

```yaml
lint:
  stage: test
  script:
    - go install github.com/dotcommander/cclint@latest
    - cclint --format json --output cclint-report.json
  artifacts:
    when: always
    paths:
      - cclint-report.json
    reports:
      lint: cclint-report.json
```

### Pre-commit Hook

Create `.git/hooks/pre-commit`:

```bash
#!/bin/bash
# Run cclint on commit
cclint || exit 1
```

Or with [husky](https://github.com/typicode/husky):

```bash
npm install --save-dev husky
npx husky set .husky/pre-commit "cclint"
```

## File Discovery

cclint searches the following locations:

```
.claude/
├── agents/**/*.md       # Agent definitions
├── commands/**/*.md     # Command definitions
├── skills/**/SKILL.md   # Skill definitions (must be named SKILL.md)
└── settings.yaml       # Project settings

agents/**/*.md           # Alternative agents location
commands/**/*.md         # Alternative commands location
skills/**/SKILL.md       # Alternative skills location
```

## Quality Scoring

Components are scored 0-100 across four categories:

| Category | Points | Description |
|----------|--------|-------------|
| Structural | 0-40 | Required fields and sections |
| Practices | 0-40 | Best practices adherence |
| Composition | 0-10 | Size and complexity |
| Documentation | 0-10 | Documentation quality |

### Tiers

| Tier | Score Range |
|------|-------------|
| A | 85-100 |
| B | 70-84 |
| C | 50-69 |
| D | 30-49 |
| F | 0-29 |

## Customization

cclint is opinionated. Fork it to match your preferences.

### Modify Line Limits

Edit `internal/cli/*.go` and search for `lines >`:

```go
// agents.go line ~260
if lines > 200 {  // Change this threshold
    suggestions = append(suggestions, ...)
}
```

### Add/Remove Checks

Edit validation functions in `internal/cli/*.go`:

```go
// Add a new check in agents.go
func validateAgentBestPractices(...) []cue.ValidationError {
    // Your custom check
    if someCondition {
        suggestions = append(suggestions, cue.ValidationError{
            File:     filePath,
            Message:  "Your custom message",
            Severity: "suggestion",
        })
    }
}
```

### Modify Schemas

Edit CUE schemas in `internal/cue/schemas/*.cue`:

```cue
#Agent: {
    name: string
    description: string
    // Add your custom fields
    customField?: string
}
```

### Rebuild

```bash
go build -o cclint .
```

## Examples

### Valid Agent

```markdown
---
name: "example-agent"
description: "Use PROACTIVELY when you need to process data. Handles ETL workflows."
model: sonnet
color: blue
tools: [Read, Write, Edit]
---

# Foundation

Skill: data-processing

## Workflow

### Phase 1: Read
### Phase 2: Transform
### Phase 3: Write

## Expected Output

Processed data file with transformations applied.

## Success Criteria

- [ ] Input file read successfully
- [ ] All transformations applied
- [ ] Output file written
```

### Valid Command (Thin)

```markdown
---
allowed-tools: Task
description: "Run comprehensive data analysis on your codebase"
argument-hint: "<directory>"
---

Task(data-specialist): Analyze the directory at $ARGUMENTS
```

### Valid Skill

```markdown
---
name: "data-processing"
description: "Data transformation patterns for ETL workflows"
---

## Quick Reference

| User Question | Action |
|---------------|--------|
| "How do I transform CSV data?" | Read(csv/transform) |
| "Need to merge datasets?" | Read(csv/merge) |

## Workflow

### Phase 1: Extraction
### Phase 2: Transformation
### Phase 3: Loading

## Anti-Patterns

| Anti-Pattern | Problem | Fix |
|--------------|---------|-----|
| Process entire file in memory | OOM on large files | Stream line by line |
| Hardcode file paths | Not portable | Use input arguments |

## Examples

### Transform CSV
```bash
cat input.csv | transform --field=email --action=lowercase
```
```

## Troubleshooting

### "No files found to validate"

Ensure your files are in the expected locations:

```bash
# Check discovery
cclint -v --root /path/to/project agents
```

### False positives on suggestions

Use `-q`/``--quiet` to suppress suggestions:

```bash
cclint -q agents  # Errors only
```

Or customize the rules by forking and editing the validation functions.

### Missing skill references

cclint validates that skills referenced in `Skill()` calls actually exist. Ensure your skill files are named `SKILL.md`:

```
skills/
└── data-processing/
    └── SKILL.md  # Correct filename
```

## License

MIT

## Contributing

Contributions welcome! Please:

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## Changelog

See [CHANGELOG.md](CHANGELOG.md) for version history.
