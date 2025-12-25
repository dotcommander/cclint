# Agent Lint Rules

Comprehensive reference for all validation rules applied to agent components.

---

## Overview

This document catalogs all validation rules enforced by `cclint` when linting agent files. Rules are organized by category and include severity levels, exact error messages, and source attribution.

**Rule Numbering:** Rules 001-021 cover structural, security, best-practice, and documentation requirements.

**Source Attribution:**
- `anthropic-docs`: Official Anthropic documentation requirements
- `cclint-observe`: Best practices observed from high-quality components

---

## Required Fields (001-004)

### Rule 001: Missing 'name' Field

**Severity:** error
**Component:** agent
**Category:** structural

**Description:**
The `name` field is required in all agent frontmatter per Anthropic documentation.

**Pass Criteria:**
- Frontmatter contains `name` field
- `name` value is a non-empty string

**Fail Message:**
`Required field 'name' is missing or empty`

**Source:** [Anthropic Docs - Subagents](https://code.claude.com/docs/en/sub-agents) - "name and description are required"

---

### Rule 002: Missing 'description' Field

**Severity:** error
**Component:** agent
**Category:** structural

**Description:**
The `description` field is required in all agent frontmatter per Anthropic documentation.

**Pass Criteria:**
- Frontmatter contains `description` field
- `description` value is a non-empty string

**Fail Message:**
`Required field 'description' is missing or empty`

**Source:** [Anthropic Docs - Subagents](https://code.claude.com/docs/en/sub-agents) - "name and description are required"

---

### Rule 003: Empty 'name' Field

**Severity:** error
**Component:** agent
**Category:** structural

**Description:**
The `name` field must contain a non-empty string value.

**Pass Criteria:**
- `name` field exists and is not an empty string

**Fail Message:**
`Required field 'name' is missing or empty`

**Source:** [Anthropic Docs - Subagents](https://code.claude.com/docs/en/sub-agents) - "name and description are required"

---

### Rule 004: Empty 'description' Field

**Severity:** error
**Component:** agent
**Category:** structural

**Description:**
The `description` field must contain a non-empty string value.

**Pass Criteria:**
- `description` field exists and is not an empty string

**Fail Message:**
`Required field 'description' is missing or empty`

**Source:** [Anthropic Docs - Subagents](https://code.claude.com/docs/en/sub-agents) - "name and description are required"

---

## Name Format (005-007)

### Rule 005: Invalid Name Format

**Severity:** error
**Component:** agent
**Category:** structural

**Description:**
Agent names must use only lowercase letters, numbers, and hyphens. This ensures consistent naming across the system per Anthropic documentation.

**Pass Criteria:**
- Name contains only characters from set: `a-z`, `0-9`, `-`
- No uppercase letters, spaces, or special characters

**Fail Message:**
`Name must contain only lowercase letters, numbers, and hyphens`

**Source:** [Anthropic Docs - Subagents](https://code.claude.com/docs/en/sub-agents) - "unique identifier using lowercase letters and hyphens"

---

### Rule 006: Reserved Word in Name

**Severity:** error
**Component:** agent
**Category:** structural

**Description:**
Agent names cannot use reserved words (`anthropic`, `claude`) per Anthropic documentation.

**Pass Criteria:**
- Name is not `anthropic` (case-insensitive)
- Name is not `claude` (case-insensitive)

**Fail Message:**
`Name '[name]' is a reserved word and cannot be used`

**Source:** [Anthropic Docs - Subagents](https://code.claude.com/docs/en/sub-agents) - reserved words mentioned

---

### Rule 007: Name Doesn't Match Filename

**Severity:** suggestion
**Component:** agent
**Category:** best-practice

**Description:**
For consistency and discoverability, the agent name should match the filename (without `.md` extension).

**Pass Criteria:**
- Agent `name` field matches filename
- Example: `my-agent.md` should have `name: my-agent`

**Fail Message:**
`Name "[name]" doesn't match filename "[filename]"`

**Source:** cclint observation - naming consistency for discoverability

---

## Frontmatter Validation (008-009)

### Rule 008: Invalid Color Value

**Severity:** suggestion
**Component:** agent
**Category:** structural

**Description:**
If the optional `color` field is specified, it must use a valid color name.

**Pass Criteria:**
- `color` field is one of: `red`, `blue`, `green`, `yellow`, `purple`, `orange`, `pink`, `cyan`
- Or `color` field is not specified (optional field)

**Fail Message:**
`Invalid color '[color]'. Valid colors are: red, blue, green, yellow, purple, orange, pink, cyan`

**Source:** cclint observation - color options beyond official documented list (purple, cyan, green, orange, blue, red)

---

### Rule 009: XML Tags in Description

**Severity:** error
**Component:** agent
**Category:** security

**Description:**
Agent descriptions must not contain XML-like tags per Anthropic documentation. This prevents injection attacks and ensures clean UI rendering.

**Pass Criteria:**
- Description does not match pattern `<[a-zA-Z][^>]*>`
- No HTML/XML tag-like structures

**Fail Message:**
`Description contains XML-like tags which are not allowed`

**Source:** [Anthropic Docs - Subagents](https://code.claude.com/docs/en/sub-agents) - XML tags not allowed in description field

---

## Size and Structure (010-015)

### Rule 010: Agent Exceeds 220 Lines

**Severity:** suggestion
**Component:** agent
**Category:** best-practice

**Description:**
Agents should stay under 220 lines (200 ± 10% tolerance). Longer agents indicate methodology should be extracted to skills per the thin-agent/fat-skill pattern.

**Pass Criteria:**
- Total line count (including frontmatter and body) ≤ 220

**Fail Message:**
`Agent is [N] lines. Best practice: keep agents under ~220 lines (200±10%) - move methodology to skills instead.`

**Source:** cclint observation - thin agent/fat skill pattern enforcement with tolerance threshold

---

### Rule 011: Missing Model Specification

**Severity:** suggestion
**Component:** agent
**Category:** best-practice

**Description:**
While Anthropic docs mark `model` as optional, specifying a model ensures optimal performance by choosing the right capability tier.

**Pass Criteria:**
- Frontmatter contains `model` field
- Common values: `sonnet`, `opus`, `haiku`

**Fail Message:**
`Agent lacks 'model' specification. Consider adding 'model: sonnet' or appropriate model for optimal performance.`

**Source:** cclint observation - explicit model selection improves performance predictability

---

### Rule 012: Bloat Section "Quick Reference"

**Severity:** suggestion
**Component:** agent
**Category:** best-practice

**Description:**
The section `## Quick Reference` belongs in skills, not agents. Agents orchestrate; skills contain reference material.

**Pass Criteria:**
- Agent body does not contain exact heading `## Quick Reference`

**Fail Message:**
`Agent has '## Quick Reference' - belongs in skill, not agent`

**Source:** cclint observation - reference material belongs in skills per thin agent pattern

---

### Rule 013: Bloat Section "When to Use"

**Severity:** suggestion
**Component:** agent
**Category:** best-practice

**Description:**
The section `## When to Use` is redundant. The caller (user or command) decides usage; agent description and triggers should be sufficient.

**Pass Criteria:**
- Agent body does not contain exact heading `## When to Use`

**Fail Message:**
`Agent has '## When to Use' - caller decides, use description triggers`

**Source:** cclint observation - redundant with description field and trigger patterns

---

### Rule 014: Bloat Section "What it does"

**Severity:** suggestion
**Component:** agent
**Category:** best-practice

**Description:**
The section `## What it does` is redundant with the required `description` field. Use the description instead.

**Pass Criteria:**
- Agent body does not contain exact heading `## What it does`

**Fail Message:**
`Agent has '## What it does' - belongs in description`

**Source:** cclint observation - redundant with required description field

---

### Rule 015: Bloat Section "Usage"

**Severity:** suggestion
**Component:** agent
**Category:** best-practice

**Description:**
The section `## Usage` either belongs in the corresponding skill or can be removed. Agents should be thin orchestration layers.

**Pass Criteria:**
- Agent body does not contain exact heading `## Usage`

**Fail Message:**
`Agent has '## Usage' - belongs in skill or remove`

**Source:** cclint observation - usage details should be in skills or omitted

---

## Inline Methodology Detection (016-019)

### Rule 016: Inline Scoring Formula

**Severity:** suggestion
**Component:** agent
**Category:** best-practice

**Description:**
Scoring formulas should be in skills, not inline in agents. Agents reference skill methodology; they don't define it.

**Pass Criteria:**
- Agent does not contain patterns like `score = (formula with 20+ chars)`

**Fail Message:**
`Inline scoring formula detected - should be 'See skill for scoring'`

**Source:** cclint observation - methodology should be in skills, not inline in agents

---

### Rule 017: Inline Priority Matrix

**Severity:** suggestion
**Component:** agent
**Category:** best-practice

**Description:**
Priority matrices (tables with CRITICAL/HIGH/etc.) belong in skills for reusability and maintenance.

**Pass Criteria:**
- Agent does not contain table patterns with `CRITICAL`, `HIGH` columns

**Fail Message:**
`Inline priority matrix detected - move to skill`

**Source:** cclint observation - decision matrices belong in skills for reusability

---

### Rule 018: Inline Tier Scoring

**Severity:** suggestion
**Component:** agent
**Category:** best-practice

**Description:**
Tier scoring details (Tier 1/2/3/4 with point bonuses) should be documented in skills, not agents.

**Pass Criteria:**
- Agent does not contain tier scoring patterns like `tier 1 +10 points`

**Fail Message:**
`Tier scoring details inline - move to skill`

**Source:** cclint observation - scoring methodology should be in skills for maintenance

---

### Rule 019: Inline Detection Patterns

**Severity:** suggestion
**Component:** agent
**Category:** best-practice

**Description:**
Regex patterns and detection logic belong in skills. Agents should reference skill methods, not implement them inline.

**Pass Criteria:**
- Agent does not contain `regexp.Compile()` or `regexp.MustCompile()` calls

**Fail Message:**
`Detection patterns inline - move to skill`

**Source:** cclint observation - regex patterns should be in skills, not agents

---

## Skill Integration (020-021)

### Rule 020: No Skill Reference Found

**Severity:** suggestion
**Component:** agent
**Category:** best-practice

**Description:**
If an agent has structured sections (Foundation, Workflow) but no skill references, consider extracting reusable methodology to a skill.

**Pass Criteria:**
- Agent contains `Skill(` tool call, OR
- Agent contains `Skill:` reference, OR
- Agent contains `Skills:` reference, OR
- Agent lacks both `## Foundation` and `## Workflow` sections (simple agent)

**Fail Message:**
`No skill reference found. If methodology is reusable, consider extracting to a skill.`

**Source:** cclint observation - agents with structured sections should reference skills

---

### Rule 021: Missing PROACTIVELY Pattern

**Severity:** suggestion
**Component:** agent
**Category:** documentation

**Description:**
Agent descriptions should include "Use PROACTIVELY when..." pattern to clarify automatic activation scenarios, per Anthropic best practices.

**Pass Criteria:**
- `description` field contains the word `PROACTIVELY`

**Fail Message:**
`Description lacks 'Use PROACTIVELY when...' pattern. Add to clarify activation scenarios.`

**Source:** [Anthropic Docs - Subagents](https://code.claude.com/docs/en/sub-agents) - recommended best practice for proactive agents

---

## Additional Validations

Beyond the 21 core rules, agents undergo additional validations:

### Allowed-Tools Validation
- **Category:** security
- Validates `allowed-tools` field syntax
- Ensures referenced tools exist
- Prevents overly permissive tool access

### Cross-File Validation
- **Category:** structural
- Validates skill references point to existing skills
- Checks for broken cross-component links

### Secrets Detection
- **Category:** security
- Scans for hardcoded API keys, passwords, tokens
- Prevents accidental secret exposure

### Quality Scoring
- **Category:** best-practice
- Structural quality (0-40 points)
- Best practices adherence (0-40 points)
- Composition quality (0-10 points)
- Documentation quality (0-10 points)
- Total: 0-100 with tier grading (A≥85, B≥70, C≥50, D≥30, F<30)

---

## Rule Summary by Severity

| Severity | Count | Rule IDs |
|----------|-------|----------|
| error | 7 | 001-006, 009 |
| suggestion | 14 | 007-008, 010-021 |

## Rule Summary by Category

| Category | Count | Rule IDs |
|----------|-------|----------|
| structural | 8 | 001-008 |
| security | 1 | 009 |
| best-practice | 11 | 007-008, 010-020 |
| documentation | 1 | 021 |

---

## Enforcement Notes

1. **Errors block passing:** Files with any error-severity violations fail linting
2. **Suggestions are advisory:** Files with only suggestions pass linting but show warnings
3. **Source attribution:** Each rule traces to either official docs or observed best practices
4. **Tolerance thresholds:** Line count uses ±10% tolerance (220 = 200 + 20)
5. **Pattern matching:** Bloat section detection uses exact heading match to avoid false positives
