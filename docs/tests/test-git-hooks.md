# Test: Git Hooks Guide

## Test Cases

### Test 1: Pre-commit hook script example exists

**Given** docs/guides/git-hooks.md exists
**When** read
**Then** pre-commit hook script example is shown

**Verification:**
```bash
grep -A 5 "#!/bin/bash" docs/guides/git-hooks.md | grep -q "cclint --staged"
```

**Expected:** File contains shell script with `cclint --staged` command

---

### Test 2: CLI flags are explained

**Given** git-hooks.md
**When** scanning content
**Then** --staged and --diff flags are explained

**Verification:**
```bash
grep -q "## CLI Flags" docs/guides/git-hooks.md
grep -q "--staged" docs/guides/git-hooks.md
grep -q "--diff" docs/guides/git-hooks.md
```

**Expected:** Documentation contains CLI Flags section with both flags explained

---

### Test 3: Husky integration example included

**Given** git-hooks.md
**When** checking sections
**Then** husky integration example is included

**Verification:**
```bash
grep -q "## Husky Integration" docs/guides/git-hooks.md
grep -q "npx husky set .husky/pre-commit" docs/guides/git-hooks.md
```

**Expected:** Documentation contains Husky Integration section with installation and configuration examples

## Run All Tests

```bash
# Test 1
grep -A 5 "#!/bin/bash" docs/guides/git-hooks.md | grep -q "cclint --staged" && echo "✅ Test 1 passed" || echo "❌ Test 1 failed"

# Test 2
grep -q "## CLI Flags" docs/guides/git-hooks.md && grep -q "--staged" docs/guides/git-hooks.md && grep -q "--diff" docs/guides/git-hooks.md && echo "✅ Test 2 passed" || echo "❌ Test 2 failed"

# Test 3
grep -q "## Husky Integration" docs/guides/git-hooks.md && grep -q "npx husky set" docs/guides/git-hooks.md && echo "✅ Test 3 passed" || echo "❌ Test 3 failed"
```
