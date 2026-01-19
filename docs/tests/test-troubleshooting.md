# Test: Troubleshooting Guide

Tests for `docs/guides/troubleshooting.md`

## Test 1: Common Issues Listed

**Given** `docs/guides/troubleshooting.md exists
**When** the file is read
**Then** common issues are listed with solutions

**Verification**:
```bash
grep -c "^## Issue:" docs/guides/troubleshooting.md
# Expected: >= 10 (at least 10 common issues)
```

---

## Test 2: Specific Scenarios Covered

**Given** `troubleshooting.md`
**When** scanning content
**Then** "file not found" and "schema validation failed" scenarios are covered

**Verification**:
```bash
# Check for file not found issue
grep -i "file not found" docs/guides/troubleshooting.md

# Check for schema validation issue
grep -i "schema validation" docs/guides/troubleshooting.md

# Check for required fields table
grep -A 5 "Common Missing Fields" docs/guides/troubleshooting.md
```

---

## Test 3: Symptom-Cause-Solution Pattern

**Given** `troubleshooting.md`
**When** checking format
**Then** each issue follows: symptom, cause, solution pattern

**Verification**:
```bash
# Count issues with all three sections
grep -Pzo "(?s)\*\*Symptom\*\*:.*?\*\*Cause\*\*:.*?\*\*Solution\*\*:" docs/guides/troubleshooting.md | grep -c "Issue:"
# Expected: All issues have complete sections
```

---

## Test 4: Code Examples Provided

**Given** `troubleshooting.md`
**When** reviewing solutions
**Then** solutions include actionable code examples

**Verification**:
```bash
# Count code blocks in solutions
grep -c '```bash' docs/guides/troubleshooting.md
# Expected: >= 10 (one per issue)
```

---

## Test 5: Cross-References Present

**Given** `troubleshooting.md`
**When** checking links
**Then** related documentation is referenced

**Verification**:
```bash
# Check for reference links
grep -c "\[.*\](../.*\.md)" docs/guides/troubleshooting.md
# Expected: >= 3 cross-references to other docs
```
