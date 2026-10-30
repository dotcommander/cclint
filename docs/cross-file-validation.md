# Cross-File Validation

How cclint detects and validates references between components.

---

## Overview

cclint performs cross-file validation to ensure components reference each other correctly. This includes:

1. **Skill Reference Validation** - Agents referencing skills that don't exist
2. **Circular Dependency Detection** - Cycles in agent → skill → agent delegation chains
3. **Orphan Detection** - Skills not referenced by any agent/command
4. **Tool Validation** - Commands using tools they haven't declared
5. **Ghost Trigger Detection** - Trigger mapping tables routing to skills or agents that don't exist
6. **Skill Reference Files** - Missing `references/*.md` targets and unmentioned reference files

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

Skill references inside markdown code blocks are also detected (the extractor does not exclude fenced blocks):

~~~markdown
## Foundation

**MANDATORY: Load skills:**
```
Skill: first-skill     # Detected ✓
Skill: second-skill    # Detected ✓
Skill: third-skill     # Detected ✓
```
~~~

All three skills are detected and validated.

### Technical Details

The detection uses this regex pattern:

```go
regexp.MustCompile(`(?m)^[^*\n]*\bSkill:[ \t]*(` + skillRefNamePattern + `)`) // plain form
```

Key components:
- `(?m)` - Multiline mode, `^` matches start of each line
- `[^*\n]*` - Match any chars except `*` (bold markers) or `\n` (newlines)
- `\bSkill:` - Word boundary followed by "Skill:"
- `skillRefNamePattern` - Capture a local or `plugin:skill`-qualified name; other regexes handle bold, function and list forms

**Important:** The `\n` exclusion is critical. Without it, Go's regex engine would greedily match across newlines, causing only the last skill in a block to be detected.

## Circular Dependency Detection

### What It Checks

Circular dependencies occur when components form a loop in their reference chain:

```
agent-a → skill-b → agent-c → skill-d → agent-a  (cycle!)
```

These cycles indicate potentially recursive delegation or loading paths; the checker reports graph references rather than executing them.

### How It Works

cclint uses DFS (Depth-First Search) with color marking to detect back edges in the component graph:

1. **White (unvisited)** - Node not yet processed
2. **Gray (in progress)** - Node is in current DFS path
3. **Black (complete)** - Node fully processed

A cycle is detected when we encounter a gray node while exploring.

### Output

Cycles are reported as errors on all involved agents:

```bash
$ cclint agents

✗ agents/agent-a.md
    ✘ Circular dependency detected: agent:agent-a → skill:shared-skill → agent:agent-a

✗ agents/agent-c.md
    ✘ Circular dependency detected: agent:agent-a → skill:shared-skill → agent:agent-a
```

### Disabling

To disable circular dependency detection (not recommended):

```bash
cclint --no-cycle-check agents
```

## Orphan Detection

### What It Checks

An orphaned skill is one with no incoming references from:
- Commands (via `Task(X-specialist)` delegation)
- Agents (via `Skill:` declarations)
- Other skills (via name mentions)
- Skill trigger routing tables in `references/*.md`

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

## Ghost Trigger Detection

### What It Checks

Trigger mapping tables in `skills/*/references/*.md` files route user input to specific skills or agents. Ghost triggers are entries in those tables that reference a skill or agent that doesn't exist — they route to nothing.

cclint validates two table formats:

**Skill routing tables** (e.g., in `triggers.md` or `routing.md`):

```markdown
| Trigger | Skill |
|---------|-------|
| "analyze code" | code-analysis |
| "write tests"  | code-testing-qa |
```

**Agent routing tables** (e.g., in agent delegation references):

```markdown
| Trigger Pattern | Route To |
|-----------------|----------|
| "deploy"        | deploy-agent |
| "review PR"     | review-agent |
```

### Heuristic

Only files that contain a `| Trigger` column header are scanned. This prevents false positives in files that happen to use similar table structures for other purposes.

