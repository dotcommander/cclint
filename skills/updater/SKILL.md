---
name: updater
description: "Local-first Claude Code lint-surface sweep for cclint: fetch the changelog gap, verify candidate shapes against the installed binary, then edit schemas, lint rules, tests, and docs directly — all parent-side, dispatch only as an escape hatch for large generation. Use when a new Claude Code version ships, when CLAUDE.md claude_code_last_updated is stale, for version gap detection and domain classification, or to re-verify a changelog claim against the binary."
---

# Updater — Claude Code Lint-Surface Sweep

Every phase runs parent-side with direct tools. The full changelog, raw binary
strings, and whole source files never enter root context — only the gap slice,
classified candidates, and edit anchors do. Root judges scope, edits directly,
and sequences; an LLM dispatch is reserved for a single step that must generate
more than ~100 lines of new code (Local-vs-Dispatch Boundary below).

**Time budget: the whole invocation finishes in under 10 minutes.** Phases 0–3
≤ 4 min combined; implementation and docs are direct edits; smoke ≤ 1 min; FUI
two batched posts ≤ 1 min. A dispatch, when it applies at all, is the dominant
cost — treat it as the exception, never the default.

## Quick Reference

| Phase | Owner | Returns to root |
|-------|-------|-----------------|
| 0. Preflight | root, one Bash block | binary identity, commit policy, row + var anchors |
| 1. Gap + classify | root fast path | gap slice + candidate table — never the full changelog |
| 2. Verify shapes | root (`rg -a` on the installed binary) | snippet + type determination per candidate |
| 3. Judge scope | root | confirmed change list; `--dry-run` stops here |
| 4. Locate anchors | root (`rg -n` per edit site) | file:line + verbatim spans |
| 5. Implement | root direct edits (dispatch only for >100-line steps) | code + tests, one logical group at a time |
| 6. Docs | root direct edits | lint-surface.md surface + AGENTS.md row, `cmd.ClaudeCodeSurface` |
| 7. Smoke + close | root | exit codes, commits, FUI handoff |

Entry-point routing:

| User question | Action |
|---------------|--------|
| "A new Claude Code version shipped — sweep it" | Full workflow, Phases 0–7 |
| "Anything lintable since the pinned version?" | `--dry-run`: Phases 0–3, candidate table only |
| "Re-scan from an older baseline" | `--from <ver>` override, then full workflow |
| "The changelog says X — does the binary agree?" | Phase 2 shape verification before any edit |

## Local-vs-Dispatch Boundary

Tasks this workflow formerly delegated to an LLM and now handles locally —
keep them local:

| Task | Was | Now |
|------|-----|-----|
| Large-gap classification (> 5 versions) | general-purpose WebFetch agent | Local grep triage of the gap slice (Phase 1) |
| Edit-site anchors | pi-agent read drive (~1.5 min) | Root `rg -n` (Phase 4) |
| Implementation | serial pi-agent write drives | Root direct edits for mechanical changes (Phase 5) |
| Docs + version row | pi-agent write drive | Root direct edits (Phase 6) |
| Commits | rode in write drives | Root commits at close (Phase 7) |
| Build / test / cclint verification | Task dispatch | Root runs the commands (Phase 7) |
| Smoke fixture run | never dispatched — keep it that way | Root (Phase 7) |

**Dispatch escape hatch:** one pi-agent write drive per logical change ONLY
when that change exceeds ~100 lines of new code (typically a new linter rule
with its test file). Spec carries: exact old→new text from Phase 4 anchors
(never paraphrase), every file a failing gate may force the worker to touch —
fixture tests asserting old behavior — declared as a TARGET up front,
`VERIFY LANDED:` grep, `BUILD CHECK: go build ./... && go test ./internal/cue/ ./internal/lint/`
(scoped, never `go test ./...`), and `COMMIT MESSAGE:` only when every TARGET
was clean at preflight. Never instruct a dispatched agent to call TaskUpdate —
it is a deferred tool that fails silently in dispatch; root owns task
lifecycle. After any drive, verify tree state with `git status` / `git log`
directly — a `baseline-unchanged` self-report can be blind to surviving edits.

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
| Docs + surface row | Any modeled change above | `docs/rules/settings.md`, `docs/reference/schemas.md`, `.claude/rules/lint-surface.md`, `AGENTS.md`, `cmd/root.go` |

