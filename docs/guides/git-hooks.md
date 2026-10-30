# Git Hooks

A local pre-commit hook can lint Claude Code components before a commit is created. Hooks run on each machine where they are installed; they do not replace checks in shared automation.

## Pre-Commit Hook

### Basic Shell Script

Create a pre-commit hook at your repository's Git hooks path (usually `.git/hooks/pre-commit`; use `git rev-parse --git-path hooks/pre-commit` to find it for worktrees or custom layouts):

```bash
#!/bin/bash

# Exit with cclint's status; only relevant staged components are selected.
exec cclint --staged
```

Make the hook executable (adjust the path if Git reports a different one):

```bash
chmod +x .git/hooks/pre-commit
```

### With Multiple Checks

```bash
#!/bin/bash

set -e

echo "Running cclint..."

# Lint staged files only
if ! cclint --staged; then
  echo ""
  echo "❌ cclint found issues. Fix them or use --no-verify to bypass."
  exit 1
fi

echo "✅ cclint passed"
```

## CLI Flags

### `--staged`

Lint only files that are staged for commit (added via `git add`).

```bash
cclint --staged
```

**Use case**: Pre-commit hooks - only validate files being committed.

**Behavior**: Selects relevant staged `.md` and `.json` component paths. A staged component deletion triggers a full project lint so broken cross-file references can be reported. Untracked files are not selected until staged.

### `--diff`

Lint all uncommitted changes, including unstaged files.

```bash
cclint --diff
```

**Use case**: Pre-commit validation when you want to catch issues in working directory.

**Behavior**: Selects relevant tracked changes against `HEAD` plus untracked component files; a relevant tracked deletion triggers a full project lint. Files outside the configured root and non-component changes are skipped.

## Husky Integration

[ Husky](https://github.com/typicode/husky) is a popular Git hooks manager for Node.js projects.

### Installation

```bash
npm install -D husky
npx husky init
```

### Configure Pre-Commit Hook

After `npx husky init`, replace the contents of `.husky/pre-commit` with the following. The older `npx husky set .husky/pre-commit` command is not used by current Husky:

```sh
cclint --staged
```

Ensure `cclint` is on the hook's `PATH`. Avoid invoking it through `lint-staged`: cclint selects staged files itself, including relevant deletions that filename-based runners may not pass through.

## Bypassing Hooks

To bypass the pre-commit hook when needed:

```bash
git commit --no-verify -m "WIP: work in progress"
```

## Exit Codes

- `0`: All checks passed
- `1`: Lint threshold reached or invocation/tool error; read the diagnostic to distinguish them.

Git hooks run only where installed; run cclint explicitly in any separate validation environment.
