# Writing Skills Guide

This guide explains how to write effective skill components that pass cclint validation and follow Claude Code best practices.

---

## Overview

Skills are reusable methodology components in the Claude Code system. They contain domain knowledge, patterns, reference material, and best practices that agents and commands can reference. A well-written skill is self-contained, discoverable, and follows the thin → fat pattern.

---

## The Thin → Fat Pattern

Skills follow the "thin skill → fat references" pattern:

### Why 500 Lines?

cclint recommends 500 lines and emits a size diagnostic beyond 550 lines. Keep the entry point navigable and move long reference material out when helpful:

- **SKILL.md is the entry point** - provides quick navigation to methodology
- **references/ for heavy content** - schemas, patterns, examples loaded on demand
- **500 lines is a guideline** for the core methodology, not a requirement to split useful content mechanically
- **Keeps skills discoverable** - short skills load faster and are easier to scan

When skills grow beyond 500 lines, it's a signal that reference material should move to a `references/` subdirectory. This keeps the core methodology accessible while preserving detailed documentation.

### Size Tolerance

cclint enforces the 500-line target with a 10% tolerance:

| Lines | Status | Meaning |
|-------|--------|---------|
| <=250 | Excellent | Very focused core methodology |
| <=400 | Good | Well within limits |
| <=550 | OK | Within tolerance (500 ± 10%) |
| >550 | Over limit | Exceeds best practice |
| >660 | Fat skill | Strongly exceeds, move content to references/ |

---

## Required Sections

### Quick Reference Table

The `## Quick Reference` table is a useful navigation aid for skills. It answers: "How do I use this skill?" and provides a navigation map for Claude to find relevant patterns.

**Example structure:**

```markdown
## Quick Reference

| Intent | Action |
|--------|--------|
| [User intent description] | [Action or skill reference] |
| [Another intent] | [Another action] |
```

**Why it helps:**
- **Semantic routing** - Claude matches user intent to actions
- **Pattern navigation** - Shows available methodology patterns
- **Skill composition** - Enables cross-references to other skills

**Best practices:**
- Keep descriptions concise (3-6 words)
- Use action verbs for actions
- Cross-reference related skills
- Include common use cases

### When to Use Section

The `## When to Use` section documents trigger patterns. It answers: "When should this skill be invoked?"

```markdown
## When to Use

- User mentions [specific domain keyword]
- [Specific problem pattern] is detected
- Context requires [specific capability]
- Proactively when [condition] suggests this skill helps
```

**Best practices:**
- Use bullet points for readability
- Include reactive or proactive triggers as appropriate
- Reference domain-specific keywords
- Mention related task patterns

---

## Skill Frontmatter

The CUE schema requires `name` and `description`:

```yaml
---
name: my-skill
description: Brief summary. Use when/for/proactively [trigger pattern].
---
```

### Name Format

- Use lowercase letters, numbers, and hyphens only: `my-skill`, `react-patterns`
- No consecutive hyphens (`--`)
- Cannot start or end with hyphen
- Should match parent directory name
- Avoid reserved words: `anthropic`, `claude`

### Description Guidelines

Trigger phrases help discovery; cclint suggests them when absent rather than requiring a particular phrase:

- `Use when...` - reactive invocation
- `Use for...` - domain scope
- `Use proactively...` - autonomous invocation

**Write in third person:**

- Good: "Analyzes API patterns for REST and GraphQL..."
- Bad: "I analyze API patterns..."

**Length and style:**

- Aim for 50+ characters for sufficient context (shorter descriptions receive a suggestion)
- No XML-like tags (`<tag>`)
- Describe what, not who
- Include domain keywords for searchability

---

## Recommended Sections

### Methodology Section

Document the core framework or pattern:

```markdown
## [Framework Name]

| Phase | Question | Output |
|-------|----------|--------|
| **A**nalyze | What's the current state? | State analysis |
| **B**uild | How do we improve it? | Solution plan |
| **C**heck | Did it work? | Verification |
```

### Anti-Patterns Section

Document common mistakes and what to avoid:

```markdown
## Anti-Patterns

| Anti-Pattern | Problem | Fix |
|--------------|---------|-----|
| **[Name]** | [What goes wrong] | [How to fix properly] |
```

### Examples Section

Demonstrate usage with concrete examples:

```markdown
## Examples

### Example 1: [Title]

**Input:** [Example input]

**Process:**
1. [Step 1]
2. [Step 2]

**Output:** [Expected result]
```

---

## The references/ Subdirectory

When to use `references/`:

- Heavy documentation (>100 lines)
- Reusable schemas or patterns
- Large examples or case studies
- API documentation or protocol details
- Formulas, calculations, lookup tables

### Structure

```
my-skill/
├── SKILL.md          # Core methodology (<500 lines)
└── references/       # Heavy documentation
    ├── pattern-a.md  # Detailed pattern documentation
    ├── schema.md     # Data structures or schemas
    └── examples.md   # Extended examples
```

### Referencing from SKILL.md

````markdown
## Deep Dive: [Topic]

For detailed [topic] documentation:

```text
Read(references/topic.md)
```

Or inline: `See [reference guide](references/topic.md) for details.`
````

### Reference Best Practices

- Use relative links so the skill remains portable
- Keep references navigable; deeper links are fine when they genuinely help
- Make each reference understandable from the entry point

---

## Complete Example