**Skip**: CLI flags, UI changes, SDK fields, env vars, performance
improvements, bug fixes to existing behavior, and anything on the
"Intentionally Not Modeled" guardrail in `.claude/rules/lint-surface.md` (runtime hook stdout contract,
PermissionRequest agent hooks, instructionFiles/projectInstructions,
claudeMdProcessingMode, yoloClassifier).

## Workflow

### Phase 0 — Preflight (root, one Bash block)

```bash
claude --version; BIN=$(readlink -f "$(command -v claude)"); echo "$BIN"
git status --porcelain --untracked-files=all | head -40
rg -n 'claude_code_last_updated' AGENTS.md | head -1
rg -n '## Validation Surface' .claude/rules/lint-surface.md | head -1
rg -n 'ClaudeCodeSurface = ' cmd/root.go
```

Decisions this block drives:

- **Binary identity**: `~/.local/bin/claude` resolves to
  `~/.local/share/claude/versions/<ver>` — a compiled binary, not cli.js.
  Every grep needs `rg -a`; plain rg reports only "binary file matches". If
  the binary is older than the changelog's newest entry, later versions are
  changelog-classified-only — record that.
- **Commit policy**: the close (Phase 7) commits only when the tree was clean
  at preflight. If any sweep target is pre-dirty (modified or untracked), make
  no commits — a half-committed sweep is worse than one uniform uncommitted
  delta — and report commit sequencing at close. Never commit the AGENTS.md
  row without the root.go var in the same commit:
  `TestClaudeCodeSurfaceMatchesAGENTSRow` breaks mid-history.
- **Dirty-tree health**: if the tree is dirty on gate-relevant paths, one
  scoped probe before editing so failures are attributable:
  `go build ./... && go test ./internal/cue/ ./internal/lint/`. Plain
  `go build` works while no ancestor go.work exists; if one reappears, prefix
  gates with `GOWORK=off` (this repo is not in it).

### Phase 1 — Gap + classify (root, no dispatch at any gap size)

Fetch, enumerate headers, then triage locally — only matching bullets and
version headers enter context, never the whole gap:

```bash
gh api -H "Accept: application/vnd.github.raw" repos/anthropics/claude-code/contents/CHANGELOG.md > /tmp/cclint-changelog.md
test -s /tmp/cclint-changelog.md && grep -qm1 '^## ' /tmp/cclint-changelog.md || { echo "fetch failed: empty or headerless changelog"; exit 1; }
grep -n '^## ' /tmp/cclint-changelog.md | head -8
sed -n "1,$((PREV_HEADER_LINE-1))p" /tmp/cclint-changelog.md > /tmp/cclint-gap.md
grep -nE '^## |settings|frontmatter|hook|tool|model|plugin|marketplace|schema|sandbox|permission' /tmp/cclint-gap.md
```

Read the ±3-line context of matching bullets only when classification needs
it. The changelog is newest-first, so the gap is everything above the previous
version's header.

- NOT `aic claude` — it returns only the newest entry (and takes no `v`
  prefix, and never wrap fetches in `2>/dev/null`: aic writes "version not
  found" to stderr; silencing it fabricates empty output). If `gh` fails,
  WebFetch on raw.githubusercontent.com works, but its small model caps
  verbatim quotes at ~125 chars — cross-check long bullets with a second
  targeted pass.
- Newest header == previous version → current. Report and stop; no Phase 6
  bump.
- Classify the slice in-root against the Domain Classification Table and the
  skip list above.

### Phase 2 — Pin exact shapes against the binary (root)

HARD GATE: never model from changelog prose — binary-grepped shapes only. One
discriminating grep per candidate:

```bash
rg -a -o '.{0,100}<candidateKey>.{0,180}' "$BIN" | head -5
```

The snippet determines string / [...string] / bool / enum members / union.
When a candidate's names are already modeled (KnownTools, knownPluginFields,
schema fields), the grep only confirms the surface didn't move — no edit
follows.

