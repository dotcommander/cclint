# cclint

> Opinionated linter for Claude Code agents, commands, and skills.

Like ESLint but for your Claude Code config. Catches structural issues, validates schemas, and yells at you about best practices.

---

## Install

```bash
git clone https://github.com/dotcommander/cclint.git && cd cclint
go build -o cclint .
ln -sf $(pwd)/cclint ~/go/bin/cclint
```

Then run from anywhere:
```bash
cclint agents
```

---

## What It Checks

| Thing | What It Yells About |
|-------|---------------------|
| **Agents** | Missing fields, bad names, too long (>200 lines), missing sections, no Skill() loading |
| **Commands** | Bad names, too long (>50 lines), missing Usage/Workflow, no semantic routing tables |
| **Skills** | Wrong filename, empty files, no frontmatter, no Quick Reference table, too long (>500 lines) |

---

## Usage Examples

```bash
# Check all agents
cclint agents

# Check just commands
cclint commands

# Check just skills
cclint skills

# Check everything (default)
cclint

# From a specific directory
cclint --root ~/my-project agents

# JSON output for CI/CD
cclint --format json --output report.json agents

# Quiet mode (errors only, no suggestions)
cclint --quiet commands

# Verbose mode (natter on about everything)
cclint --verbose skills
```

---

## What The Output Looks Like

```
💡 agents/architecture-specialist.md
    💡 Agent is 251 lines. Best practice: keep agents under 200 lines - move methodology to skills instead.
    💡 Agent has 'triggers' but no 'proactive_triggers'. Consider adding proactive trigger phrases.
    💡 Agent lacks 'context_isolation: true'. Consider adding for cleaner context management.

💡 commands/commit.md
    💡 Command is 129 lines. Best practice: keep commands under 50 lines - delegate to specialist agents.
    💡 Consider adding a semantic routing table for discoverability (| User Question | Action |).

70/70 passed, 0 errors, 308 suggestions (0s)
```

💡 = Suggestion (won't fail lint)
✗ = Error (will fail lint)

---

## Opinionated Best Practices

This tool enforces patterns from real-world Claude Code setups:

### Commands should be tiny (<50 lines)

```markdown
---
allowed-tools: Task(specialist)
---
Delegate to specialist for the actual work.
Task(specialist): $ARGUMENTS
```

If your command is >50 lines, you're doing it wrong. Extract methodology to a skill.

### Agents should load skills

```markdown
---
model: sonnet
triggers: ["performance", "slow"]
proactive_triggers: ["investigate slowness", "optimize"]
---
## Foundation
Skill(performance-patterns)

## Workflow
Phase 1: Profile → Phase 2: Analyze → Phase 3: Optimize
```

Don't embed methodology in agents. Put it in skills, load with `Skill()`.

### Skills need routing tables

```markdown
## Quick Reference

| User Question | Action |
|---------------|--------|
| "How do I cache?" | Read(references/caching.md) |
| "What's the cache strategy?" | Read(references/strategy.md) |
```

This makes skills discoverable. Without it, nobody knows what your skill does.

---

## Make It Your Own

This linter is opinionated—but they're *my* opinions. Fork it and make it yours.

### Customize the rules

The validation schemas live in `internal/cue/schemas/`:

```
internal/cue/schemas/
├── agent.cue      # Agent frontmatter schema
├── command.cue    # Command frontmatter schema
├── settings.cue   # Settings validation
└── claude_md.cue  # CLAUDE.md structure
```

Edit these CUE files to match your team's conventions:

```cue
// Example: Change allowed colors in agent.cue
#Color: "red" | "blue" | "green"  // Your brand colors only

// Example: Require specific fields
#Agent: {
    name: string
    description: string
    owner: string        // Add required fields
    team?: string        // Add optional fields
    ...
}
```

### Adjust size limits

In `internal/cli/agents.go`, `commands.go`, and `skills.go`, find the line limits:

```go
if lines > 200 {  // Change to your preference
```

### Add new checks

The validation functions in `internal/cli/*.go` are straightforward Go. Add checks that matter to your team, remove ones that don't.

After changes:
```bash
go build -o cclint .
```

---

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | All good (suggestions don't count) |
| 1 | Errors found |

---

## Contributing

Found a false positive? Missing check? Open an issue or PR.

---

## License

MIT
