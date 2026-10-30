# Writing Agents Guide

This guide explains how to write effective agent components that pass cclint validation and follow Claude Code best practices.

---

## Overview

Agents are specialized Claude Code instructions for delegated tasks. They need not delegate further; keep the agent focused and link reusable methodology from a skill when it exists.

---

## The Thin Agent Pattern

Agents should be orchestration layers, not methodology repositories. This is the core principle behind agent design.

### Why 200 Lines?

cclint recommends 200 lines, with a 10% tolerance before its size diagnostic:

- **Agents orchestrate** - they coordinate work, make decisions, and delegate
- **Skills contain methodology** - they hold reusable patterns, reference material, and domain knowledge
- **200 lines is a guideline**; move lengthy reusable methodology to a skill when useful

When agents grow beyond 200 lines, it's a signal that methodology should move to a skill. This keeps agents focused on delegation and skills focused on reusable knowledge.

### Size Tolerance

cclint enforces the 200-line target with a 10% tolerance:

| Lines | Status | Meaning |
|-------|--------|---------|
| <=120 | Excellent | Very lean orchestration |
| <=180 | Good | Well within limits |
| <=220 | OK | Within tolerance (200 ± 10%) |
| >220 | Over limit | Exceeds best practice |
| >275 | Fat agent | Strongly exceeds, needs splitting |

---

## Required Sections

### Foundation Section

The `## Foundation` section establishes the agent's purpose, capabilities, and constraints. It answers: "What does this agent do and how should it work?"

Useful elements include purpose, capabilities, and constraints; do not add them solely to satisfy a template.

Example:

```markdown
## Foundation

**Purpose:** Orchestrate code-improvement workflows by analyzing code quality, extracting reusable patterns, and applying SOLID/DRY principles.

**Capabilities:**
- Detect architectural violations (coupling, duplication)
- Identify repeated patterns before proposing an extraction
- Apply focused changes within the requested scope

**Constraints:**
- Don't add features beyond improvement scope
- Preserve existing behavior - no functional changes
- Use agent delegation for complex multi-file operations
```

### Workflow Section

The `## Workflow` section defines the execution phases. It answers: "How does this agent accomplish its work step by step?"

Agents should use phase-based workflows:

```markdown
## Workflow

### Phase 1: Analyze Current State

1. Read target files to understand structure
2. Identify patterns and violations
3. Score complexity and coupling

### Phase 2: Plan Refactoring

1. Select only patterns relevant to the requested change
2. Apply focused changes where useful
3. Minimize changes to address specific issues

### Phase 3: Verify

1. Run the focused check for the changed behavior
2. Report any unverified boundaries
```

---

## Agent Frontmatter

Required by the CUE schema: `name` and `description`. Other fields below are optional, although cclint may suggest a model and award scoring points for tools.

```yaml
---
name: my-agent
description: Use PROACTIVELY when... (include trigger pattern)
model: sonnet
tools:
  - Tool1
  - Tool2
---
```

### Name Format

- Use lowercase letters, numbers, and hyphens only: `my-agent`, `code-poet-agent`
- No spaces, underscores, or special characters
- Avoid reserved words: `anthropic`, `claude`

### Description Guidelines

Describe concrete activation conditions. cclint rewards proactive/"use when" triggers in scoring, but does not require the exact phrase `Use PROACTIVELY when...`:

```yaml
description: Comprehensive code-quality specialist for SOLID principles and DRY analysis. Use PROACTIVELY when improving code structure, extracting abstractions, or applying design patterns.
```

**Write in third person** - describe what the agent does, not what "I" do:

- Good: "Applies SOLID principles to code structure"
- Bad: "I will apply SOLID principles"

### Model Selection

Choose the appropriate model for the task:

| Model | Use When |
|-------|----------|
| `haiku` | Quick lookups, single-file operations |
| `sonnet` | Default for most coding tasks |
| `opus` | Complex reasoning, architecture, multi-file changes |

---

## Skill References

Agents should reference skills rather than defining methodology inline. This enables:

- **Reusability** - other agents can use the same skill
- **Maintainability** - methodology updates in one place
- **Testability** - skills can be validated independently

### Reference Formats

