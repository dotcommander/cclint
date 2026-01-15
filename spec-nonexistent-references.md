# Spec: Validate Non-Existent Agent and Skill References

## TL;DR

| Issue | Root Cause | File:Line | Fix |
|-------|------------|-----------|-----|
| Skills referenced via `Skill()` calls aren't validated | `ValidateAgent()` only checks `Skill:` pattern, not `Skill()` | `internal/cli/crossfile.go:228` | Add `Skill()` pattern to `findSkillReferences()` |
| Agent references in skills aren't exhaustive | `ValidateSkill()` only checks 4 patterns | `internal/cli/crossfile.go:247-283` | Add missing patterns: `Task()` without suffix filter, `delegate via`, narrative mentions |
| Commands check agent refs but not skill refs | `ValidateCommand()` only validates Task() → agents | `internal/cli/crossfile.go:60-104` | Add skill reference validation to commands |
| No validation for direct skill invocation | Skills can be called like `Skill(foo-bar)` in any component | N/A | Add pattern to all component validators |

---

## Priority Matrix

| Priority | Task | Complexity | Impact |
|----------|------|------------|--------|
| P0 | Extend `findSkillReferences()` to detect `Skill()` calls | Low | Fixes false negatives in agent validation |
| P0 | Add skill reference validation to `ValidateCommand()` | Low | Detects broken skill refs in commands |
| P1 | Add comprehensive agent patterns to `ValidateSkill()` | Medium | Catches all agent reference styles |
| P2 | Test coverage for new patterns | Low | Prevents regressions |

---

## Background

### Current State

Cross-file validation exists in `/Users/vampire/go/src/cclint/internal/cli/crossfile.go` with:

1. **Skill Reference Detection** (`findSkillReferences()` lines 8-41)
   - ✅ Detects: `Skill: name`, `**Skill**: name`, `Skills:\n- name`
   - ❌ Missing: `Skill("name")`, `Skill(name)`

2. **Agent Reference Validation** (`ValidateAgent()` lines 224-241)
   - ✅ Uses `findSkillReferences()` to check skills exist
   - ❌ Misses `Skill()` function call pattern

3. **Command Validation** (`ValidateCommand()` lines 60-104)
   - ✅ Validates `Task(agent-name)` → agent exists
   - ❌ No skill reference validation

4. **Skill Validation** (`ValidateSkill()` lines 243-283)
   - ✅ Checks 4 agent patterns
   - ❌ Incomplete pattern coverage

### Gap Analysis

**Pattern Coverage Matrix:**

| Reference Type | Agent Validator | Command Validator | Skill Validator |
|----------------|-----------------|-------------------|-----------------|
| `Skill: name` | ✅ | ❌ | N/A |
| `Skill(name)` | ❌ | ❌ | N/A |
| `Skill("name")` | ❌ | ❌ | N/A |
| `Task(agent)` | N/A | ✅ | ✅ (suffix-only) |
| `delegate to agent` | N/A | ❌ | ✅ |
| Narrative mention | N/A | ❌ | ❌ |

---

## Root Cause: Regex Pattern Gaps

### Issue 1: `findSkillReferences()` Incomplete

**File:** `internal/cli/crossfile_refs.go:14-25`