For bare skill names (no path separator), a hyphen guard is applied: single-word tokens without a hyphen are skipped, since they are more likely to be prose labels than skill identifiers.

### Output

Ghost triggers are reported as errors on the reference file that contains them:

```bash
$ cclint skills --verbose

✗ skills/agent-claude-code/references/triggers.md
    ✘ Trigger map references non-existent skill 'old-skill'. Create skills/old-skill/SKILL.md
    ✘ Trigger map references non-existent agent 'retired-agent'. Create agents/retired-agent.md
```

### Severity

Ghost trigger detection produces `error`-level findings (unlike orphans which are `info`). A ghost trigger is an active misconfiguration: any agent or user that follows the routing table will be silently misdirected.

### Built-in Exclusions

Built-in agent types and model names are excluded from ghost trigger validation. The same exclusion list used for agent reference validation applies here (e.g., `haiku`, `sonnet`, `opus`, `Explore`, `Plan`).

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

## Agent Reference Detection (Skills)

Skills can reference agents using multiple patterns. All are validated:

| Pattern | Example | Validated |
|---------|---------|-----------|
| Task() specialist | `Task(foo-specialist)` | Yes |
| Task() generic | `Task(agent-name)` | Yes |
| Task() quoted | `Task("agent-name")` | Yes |
| Delegate via | `delegate via agent-name` | Yes |
| Agent handles | `foo-agent handles...` | Yes |
| Use/See specialist | `use foo-specialist` | Yes |

### Built-in Agents

Built-in agents and model names are automatically excluded from validation:

**Subagent Types:**
- `Explore` - File exploration subagent
- `Plan` - Planning subagent
- `general-purpose` - General purpose subagent
- `claude-code-guide` - Claude Code guide
- `statusline-setup` - Statusline setup

**Model Names:**
- `haiku` - Claude Haiku model
- `sonnet` - Claude Sonnet model
- `opus` - Claude Opus model

### Dynamic References

Dynamic/variable references are skipped:

```markdown
Task(args.agent)        # Skipped - contains dot
Task(subagent_type)     # Skipped - dynamic type
Task(items[0])          # Skipped - contains bracket
```

## Skill Reference Detection (All Components)

Both agents and commands can reference skills. All patterns are validated:

| Pattern | Example | Validated In |
|---------|---------|--------------|
| Colon | `Skill: foo-bar` | Agents, Commands |
| Function | `Skill(foo-bar)` | Agents, Commands |
| Quoted | `Skill("foo-bar")` | Agents, Commands |
| List | `Skills:\n- foo-bar` | Agents, Commands |

### Error Messages

When a non-existent skill is referenced:

```bash
$ cclint agents
✗ agents/my-agent.md
    ✘ Skill: missing-skill references non-existent skill. Create skills/missing-skill/SKILL.md

$ cclint commands
✗ commands/my-command.md
    ✘ References non-existent skill 'missing-skill'. Create skills/missing-skill/SKILL.md
```

## Notes

### Skills in Comments

An inline comment prefixed with `<!--` does not match the plain `Skill:` regex:

```markdown
<!-- Skill: hidden-skill -->  ← Not detected (inside comment)
```

Do not rely on Markdown comments to hide references: extraction is regex-based, and a `Skill:` line inside a multiline comment can still match. Remove stale references rather than commenting them out.

## Best Practices

### Wire Skills to Agents

To avoid an orphan suggestion, reference the skill from an agent, command, another skill, or trigger map. For example, in an agent:

~~~markdown
## Foundation

**MANDATORY: Load skills based on context:**
```
Skill: primary-skill      # Always loaded
Skill: secondary-skill    # When applicable
```
~~~

### Use Consistent Formatting

Prefer the plain `Skill: name` format for clarity:

```markdown
# Good
Skill: my-skill

# Also good (in code blocks)
~~~text
Skill: my-skill
~~~

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
