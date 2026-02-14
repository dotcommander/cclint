# cclint

cclint validates Claude Code components before they break in runtime.
It lints agents, commands, skills, plugins, and settings.

## Why install

- Catch frontmatter and schema errors before commit
- Enforce naming, structure, and security rules consistently
- Track quality with component scoring (0-100)
- Support gradual rollout with baseline mode
- Generate machine-readable reports for CI

## Quick start

1. Install:

```bash
go install github.com/dotcommander/cclint@latest
```

2. Lint your project:

```bash
cclint
```

3. Lint staged files in a commit workflow:

```bash
cclint --staged
```

## First successful run

```bash
cclint --scores
```

If your files are valid, you should see passing checks and quality scores.

## Uninstall / rollback

```bash
rm "$(go env GOPATH)/bin/cclint"
```

## Next docs step

Go to `docs/README.md` for setup, integration, and contributor guides.