Here is an illustrative skill; run cclint on an actual `SKILL.md` to see project-specific findings:

````markdown
---
name: debugging-methodology
description: Systematic debugging with RPI methodology (Research-Plan-Implement). Use when investigating bugs, analyzing stack traces, diagnosing test failures, or troubleshooting production issues. Proactively use when errors lack clear root cause.
---

# Debugger

Systematic debugging methodology with RPI framework.

## Quick Reference

| Intent | Action |
|--------|--------|
| New bug, unknown cause | Use RPI framework below |
| Stack trace analysis | See Stack Trace Patterns |
| Test failure | Read(references/test-failures.md) |
| Performance issue | Read(references/performance.md) |

## When to Use

- User reports a bug with unknown root cause
- Stack traces need analysis
- Tests fail with unclear error messages
- Production issues require diagnosis
- Proactively when encountering error patterns

## RPI Framework

| Phase | Action | Output |
|-------|--------|--------|
| **R**esearch | Gather facts, read code, analyze errors | Problem understanding |
| **P**lan | Hypothesis root cause, design fix approach | Action plan |
| **I**mplement | Apply fix, verify, add regression test | Resolved issue |

---

## Phase 1: Research

**Goal**: Understand the problem before proposing solutions.

| Step | Action | Output |
|------|--------|--------|
| 1 | Reproduce the issue | Confirmed behavior |
| 2 | Read relevant code | Code understanding |
| 3 | Analyze error messages | Error context |
| 4 | Check recent changes | Possible causes |

**Template:**
```markdown
## Research

**Symptom**: [What happens]
**Reproduction**: [Steps to reproduce]
**Error**: [Exact error message]
**Context**: [Relevant code/state]
```

---

## Phase 2: Plan

**Goal**: Form hypothesis and design minimal fix.

| Question | Answer |
|----------|--------|
| Root cause hypothesis | [What's broken] |
| Fix approach | [How to fix] |
| Risk assessment | [What could break] |
| Test strategy | [How to verify] |

---

## Phase 3: Implement

**Goal**: Apply fix and verify it works.

1. Implement the fix
2. Add/confirm regression test
3. Run test suite
4. Verify production behavior

---

## Stack Trace Patterns

| Pattern | Meaning | Action |
|---------|---------|--------|
| `NullPointerException` | Null dereference | Find null variable |
| `TypeError: undefined` | Missing property | Check object structure |
| `404 Not Found` | Missing endpoint | Verify route/API path |
| `Connection refused` | Service down | Check service status |

---

## Anti-Patterns

| Anti-Pattern | Problem | Fix |
|--------------|---------|-----|
| **Fix without understanding** | Treats symptoms, not cause | Always Research first |
| **Shotgun debugging** | Changing random things | Use RPI systematically |
| **Ignoring tests** | Regression risk | Add tests before fix |
| **Premature optimization** | Wrong problem | Measure before optimizing |

---

## Examples

### Example 1: Simple Bug

**Input:** "Login fails on Safari"

**Research:**
- Reproduced on Safari 17
- Error: `localStorage is undefined`
- Code uses `localStorage` without check

**Plan:**
- Hypothesis: Safari private mode blocks localStorage
- Fix: Add try/catch with fallback

**Implement:**
```javascript
try {
  return localStorage.getItem(key);
} catch {
  return sessionStorage.getItem(key) || null;
}
```

**Verify:** Test on Safari, add regression test

### Example 2: Test Failure

**Input:** Test fails with `expected 5, got undefined`

**Research:**
- Read test to understand expectation
- Read code under test
- Function returns `undefined` instead of number

**Plan:**
- Missing `return` statement
- Add return, re-run test

**Result:** Test passes

---

## Success Criteria

**Debugging succeeds when:**
- [ ] Root cause identified (not just symptom)
- [ ] Fix is minimal and targeted
- [ ] Regression test added
- [ ] All tests pass
- [ ] Issue resolved in production

**Debugging fails when:**
- Issue recurs after fix
- Fix breaks other functionality
- No test added for regression
- Root cause unknown
````

This skill:
- Has a Quick Reference table for semantic routing
- Has When to Use section with trigger patterns
- Includes trigger phrases in description
- Has Anti-Patterns and Examples sections
- Uses third-person description
- Demonstrates references/ usage
- Can be adapted to the actual task and checked with cclint

---

## Validation Checklist

Before considering a skill complete, verify:

- [ ] File is named `SKILL.md` (case-sensitive)
- [ ] Name uses lowercase, numbers, hyphens only
- [ ] No consecutive hyphens in name
- [ ] Name does not start or end with hyphen
- [ ] Name matches parent directory
- [ ] Description explains when this skill helps
- [ ] Description gives enough context (aim for 50+ characters)
- [ ] Description is in third person
- [ ] Description contains no XML tags
- [ ] Total lines ≤ 550 (preferably ≤ 500)
- [ ] `## Quick Reference` table present
- [ ] Usage conditions are clear (a separate section is optional)
- [ ] Anti-patterns and examples are included when useful (cclint suggests them when absent)
- [ ] No reserved words in name (anthropic, claude)
- [ ] References use relative paths when the skill should be portable

---

## Further Reading

- [Skill Lint Rules](../rules/skills.md) - Complete rule reference
- [Configuration Guide](./configuration.md) - Setting up cclint
- [Scoring Documentation](../scoring/README.md) - Quality metrics