### Phase 3 — Root judgment (no dispatch)

Drop anything hitting the scope guardrail. Genuinely ambiguous →
AskUserQuestion. HARD GATE: emit a user-visible classification table BEFORE
any implementation — one row per changelog entry across the whole gap, not
just the lintable ones:

```
| Version | Change | Lintable? | cclint action |
|---|---|---|---|
| 2.1.X | <one-line description> | ✅ Yes / ⚠️ Partial / ❌ No | <action or skip reason> |
```

- ✅ Yes — schema/lint surface change, will be implemented
- ⚠️ Partial — adjacent to lintable surface but out of scope; document in
  `.claude/rules/lint-surface.md` operational context
- ❌ No — runtime/UX/CLI/SDK/bugfix, no lint surface

End with a one-line summary: `N lintable, M actioned, K documented, rest
skipped (runtime/UX/bugfixes).` `--dry-run` stops here. If zero candidates,
still emit the table, then report and stop.

### Phase 4 — Locate anchors (root `rg -n`)

For every edit site the confirmed list touches: CUE schemas, Go validators,
test tables, doc pages, `.claude/rules/lint-surface.md`, the AGENTS.md
`claude_code_last_updated` row, and
`cmd/root.go` `ClaudeCodeSurface`. Grep each, record `file:line` plus
verbatim ±3-line spans — those spans are the exact old→new text for Phase 5/6
edits. A docs-only sweep needs only the row + var anchors, both already in
hand from Phase 0.

### Phase 5 — Implement (root direct edits)

One logical group at a time (e.g. new settings keys + their test file; model
aliases + test entries), using the Phase 4 spans as literal old→new edits:

- Code + tests land together in the same group — never separate them.
- **Minimal-edit constraint:** make only the edits the confirmed list
  specifies; do not refactor, split, or reorganize files beyond them.
- **Doc-sync**: when changing code values, update the matching docs in the
  same group — hook events → `docs/rules/settings.md` +
  `docs/reference/schemas.md`; Known tools → `docs/reference/schemas.md`;
  schema fields → `docs/reference/schemas.md`.
- After all groups: `go build ./... && go test ./internal/cue/ ./internal/lint/`
  (scoped). Fix task-owned failures; rerun only what failed.

Use the dispatch escape hatch only for a >100-line single change, per the
Local-vs-Dispatch Boundary.

### Phase 6 — Docs (root direct edits)

Split across two files — never target the CLAUDE.md symlink, always the
physical repo-root AGENTS.md:

- `.claude/rules/lint-surface.md` (paths-conditional rule, the lint-surface
  ledger): Validation Surface entries, enum lines, and an Operational
  Context sweep bullet (versions covered, what was modeled, what was excluded
  and why).
- AGENTS.md: only the `claude_code_last_updated` table row.

Plus the matching `cmd.ClaudeCodeSurface` var in `cmd/root.go` in the same
change
(`TestClaudeCodeSurfaceMatchesAGENTSRow` guards the sync). Check
`docs/cross-file-validation.md` and `docs/common-tasks.md` only when the
change adds a validation type or component surface.

### Phase 7 — Smoke + close (root)

Rebuild first (`go build -o cclint .` — stale-binary trap; the output is
gitignored). Then:

- Full sweep: lint a temp fixture exercising each new key/alias via `--root`;
  expect exit 0 with no unknown-field or "Unknown model" output.
- Docs-only sweep: `./cclint --version` must print the new surface, and one
  `--root` fixture lint exits 0.
