# Cross-File Validation

How cclint detects and validates references between components.

---

## Overview

cclint performs cross-file validation to ensure components reference each other correctly. This includes:

1. **Skill Reference Validation** - Agents referencing skills that don't exist
2. **Orphan Detection** - Skills not referenced by any agent/command
3. **Tool Validation** - Commands using tools they haven't declared

## Skill Reference Detection

### Patterns Recognized

cclint uses multiple regex patterns to detect skill references in agent files:

| Pattern | Example | Description |
|---------|---------|-------------|
| Plain | `Skill: foo-bar` | Standard skill reference |
| Bold | `**Skill**: foo-bar` | Bold-formatted reference |
| Function | `Skill(foo-bar)` | Tool-call style |
| List | `Skills:\n  - foo-bar` | List format |

### Code Block Support

Skill references inside markdown code blocks are fully supported:

```markdown
## Foundation

**MANDATORY: Load skills:**
```
Skill: first-skill     # Detected ✓
Skill: second-skill    # Detected ✓
Skill: third-skill     # Detected ✓
```
```

All three skills are detected and validated.

### Technical Details

The detection uses this regex pattern:

```go
regexp.MustCompile(`(?m)^[^*\n]*\bSkill:\s*([a-z0-9][a-z0-9-]*)`)
```

Key components:
- `(?m)` - Multiline mode, `^` matches start of each line
- `[^*\n]*` - Match any chars except `*` (bold markers) or `\n` (newlines)
- `\bSkill:` - Word boundary followed by "Skill:"
- `([a-z0-9][a-z0-9-]*)` - Capture skill name (lowercase, hyphens)

**Important:** The `\n` exclusion is critical. Without it, Go's regex engine would greedily match across newlines, causing only the last skill in a block to be detected.

## Orphan Detection

### What It Checks

An orphaned skill is one with no incoming references from:
- Commands (via `Task(X-specialist)` delegation)
- Agents (via `Skill:` declarations)
- Other skills (via cross-references)

### Output

Orphaned skills appear as suggestions in verbose mode:

```bash
$ cclint skills --verbose

💡 skills/unused-skill/SKILL.md
    💡 Skill 'unused-skill' has no incoming references - consider adding crossrefs
```

### Severity

Orphan detection produces `info`-level suggestions (not errors). Skills may legitimately be:
- New and not yet integrated
- Used for reference/documentation only
- Invoked dynamically by name

## Cross-Reference Tracking

### How It Works

1. **Discovery Phase** - Find all agents, commands, and skills
2. **Reference Extraction** - Parse each file for skill references
3. **Graph Building** - Build a map of which skills are referenced
4. **Orphan Detection** - Find skills with zero incoming edges

### Data Structures

```go
type CrossFileValidator struct {
    agents   map[string]File  // agent-name -> file info
    skills   map[string]File  // skill-name -> file info
    commands map[string]File  // command-name -> file info
}
```

### Validation Flow

```
Agent File
    │
    ├─→ findSkillReferences(content)
    │   └─→ Returns []string of skill names
    │
    └─→ For each skill reference:
        └─→ Check if skill exists in skills map
            ├─→ Yes: Mark skill as referenced
            └─→ No: Error: "Skill: X references non-existent skill"
```

## Common Issues

### False Positives (Fixed in v1.1.0)

**Issue:** Only detecting one skill per file when multiple are declared.

**Cause:** The regex `[^*]*` matched newlines in Go, causing greedy matching across lines.

**Fix:** Changed to `[^*\n]*` to prevent matching across newlines.

**Before:** 100 "orphaned" skills (mostly false positives)
**After:** 19 genuinely unwired skills

### Skills in Comments

Skill references in markdown comments are NOT detected:

```markdown
<!-- Skill: hidden-skill -->  ← Not detected (inside comment)
```

This is intentional - commented references shouldn't count.

## Best Practices

### Wire Skills to Agents

Every skill should be referenced by at least one agent:

```markdown
## Foundation

**MANDATORY: Load skills based on context:**
```
Skill: primary-skill      # Always loaded
Skill: secondary-skill    # When applicable
```
```

### Use Consistent Formatting

Prefer the plain `Skill: name` format for clarity:

```markdown
# Good
Skill: my-skill

# Also good (in code blocks)
```
Skill: my-skill
```

# Avoid (harder to read)
**Skill**: my-skill
Skill("my-skill")
```

### Check for Orphans Regularly

Run verbose mode periodically to find unwired skills:

```bash
cclint skills --verbose | grep "no incoming references"
```

## Related Documentation

- [Agent Lint Rules](rules/agents.md) - Agent validation rules
- [Skill Lint Rules](rules/skills.md) - Skill validation rules
- [Quality Scoring](scoring/README.md) - Component scoring system
