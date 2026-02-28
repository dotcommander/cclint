# Common Tasks

## Purpose

Copy/paste commands for daily cclint workflows.

## Prerequisites

- cclint installed and available in `PATH`
- Run from your project root

## Main workflow

Run all components:

```bash
cclint
```

Run one component type:

```bash
cclint agents
cclint commands
cclint skills
cclint settings
cclint rules
cclint context
cclint plugins
cclint output-styles
cclint summary
cclint fmt
```

Run a single file:

```bash
cclint path/to/agent.md
cclint a.md b.md c.md
cclint --type agent ./custom/file.md
```

Run on changed files:

```bash
cclint --staged
cclint --diff
```

Generate CI output:

```bash
cclint --format json --output cclint-report.json
```

Check quality scoring:

```bash
cclint --scores
cclint --improvements
```

## Verification

- JSON report file exists after CI command
- `--staged` returns only files in the index
- score output includes grade bands (`A` to `F`)

## Related docs

- Setup path: `docs/setup.md`
- Command reference: `docs/reference/commands/commands.md`
- Rule reference: `docs/rules/README.md`
