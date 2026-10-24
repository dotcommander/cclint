# Common Tasks

## Purpose

Copy/paste commands for daily cclint workflows.

## Prerequisites

- cclint installed and available in `PATH`
- Run from your project root

## Main workflow

Run the default component set (context files are opt-in with `cclint context`):

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

Target a project from another directory:

```bash
cclint --root /path/to/project .claude/agents/helper.md
cclint --root /path/to/project fmt --check .claude/agents/
cclint --root /path/to/project fmt --write --file .claude/agents/helper.md
```

With an explicit root (CLI, configuration, or `CCLINT_ROOT`), relative lint
and formatter selections are relative to that root. Absolute paths are unchanged.
Without an explicit root, file and directory selections remain relative to the
current directory; component discovery can still use an automatically detected root.

Format selected component groups:

```bash
cclint fmt --check agents skills
cclint fmt --write agents skills
```

Multiple component selectors form one selection. Explicit file or directory
arguments override component selectors; `--file` overrides positional selections.

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

`summary` aggregates agents, commands, and skills. Its default output is a console
quality overview. Use `cclint summary --format json` (or `markdown`) for the shared
lint report format; JSON includes each component's quality score, while Markdown
shows diagnostics (add `--verbose` to list clean files too). `--output` saves the
report to a file. `--quiet` suppresses reports and `--fail-on` selects the failure
threshold, as for lint.

Inspect commands and the installed version:

```bash
cclint --help
cclint lint --help
cclint fmt --help
cclint summary --help
cclint --version
```

## Combine plugins, exclusions, baselines, and Git selection

Run from the project root containing `.claude/` and your plugin layout. A local
`.cclintrc.yaml` can exclude a plugin you do not maintain:

```yaml
format: json
failOn: error
exclude:
  - "plugins/marketplaces/acme/plugins/vendor/**"
```

Exclusions are root-relative globs. Full-project and Git-selected lint both honor
them. Other discovered, non-excluded components remain available to resolve skill
and agent references even when they are not selected for reporting.

Review existing findings before accepting them into a baseline: snapshots include
errors as well as advisories.

```bash
cclint --format json
cclint --baseline-create
cclint --baseline
```

The filtered report still lists the selected files, but accepted diagnostics no
longer count toward failure. After editing and staging components with your usual
Git workflow, check only new findings:

```bash
cclint --staged --baseline --format json
cclint --diff --baseline --format json
```

`--staged` selects staged paths; `--diff` also includes unstaged and untracked
paths. Both lint the current file contents, not a separate copy from the index.
A new error remains in the report and exits 1 with `failOn: error`; fixing it
restores exit 0 when no other unaccepted errors remain. Do not recreate the
baseline merely to hide the new error.

Configuration precedence is defaults → file → environment → CLI. For example,
this still emits JSON despite the environment requesting Markdown:

```bash
CCLINT_FORMAT=markdown cclint --staged --baseline --format json
```

Use `--quiet` for exit-status-only checks, or `--output report.json` to save a
non-quiet report. Baselines filter diagnostics; exclusions remove files from
automatic selection and reference discovery.

## Verification

- JSON report file exists after CI command
- `--staged` normally reports staged paths; a relevant deletion triggers
  project-wide validation to detect references to the removed component
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
