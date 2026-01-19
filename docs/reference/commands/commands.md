# cclint commands

Lint and validate command definition files.

---

## Usage

```bash
cclint commands [flags]
```

---

## Description

The `commands` subcommand scans for and validates command definition files in
your Claude Code project. Commands are the thin orchestration layer that route
user invocations to appropriate specialist agents.

### Supported File Patterns

- `.claude/commands/**/*.md`
- `commands/**/*.md`

### What Gets Validated

- Frontmatter structure and required fields
- Name format (lowercase alphanumeric with hyphens)
- Line count limit (50 lines ±10% tolerance)
- Delegation pattern compliance
- Bloat section detection

---

## Line Limit Rule

Commands must stay under **~55 lines** (50 lines ±10% tolerance).

### Rationale

Commands follow the **thin delegation pattern**: they should route to specialist
agents rather than implementing logic directly. If a command exceeds 55 lines,
it indicates that methodology should be moved to an agent or skill.

### Example Violation Message

```text
my-command.md:60
suggestion: Command is 60 lines. Best practice: keep commands under ~55
lines (50±10%) - delegate to specialist agents instead of implementing
logic directly.
```

### What To Do Instead

Extract implementation logic to a specialist agent, then delegate via `Task()`:

```yaml
---
name: my-command
description: Does complex work
allowed-tools: Task
---

Delegate to specialist agent for implementation.
```

---

## Delegation Pattern Requirement

Commands should delegate work to specialist agents rather than implementing
logic inline.

### Thin Command Pattern

A "thin" command uses `Task()` to delegate to an agent:

```markdown
---
name: my-command
description: Example thin command
allowed-tools: Task
---

Use the specialist-agent for this task.
```

### Fat Command Anti-Pattern

A "fat" command contains implementation sections:

- `## Implementation` steps
- `### Steps` sections
- Inline code examples without delegation

These sections indicate the command should delegate to an agent instead.

### Validation

- Rule 025 detects implementation sections
- Rule 026 checks for `allowed-tools: Task` when using `Task()` delegation
- Rule 027-031 detect bloat sections in thin commands

---

## Example Output

### Passing Command

```text
✓ my-command.md (42 lines)
  Score: 85/100 (B)
```

### Failing Command

```text
✗ complex-command.md
  errors:
    - Line 15: Name must be lowercase alphanumeric with hyphens only
    - Line 1: Required field 'description' is missing or empty

  suggestions:
    - Line 60: Command is 60 lines. Best practice: keep commands under
      ~55 lines (50±10%) - delegate to specialist agents instead of
      implementing logic directly.
```

---

## See Also

- [Agents Reference](agents.md)
- [CLI Reference](../cli.md)
- [Global Flags](../flags.md)
- [Command Writing Guide](../../guides/writing-commands.md)
- [Command Lint Rules](../../../rules/commands.md)
