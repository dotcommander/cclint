# Writing Agents Guide

This guide explains how to write effective agent components that pass cclint validation and follow Claude Code best practices.

---

## Overview

Agents are autonomous task execution components in the Claude Code system. They orchestrate work by delegating to skills and other specialized agents. A well-written agent is thin, focused, and references reusable methodology from skills.

---

## The Thin Agent Pattern

Agents should be orchestration layers, not methodology repositories. This is the core principle behind agent design.

### Why 200 Lines?

The 200-line limit (with 220-line tolerance) exists to enforce the thin-agent/fat-skill pattern:

- **Agents orchestrate** - they coordinate work, make decisions, and delegate
- **Skills contain methodology** - they hold reusable patterns, reference material, and domain knowledge
- **200 lines is enough** for any orchestration task if methodology is properly extracted

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

Required elements:

- **Purpose statement** - What problem does this agent solve?
- **Key capabilities** - What tools and patterns does it use?
- **Constraints** - What are the boundaries and limitations?

Example:

```markdown
## Foundation

**Purpose:** Orchestrate code-improvement workflows by analyzing code quality, extracting reusable patterns, and applying SOLID/DRY principles.

**Capabilities:**
- Detect architectural violations (coupling, duplication)
- Extract abstractions on 3rd occurrence (YAGNI principle)
- Apply SOLID design patterns with minimal changes

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

1. Extract common patterns (3rd occurrence threshold)
2. Apply SOLID principles where applicable
3. Minimize changes to address specific issues

### Phase 3: Verify

1. Run tests to confirm behavior preserved
2. Run lint/typecheck/build
3. Mark task complete when all pass
```

---

## Agent Frontmatter

Required fields:

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

**Must include** `Use PROACTIVELY when...` pattern for discoverability:

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

All these formats are detected as skill references:

```markdown
- Direct reference: `Skill: react-patterns`
- Inline call: `Skill(react-patterns)`
- List format: `Skills: react-patterns, svelte`
- Bold reference: `**Skill**: react-patterns`
```

### When to Extract to Skills

Move methodology to a skill when:

1. **Pattern appears 3+ times** across agents/commands
2. **Agent exceeds 200 lines** - likely contains methodology
3. **Reference material needed** - lookup tables, formulas, patterns
4. **Cross-domain knowledge** - API docs, protocol details

---

## Best Practices

### DO

- Keep agents under 200 lines (220 with tolerance)
- Reference skills for methodology
- Use phase-based workflows (Foundation + Workflow)
- Include `Use PROACTIVELY when...` in description
- Write descriptions in third person
- Name agents after what they do, not how they work
- Include `## Success Criteria` and `## Edge Cases` sections
- Mark hard gates with `HARD GATE` markers

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

Here is a passing agent that demonstrates all best practices:

```markdown
---
name: code-poet-agent
description: Deep code-quality specialist for architectural pattern detection, tiered analysis, and iterative extraction. Use PROACTIVELY for code improvement, DRY analysis, SOLID improvements, or code simplification.
model: opus
tools:
  - Read
  - Write
  - Edit
  - MultiEdit
  - Bash
  - Grep
  - Glob
  - Task
  - AskUserQuestion
  - TodoWrite
---

## Foundation

**Purpose:** Orchestrate code improvement using systematic pattern detection and SOLID/DRY analysis.

**Capabilities:**
- Detect architectural violations (tight coupling, duplication)
- Extract abstractions on 3rd occurrence (YAGNI principle)
- Apply SOLID patterns with minimal targeted changes
- Iterative analysis with verification cycles

**Constraints:**
- Abstract on 3rd occurrence, not before
- Preserve existing behavior - no functional changes
- Don't add features beyond improvement scope
- Don't add comments/docstrings to unchanged code
- Use agent delegation for multi-file operations

## Workflow

### Phase 1: Analyze

1. Read target files to understand current structure
2. Identify patterns and architectural violations
3. Score complexity using SOLID/DRY rubric

### Phase 2: Extract

1. Extract common patterns (3rd occurrence threshold)
2. Apply SOLID principles where applicable
3. Minimize changes to address specific violations

### Phase 3: Verify

1. Run tests to confirm behavior preserved
2. Run lint/typecheck/build
3. Complete when all pass

## Success Criteria

- All tests pass
- Lint/typecheck/build succeed
- Code follows SOLID principles (≥80% score)
- Duplication eliminated (DRY satisfied)

## Edge Cases

- Single occurrence: don't extract (YAGNI)
- Test files: preserve test isolation
- Generated code: don't modify
- External dependencies: can't be improved safely
```

This agent:
- Is 42 lines (well under 200)
- Has Foundation and Workflow sections
- References SOLID/DRY methodology (would be in a skill)
- Includes `Use PROACTIVELY when...` in description
- Has clear Success Criteria and Edge Cases
- Uses third-person description

---

## Validation Checklist

Before considering an agent complete, verify:

- [ ] Name uses lowercase, numbers, hyphens only
- [ ] Description includes `Use PROACTIVELY when...`
- [ ] Description is in third person
- [ ] Total lines ≤ 220 (preferably ≤ 200)
- [ ] `## Foundation` section present
- [ ] `## Workflow` section present with phases
- [ ] No `## Quick Reference` section
- [ ] No `## When to Use` section
- [ ] No `## What it does` section
- [ ] No `## Usage` section
- [ ] Skill references present for methodology
- [ ] No inline scoring formulas
- [ ] No inline priority matrices
- [ ] No inline regex patterns (unless trivial)

---

## Further Reading

- [Agent Lint Rules](../rules/agents.md) - Complete rule reference
- [Configuration Guide](./configuration.md) - Setting up cclint
- [Scoring Documentation](../scoring/README.md) - Quality metrics
