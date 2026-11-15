# cclint

A linter for Claude Code components. Catches schema errors, enforces structure, and flags cross-file problems before they surface as confusing runtime behavior.

It checks agents, commands, skills, plugins, settings, rules, and output styles.
Context files can be checked explicitly with `cclint context`.

## Install

Requires Go 1.26+.

```bash
go install github.com/dotcommander/cclint@latest
```

`cclint --version` also prints the Claude Code release whose lint surface cclint models.

## Usage

```bash
cclint                          # lint the enclosing project (~/.claude if none found)
cclint agents commands          # one or more component types
cclint ./agents/reviewer.md     # specific files or directories
cclint --root ~/work/proj       # lint an explicit tree
cclint --staged                 # only staged files (pre-commit)
cclint fmt --write              # auto-format component files
```

With no `--root`, cclint walks up from the working directory to the nearest
`.claude/` directory or `.claude-plugin/plugin.json` and lints that project;
outside any project it lints `~/.claude`.

Discovered per root:

- Agent, command, and skill Markdown under `.claude/` and at the project root
- `.claude/settings.json` and `.claude/settings.local.json`
- `.claude-plugin/plugin.json` manifests
- Output styles under `.claude/output-styles/` or `output-styles/`

Context files (`CLAUDE.md`, `CLAUDE.local.md`, `AGENTS.md`, `.claude/CLAUDE.md`,
`.claude/AGENTS.md`) are linted explicitly with `cclint context`.

## Gating a run

cclint exits non-zero when findings meet the `--fail-on` threshold — `error`
by default, tightened to `warning` or `suggestion`:

```bash
cclint --quiet --fail-on error          # exit status only
cclint --format json -o report.json     # machine-readable report
```

For pre-commit and CI recipes (for example `cclint --staged --quiet` in
`.git/hooks/pre-commit`), see [connect an assistant](docs/connect-assistant.md).

## What it catches

- Frontmatter schema errors (embedded CUE schemas)
- Size, structure, and naming violations per component type
- Cross-file issues: broken references, orphaned skills, ghost triggers
- Settings validation: hooks, permissions, MCP servers, security
- Delegation anti-patterns in commands and agents

## Fixing what it finds

- `cclint fmt --write` — canonicalize component files
- `cclint --improvements` — suggested fixes alongside findings
- `cclint --scores` — 0-100 quality scores to prioritize what to fix first

For projects with existing violations, adopt incrementally:

```bash
cclint --baseline-create   # snapshot current state
cclint --baseline          # only fail on new issues
```

Commit `.cclintbaseline.json` and tighten over time.

## Quality scoring

Agents, commands, skills, plugins, and output styles get a 0-100 score across structural completeness, practices, composition, and documentation. Tier grades A-F. Useful for prioritizing what to fix first.

## Documentation and maintenance

- [Documentation index](docs/README.md): setup, workflows, and reference material.
- [Maintainer guide](docs/change-cclint.md): development and keeping cclint aligned with Claude Code.
- [Updater procedure](skills/updater/SKILL.md): an assistant-guided workflow for reviewing upstream changes; not an autonomous cclint command.