The scorer recognizes these patterns (cross-file reference validation has its own matching rules):

```markdown
- Direct reference: `Skill: react-patterns`
- Inline call: `Skill(react-patterns)`
- List format: `Skills:` followed by skill names on the next line
- Bold reference: `**Skill**: react-patterns`
```

### When to Extract to Skills

Move methodology to a skill when:

1. **Methodology is reused** across agents/commands
2. **Agent exceeds 200 lines** - likely contains methodology
3. **Reference material needed** - lookup tables, formulas, patterns
4. **Cross-domain knowledge** - API docs, protocol details

---

## Best Practices

### DO

- Aim for fewer than 200 lines (size diagnostic beyond 220)
- Reference skills for methodology
- Use phase-based workflows (Foundation + Workflow)
- Describe when the agent should be used
- Write descriptions in third person
- Name agents after what they do, not how they work
- Include `## Success Criteria` and `## Edge Cases` sections
- Mark genuine hard gates clearly; the scorer rewards `HARD GATE` text

### DON'T

- Include `## Quick Reference` sections (belongs in skills)
- Include `## When to Use` sections (redundant with description)
- Include `## What it does` sections (use description field)
- Include `## Usage` sections (belongs in skill or omit)
- Define scoring formulas inline (should reference skill)
- Define priority matrices inline (move to skill)
- Define detection patterns inline (move to skill)
- Write descriptions in first person ("I will...")

---

## Complete Example

Here is an illustrative agent. Its exact score and validation result depend on the current project and linked skills:

```markdown
---
name: code-poet-agent
description: Deep code-quality specialist for architectural pattern detection, tiered analysis, and iterative extraction. Use PROACTIVELY for code improvement, DRY analysis, SOLID improvements, or code simplification.
model: opus
tools:
  - Read
  - Write
  - Edit
  - Bash
  - Grep
  - Glob
  - Task
  - AskUserQuestion
---

## Foundation

**Purpose:** Orchestrate code improvement using systematic pattern detection and SOLID/DRY analysis.

**Capabilities:**
- Detect architectural violations (tight coupling, duplication)
- Identify repeated patterns before proposing an extraction
- Apply focused changes within the requested scope
- Verify behavior with the relevant check

**Constraints:**
- Avoid speculative abstraction
- Preserve existing behavior - no functional changes
- Don't add features beyond improvement scope
- Don't add comments/docstrings to unchanged code
- Delegate only when it helps the task

## Workflow

### Phase 1: Analyze

1. Read target files to understand current structure
2. Identify patterns and architectural violations
3. Score complexity using SOLID/DRY rubric

### Phase 2: Extract

1. Extract only patterns needed for the requested change
2. Apply focused changes where useful
3. Minimize changes to address specific violations

### Phase 3: Verify

1. Run the focused check for the changed behavior
2. Report any unverified boundaries

## Success Criteria

- Relevant check confirms the requested behavior
- Requested code improvement is delivered without unrelated changes

## Edge Cases

- Single occurrence: don't extract (YAGNI)
- Test files: preserve test isolation
- Generated code: don't modify
- External dependencies: can't be improved safely
```

This example has Foundation and Workflow sections, third-person description triggers, and a focused success condition. If a skill actually owns the methodology, add a real `Skill: skill-name` reference rather than mentioning it only in commentary.

---

## Validation Checklist

Before considering an agent complete, verify:

- [ ] Name uses lowercase, numbers, hyphens only
- [ ] Description explains useful activation conditions
- [ ] Description is in third person
- [ ] Total lines ≤ 220 (preferably ≤ 200)
- [ ] `## Foundation` section present
- [ ] `## Workflow` section present with phases
- [ ] No `## Quick Reference` section
- [ ] No `## When to Use` section
- [ ] No `## What it does` section
- [ ] No `## Usage` section
- [ ] Reusable methodology links to an existing skill, if applicable
- [ ] No inline scoring formulas
- [ ] No inline priority matrices
- [ ] No inline regex patterns (unless trivial)

---

## Further Reading

- [Agent Lint Rules](../rules/agents.md) - Complete rule reference
- [Configuration Guide](./configuration.md) - Setting up cclint
- [Scoring Documentation](../scoring/README.md) - Quality metrics