- Remove the fixture (literal path, `rm -r`).

Close:

- **Commits** per the Phase 0 policy — clean tree → logical groups (code,
  docs, AGENTS.md row + var together); pre-dirty tree → no commits, report
  sequencing. No push, no tag unless explicitly requested.
- **FUI**: two batched posts on the sweep thread — the classification record
  and the final handoff. Before re-posting after a suspected delivery
  failure, check `fui post list`: exit 0 ≠ indexed (apply-lock expiry);
  `fui scan` reconciles.

## Gotchas

| Gotcha | Problem | Fix |
|--------|---------|-----|
| `gh api ... --jq '.content' \| base64 -d` | Contents API omits `content` for files > 1 MB — CHANGELOG.md crossed 1 MB; 200 OK with an empty body | Raw Accept header form (Phase 1 command) |
| `aic claude` for multi-version gaps | Returns only the newest entry | `gh api` raw form; WebFetch fallback |
| Plain `rg` on the claude binary | Versioned install is a compiled binary; plain rg says "binary file matches" | `rg -a` |
| Modeling from changelog prose | Prose misstates shapes | Phase 2 binary grep is the only shape source |
| Skipping the classification table | Wastes effort on non-lintable changes | Phase 3 HARD GATE: full-entry table first |
| Returning the full changelog into root context | Token sink | Slice + grep-triage only the gap |
| WebFetch for the whole gap | Small model caps verbatim quotes at ~125 chars | Second targeted pass on long bullets |
| Stale `./cclint` binary | Phantom results | `go build -o cclint .` before smoke |
| Unscoped `go test ./...` | Slow; scope creep | `./internal/cue/ ./internal/lint/` |
| Committing AGENTS.md row without the root.go var | `TestClaudeCodeSurfaceMatchesAGENTSRow` breaks mid-history | Same commit, always |
| Editing CLAUDE.md (symlink) | Edits the symlink, not the file | Edit physical AGENTS.md |
| Committing over a pre-dirty tree | Half-committed sweep | Phase 0 policy: no commits, report sequencing |
| Dispatching small mechanical edits | ~1.5 min + drift risk per drive | Root direct edits; hatch is >100-line only |

## Examples

### Hook event addition

Changelog says: "Added `WorktreeCreate` hook event"

1. Phase 2: `rg -a -o '.{0,100}WorktreeCreate.{0,180}' "$BIN"` — confirm event name
2. Edit `internal/lint/settings.go` valid events map + `internal/cue/schemas/settings.cue` hook event enum
3. Add test case in `internal/lint/settings_hooks_test.go`
4. Doc-sync: `docs/rules/settings.md` + `docs/reference/schemas.md`
5. Phase 6: lint-surface.md surface entry + Operational Context bullet

### New tool name

Changelog says: "Added `EnterWorktree` tool"

1. Phase 2: grep the binary for the tool name and its enum neighborhood
2. Edit `internal/textutil/lineutil.go` `KnownTools` (the CUE `#KnownTool`
   union is generated from it — no `.cue` edit)
3. Add test case in `internal/textutil/lineutil_test.go`
4. Doc-sync: `docs/reference/schemas.md` Known Tools section
5. Phase 6: lint-surface.md surface entry

## Success Criteria

- [ ] Version gap identified or confirmed current
- [ ] Candidate shapes verified against the installed binary, not changelog prose
- [ ] Classification table covered every gap entry before implementation
- [ ] Code changes include tests; scoped `go build` + `go test` pass clean
- [ ] Every mechanical task handled root-side; dispatch used only for a >100-line step, if any
- [ ] `.claude/rules/lint-surface.md` surface + AGENTS.md `claude_code_last_updated` row updated, `cmd.ClaudeCodeSurface` bumped to match, same commit
- [ ] Smoke fixture passed and was removed; commits (if any) follow the Phase 0 policy; no push or tag
- [ ] FUI sweep thread carries the classification record and final handoff; whole invocation under 10 minutes
