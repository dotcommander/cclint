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

Inspect commands and the installed version:

```bash
cclint --help
cclint lint --help
cclint fmt --help
cclint summary --help
cclint --version
```

## Verification

- JSON report file exists after CI command
- `--staged` returns only files in the index
- score output includes grade bands (`A` to `F`)

## Related docs

- Setup path: `docs/setup.md`
- Command reference: `docs/reference/commands/commands.md`
- Rule reference: `docs/rules/README.md`

Baseline snapshots created now use version 1.1 fingerprints. Version 1.0 snapshots
still match legacy fingerprints for the same file and diagnostic source; loading a
snapshot does not rewrite it. Legacy numeric over-suppression remains until you
refresh with `cclint --baseline-create`. Historical CUE entries with blank filenames
cannot safely match a real file. Refresh those entries explicitly by selecting the
affected files and using a dedicated `--baseline-path`, then review the snapshot.
Explicit file, directory, and Git selections snapshot or filter only their selected
results. Multi-type runs produce one report and one baseline operation.
