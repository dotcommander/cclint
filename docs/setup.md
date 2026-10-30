# Setup

## Purpose

Set up cclint and run your first reliable lint loop.

## Prerequisites

- Go 1.26+
- A Claude Code project containing the components you want to check

## Main workflow

1. Install cclint:

```bash
go install github.com/dotcommander/cclint@latest
```

2. Run the default lint pass from your project root:

```bash
cclint
```

Context files are not part of the default pass; use `cclint context` to check
them. If project discovery does not select the intended root, pass
`--root /path/to/project`.

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

- [Common workflows](common-tasks.md)
- [Configuration](guides/configuration.md)
- [Git hooks](guides/git-hooks.md)
- [Troubleshooting](guides/troubleshooting.md)