Current patterns:
```go
skillPatterns := []*regexp.Regexp{
    regexp.MustCompile(`(?m)^[^*\n]*\bSkill:\s*([a-z0-9][a-z0-9-]*)`),      // Skill: foo
    regexp.MustCompile(`(?m)\*\*Skill\*\*:\s*([a-z0-9][a-z0-9-]*)`),        // **Skill**: foo
    regexp.MustCompile(`(?m)Skill\(\s*["']?([a-z0-9][a-z0-9-]*)["']?\s*\)`), // Skill(foo) ✅ PRESENT
    regexp.MustCompile(`(?m)Skills?:\s*\n\s*[-*]\s*([a-z0-9][a-z0-9-]*)`),  // Skills:\n- foo
}
```

**Observation:** Pattern 3 actually EXISTS (line 22) but may not be capturing all cases.

**Re-analysis needed:** Check if pattern 3 works correctly for:
- `Skill("foo-bar")` with quotes
- `Skill(foo-bar)` without quotes
- `Skill( foo-bar )` with whitespace

### Issue 2: `ValidateSkill()` Narrow Agent Patterns

**File:** `internal/cli/crossfile.go:247-255`

Current patterns only match `-specialist` suffix:
```go
agentPatterns := []struct {
    pattern string
    example string
}{
    {`delegate to\s+([a-z0-9][a-z0-9-]*-specialist)`, "delegate to foo-specialist"},
    {`use\s+([a-z0-9][a-z0-9-]*-specialist)`, "use foo-specialist"},
    {`see\s+([a-z0-9][a-z0-9-]*-specialist)`, "see foo-specialist"},
    {`Task\(([a-z0-9][a-z0-9-]*-specialist)`, "Task(foo-specialist)"},
}
```

**Problem:** Agents without `-specialist` suffix (e.g., `Task(Explore)`, `Task(Plan)`) aren't detected.

**Missing patterns:**
- `Task(agent-name)` without suffix requirement
- `agent-name handles` (narrative)
- `via agent-name` (indirect delegation)

---

## Implementation Plan

### Task 1: Verify `Skill()` Detection (P0)

**Test First:**
```go
// File: internal/cli/crossfile_refs_test.go (create if needed)
func TestFindSkillReferences_FunctionCalls(t *testing.T) {
    tests := []struct {
        name     string
        content  string
        expected []string
    }{
        {
            name:     "Skill() without quotes",
            content:  "Load via Skill(foo-bar)",
            expected: []string{"foo-bar"},
        },
        {
            name:     "Skill() with double quotes",
            content:  `Use Skill("my-skill")`,
            expected: []string{"my-skill"},
        },
        {
            name:     "Skill() with single quotes",
            content:  `Invoke Skill('test-skill')`,
            expected: []string{"test-skill"},
        },
        {
            name:     "Skill() with whitespace",
            content:  "Call Skill( some-skill )",
            expected: []string{"some-skill"},
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := findSkillReferences(tt.content)
            if len(got) != len(tt.expected) {
                t.Errorf("findSkillReferences() = %v, want %v", got, tt.expected)
            }
            for i, skill := range tt.expected {
                if got[i] != skill {
                    t.Errorf("skill[%d] = %s, want %s", i, got[i], skill)
                }
            }
        })
    }
}
```

**If tests PASS:** Pattern already works, document it.
**If tests FAIL:** Fix regex in `crossfile_refs.go:22`.

---

### Task 2: Add Skill Validation to Commands (P0)

**File:** `internal/cli/crossfile.go:60-104`

**Current:** Only validates agent references via `Task()`.

**Add after line 104:**
```go
// Check for skill references (Skill: or Skill() patterns)
skillRefs := findSkillReferences(contents)
seenSkillErrors := make(map[string]bool)

