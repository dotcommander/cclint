# cclint

A linter for Claude Code components: agents, commands, skills, plugins, and settings.

Validates YAML frontmatter, enforces structural patterns, detects security issues, and provides quality scoring with improvement recommendations.

## Installation

```bash
go install github.com/dotcommander/cclint@latest
```

Ensure `~/go/bin` is in your `PATH`.

## Usage

```bash
cclint                    # lint all components
cclint agents             # lint specific type
cclint --staged           # lint only git-staged files
cclint --scores           # show quality scores (0-100)
cclint --format json      # JSON output for CI
```

### Git Integration

```bash
cclint --staged           # pre-commit: staged files only
cclint --diff             # all uncommitted changes
```

### Baseline Mode

Adopt cclint incrementally in legacy projects:

```bash
cclint --baseline-create  # snapshot current issues
cclint --baseline         # only new issues fail
```

### Formatting

```bash
cclint fmt file.md        # preview formatted output
cclint fmt -w file.md     # format in place
cclint fmt --check        # CI: exit 1 if unformatted
```

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | No errors |
| 1 | Errors found |

## Documentation

| Document | Description |
|----------|-------------|
| [Rules Reference](docs/RULES.md) | All 124 lint rules with severity and sources |
| [Quality Scoring](docs/scoring/README.md) | Scoring methodology and tier grades |
| [Cross-File Validation](docs/cross-file-validation.md) | Skill reference and dependency checking |
| [Anthropic Requirements](docs/ANTHROPIC_REQUIREMENTS.md) | Official vs opinionated rules |

### Component-Specific Rules

- [Agent Rules](docs/rules/agents.md)
- [Command Rules](docs/rules/commands.md)
- [Skill Rules](docs/rules/skills.md)
- [Settings Rules](docs/rules/settings.md)
- [Plugin Rules](docs/rules/plugins.md)
- [Security Rules](docs/rules/security.md)

## CI/CD

### GitHub Actions

```yaml
- uses: actions/setup-go@v5
  with:
    go-version: '1.23'
- run: go install github.com/dotcommander/cclint@latest
- run: cclint --format json --output report.json
```

### Pre-commit Hook

```bash
#!/bin/bash
cclint --staged || exit 1
```

## Configuration

Create `.cclintrc.json` in project root:

```json
{
  "root": ".",
  "format": "console",
  "verbose": false,
  "showScores": false
}
```

Environment variables: prefix with `CCLINT_` (e.g., `CCLINT_VERBOSE=true`).

## License

MIT
