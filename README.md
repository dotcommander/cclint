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
cclint --staged           # only staged files (pre-commit)
cclint --scores           # include quality scores (0-100)
cclint --improvements     # show what would move the needle
```

## What it catches

- Frontmatter schema errors (via embedded CUE schemas)
- Naming, size, and structure violations per component type
- Cross-file issues: broken skill references, orphaned skills, ghost triggers
- Model field mutations and delegation anti-patterns

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