for _, skillRef := range skillRefs {
    if seenSkillErrors[skillRef] {
        continue
    }
    if _, exists := v.skills[skillRef]; !exists {
        seenSkillErrors[skillRef] = true
        errors = append(errors, cue.ValidationError{
            File:     filePath,
            Message:  fmt.Sprintf("Command references non-existent skill '%s'. Create skills/%s/SKILL.md", skillRef, skillRef),
            Severity: "error",
            Source:   cue.SourceCClintObserve,
        })
    }
}
```

**Test Case:**
```go
// File: internal/cli/crossfile_test.go (expand existing tests)
func TestValidateCommand_SkillReferences(t *testing.T) {
    validator := NewCrossFileValidator([]discovery.File{
        {
            RelPath:  "skills/existing-skill/SKILL.md",
            Type:     discovery.FileTypeSkill,
            Contents: "Skill content",
        },
    })

    tests := []struct {
        name        string
        contents    string
        wantErrors  int
        wantMessage string
    }{
        {
            name:       "valid skill reference",
            contents:   "Use Skill: existing-skill",
            wantErrors: 0,
        },
        {
            name:        "missing skill reference",
            contents:    "Use Skill: nonexistent-skill",
            wantErrors:  1,
            wantMessage: "non-existent skill 'nonexistent-skill'",
        },
        {
            name:        "Skill() function call missing",
            contents:    `Invoke Skill("missing-skill")`,
            wantErrors:  1,
            wantMessage: "non-existent skill 'missing-skill'",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            errors := validator.ValidateCommand("commands/test.md", tt.contents, map[string]interface{}{})
            if len(errors) != tt.wantErrors {
                t.Errorf("ValidateCommand() errors = %d, want %d", len(errors), tt.wantErrors)
            }
            if tt.wantMessage != "" && len(errors) > 0 {
                if !strings.Contains(errors[0].Message, tt.wantMessage) {
                    t.Errorf("Error message = %s, want containing %s", errors[0].Message, tt.wantMessage)
                }
            }
        })
    }
}
```

---

### Task 3: Expand Agent Patterns in Skills (P1)

**File:** `internal/cli/crossfile.go:247-255`

**Current issue:** Only matches `-specialist` suffix.

**Fix:** Add broader patterns and filter against built-ins.

```go
agentPatterns := []struct {
    pattern string
    example string
}{
    // Existing patterns (keep for specificity)
    {`delegate to\s+([a-z0-9][a-z0-9-]*-specialist)`, "delegate to foo-specialist"},
    {`use\s+([a-z0-9][a-z0-9-]*-specialist)`, "use foo-specialist"},
    {`see\s+([a-z0-9][a-z0-9-]*-specialist)`, "see foo-specialist"},
    {`Task\(([a-z0-9][a-z0-9-]*-specialist)`, "Task(foo-specialist)"},

    // NEW: Broader patterns for non-specialist agents
    {`Task\(\s*["']?([a-z0-9][a-z0-9-]*)["']?\s*[,)]`, "Task(agent-name)"},
    {`delegate via\s+([a-z0-9][a-z0-9-]*)`, "delegate via agent-name"},
    {`([a-z0-9][a-z0-9-]*-agent)\s+handles`, "foo-agent handles"},
}

