---
name: updater
description: "Check Claude Code changelog for lintable changes and implement cclint updates — version gap detection, domain classification, parallel agent dispatch"
user-invocable: false
---

## Quick Reference

| Step | Action | Tool |
|------|--------|------|
| 1. Version gap | Read CLAUDE.md → extract `claude_code_last_updated` | Read |
| 2. Fetch changelog | `aic claude` (primary) or `gh api` (fallback) | Bash |
| 3. Classify | Map entries to cclint domains | — |
| 4. Task list | TaskCreate per domain group | TaskCreate |
| 5. Implement | Parallel `Task(sonnet)` per task | Task |
| 6. Verify | `Task(sonnet)`: build, test, cclint | Task |
| 7. Review docs | Check docs for stale references | Task |
| 8. Ship | Update version, delegate commit | Task |

## Domain Classification Table

For each changelog entry, classify against these domains:

| cclint Domain | Signal in Changelog | Source Files |
|---------------|---------------------|--------------|
| CUE schema (agent) | New agent frontmatter field, new allowed value | `internal/cue/schemas/agent.cue` |
| CUE schema (command) | New command frontmatter field | `internal/cue/schemas/command.cue` |
| CUE schema (skill) | New skill frontmatter field | `internal/cue/schemas/skill.cue` |
| CUE schema (settings) | New settings field, new setting type | `internal/cue/schemas/settings.cue` |
| Hook events | New hook event name, changed hook behavior | `internal/lint/settings.go` |
| Known tools | New tool name in allowed-tools | `internal/textutil/lineutil.go` |
| Lint rules (agents) | New agent validation requirement, naming rule | `internal/lint/agent_linter.go` |
| Lint rules (skills) | New skill validation requirement | `internal/lint/skill_linter.go` |
| Lint rules (commands) | New command validation requirement | `internal/lint/command_linter.go` |
| Cross-file validation | New reference pattern, new component type | `internal/lint/crossfile/` |
| Discovery | New file path pattern, new component directory | `internal/discovery/` |

**Skip**: CLI flags, UI changes, SDK fields, performance improvements, bug fixes to existing behavior.

Output: `| Changelog Entry | Domain | Action Required |`

## Workflow

### Step 1: Version gap

Read `CLAUDE.md` Operational Context → extract `claude_code_last_updated`.
Report: `v{old} → v{latest}`. If no gap, report and stop (unless `--from` overrides).

### Step 2: Fetch changelog

Primary method — fetch via `aic claude` (Homebrew CLI):

```bash
aic claude > .work/CHANGELOG.upstream.md
```

Fallback — if `aic` is not available, use GitHub API:

```bash
gh api repos/anthropics/claude-code/contents/CHANGELOG.md \
  -H "Accept: application/vnd.github.raw" > .work/CHANGELOG.upstream.md
```

Then Read `.work/CHANGELOG.upstream.md` and extract entries between `## {old_version}` and `## {new_version}` headers.

### Step 3: Classify lintable items

For each changelog entry, classify against the Domain Classification Table above. If zero lintable items found, report and stop.

### Step 4: Task list

TaskCreate one task per domain group from Step 3. Include:
- Schema changes (CUE files)
- Go lint rule changes (`internal/lint/`)
- Known tools/events updates (`internal/textutil/lineutil.go`, `internal/lint/settings.go`)
- Test additions (matching each code change)

Set dependencies: verification task blocked by all implementation tasks.

### Step 5: Implement

Dispatch parallel agents. For each task:
- Use `Task(sonnet)` for code + tests together (not separate agents)
- Include in every agent prompt: **"TaskUpdate task #N to in_progress before starting, completed when done."**
- Each agent must add tests alongside code changes
- **Doc-sync**: When changing code values (hook events, tools, schema fields), agents MUST also update:
  - Hook events → `docs/rules/settings.md` + `docs/reference/schemas.md`
  - Known tools → `docs/reference/schemas.md`
  - Schema fields → `docs/reference/schemas.md`

### Step 6: Verify

Delegate to `Task(sonnet)`:
1. `go build ./...`
2. `go test ./...`
3. `go build -o cclint . && ./cclint` — confirm zero errors
4. **Doc-code sync**: grep `docs/` for old value lists and confirm new values appear

If suggestions appear from the new changes, fix them before proceeding.

### Step 7: Review docs

Check if changes affect user-facing documentation:
- `docs/rules/settings.md` — valid events, tool names, fail messages
- `docs/reference/schemas.md` — schema field tables, hook events, tool whitelists
- `docs/cross-file-validation.md` — new validation types
- `docs/common-tasks.md` — new component types or flags
- `README.md` / `CLAUDE.md` — operational context, feature list

Delegate doc fixes to `Task(sonnet)` if needed.

### Step 8: Ship

Ask user before proceeding. Then:
- Update CLAUDE.md `claude_code_last_updated` to new version
- Delegate to `Task(sonnet)`: group into logical commits (code, docs, CLAUDE.md)
- Do NOT push or tag unless user explicitly requests it

## Anti-Patterns

| Mistake | Why it fails |
|---------|-------------|
| Skipping Step 3 classification | Wastes agent time on non-lintable changes |
| Implementing without TaskCreate | No dependency tracking, verification runs too early |
| Separate test tasks | Tests must ship with code in the same agent |
| Hardcoding `internal/cli/` paths | Package moved to `internal/lint/` — use Domain Classification Table |
| Using `WebFetch` for changelog | `aic claude` and `gh api` are more reliable; WebFetch may truncate |
| Pushing without user confirmation | Step 8 requires explicit user approval |

## Examples

### Hook event addition

Changelog says: "Added `WorktreeCreate` hook event"

1. Add to `internal/lint/settings.go` valid events map
2. Add to `internal/cue/schemas/settings.cue` hook event enum
3. Update `docs/rules/settings.md` Rule 049 description
4. Update `docs/reference/schemas.md` Hook Events list
5. Add test case in `internal/lint/settings_hooks_test.go`

### New tool name

Changelog says: "Added `EnterWorktree` tool"

1. Add to `internal/textutil/lineutil.go` `KnownTools` map
2. Add to `internal/cue/schemas/agent.cue` `#KnownTool` union
3. Update `docs/reference/schemas.md` Known Tools section
4. Add test case in `internal/textutil/lineutil_test.go`

## Success Criteria

- [ ] Version gap identified or confirmed current
- [ ] Changelog fetched and parsed (with fallback if truncated)
- [ ] Lintable items classified with domain and action
- [ ] Tasks created with proper dependencies and agent instructions
- [ ] All code changes include tests in the same agent
- [ ] `go build`, `go test`, and `cclint` all pass clean
- [ ] Docs reviewed and updated if affected
- [ ] CLAUDE.md version updated
- [ ] Changes committed in logical groups
