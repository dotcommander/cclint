# cclint

A linter for Claude Code components. Catches schema errors, enforces structure, and flags cross-file problems before they surface as confusing runtime behavior.

Handles agents, commands, skills, plugins, and settings.

## Install

```bash
go install github.com/dotcommander/cclint@latest
```

## Usage

```bash
cclint                    # lint everything under ~/.claude
cclint agents             # one component type
cclint ./path/to/file.md  # lint specific files
cclint --staged           # only staged files (pre-commit)
cclint --scores           # quality scores (0-100)
cclint fmt --write        # auto-format component files
```

## What it catches

- Frontmatter schema errors (embedded CUE schemas)
- Size, structure, and naming violations per component type
- Cross-file issues: broken references, orphaned skills, ghost triggers
- Settings validation: hooks, permissions, MCP servers, security
- Delegation anti-patterns in commands and agents

## Baseline mode

For projects with existing violations, adopt incrementally:

```bash
cclint --baseline-create   # snapshot current state
cclint --baseline          # only fail on new issues
```

Commit `.cclintbaseline.json` and tighten over time.

## Quality scoring

Every component gets a 0-100 score across structural completeness, practices, composition, and documentation. Tier grades A-F. Useful for prioritizing what to fix first.

## More

See `docs/README.md` for setup, CI integration, and contributor guides.
