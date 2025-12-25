# cclint Rules Reference

Complete reference of all 124 lint rules enforced by cclint.

---

## Table of Contents

- [Agent Rules (001-021)](#agent-rules-001-021)
- [Command Rules (022-034)](#command-rules-022-034)
- [Skill Rules (035-047)](#skill-rules-035-047)
- [Settings Rules (048-074)](#settings-rules-048-074)
- [Plugin Rules (075-092)](#plugin-rules-075-092)
- [Security Rules (093-104)](#security-rules-093-104)
- [Schema Constraints (105-124)](#schema-constraints-105-124)

## Severity Legend

- **error** - Must fix (blocks component usage)
- **warning** - Should fix (security/quality concern)
- **suggestion** - Optional (best practice recommendation)

## Source Attribution

Each rule includes a source indicating its authority:

| Source Type | Meaning | Enforced? |
|-------------|---------|-----------|
| **anthropic-docs** | Official Anthropic documentation requirement | Yes (error severity) |
| **cclint-observe** | cclint's opinionated best practice | Advisory (suggestion severity) |
| **industry-standard** | Industry security standards (OWASP, CWE, RFC) | Warning (security concern) |

### About cclint Observations

Rules marked `cclint-observe` are **NOT** Anthropic requirements. They are opinionated best practices based on:
- Patterns observed in high-quality Claude Code components
- Maintainability and readability improvements
- The thin-agent/fat-skill architecture pattern (cclint's recommendation)

**These are suggestions, not mandates.** Components can pass linting while having suggestion-severity violations.

### Anthropic-Documented vs cclint-Observed

| What Anthropic Documents | What cclint Adds |
|--------------------------|------------------|
| name/description required | Line count limits |
| Lowercase hyphenated names | Bloat section detection |
| Reserved words blocked | Inline methodology detection |
| "Use PROACTIVELY" patterns | Thin/fat architecture enforcement |
| "Use when" trigger phrases | Skill reference requirements |
| XML tags not allowed | Name-filename matching |

For detailed sources with URLs, see the individual rule files in `docs/rules/`.

---

## Agent Rules (001-021)

### Required Fields (001-004)

#### Rule 001: Missing 'name' Field

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

**Source:** anthropic-docs

---

#### Rule 002: Missing 'description' Field

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

**Source:** anthropic-docs

---

#### Rule 003: Empty 'name' Field

**Severity:** error
**Component:** agent
**Category:** structural

**Description:**
The `name` field must contain a non-empty string value.

**Pass Criteria:**
- `name` field exists and is not an empty string

**Fail Message:**
`Required field 'name' is missing or empty`

**Source:** anthropic-docs

---

#### Rule 004: Empty 'description' Field

**Severity:** error
**Component:** agent
**Category:** structural

**Description:**
The `description` field must contain a non-empty string value.

**Pass Criteria:**
- `description` field exists and is not an empty string

**Fail Message:**
`Required field 'description' is missing or empty`

**Source:** anthropic-docs

---

### Name Format (005-007)

#### Rule 005: Invalid Name Format

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

**Source:** anthropic-docs

---

#### Rule 006: Reserved Word in Name

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

**Source:** anthropic-docs

---

#### Rule 007: Name Doesn't Match Filename

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

**Source:** cclint-observe

---

### Frontmatter Validation (008-009)

#### Rule 008: Invalid Color Value

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

**Source:** cclint-observe

---

#### Rule 009: XML Tags in Description

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

**Source:** anthropic-docs

---

### Size and Structure (010-015)

#### Rule 010: Agent Exceeds 220 Lines

**Severity:** suggestion
**Component:** agent
**Category:** best-practice

**Description:**
Agents should stay under 220 lines (200 ± 10% tolerance). Longer agents indicate methodology should be extracted to skills per the thin-agent/fat-skill pattern.

**Pass Criteria:**
- Total line count (including frontmatter and body) ≤ 220

**Fail Message:**
`Agent is [N] lines. Best practice: keep agents under ~220 lines (200±10%) - move methodology to skills instead.`

**Source:** cclint-observe

---

#### Rule 011: Missing Model Specification

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

**Source:** cclint-observe

---

#### Rule 012: Bloat Section "Quick Reference"

**Severity:** suggestion
**Component:** agent
**Category:** best-practice

**Description:**
The section `## Quick Reference` belongs in skills, not agents. Agents orchestrate; skills contain reference material.

**Pass Criteria:**
- Agent body does not contain exact heading `## Quick Reference`

**Fail Message:**
`Agent has '## Quick Reference' - belongs in skill, not agent`

**Source:** cclint-observe

---

#### Rule 013: Bloat Section "When to Use"

**Severity:** suggestion
**Component:** agent
**Category:** best-practice

**Description:**
The section `## When to Use` is redundant. The caller (user or command) decides usage; agent description and triggers should be sufficient.

**Pass Criteria:**
- Agent body does not contain exact heading `## When to Use`

**Fail Message:**
`Agent has '## When to Use' - caller decides, use description triggers`

**Source:** cclint-observe

---

#### Rule 014: Bloat Section "What it does"

**Severity:** suggestion
**Component:** agent
**Category:** best-practice

**Description:**
The section `## What it does` is redundant with the required `description` field. Use the description instead.

**Pass Criteria:**
- Agent body does not contain exact heading `## What it does`

**Fail Message:**
`Agent has '## What it does' - belongs in description`

**Source:** cclint-observe

---

#### Rule 015: Bloat Section "Usage"

**Severity:** suggestion
**Component:** agent
**Category:** best-practice

**Description:**
The section `## Usage` either belongs in the corresponding skill or can be removed. Agents should be thin orchestration layers.

**Pass Criteria:**
- Agent body does not contain exact heading `## Usage`

**Fail Message:**
`Agent has '## Usage' - belongs in skill or remove`

**Source:** cclint-observe

---

### Inline Methodology Detection (016-019)

#### Rule 016: Inline Scoring Formula

**Severity:** suggestion
**Component:** agent
**Category:** best-practice

**Description:**
Scoring formulas should be in skills, not inline in agents. Agents reference skill methodology; they don't define it.

**Pass Criteria:**
- Agent does not contain patterns like `score = (formula with 20+ chars)`

**Fail Message:**
`Inline scoring formula detected - should be 'See skill for scoring'`

**Source:** cclint-observe

---

#### Rule 017: Inline Priority Matrix

**Severity:** suggestion
**Component:** agent
**Category:** best-practice

**Description:**
Priority matrices (tables with CRITICAL/HIGH/etc.) belong in skills for reusability and maintenance.

**Pass Criteria:**
- Agent does not contain table patterns with `CRITICAL`, `HIGH` columns

**Fail Message:**
`Inline priority matrix detected - move to skill`

**Source:** cclint-observe

---

#### Rule 018: Inline Tier Scoring

**Severity:** suggestion
**Component:** agent
**Category:** best-practice

**Description:**
Tier scoring details (Tier 1/2/3/4 with point bonuses) should be documented in skills, not agents.

**Pass Criteria:**
- Agent does not contain tier scoring patterns like `tier 1 +10 points`

**Fail Message:**
`Tier scoring details inline - move to skill`

**Source:** cclint-observe

---

#### Rule 019: Inline Detection Patterns

**Severity:** suggestion
**Component:** agent
**Category:** best-practice

**Description:**
Regex patterns and detection logic belong in skills. Agents should reference skill methods, not implement them inline.

**Pass Criteria:**
- Agent does not contain `regexp.Compile()` or `regexp.MustCompile()` calls

**Fail Message:**
`Detection patterns inline - move to skill`

**Source:** cclint-observe

---

### Skill Integration (020-021)

#### Rule 020: No Skill Reference Found

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

**Source:** cclint-observe

---

#### Rule 021: Missing PROACTIVELY Pattern

**Severity:** suggestion
**Component:** agent
**Category:** documentation

**Description:**
Agent descriptions should include "Use PROACTIVELY when..." pattern to clarify automatic activation scenarios, per Anthropic best practices.

**Pass Criteria:**
- `description` field contains the word `PROACTIVELY`

**Fail Message:**
`Description lacks 'Use PROACTIVELY when...' pattern. Add to clarify activation scenarios.`

**Source:** anthropic-docs

---

## Command Rules (022-034)

### Rule 022: Invalid Name Format

**Severity:** error
**Component:** command
**Category:** structural

**Description:**
Validates that the `name` field in command frontmatter (if present) follows the required format: lowercase alphanumeric characters with hyphens only.

**Pass Criteria:**
- Name contains only characters: `a-z`, `0-9`, `-`
- No uppercase letters, underscores, or special characters
- Name field is optional (derived from filename per Anthropic docs)

**Fail Message:**
`Name must be lowercase alphanumeric with hyphens only`

**Source:** Anthropic Documentation

---

### Rule 023: XML Tags in Description

**Severity:** error
**Component:** command
**Category:** structural

**Description:**
Detects XML-like tags in the `description` field of command frontmatter. XML tags are not allowed in command descriptions per Anthropic documentation.

**Pass Criteria:**
- Description field contains no XML-like tags (e.g., `<tag>`, `<div>`, `<example>`)
- Plain text descriptions only

**Fail Message:**
`Description contains XML-like tags which are not allowed`

**Source:** Anthropic Documentation

---

### Rule 024: Command Exceeds Line Limit

**Severity:** suggestion
**Component:** command
**Category:** best-practice

**Description:**
Checks that command files stay within the recommended size limit of ~55 lines (50 lines ±10% tolerance). Commands exceeding this threshold should delegate to specialist agents instead of implementing logic directly.

**Pass Criteria:**
- Total line count ≤ 55 lines
- Commands follow thin delegation pattern

**Fail Message:**
`Command is [N] lines. Best practice: keep commands under ~55 lines (50±10%) - delegate to specialist agents instead of implementing logic directly.`

**Source:** cclint observation (thin/fat architecture pattern)

---

### Rule 025: Direct Implementation Patterns Detected

**Severity:** suggestion
**Component:** command
**Category:** best-practice

**Description:**
Detects implementation sections (`## Implementation`, `### Steps`) in command files. Commands should delegate to specialist agents rather than containing step-by-step implementation details.

**Pass Criteria:**
- No `## Implementation` section
- No `### Steps` section
- Command delegates to agents instead of implementing logic

**Fail Message:**
`Command contains implementation steps. Consider delegating to a specialist agent instead.`

**Source:** cclint observation (thin command pattern)

---

### Rule 026: Missing allowed-tools with Task() Delegation

**Severity:** suggestion
**Component:** command
**Category:** security

**Description:**
Checks that commands using the `Task()` tool in their body include the `allowed-tools` permission in frontmatter. This ensures explicit permission for agent delegation.

**Pass Criteria:**
- If `Task(` appears in command body, `allowed-tools:` must be present in frontmatter
- Typically `allowed-tools: Task` or specific agent permissions

**Fail Message:**
`Command uses Task() but lacks 'allowed-tools' permission. Add 'allowed-tools: Task' to frontmatter.`

**Source:** cclint observation (security best practice)

---

### Rule 027: Bloat Section "Quick Reference"

**Severity:** suggestion
**Component:** command
**Category:** best-practice

**Description:**
Detects `## Quick Reference` section in thin commands (those delegating via `Task()`). Quick reference tables belong in skills, not commands.

**Pass Criteria:**
- Thin commands (with `Task()` delegation) do not contain `## Quick Reference`
- Reference content resides in skill files

**Fail Message:**
`Thin command has '## Quick Reference' - belongs in skill, not command`

**Source:** cclint observation (thin/fat architecture pattern)

---

### Rule 028: Bloat Section "Usage"

**Severity:** suggestion
**Component:** command
**Category:** best-practice

**Description:**
Detects `## Usage` section in thin commands (those delegating via `Task()`). Since thin commands delegate to agents that have full context, usage sections are redundant.

**Pass Criteria:**
- Thin commands (with `Task()` delegation) do not contain `## Usage`
- Usage documentation resides in agent or skill files

**Fail Message:**
`Thin command has '## Usage' - agent has full context, remove`

**Source:** cclint observation (thin/fat architecture pattern)

---

### Rule 029: Bloat Section "Workflow"

**Severity:** suggestion
**Component:** command
**Category:** best-practice

**Description:**
Detects `## Workflow` section in thin commands (those delegating via `Task()`). Workflow details duplicate agent content and should not appear in thin commands.

**Pass Criteria:**
- Thin commands (with `Task()` delegation) do not contain `## Workflow`
- Workflow documentation resides in agent files

**Fail Message:**
`Thin command has '## Workflow' - duplicates agent content, remove`

**Source:** cclint observation (thin/fat architecture pattern)

---

### Rule 030: Bloat Section "When to use"

**Severity:** suggestion
**Component:** command
**Category:** best-practice

**Description:**
Detects `## When to use` section in thin commands. This information should be captured in the command's `description` field instead.

**Pass Criteria:**
- Thin commands (with `Task()` delegation) do not contain `## When to use`
- Usage context documented in description field

**Fail Message:**
`Thin command has '## When to use' - belongs in description, remove`

**Source:** cclint observation (thin/fat architecture pattern)

---

### Rule 031: Bloat Section "What it does"

**Severity:** suggestion
**Component:** command
**Category:** best-practice

**Description:**
Detects `## What it does` section in thin commands. This information should be captured in the command's `description` field instead.

**Pass Criteria:**
- Thin commands (with `Task()` delegation) do not contain `## What it does`
- Functionality described in description field

**Fail Message:**
`Thin command has '## What it does' - belongs in description, remove`

**Source:** cclint observation (thin/fat architecture pattern)

---

### Rule 032: Excessive Code Examples

**Severity:** suggestion
**Component:** command
**Category:** best-practice

**Description:**
Checks that commands contain no more than 2 code examples. Excessive examples indicate the command may be too complex and should delegate to an agent.

**Pass Criteria:**
- Maximum of 2 code blocks with ```bash or ```shell fences
- Concise example usage only

**Fail Message:**
`Command has [N] code examples. Best practice: max 2 examples.`

**Source:** cclint observation (documentation best practice)

---

### Rule 033: Success Criteria Not in Checkbox Format

**Severity:** suggestion
**Component:** command
**Category:** documentation

**Description:**
Checks that success criteria sections use checkbox format (`- [ ]`) instead of prose. This makes success conditions explicit and testable.

**Pass Criteria:**
- If `## Success` or `Success criteria:` section exists, it must contain `- [ ]` checkboxes
- Success criteria are actionable and verifiable

**Fail Message:**
`Success criteria should use checkbox format '- [ ]' not prose`

**Source:** cclint observation (documentation best practice)

---

### Rule 034: Missing Usage Section in Fat Command

**Severity:** suggestion
**Component:** command
**Category:** documentation

**Description:**
Detects fat commands (>40 lines without `Task()` delegation) that lack a `## Usage` or `## Workflow` section. Fat commands implementing logic directly should document their usage, or better yet, delegate to a specialist agent.

**Pass Criteria:**
- Fat commands (>40 lines, no `Task()`) contain `## Usage` or `## Workflow` section
- OR command is refactored to delegate to an agent (preferred)

**Fail Message:**
`Fat command without Task delegation lacks '## Usage' section. Consider delegating to a specialist agent.`

**Source:** cclint observation (documentation best practice)

---

## Skill Rules (035-047)

### Rule 035: Skill file must be named SKILL.md

**Severity:** error
**Component:** skill
**Category:** structural

**Description:**
Skill files must be named `SKILL.md` (case-sensitive) according to Anthropic's documentation standards. This naming convention is critical for skill discovery and loading.

**Pass Criteria:**
- File path ends with `/SKILL.md` (Unix) or `\SKILL.md` (Windows)

**Fail Message:**
`Skill file must be named SKILL.md`

**Source:** Anthropic Documentation

---

### Rule 036: Empty skill file

**Severity:** error
**Component:** skill
**Category:** structural

**Description:**
Skill files must contain content. Empty files provide no value and cannot be loaded by the skill system.

**Pass Criteria:**
- File contains non-whitespace content

**Fail Message:**
`Skill file is empty`

**Source:** Anthropic Documentation

---

### Rule 037: Missing frontmatter

**Severity:** suggestion
**Component:** skill
**Category:** documentation

**Description:**
Skills should include YAML frontmatter with `name` and `description` fields. The description is critical for skill discovery, allowing Claude to find and load appropriate skills based on task requirements.

**Pass Criteria:**
- File starts with `---` (YAML frontmatter delimiter)

**Fail Message:**
`Add YAML frontmatter with name and description (description is critical for skill discovery)`

Or when skill name can be extracted:
`Add frontmatter: ---\nname: {skill-name}\ndescription: Brief summary of what this skill does\n--- (description critical for discovery)`

**Source:** Anthropic Documentation

---

### Rule 038: Reserved word in name

**Severity:** error
**Component:** skill
**Category:** best-practice

**Description:**
Skill names cannot use reserved words like "anthropic" or "claude". These are reserved for official components.

**Pass Criteria:**
- `name` field does not equal "anthropic" or "claude" (case-insensitive)

**Fail Message:**
`Name '{name}' is a reserved word and cannot be used`

**Source:** Anthropic Documentation

---

### Rule 039: XML tags in description

**Severity:** error
**Component:** skill
**Category:** best-practice

**Description:**
Skill descriptions must not contain XML-like tags (e.g., `<tag>`). These interfere with skill discovery and processing.

**Pass Criteria:**
- `description` field does not match pattern `<[a-zA-Z][^>]*>`

**Fail Message:**
`Description contains XML-like tags which are not allowed`

**Source:** Anthropic Documentation

---

### Rule 040: First-person description

**Severity:** suggestion
**Component:** skill
**Category:** best-practice

**Description:**
Skill descriptions should use third-person perspective (e.g., "Analyzes..." not "I analyze..."). This maintains consistency with Anthropic's documentation standards and improves skill discovery.

**Pass Criteria:**
- `description` does not start with: "I ", "I'm ", "I'll ", "I've ", "My ", "We ", "We're "

**Fail Message:**
`Skill description should use third person (e.g., 'Analyzes...' not 'I analyze...')`

**Source:** Anthropic Documentation

---

### Rule 041: Addressing user in description

**Severity:** suggestion
**Component:** skill
**Category:** best-practice

**Description:**
Skill descriptions should describe what the skill does, not address the user directly (e.g., "You can..."). This improves clarity and skill discovery.

**Pass Criteria:**
- `description` does not start with "You "

**Fail Message:**
`Skill description should describe what it does, not address the user`

**Source:** Anthropic Documentation

---

### Rule 042: Description too short

**Severity:** suggestion
**Component:** skill
**Category:** documentation

**Description:**
Skill descriptions should be at least 50 characters to provide sufficient detail for skill discovery. Shorter descriptions may not provide enough context for Claude to determine when to use the skill.

**Pass Criteria:**
- `description` field is 50+ characters

**Fail Message:**
`Description is only {length} chars. Aim for 50+ to help with skill discovery.`

**Source:** Anthropic Documentation

---

### Rule 043: Missing trigger phrases

**Severity:** suggestion
**Component:** skill
**Category:** documentation

**Description:**
Skill descriptions should include trigger phrases like "Use when...", "Use for...", or "Use proactively..." to help Claude understand when to invoke the skill.

**Pass Criteria:**
- `description` contains at least one of: "use when", "use for", "use proactively" (case-insensitive)

**Fail Message:**
`Consider adding trigger phrases like 'Use when...' or 'Use for...' to help skill discovery`

**Source:** Anthropic Documentation

---

### Rule 044: Invalid semver version format

**Severity:** warning
**Component:** skill
**Category:** best-practice

**Description:**
If a `version` field is provided in frontmatter, it should follow semantic versioning format (e.g., "1.0.0"). This enables version tracking and compatibility checking.

**Pass Criteria:**
- `version` field matches pattern: `^\d+\.\d+\.\d+(-[a-zA-Z0-9.]+)?(\+[a-zA-Z0-9.]+)?$`
- Examples: "1.0.0", "2.1.3-beta", "1.0.0+build.123"

**Fail Message:**
`Version '{version}' should follow semver format (e.g., '1.0.0')`

**Source:** cclint observation

---

### Rule 045: Missing Anti-Patterns section

**Severity:** suggestion
**Component:** skill
**Category:** documentation

**Description:**
Skills should include an "Anti-Patterns" section documenting common mistakes, incorrect usage patterns, or approaches to avoid. This helps users understand not just how to use the skill correctly, but what to avoid.

**Pass Criteria:**
Document contains at least one of:
- `## Anti-Pattern` or `## Anti-Patterns`
- `### Anti-Pattern`
- `## Best Practices` with `### don't` subsection
- Table with `| Anti-Pattern` column

**Fail Message:**
`Consider adding '## Anti-Patterns' section to document common mistakes.`

**Source:** cclint observation

---

### Rule 046: Skill exceeds 550 lines

**Severity:** suggestion
**Component:** skill
**Category:** structural

**Description:**
Skills should be kept under ~550 lines (500 lines with 10% tolerance). Larger skills become difficult to maintain and slow to load. Heavy documentation, schemas, or examples should be moved to a `references/` subdirectory and loaded on demand.

**Pass Criteria:**
- File has 550 or fewer lines

**Fail Message:**
`Skill is {lines} lines. Best practice: keep skills under ~550 lines (500±10%) - move heavy docs to references/ subdirectory.`

**Source:** cclint observation

---

### Rule 047: Missing Examples section

**Severity:** suggestion
**Component:** skill
**Category:** documentation

**Description:**
Skills should include examples demonstrating usage. This helps users understand how to apply the skill in practice.

**Pass Criteria:**
Document contains at least one of:
- `## Example` or `## Examples`
- `## Expected Output`
- `## Usage`
- `### Example`

**Fail Message:**
`Consider adding '## Examples' section to illustrate skill usage.`

**Source:** cclint observation

---

## Settings Rules (048-074)

### Hook Structure Rules (048-057)

#### Rule 048: JSON Parse Error

**Severity:** error
**Component:** settings
**Category:** structural

**Description:**
Settings file must contain valid JSON. This rule catches syntax errors in JSON parsing.

**Pass Criteria:**
File contains well-formed JSON with proper syntax (valid braces, quotes, commas, etc.)

**Fail Message:**
`Error parsing JSON: [error details]`

---

#### Rule 049: Unknown Hook Event Name

**Severity:** error
**Component:** settings
**Category:** structural

**Description:**
Hook event names must match Anthropic's documented event types. Valid events: PreToolUse, PostToolUse, PermissionRequest, Notification, UserPromptSubmit, Stop, SubagentStop, PreCompact, SessionStart, SessionEnd.

**Pass Criteria:**
Every key in the `hooks` object matches a valid hook event name from Anthropic documentation.

**Fail Message:**
`Unknown hook event '[eventName]'. Valid events: PreToolUse, PostToolUse, PermissionRequest, Notification, UserPromptSubmit, Stop, SubagentStop, PreCompact, SessionStart, SessionEnd`

**Source:** anthropic-docs

---

#### Rule 050: Hook Configuration Not Array

**Severity:** error
**Component:** settings
**Category:** structural

**Description:**
Each hook event's configuration must be an array of hook matchers. The structure is `{ "EventName": [...] }` where the value is an array.

**Pass Criteria:**
The value for each hook event key is a JSON array.

**Fail Message:**
`Event '[eventName]': hook configuration must be an array`

**Source:** anthropic-docs

---

#### Rule 051: Hook Matcher Not Object

**Severity:** error
**Component:** settings
**Category:** structural

**Description:**
Each element in a hook event's array must be an object containing `matcher` and `hooks` fields.

**Pass Criteria:**
Each array element is a JSON object (not a string, number, or array).

**Fail Message:**
`Event '[eventName]' hook [index]: must be an object with 'matcher' and 'hooks' fields`

**Source:** anthropic-docs

---

#### Rule 052: Missing 'matcher' Field

**Severity:** error
**Component:** settings
**Category:** structural

**Description:**
Each hook matcher object must contain a `matcher` field that defines when the hook should trigger.

**Pass Criteria:**
Hook matcher object contains the required `matcher` field.

**Fail Message:**
`Event '[eventName]' hook [index]: missing required field 'matcher'`

**Source:** anthropic-docs

---

#### Rule 053: Missing 'hooks' Field

**Severity:** error
**Component:** settings
**Category:** structural

**Description:**
Each hook matcher object must contain a `hooks` field that defines the array of hooks to execute when the matcher matches.

**Pass Criteria:**
Hook matcher object contains the required `hooks` field.

**Fail Message:**
`Event '[eventName]' hook [index]: missing required field 'hooks'`

**Source:** anthropic-docs

---

#### Rule 054: Inner Hooks Not Array

**Severity:** error
**Component:** settings
**Category:** structural

**Description:**
The `hooks` field within a matcher must be an array of hook objects.

**Pass Criteria:**
The `hooks` field value is a JSON array.

**Fail Message:**
`Event '[eventName]' hook [index]: 'hooks' field must be an array`

**Source:** anthropic-docs

---

#### Rule 055: Hook Object Not Object

**Severity:** error
**Component:** settings
**Category:** structural

**Description:**
Each individual hook within the `hooks` array must be an object containing hook configuration (type, command/prompt, etc.).

**Pass Criteria:**
Each element in the inner `hooks` array is a JSON object.

**Fail Message:**
`Event '[eventName]' hook [index] inner hook [innerIndex]: must be an object`

**Source:** anthropic-docs

---

#### Rule 056: Missing 'type' Field

**Severity:** error
**Component:** settings
**Category:** structural

**Description:**
Each hook object must specify its `type` field (either "command" or "prompt").

**Pass Criteria:**
Hook object contains the required `type` field.

**Fail Message:**
`Event '[eventName]' hook [index] inner hook [innerIndex]: missing required field 'type'`

**Source:** anthropic-docs

---

#### Rule 057: Type Not String

**Severity:** error
**Component:** settings
**Category:** structural

**Description:**
The `type` field must be a string value, not a number, boolean, or object.

**Pass Criteria:**
The `type` field contains a JSON string value.

**Fail Message:**
`Event '[eventName]' hook [index] inner hook [innerIndex]: 'type' must be a string`

**Source:** anthropic-docs

---

### Hook Type Rules (058-061)

#### Rule 058: Invalid Hook Type

**Severity:** error
**Component:** settings
**Category:** structural

**Description:**
Hook type must be either "command" (executes a shell command) or "prompt" (modifies Claude's prompt). No other values are valid.

**Pass Criteria:**
The `type` field value is exactly "command" or "prompt".

**Fail Message:**
`Event '[eventName]' hook [index] inner hook [innerIndex]: invalid type '[hookType]'. Valid types: command, prompt`

**Source:** anthropic-docs

---

#### Rule 059: Command Type Missing 'command' Field

**Severity:** error
**Component:** settings
**Category:** structural

**Description:**
When hook type is "command", the hook object must include a `command` field containing the shell command to execute.

**Pass Criteria:**
Hook object with `"type": "command"` contains a `command` field.

**Fail Message:**
`Event '[eventName]' hook [index] inner hook [innerIndex]: type 'command' requires 'command' field`

**Source:** anthropic-docs

---

#### Rule 060: Prompt Type on Unsupported Event

**Severity:** error
**Component:** settings
**Category:** structural

**Description:**
Prompt hooks are only supported for specific events: Stop, SubagentStop, UserPromptSubmit, PreToolUse, PermissionRequest. Other events cannot use prompt-type hooks.

**Pass Criteria:**
If hook type is "prompt", the parent event must be one of the supported prompt hook events.

**Fail Message:**
`Event '[eventName]' hook [index] inner hook [innerIndex]: event '[eventName]' does not support prompt hooks. Prompt hooks only supported for: Stop, SubagentStop, UserPromptSubmit, PreToolUse, PermissionRequest`

**Source:** anthropic-docs

---

#### Rule 061: Prompt Type Missing 'prompt' Field

**Severity:** error
**Component:** settings
**Category:** structural

**Description:**
When hook type is "prompt", the hook object must include a `prompt` field containing the prompt text to inject.

**Pass Criteria:**
Hook object with `"type": "prompt"` contains a `prompt` field.

**Fail Message:**
`Event '[eventName]' hook [index] inner hook [innerIndex]: type 'prompt' requires 'prompt' field`

**Source:** anthropic-docs

---

### Hook Security Rules (062-074)

#### Rule 062: Unquoted Variable Expansion

**Severity:** warning
**Component:** settings
**Category:** security

**Description:**
Variable expansions ($VAR or ${VAR}) should be quoted to prevent word splitting and pathname expansion. Unquoted variables can lead to unexpected behavior or security issues if they contain spaces or special characters.

**Pass Criteria:**
All variable references in command strings are quoted: `"$VAR"` or `'$VAR'`.

**Fail Message:**
`Event '[eventName]' hook [index] inner hook [innerIndex]: Unquoted variable expansion detected. Use "$VAR" to prevent word splitting`

**Source:** cclint-observe

---

#### Rule 063: Path Traversal Detected

**Severity:** warning
**Component:** settings
**Category:** security

**Description:**
The `..` pattern in paths can be used for path traversal attacks, accessing files outside intended directories. This is a potential security risk.

**Pass Criteria:**
Command string does not contain `..` sequences.

**Fail Message:**
`Event '[eventName]' hook [index] inner hook [innerIndex]: Path traversal '..' detected in hook command - potential security risk`

**Source:** cclint-observe

---

#### Rule 064: Hardcoded Absolute Path

**Severity:** warning
**Component:** settings
**Category:** best-practice

**Description:**
Hardcoded absolute paths (like `/Users/username/project` or `/home/user/file`) make hooks non-portable. Use `$CLAUDE_PROJECT_DIR` for project-relative paths to ensure hooks work across machines.

**Pass Criteria:**
If absolute paths are present (starting with /Users, /home, /var, /tmp, /etc), the command also uses `$CLAUDE_PROJECT_DIR` or the path is legitimately system-wide.

**Fail Message:**
`Event '[eventName]' hook [index] inner hook [innerIndex]: Hardcoded absolute path detected. Consider using $CLAUDE_PROJECT_DIR for portability`

**Source:** cclint-observe

---

#### Rule 065: .env File Access

**Severity:** warning
**Component:** settings
**Category:** security

**Description:**
Accessing `.env` files in hooks may expose secrets. Ensure that hook commands do not log or transmit environment secrets.

**Pass Criteria:**
Command does not reference `.env` files, or access is verified to be secure.

**Fail Message:**
`Event '[eventName]' hook [index] inner hook [innerIndex]: Accessing .env file - ensure secrets are not logged`

**Source:** cclint-observe

---

#### Rule 066: .git Directory Access

**Severity:** warning
**Component:** settings
**Category:** security

**Description:**
Accessing the `.git` directory can expose repository internals, credentials, or sensitive history. This is a potential security concern.

**Pass Criteria:**
Command does not access `.git/` paths.

**Fail Message:**
`Event '[eventName]' hook [index] inner hook [innerIndex]: Accessing .git directory - potential security concern`

**Source:** cclint-observe

---

#### Rule 067: Credentials File Access

**Severity:** warning
**Component:** settings
**Category:** security

**Description:**
Accessing files with "credentials" in the name suggests handling of sensitive authentication data. Ensure secure handling and no exposure of secrets.

**Pass Criteria:**
Command does not reference files containing the word "credentials".

**Fail Message:**
`Event '[eventName]' hook [index] inner hook [innerIndex]: Accessing credentials file - ensure secure handling`

**Source:** cclint-observe

---

#### Rule 068: .ssh Directory Access

**Severity:** warning
**Component:** settings
**Category:** security

**Description:**
The `.ssh` directory contains private keys and SSH configuration. Accessing it in hooks is a high security risk and should be carefully reviewed.

**Pass Criteria:**
Command does not access `.ssh/` paths.

**Fail Message:**
`Event '[eventName]' hook [index] inner hook [innerIndex]: Accessing .ssh directory - high security risk`

**Source:** cclint-observe

---

#### Rule 069: .aws Directory Access

**Severity:** warning
**Component:** settings
**Category:** security

**Description:**
The `.aws` directory contains AWS credentials and configuration. Accessing it in hooks can expose cloud credentials. Ensure no secrets are logged or transmitted.

**Pass Criteria:**
Command does not access `.aws/` paths.

**Fail Message:**
`Event '[eventName]' hook [index] inner hook [innerIndex]: Accessing AWS config directory - ensure no secrets exposed`

**Source:** cclint-observe

---

#### Rule 070: SSH Private Key Access

**Severity:** warning
**Component:** settings
**Category:** security

**Description:**
Accessing SSH private key files (id_rsa, id_ed25519, id_dsa) in hooks is a high security risk. Private keys should never be logged, transmitted, or exposed.

**Pass Criteria:**
Command does not reference SSH private key filenames.

**Fail Message:**
`Event '[eventName]' hook [index] inner hook [innerIndex]: Accessing SSH private key - high security risk`

**Source:** cclint-observe

---

#### Rule 071: eval Command Detected

**Severity:** warning
**Component:** settings
**Category:** security

**Description:**
The `eval` command executes arbitrary strings as shell commands, creating a potential command injection vulnerability if user input or external data is involved.

**Pass Criteria:**
Command does not use the `eval` keyword.

**Fail Message:**
`Event '[eventName]' hook [index] inner hook [innerIndex]: eval command detected - potential command injection risk`

**Source:** cclint-observe

---

#### Rule 072: Command Substitution Detected

**Severity:** warning
**Component:** settings
**Category:** security

**Description:**
Command substitution using `$(...)` can execute arbitrary commands. Ensure that any substituted commands do not include unsanitized input to prevent command injection.

**Pass Criteria:**
Command does not use `$(...)` syntax, or usage is verified to be safe.

**Fail Message:**
`Event '[eventName]' hook [index] inner hook [innerIndex]: Command substitution detected - ensure input is sanitized`

**Source:** cclint-observe

---

#### Rule 073: Backtick Substitution Detected

**Severity:** warning
**Component:** settings
**Category:** security

**Description:**
Backtick command substitution (`` `command` ``) can execute arbitrary commands. This is an older shell syntax with the same risks as `$(...)`. Ensure input is sanitized.

**Pass Criteria:**
Command does not use backtick (`` ` ``) command substitution, or usage is verified to be safe.

**Fail Message:**
`Event '[eventName]' hook [index] inner hook [innerIndex]: Backtick command substitution detected - ensure input is sanitized`

**Source:** cclint-observe

---

#### Rule 074: Redirecting to /dev/

**Severity:** warning
**Component:** settings
**Category:** security

**Description:**
Redirecting output to `/dev/` paths (like `/dev/null`, `/dev/tcp/`, or device files) should be reviewed to ensure it's intentional and not masking malicious activity or causing unintended side effects.

**Pass Criteria:**
Command does not redirect to `/dev/` paths, or redirection is verified to be intentional.

**Fail Message:**
`Event '[eventName]' hook [index] inner hook [innerIndex]: Redirecting to /dev/ - verify this is intentional`

**Source:** cclint-observe

---

## Plugin Rules (075-092)

### Required Fields (Rules 075-083)

#### Rule 075: Missing 'name' field

**Severity:** error
**Component:** plugin
**Category:** structural

**Description:**
The plugin manifest must include a `name` field that identifies the plugin.

**Pass Criteria:**
- The `name` field exists
- The `name` field is not empty

**Fail Message:**
`Required field 'name' is missing or empty`

**Source:** Anthropic Docs

---

#### Rule 076: Empty 'name' field

**Severity:** error
**Component:** plugin
**Category:** structural

**Description:**
The `name` field cannot be an empty string.

**Pass Criteria:**
- The `name` field contains at least one character

**Fail Message:**
`Required field 'name' is missing or empty`

**Source:** Anthropic Docs

---

#### Rule 077: Missing 'description' field

**Severity:** error
**Component:** plugin
**Category:** structural

**Description:**
The plugin manifest must include a `description` field that explains what the plugin does.

**Pass Criteria:**
- The `description` field exists
- The `description` field is not empty

**Fail Message:**
`Required field 'description' is missing or empty`

**Source:** Anthropic Docs

---

#### Rule 078: Empty 'description' field

**Severity:** error
**Component:** plugin
**Category:** structural

**Description:**
The `description` field cannot be an empty string.

**Pass Criteria:**
- The `description` field contains at least one character

**Fail Message:**
`Required field 'description' is missing or empty`

**Source:** Anthropic Docs

---

#### Rule 079: Missing 'version' field

**Severity:** error
**Component:** plugin
**Category:** structural

**Description:**
The plugin manifest must include a `version` field following semantic versioning.

**Pass Criteria:**
- The `version` field exists
- The `version` field is not empty

**Fail Message:**
`Required field 'version' is missing or empty`

**Source:** Anthropic Docs

---

#### Rule 080: Empty 'version' field

**Severity:** error
**Component:** plugin
**Category:** structural

**Description:**
The `version` field cannot be an empty string.

**Pass Criteria:**
- The `version` field contains at least one character

**Fail Message:**
`Required field 'version' is missing or empty`

**Source:** Anthropic Docs

---

#### Rule 081: Missing 'author' field

**Severity:** error
**Component:** plugin
**Category:** structural

**Description:**
The plugin manifest must include an `author` object.

**Pass Criteria:**
- The `author` field exists as an object/map

**Fail Message:**
`Required field 'author' is missing`

**Source:** Anthropic Docs

---

#### Rule 082: Missing 'author.name' field

**Severity:** error
**Component:** plugin
**Category:** structural

**Description:**
The `author` object must include a `name` field identifying the plugin author.

**Pass Criteria:**
- The `author.name` field exists
- The `author.name` field is not empty

**Fail Message:**
`Required field 'author.name' is missing or empty`

**Source:** Anthropic Docs

---

#### Rule 083: Empty 'author.name' field

**Severity:** error
**Component:** plugin
**Category:** structural

**Description:**
The `author.name` field cannot be an empty string.

**Pass Criteria:**
- The `author.name` field contains at least one character

**Fail Message:**
`Required field 'author.name' is missing or empty`

**Source:** Anthropic Docs

---

### Constraints (Rules 084-087)

#### Rule 084: Reserved word in name

**Severity:** error
**Component:** plugin
**Category:** security

**Description:**
Plugin names cannot use reserved words that conflict with core Claude Code namespaces.

**Pass Criteria:**
- The `name` field does not contain (case-insensitive):
  - `anthropic`
  - `claude`

**Fail Message:**
`Name '[name]' is a reserved word and cannot be used`

**Source:** Anthropic Docs

---

#### Rule 085: Name exceeds 64 characters

**Severity:** error
**Component:** plugin
**Category:** structural

**Description:**
Plugin names must not exceed 64 characters to ensure compatibility and usability.

**Pass Criteria:**
- The `name` field is 64 characters or fewer

**Fail Message:**
`Name exceeds 64 character limit ([N] chars)`

**Source:** Anthropic Docs

---

#### Rule 086: Description exceeds 1024 characters

**Severity:** error
**Component:** plugin
**Category:** structural

**Description:**
Plugin descriptions must not exceed 1024 characters to ensure readability and UI compatibility.

**Pass Criteria:**
- The `description` field is 1024 characters or fewer

**Fail Message:**
`Description exceeds 1024 character limit ([N] chars)`

**Source:** Anthropic Docs

---

#### Rule 087: Invalid semver version format

**Severity:** error
**Component:** plugin
**Category:** structural

**Description:**
The `version` field must follow semantic versioning format (MAJOR.MINOR.PATCH).

**Pass Criteria:**
- The `version` matches the pattern: `^\d+\.\d+\.\d+(-[a-zA-Z0-9.]+)?(\+[a-zA-Z0-9.]+)?$`
- Examples: `1.0.0`, `2.3.4-beta.1`, `1.0.0+build.123`

**Fail Message:**
`Version '[version]' must be in semver format (e.g., 1.0.0)`

**Source:** Anthropic Docs

---

### Best Practices (Rules 088-092)

#### Rule 088: Missing 'homepage' field

**Severity:** suggestion
**Component:** plugin
**Category:** documentation

**Description:**
Adding a `homepage` field helps users discover more information about the plugin.

**Pass Criteria:**
- The `homepage` field exists with a valid URL

**Fail Message:**
`Consider adding 'homepage' field with project URL`

**Source:** cclint observe

---

#### Rule 089: Missing 'repository' field

**Severity:** suggestion
**Component:** plugin
**Category:** documentation

**Description:**
Adding a `repository` field helps users find the source code and contribute to the plugin.

**Pass Criteria:**
- The `repository` field exists with a valid repository URL

**Fail Message:**
`Consider adding 'repository' field with source code URL`

**Source:** cclint observe

---

#### Rule 090: Missing 'license' field

**Severity:** suggestion
**Component:** plugin
**Category:** documentation

**Description:**
Adding a `license` field clarifies usage rights and legal terms for the plugin.

**Pass Criteria:**
- The `license` field exists with a valid SPDX license identifier

**Fail Message:**
`Consider adding 'license' field (e.g., MIT, Apache-2.0)`

**Source:** cclint observe

---

#### Rule 091: Missing 'keywords' array

**Severity:** suggestion
**Component:** plugin
**Category:** documentation

**Description:**
Adding a `keywords` array improves plugin discoverability in search and catalogs.

**Pass Criteria:**
- The `keywords` field exists as a non-empty array

**Fail Message:**
`Consider adding 'keywords' array for discoverability`

**Source:** cclint observe

---

#### Rule 092: Description too short

**Severity:** suggestion
**Component:** plugin
**Category:** best-practice

**Description:**
Plugin descriptions should be at least 50 characters to provide adequate context for users.

**Pass Criteria:**
- The `description` field is 50 characters or longer

**Fail Message:**
`Description is only [N] chars - consider expanding for clarity`

**Source:** cclint observe

---

## Security Rules (093-104)

### Tool Validation

#### Rule 093: Unknown tool in allowed-tools

**Severity:** warning
**Component:** all
**Category:** security

**Description:**
Validates that tool names in `allowed-tools` and `tools` frontmatter fields are recognized Claude Code tools. Prevents typos and ensures proper tool restrictions.

**Pass Criteria:**
- Tool names must match known tools: Read, Write, Edit, MultiEdit, Glob, Grep, LS, Bash, Task, WebFetch, WebSearch, AskUserQuestion, TodoWrite, Skill, LSP, NotebookEdit, EnterPlanMode, ExitPlanMode, KillShell, TaskOutput, or `*`
- Patterns like `Task(specialist-name)` and `Bash(npm:*)` are validated by base tool name
- Empty or whitespace-only tool names are ignored

**Fail Message:**
`Unknown tool '[tool-name]' in [field-name]. Check spelling or verify it's a valid tool.`

---

### Secrets Detection

#### Rule 094: Hardcoded API key pattern

**Severity:** warning
**Component:** all
**Category:** security

**Description:**
Detects potential hardcoded API keys using common assignment patterns with key/value lengths typical of API credentials.

**Pass Criteria:**
- No matches for pattern: `(?i)(api[_-]?key|apikey)\s*[:=]\s*["'][^"']{10,}["']`
- API keys should be loaded from environment variables or secrets management

**Fail Message:**
`Possible hardcoded API key detected - use environment variables`

---

#### Rule 095: Hardcoded password pattern

**Severity:** warning
**Component:** all
**Category:** security

**Description:**
Detects potential hardcoded passwords in variable assignments.

**Pass Criteria:**
- No matches for pattern: `(?i)(password|passwd|pwd)\s*[:=]\s*["'][^"']+["']`
- Passwords should be loaded from environment variables or secrets management

**Fail Message:**
`Possible hardcoded password detected - use secrets management`

---

#### Rule 096: Hardcoded secret/token pattern

**Severity:** warning
**Component:** all
**Category:** security

**Description:**
Detects potential hardcoded secrets or tokens with sufficient length to be real credentials.

**Pass Criteria:**
- No matches for pattern: `(?i)(secret|token)\s*[:=]\s*["'][^"']{10,}["']`
- Secrets should be loaded from environment variables or secrets management

**Fail Message:**
`Possible hardcoded secret/token detected - use environment variables`

---

#### Rule 097: OpenAI API key pattern (sk-)

**Severity:** warning
**Component:** all
**Category:** security

**Description:**
Detects OpenAI API keys by their distinctive `sk-` prefix format.

**Pass Criteria:**
- No matches for pattern: `sk-[a-zA-Z0-9]{20,}`
- Never commit API keys to version control

**Fail Message:**
`OpenAI API key pattern detected - never commit API keys`

---

#### Rule 098: Slack bot token pattern (xoxb-)

**Severity:** warning
**Component:** all
**Category:** security

**Description:**
Detects Slack bot tokens by their distinctive `xoxb-` prefix format.

**Pass Criteria:**
- No matches for pattern: `xoxb-[a-zA-Z0-9-]+`
- Use environment variables for Slack tokens

**Fail Message:**
`Slack bot token pattern detected - use environment variables`

---

#### Rule 099: GitHub PAT pattern (ghp_)

**Severity:** warning
**Component:** all
**Category:** security

**Description:**
Detects GitHub personal access tokens by their `ghp_` prefix and 36-character format.

**Pass Criteria:**
- No matches for pattern: `ghp_[a-zA-Z0-9]{36}`
- Never commit GitHub tokens to version control

**Fail Message:**
`GitHub personal access token pattern detected`

---

#### Rule 100: Google API key pattern (AIza)

**Severity:** warning
**Component:** all
**Category:** security

**Description:**
Detects Google API keys by their distinctive `AIza` prefix and length format.

**Pass Criteria:**
- No matches for pattern: `AIza[0-9A-Za-z\-_]{35}`
- Use environment variables for Google API keys

**Fail Message:**
`Google API key pattern detected - use environment variables`

---

#### Rule 101: Google OAuth client ID pattern

**Severity:** warning
**Component:** all
**Category:** security

**Description:**
Detects Google OAuth client IDs by their distinctive format ending in `.apps.googleusercontent.com`.

**Pass Criteria:**
- No matches for pattern: `[0-9]+-[0-9A-Za-z_]{32}\.apps\.googleusercontent\.com`
- While less sensitive than secrets, client IDs should typically be in configuration

**Fail Message:**
`Google OAuth client ID pattern detected`

---

#### Rule 102: Private key detected (-----BEGIN)

**Severity:** warning
**Component:** all
**Category:** security

**Description:**
Detects private cryptographic keys in PEM format (RSA, DSA, or generic private keys).

**Pass Criteria:**
- No matches for pattern: `-----BEGIN (RSA |DSA )?PRIVATE KEY-----`
- Never commit private keys to version control

**Fail Message:**
`Private key detected - never commit private keys`

---

#### Rule 103: AWS access key ID pattern

**Severity:** warning
**Component:** all
**Category:** security

**Description:**
Detects AWS access key IDs by their assignment pattern and 20-character uppercase alphanumeric format.

**Pass Criteria:**
- No matches for pattern: `aws_access_key_id\s*[:=]\s*["']?[A-Z0-9]{20}["']?`
- Use environment variables or AWS IAM roles for credentials

**Fail Message:**
`AWS access key ID detected - use environment variables`

---

#### Rule 104: AWS secret access key pattern

**Severity:** warning
**Component:** all
**Category:** security

**Description:**
Detects AWS secret access keys by their assignment pattern and 40-character base64 format.

**Pass Criteria:**
- No matches for pattern: `aws_secret_access_key\s*[:=]\s*["']?[A-Za-z0-9/+=]{40}["']?`
- Use environment variables or AWS IAM roles for credentials

**Fail Message:**
`AWS secret access key detected - use environment variables`

---

## Schema Constraints (105-124)

### Agent Schema Constraints (Rules 105-112)

#### Rule 105: Agent Name Pattern

**Severity:** error
**Component:** agent
**Category:** schema

**Description:**
Agent names must contain only lowercase letters, numbers, and hyphens. No uppercase letters, underscores, or special characters are permitted.

**Constraint:**
`name: string & =~("^[a-z0-9-]+$")`

**Valid Values:**
- `memory-specialist`
- `error-chain-1`
- `web3-agent`

**Invalid Values:**
- `Memory-Specialist` (uppercase)
- `error_chain` (underscore)
- `web3.agent` (period)
- `agent@service` (special char)

---

#### Rule 106: Agent Name Max Length

**Severity:** error
**Component:** agent
**Category:** schema

**Description:**
Agent names are limited to 64 characters to ensure readability and compatibility with filesystem paths.

**Constraint:**
`name: string & strings.MaxRunes(64)`

**Valid Values:**
Maximum 64 Unicode characters (runes).

---

#### Rule 107: Agent Description Required

**Severity:** error
**Component:** agent
**Category:** schema

**Description:**
Every agent must have a non-empty description field in its frontmatter. This description is used for documentation and agent discovery.

**Constraint:**
`description: string & !=""`

**Valid Values:**
Any non-empty string.

---

#### Rule 108: Agent Description Max Length

**Severity:** error
**Component:** agent
**Category:** schema

**Description:**
Agent descriptions are limited to 1024 characters to prevent excessive frontmatter bloat.

**Constraint:**
`description: string & strings.MaxRunes(1024)`

**Valid Values:**
Maximum 1024 Unicode characters (runes).

---

#### Rule 109: Agent Model Enum Values

**Severity:** error
**Component:** agent
**Category:** schema

**Description:**
The model field must be one of the predefined Claude Code model identifiers. Custom model names are not permitted.

**Constraint:**
`model?: "sonnet" | "opus" | "haiku" | "sonnet[1m]" | "opusplan" | "inherit"`

**Valid Values:**
- `sonnet` - Claude Sonnet 3.5
- `opus` - Claude Opus 3.5
- `haiku` - Claude Haiku 3.5
- `sonnet[1m]` - Sonnet with 1 million token context
- `opusplan` - Opus with planning mode
- `inherit` - Inherit from parent/settings

---

#### Rule 110: Agent Color Enum Values

**Severity:** error
**Component:** agent
**Category:** schema

**Description:**
Agent color must be one of the 8 standard Claude Code colors for visual identification.

**Constraint:**
`color?: "red" | "blue" | "green" | "yellow" | "purple" | "orange" | "pink" | "cyan"`

**Valid Values:**
- `red`
- `blue`
- `green`
- `yellow`
- `purple`
- `orange`
- `pink`
- `cyan`

---

#### Rule 111: Agent Tools Field Variants

**Severity:** error
**Component:** agent
**Category:** schema

**Description:**
The tools (or allowed-tools) field can be specified in three formats: wildcard for all tools, comma-separated string, or array of specific tool names.

**Constraint:**
`tools?: "*" | string | [...#KnownTool]`
`"allowed-tools"?: "*" | string | [...#KnownTool]`

**Valid Values:**
- `"*"` - All tools allowed
- `"Read,Write,Edit"` - Comma-separated string
- `["Read", "Write", "Edit"]` - Array of tool names

---

#### Rule 112: Known Tools Whitelist

**Severity:** error
**Component:** agent
**Category:** schema

**Description:**
When tools are specified as an array, each tool name must be from the known Claude Code tools whitelist.

**Constraint:**
```
#KnownTool: "Read" | "Write" | "Edit" | "MultiEdit" | "Bash" |
	"Grep" | "Glob" | "LS" | "Task" | "NotebookEdit" |
	"WebFetch" | "WebSearch" | "TodoWrite" | "BashOutput" |
	"KillBash" | "ExitPlanMode" | "AskUserQuestion" |
	"LSP" | "Skill" | "DBClient"
```

**Valid Values:**
Read, Write, Edit, MultiEdit, Bash, Grep, Glob, LS, Task, NotebookEdit, WebFetch, WebSearch, TodoWrite, BashOutput, KillBash, ExitPlanMode, AskUserQuestion, LSP, Skill, DBClient

---

### Command Schema Constraints (Rules 113-117)

#### Rule 113: Command Name Pattern

**Severity:** error
**Component:** command
**Category:** schema

**Description:**
Command names must contain only lowercase letters, numbers, and hyphens. Same pattern as agent names.

**Constraint:**
`name?: string & =~("^[a-z0-9-]+$")`

**Valid Values:**
- `refactor`
- `test-runner`
- `db-migrate`

**Invalid Values:**
- `Refactor` (uppercase)
- `test_runner` (underscore)
- `db.migrate` (period)

---

#### Rule 114: Command Name Max Length

**Severity:** error
**Component:** command
**Category:** schema

**Description:**
Command names are limited to 64 characters for consistency with agent names.

**Constraint:**
`name?: string & strings.MaxRunes(64)`

**Valid Values:**
Maximum 64 Unicode characters (runes).

---

#### Rule 115: Command Description Max Length

**Severity:** error
**Component:** command
**Category:** schema

**Description:**
Command descriptions are limited to 1024 characters. Unlike agents, command descriptions are optional.

**Constraint:**
`description?: string & strings.MaxRunes(1024)`

**Valid Values:**
Maximum 1024 Unicode characters (runes), or omitted entirely.

---

#### Rule 116: Command Allowed-Tools Variants

**Severity:** error
**Component:** command
**Category:** schema

**Description:**
The allowed-tools field can be specified as wildcard, comma-separated string, or array of specific tool names.

**Constraint:**
`"allowed-tools"?: "*" | string | [...#KnownTool]`

**Valid Values:**
- `"*"` - All tools allowed
- `"Task,TodoWrite"` - Comma-separated string
- `["Task", "TodoWrite"]` - Array of tool names

---

#### Rule 117: Command Model Enum

**Severity:** error
**Component:** command
**Category:** schema

**Description:**
Command model field must be one of the predefined Claude Code model identifiers. Same values as agent model.

**Constraint:**
`model?: "sonnet" | "opus" | "haiku" | "sonnet[1m]" | "opusplan" | "inherit"`

**Valid Values:**
- `sonnet`
- `opus`
- `haiku`
- `sonnet[1m]`
- `opusplan`
- `inherit`

---

### Settings Schema Constraints (Rules 118-121)

#### Rule 118: Hook Event Structure

**Severity:** error
**Component:** settings
**Category:** schema

**Description:**
The hooks configuration must be structured as a map of event names to arrays of hook definitions. Each event can have multiple hooks.

**Constraint:**
```
hooks?: {
	[string]: [...#Hook]
}
```

**Valid Values:**
```json
{
  "hooks": {
    "PreToolUse": [
      { "matcher": "...", "hooks": [...] }
    ],
    "PostToolUse": [
      { "matcher": "...", "hooks": [...] }
    ]
  }
}
```

---

#### Rule 119: Hook Matcher Required

**Severity:** error
**Component:** settings
**Category:** schema

**Description:**
Every hook definition must include a matcher field that determines when the hook triggers.

**Constraint:**
```
#Hook: {
	matcher: #Matcher
	hooks: [...#HookCommand]
}
```

**Valid Values:**
Any string pattern that matches hook trigger conditions (glob patterns, regex, etc.).

---

#### Rule 120: Hook Type Enum

**Severity:** error
**Component:** settings
**Category:** schema

**Description:**
Hook commands must specify type as "command". This is currently the only supported hook type.

**Constraint:**
```
#HookCommand: {
	type: "command"
	...
}
```

**Valid Values:**
- `command` (only valid value)

---

#### Rule 121: Hook Command Field Required

**Severity:** error
**Component:** settings
**Category:** schema

**Description:**
Hook command definitions must include a command field specifying what to execute.

**Constraint:**
```
#HookCommand: {
	type:    "command"
	command: string
	...
}
```

**Valid Values:**
Any non-empty string representing a shell command.

---

### CLAUDE.md Schema Constraints (Rules 122-124)

#### Rule 122: Section Structure

**Severity:** error
**Component:** claude_md
**Category:** schema

**Description:**
CLAUDE.md sections must include heading and content fields. References are optional.

**Constraint:**
```
#Section: {
	heading: string
	content: string
	references?: [...#RefLink]
}
```

**Valid Values:**
- `heading` - Section heading text (required)
- `content` - Section content/body (required)
- `references` - Array of reference links (optional)

---

#### Rule 123: Rule Name/Description Required

**Severity:** error
**Component:** claude_md
**Category:** schema

**Description:**
CLAUDE.md rule definitions must include both name and description fields.

**Constraint:**
```
#Rule: {
	name:        string
	description: string
}
```

**Valid Values:**
Both `name` and `description` must be non-empty strings.

---

#### Rule 124: Reference Path Required

**Severity:** error
**Component:** claude_md
**Category:** schema

**Description:**
Reference links in CLAUDE.md sections must include a path field. The title is optional.

**Constraint:**
```
#RefLink: {
	path:   string
	title?: string
}
```

**Valid Values:**
- `path` - File path or URL (required)
- `title` - Display title (optional)

---

## Summary Statistics

| Component | Rule Range | Count | Error | Warning | Suggestion |
|-----------|------------|-------|-------|---------|------------|
| Agent | 001-021 | 21 | 7 | 0 | 14 |
| Command | 022-034 | 13 | 2 | 0 | 11 |
| Skill | 035-047 | 13 | 4 | 1 | 8 |
| Settings | 048-074 | 27 | 17 | 10 | 0 |
| Plugin | 075-092 | 18 | 13 | 0 | 5 |
| Security | 093-104 | 12 | 0 | 12 | 0 |
| Schema | 105-124 | 20 | 20 | 0 | 0 |
| **Total** | **001-124** | **124** | **63** | **23** | **38** |

## Source Documentation URLs

| Component | Source URL |
|-----------|------------|
| Agents | https://code.claude.com/docs/en/sub-agents |
| Commands | https://code.claude.com/docs/en/slash-commands |
| Skills | https://code.claude.com/docs/en/skills |
| Hooks | https://code.claude.com/docs/en/hooks |
| Plugins | https://code.claude.com/docs/en/plugins-reference |
| Security | OWASP, CWE, Platform-specific standards |

---

## Related Documentation

- [Agent Rules](rules/agents.md)
- [Command Rules](rules/commands.md)
- [Skill Rules](rules/skills.md)
- [Settings Rules](rules/settings.md)
- [Plugin Rules](rules/plugins.md)
- [Security Rules](rules/security.md)
- [Schema Constraints](rules/schema-constraints.md)
- [Quality Scoring](scoring.md)
- [Cross-File Validation](cross-file-validation.md)