seenAgents := make(map[string]bool)
for _, agentPattern := range agentPatterns {
    re := regexp.MustCompile(agentPattern.pattern)
    matches := re.FindAllStringSubmatch(contents, -1)
    for _, match := range matches {
        if len(match) < 2 {
            continue
        }
        agentRef := strings.TrimSpace(match[1])

        // Filter out built-in agents
        if builtInSubagentTypes[agentRef] {
            continue
        }

        if seenAgents[agentRef] {
            continue
        }
        seenAgents[agentRef] = true

        if _, exists := v.agents[agentRef]; !exists {
            errors = append(errors, cue.ValidationError{
                File:     filePath,
                Message:  fmt.Sprintf("Skill references '%s' but agent doesn't exist. Create agents/%s.md", agentRef, agentRef),
                Severity: "error",
                Source:   cue.SourceCClintObserve,
            })
        }
    }
}
```

**Test Case:**
```go
func TestValidateSkill_BroadAgentPatterns(t *testing.T) {
    validator := NewCrossFileValidator([]discovery.File{
        {
            RelPath:  "agents/test-agent.md",
            Type:     discovery.FileTypeAgent,
            Contents: "Agent content",
        },
    })

    tests := []struct {
        name        string
        contents    string
        wantErrors  int
        wantMessage string
    }{
        {
            name:       "Task() without specialist suffix - exists",
            contents:   "Delegates via Task(test-agent)",
            wantErrors: 0,
        },
        {
            name:        "Task() without specialist suffix - missing",
            contents:    "Delegates via Task(missing-agent)",
            wantErrors:  1,
            wantMessage: "agent doesn't exist",
        },
        {
            name:       "built-in agent - no error",
            contents:   "Use Task(Explore) to discover",
            wantErrors: 0,
        },
        {
            name:        "delegate via pattern - missing",
            contents:    "delegate via nonexistent-agent",
            wantErrors:  1,
            wantMessage: "agent doesn't exist",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            errors := validator.ValidateSkill("skills/test/SKILL.md", tt.contents)
            if len(errors) != tt.wantErrors {
                t.Errorf("ValidateSkill() errors = %d, want %d", len(errors), tt.wantErrors)
                for _, e := range errors {
                    t.Logf("  Error: %s", e.Message)
                }
            }
            if tt.wantMessage != "" && len(errors) > 0 {
                if !strings.Contains(errors[0].Message, tt.wantMessage) {
                    t.Errorf("Error message = %s, want containing %s", errors[0].Message, tt.wantMessage)
                }
            }
        })
    }
}
```

---

### Task 4: Integration Tests (P2)

**File:** `internal/cli/crossfile_test.go` (create if needed)

End-to-end test with realistic component structure:

```go
func TestCrossFileValidation_EndToEnd(t *testing.T) {
    files := []discovery.File{
        // Agent referencing skill
        {
            RelPath:  "agents/test-agent.md",
            Type:     discovery.FileTypeAgent,
            Contents: `
Skill: real-skill
Skill(another-skill)
Task(helper-agent)
`,
        },
        // Command referencing agent and skill
        {
            RelPath:  "commands/test-cmd.md",
            Type:     discovery.FileTypeCommand,
            Contents: `
Task(test-agent)
Use Skill: real-skill
`,
        },
        // Skill referencing agent
        {
            RelPath:  "skills/real-skill/SKILL.md",
            Type:     discovery.FileTypeSkill,
            Contents: `
Delegate to helper-agent via Task(helper-agent)
`,
        },
        // Helper agent
        {
            RelPath:  "agents/helper-agent.md",
            Type:     discovery.FileTypeAgent,
            Contents: "Helper agent content",
        },
    }

    validator := NewCrossFileValidator(files)

    // Test agent validation
    agentErrors := validator.ValidateAgent("agents/test-agent.md", files[0].Contents)
    if len(agentErrors) != 1 { // Should only error on 'another-skill' (missing)
        t.Errorf("ValidateAgent() errors = %d, want 1", len(agentErrors))
    }
    if !strings.Contains(agentErrors[0].Message, "another-skill") {
        t.Errorf("Expected error about 'another-skill', got: %s", agentErrors[0].Message)
    }

    // Test command validation
    cmdErrors := validator.ValidateCommand("commands/test-cmd.md", files[1].Contents, map[string]interface{}{})
    if len(cmdErrors) != 0 { // All references should be valid
        t.Errorf("ValidateCommand() errors = %d, want 0", len(cmdErrors))
        for _, e := range cmdErrors {
            t.Logf("  Error: %s", e.Message)
        }
    }

    // Test skill validation
    skillErrors := validator.ValidateSkill("skills/real-skill/SKILL.md", files[2].Contents)
    if len(skillErrors) != 0 { // helper-agent exists
        t.Errorf("ValidateSkill() errors = %d, want 0", len(skillErrors))
    }
}
```

---

## Edge Cases

### 1. Built-in Agents

**Pattern:** `Task(Explore)`, `Task(Plan)`, etc.

**Handling:** Already covered by `builtInSubagentTypes` map (lines 20-26).

**Test:**
```go
contents := "Use Task(Explore) to discover files"
errors := validator.ValidateCommand("test.md", contents, map[string]interface{}{})
// Should NOT error - Explore is built-in
```

### 2. Dynamic/Variable References

**Pattern:** `Task(subagent_type: "foo")` or `Task(args.agent)`

**Handling:** Skip validation for dynamic patterns.

**Current code already handles this:**
```go
if strings.Contains(agentRef, "subagent_type") {
    continue
}
```

**Expand to:**
```go
if strings.Contains(agentRef, "subagent_type") ||
   strings.Contains(agentRef, ".") ||
   strings.Contains(agentRef, "[") {
    continue // Skip variable/dynamic references
}
```

### 3. Skill References in Comments

**Pattern:** `<!-- Skill: foo -->` or `// Skill: bar`

**Current behavior:** Regex detects these.

**Decision:** Should comments count?

**Recommendation:** NO - add negative lookahead for comments.

**Updated regex:**
```go
// Exclude HTML comments: (?!<!--)
// Exclude markdown code with language: (?!```)
regexp.MustCompile(`(?m)^(?!<!--)[^*\n]*\bSkill:\s*([a-z0-9][a-z0-9-]*)`)
```

---

## Error Message Format

**Consistency with existing errors:**

```go
// Current agent error format (crossfile.go:90)
fmt.Sprintf("Task(%s) references non-existent agent. Create agents/%s.md", agentRef, agentRef)

