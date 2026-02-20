# Setup

## Purpose

Set up cclint and run your first reliable lint loop.

## Prerequisites

- Go 1.25+
- A Claude Code project with component files (`agents/`, `commands/`, `skills/`, `.claude-plugin/plugin.json`)

## Main workflow

1. Install cclint:

```bash
go install github.com/dotcommander/cclint@latest
```

2. Run a full lint pass:

```bash
cclint
```

3. Run a targeted pass while you work:

```bash
cclint --staged
```

4. Enable baseline mode for gradual adoption:

```bash
cclint --baseline-create
cclint --baseline
```

## Verification

- `cclint --version` prints a version
- `cclint` exits `0` when no errors are found
- `cclint --baseline` reports ignored baseline issues instead of failing on legacy debt

## Related docs

- Common workflows: `docs/common-tasks.md`
- Config options: `docs/guides/configuration.md`
- Git hooks: `docs/guides/git-hooks.md`
- Troubleshooting: `docs/guides/troubleshooting.md`
