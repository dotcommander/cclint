# Git Hooks

Automated linting with Git hooks ensures code quality before commits are pushed.

## Pre-Commit Hook

### Basic Shell Script

Create `.git/hooks/pre-commit`:

```bash
#!/bin/bash

# Run cclint on staged files
cclint --staged

# Exit with cclint's status
exit $?
```

Make it executable:

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

**Behavior**: Uses `git diff --cached --name-only --diff-filter=ACM` to find staged files.

### `--diff`

Lint all uncommitted changes, including unstaged files.

```bash
cclint --diff
```

**Use case**: Pre-commit validation when you want to catch issues in working directory.

**Behavior**: Uses `git diff --name-only --diff-filter=ACM` to find all modified files.

## Husky Integration

[ Husky](https://github.com/typicode/husky) is a popular Git hooks manager for Node.js projects.

### Installation

```bash
npm install -D husky
npx husky init
```

### Configure Pre-Commit Hook

```bash
npx husky set .husky/pre-commit "cclint --staged"
```

This creates `.husky/pre-commit`:

```bash
#!/usr/bin/env sh
. "$(dirname -- "$0")/_/husky.sh"

cclint --staged
```

### With Lint-Staged (Optional)

For better performance with large repositories, combine with [lint-staged](https://github.com/okonet/lint-staged):

```bash
npm install -D lint-staged
```

Add to `package.json`:

```json
{
  "lint-staged": {
    "*.md": "cclint --file"
  }
}
```

Update `.husky/pre-commit`:

```bash
npx husky set .husky/pre-commit "npx lint-staged"
```

## Bypassing Hooks

To bypass the pre-commit hook when needed:

```bash
git commit --no-verify -m "WIP: work in progress"
```

## Exit Codes

- `0`: All checks passed
- `1`: Linting errors found
- Other: Tool error (missing config, invalid arguments, etc.)

## CI/CD Integration

For CI pipelines, prefer explicit checks over Git hooks:

```bash
# GitHub Actions example
- name: Lint with cclint
  run: cclint --format json --output cclint-report.json
```

Git hooks are not executed in CI environments.
