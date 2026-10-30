# cclint Lint Rules Reference

This directory is a topical reference for cclint lint findings. Rule numbers are documentation labels, not IDs emitted by the CLI; some ranges overlap and newer checks may not have a numbered entry. For the current validation behavior, consult `internal/lint/`, `internal/cue/schemas/`, and `internal/textutil/`.

## Rule Categories

| File | Rules | Component | Description |
|------|-------|-----------|-------------|
| [agents.md](agents.md) | 001-021 | Agent | Agent frontmatter and structure validation |
| [commands.md](commands.md) | 022-034 | Command | Command frontmatter and delegation patterns |
| [skills.md](skills.md) | 035-061 (gaps) | Skill | Skill structure and best practices |
| [settings.md](settings.md) | 048-074 (overlaps skills) | Settings | Hook configuration and security |
| [plugins.md](plugins.md) | 075-092 | Plugin | Plugin manifest validation |
| [security.md](security.md) | 093-104 | All | Secrets detection and tool validation |
| [schema-constraints.md](schema-constraints.md) | 105-124 | All | CUE schema constraints |

## Severity Levels

| Severity | Description | Exit Code |
|----------|-------------|-----------|
| **error** | Lint finding that fails by default | 1 |
| **warning** | Advisory by default; can fail with `--fail-on warning` | 0 by default |
| **suggestion** | Advisory by default; can fail with `--fail-on suggestion` | 0 by default |

## Source Attribution

Each rule includes a `Source:` field for tracking where the rule originated:
- **SourceAnthropicDocs** - From official Anthropic documentation
- **SourceCClintObserve** - From cclint observations (emitted as `cclint-observation`)

## Quick Reference

See [../scoring/README.md](../scoring/README.md) for quality scoring metrics (separate from lint rules).