// Current skill error format (crossfile.go:233)
fmt.Sprintf("Skill: %s references non-existent skill. Create skills/%s/SKILL.md", skillRef, skillRef)
```

**New command → skill error:**
```go
fmt.Sprintf("Command references non-existent skill '%s'. Create skills/%s/SKILL.md", skillRef, skillRef)
```

**New skill → agent error (already exists, just broader):**
```go
fmt.Sprintf("Skill references '%s' but agent doesn't exist. Create agents/%s.md", agentRef, agentRef)
```

---

## Acceptance Criteria

- [x] `findSkillReferences()` detects `Skill()` function calls (verify existing pattern)
- [ ] Commands validate skill references and report missing skills
- [ ] Skills validate all agent reference patterns (not just `-specialist`)
- [ ] Built-in agents (`Explore`, `Plan`) don't trigger false errors
- [ ] Dynamic references (`subagent_type`, variables) are skipped
- [ ] Test coverage ≥80% for new validation logic
- [ ] All existing tests pass
- [ ] Documentation updated: `/Users/vampire/go/src/cclint/docs/cross-file-validation.md`

---

## Validation Steps

After implementation:

```bash
# 1. Run tests
go test ./internal/cli/... -v

# 2. Test on real codebase
cclint agents --verbose
cclint commands --verbose
cclint skills --verbose

# 3. Check for false positives
cclint agents 2>&1 | grep "non-existent" | wc -l  # Should be low

# 4. Test built-in agent filtering
grep -r "Task(Explore)" ~/.claude/agents/*.md | wc -l
# Run linter, count errors mentioning "Explore" - should be 0

# 5. Baseline comparison
cclint --baseline-create
# Make change to reference non-existent skill
# Should detect new error
```

---

## Documentation Updates

**File:** `/Users/vampire/go/src/cclint/docs/cross-file-validation.md`

Add section:

```markdown
## Agent Reference Detection (Skills)

Skills can reference agents using multiple patterns. All are validated:

| Pattern | Example | Validated |
|---------|---------|-----------|
| Task() specialist | `Task(foo-specialist)` | ✅ |
| Task() generic | `Task(agent-name)` | ✅ |
| Narrative | `foo-agent handles...` | ✅ |
| Delegation | `delegate via agent-name` | ✅ |

Built-in agents are automatically excluded: `Explore`, `Plan`, `general-purpose`, etc.

## Skill Reference Detection (All Components)

Both agents and commands can reference skills:

| Pattern | Example | Validated |
|---------|---------|-----------|
| Colon | `Skill: foo-bar` | ✅ |
| Function | `Skill(foo-bar)` | ✅ |
| Quoted | `Skill("foo-bar")` | ✅ |
| List | `Skills:\n- foo-bar` | ✅ |
```

---

## Rollout Plan

1. **Phase 1 (P0):** Verify `Skill()` detection works
2. **Phase 2 (P0):** Add skill validation to commands + tests
3. **Phase 3 (P1):** Expand agent patterns in skill validator + tests
4. **Phase 4 (P2):** Integration tests + documentation
5. **Phase 5:** Release with updated baseline handling

**Risk:** Existing codebases may have many broken references.

**Mitigation:**
- Use `--baseline-create` to snapshot current state
- Only fail on NEW broken references
- Provide clear fix instructions in error messages

---

## Related Files

| File | Changes |
|------|---------|
| `internal/cli/crossfile.go` | Add skill validation to `ValidateCommand()`, expand agent patterns in `ValidateSkill()` |
| `internal/cli/crossfile_refs.go` | Verify/fix `Skill()` pattern detection |
| `internal/cli/crossfile_test.go` | Add comprehensive test cases |
| `docs/cross-file-validation.md` | Document new validation behavior |

---

## Questions for Review

1. Should skill references in markdown comments be validated? (Recommend: NO)
2. Should we validate references in code examples (triple backticks)? (Current: YES)
3. Should narrative mentions count? e.g., "The foo-agent handles..." (Recommend: YES for skills, add pattern)
4. Built-in agents list complete? Missing any? (Check Claude Code docs)

---

## Implementation Notes

**Go Regex Quirk:** Character classes like `[^*]` match newlines by default. Always use `[^*\n]` to prevent greedy cross-line matching. See crossfile.go:18 comments.

**DRY Opportunity:** `findSkillReferences()` and agent pattern matching logic could share a common `findReferences(patterns []string) []string` helper.

**Performance:** All validators use `make(map[string]bool)` deduplication to prevent duplicate errors. Keep this pattern.
