# Connect Assistant

## Purpose

Integrate cclint into assistant-driven workflows and CI so broken components fail fast.

## Prerequisites

- cclint installed on the runner or local machine
- A project containing Claude Code components to validate

## Main workflow

Use staged linting in local hook flows:

```bash
cclint --staged
```

Use changed-file linting during review:

```bash
cclint --diff
```

Use machine-readable output in CI:

```bash
cclint --format json --output report.json
```

Use baseline mode when onboarding legacy repositories:

```bash
cclint --baseline-create
cclint --baseline --format json --output report.json
```

## Verification

- CI job fails on new errors (`exit 1`)
- JSON output is generated at the configured path
- baseline mode ignores known issues and reports only new violations

## Related docs

- [Hook setup](guides/git-hooks.md)
- [Schema and frontmatter reference](reference/schemas.md)
- [Cross-file checks](cross-file-validation.md)

Git selection normally reports selected files, but a relevant deletion triggers
project-wide validation. Both Git modes read current working-tree contents, not
file contents from the Git index; see [common tasks](common-tasks.md).
