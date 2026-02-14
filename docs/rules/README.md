# cclint Lint Rules Reference

This directory contains documentation for all lint rules enforced by cclint.

## Rule Categories

| File | Rules | Component | Description |
|------|-------|-----------|-------------|
| [agents.md](agents.md) | 001-021 | Agent | Agent frontmatter and structure validation |
| [commands.md](commands.md) | 022-034 | Command | Command frontmatter and delegation patterns |
| [skills.md](skills.md) | 035-060 | Skill | Skill structure and best practices |
| [settings.md](settings.md) | 048-074 | Settings | Hook configuration and security |
| [plugins.md](plugins.md) | 075-092 | Plugin | Plugin manifest validation |
| [security.md](security.md) | 093-104 | All | Secrets detection and tool validation |
| [schema-constraints.md](schema-constraints.md) | 105-124 | All | CUE schema constraints |

## Severity Levels

| Severity | Description | Exit Code |
|----------|-------------|-----------|
| **error** | Must fix before deployment | 1 |
| **warning** | Should fix, may indicate problems | 0 |
| **suggestion** | Optional improvement | 0 |

## Source Attribution

Each rule includes a `Source:` field for tracking where the rule originated:
- **SourceAnthropicDocs** - From official Anthropic documentation
- **SourceCClintObserve** - From cclint best practice observations

## Quick Reference

See [../scoring/README.md](../scoring/README.md) for quality scoring metrics (separate from lint rules).
