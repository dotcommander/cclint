---
name: updater
description: "Assistant-guided Claude Code lint-surface sweep for cclint: compare the changelog gap with an available installed binary, then review and update relevant schemas, lint rules, tests, and docs. Use when a new Claude Code version ships, the AGENTS.md surface row is stale, or a changelog claim needs checking."
---

# Updater — Claude Code Lint-Surface Sweep

This is an **assistant-guided maintainer procedure**, not a cclint command or
an autonomous update. To use it, point a coding assistant with access to this
checkout at `skills/updater/SKILL.md` and ask it to follow the relevant phases.
The maintainer supplies the desired version range and scope, an installed Claude
Code binary when binary verification is needed, and access to the upstream
changelog (or a trusted copy). Classification-only and an older starting
version are requests to the assistant, **not flags implemented by cclint or
this skill**.

Review the classification and binary evidence before accepting any modeled
change; review the scoped source/docs diff and verification results before
release or commit. Missing tools or a binary that predates a candidate leave
that candidate unverified; report the gap rather than inferring its shape.
The changelog is a candidate source, not a schema authority. Keep large raw
changelog and binary output out of assistant context; work from bounded excerpts.
Use direct edits by default; delegation is optional when a change is too large
for a useful direct edit. The phases are a checklist, not a time guarantee.

## Quick Reference

| Phase | Owner | Output |
|-------|-------|--------|
| 0. Preflight | assistant | available binary identity, working-tree state, row + var anchors |
| 1. Gap + classify | assistant | gap slice + candidate table — never the full changelog |
| 2. Verify shapes | assistant (`rg -a` on a compiled binary) | snippet + type determination per candidate |
| 3. Judge scope | assistant + maintainer review | candidate list; a classification-only request stops here |
| 4. Locate anchors | assistant (`rg -n` per edit site) | file:line + verbatim spans |
| 5. Implement | assistant direct edits | code + tests, one logical group at a time |
| 6. Docs | assistant direct edits | lint-surface.md surface + AGENTS.md row, `cmd.ClaudeCodeSurface` |
| 7. Smoke + close | assistant | focused evidence and handoff; commit only if authorized |

Entry-point routing:

| User question | Action |
|---------------|--------|
| "A new Claude Code version shipped — sweep it" | Full workflow, Phases 0–7 |
| "Anything lintable since the pinned version?" | Ask for classification only (Phases 0–3); no edits |
| "Re-scan from an older baseline" | Tell the assistant which version to start from, then follow the requested workflow |
| "The changelog says X — does the binary agree?" | Phase 2 shape verification before any edit |

## Local-vs-Dispatch Boundary

The invoking assistant handles these tasks directly when practical:

| Task | Approach |
|------|----------|
| Large-gap classification | Triage the gap slice locally (Phase 1) |
| Edit-site anchors | Use `rg -n` (Phase 4) |
| Implementation and docs | Edit confirmed changes directly (Phases 5–6) |
| Build, test and fixture smoke | Run focused checks locally (Phase 7) |
| Commits | Only with explicit authorization; otherwise leave the diff uncommitted |

**Optional dispatch:** for a large change (for example, a new rule and tests),
use an available assistant/agent only if useful. Supply exact edit anchors,
scoped targets and a focused verification request. Keep task ownership with the
invoking assistant; inspect the resulting diff yourself. Do not ask a worker to
commit unless the maintainer explicitly authorized that commit.

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

### Phase 0 — Preflight

From the repository root, check the available tools and locate the installed
binary without assuming a particular installation path or `readlink -f` support.
For example:

```bash
command -v claude && claude --version
command -v gh; command -v rg
git status --porcelain --untracked-files=all
rg -n 'claude_code_last_updated' AGENTS.md
rg -n '## Validation Surface' .claude/rules/lint-surface.md
rg -n 'ClaudeCodeSurface = ' cmd/root.go
```

Set `BIN` to a verified readable path to the installed Claude Code binary if
binary inspection is possible (`command -v claude` may return a wrapper or
symlink; resolve it using tools available on the host). Record its version.

Decisions this block drives:

- **Binary identity**: when inspecting a compiled binary, use `rg -a`; plain
  `rg` may report only "binary file matches". If the available binary is older
  than a candidate, mark that candidate changelog-only, not binary-verified.
- **Preservation and commits**: preserve pre-existing changes. Do not commit
  merely because the tree was clean or this procedure says to close. A commit
  requires explicit authorization; if authorized, review its exact contents and
  keep the AGENTS.md row and `cmd/root.go` var together
  (`TestClaudeCodeSurfaceMatchesAGENTSRow` guards their agreement).
- **Dirty-tree health**: if existing changes overlap a relevant check, establish
  a focused baseline where needed so failures can be attributed. Check
  `go env GOWORK`; use `GOWORK=off` for repo-local Go commands when this module
  is not included in an active workspace.

### Phase 1 — Gap + classify (root, no dispatch at any gap size)

If `gh` is available and authenticated, fetch and triage locally. First set
`CHANGELOG` and `GAP` to unique temporary file paths you own; clean them up
when finished. Locate the exact prior-version header before setting
`PREV_HEADER_LINE` (do not run the slice command with an unset value). Only
matching bullets and version headers need enter assistant context:

```bash
gh api -H "Accept: application/vnd.github.raw" repos/anthropics/claude-code/contents/CHANGELOG.md > "$CHANGELOG"
test -s "$CHANGELOG" && grep -qm1 '^## ' "$CHANGELOG" || { echo "fetch failed: empty or headerless changelog"; exit 1; }
grep -n '^## ' "$CHANGELOG" | head -8
# Locate the exact prior-version header; set PREV_HEADER_LINE from that match.
sed -n "1,$((PREV_HEADER_LINE-1))p" "$CHANGELOG" > "$GAP"
grep -nE '^## |settings|frontmatter|hook|tool|model|plugin|marketplace|schema|sandbox|permission' "$GAP"
```

Read the ±3-line context of matching bullets only when classification needs
it. The changelog is newest-first, so the gap is everything above the previous
version's header.

- A maintainer-provided changelog copy or another available fetch route may
  replace `gh`; check that it includes the entire requested gap. Do not rely
  on `aic claude` for multi-version gaps: this setup observed only the newest
  entry. Do not suppress fetch errors or infer coverage from a partial result.
- Newest header == previous version → current. Report and stop; no Phase 6
  bump.
- Classify the slice in-root against the Domain Classification Table and the
  skip list above.

### Phase 2 — Pin exact shapes against the binary (root)

HARD GATE: never model a new shape from changelog prose. Verify against an
installed binary at or beyond the candidate version; without one, classify
only and report what remains unverified. For compiled binaries, use one
discriminating grep per candidate:

```bash
rg -a -o '.{0,100}<candidateKey>.{0,180}' "$BIN" | head -5
```

Use the surrounding evidence to determine the shape (string, array, bool,
enum, or union). A matching snippet alone may be ambiguous or incomplete; if
it does not establish the shape, leave the candidate unverified. For names
already modeled (KnownTools, knownPluginFields, schema fields), check for
actual changes before editing.

### Phase 3 — Root judgment (no dispatch)

Drop anything hitting the scope guardrail. Ask the maintainer to resolve
materially ambiguous scope before editing. HARD GATE: emit a user-visible classification table BEFORE
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
skipped (runtime/UX/bugfixes).` A classification-only request stops here. If
zero candidates, still emit the table, then report and stop.

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
- After all groups, select a focused build and test for the affected packages
  (for example, `go build ./...` and `go test ./internal/cue/ ./internal/lint/`
  when those packages changed). Fix task-owned failures; rerun only affected
  checks. Do not run the full repository test suite by default.

Use optional dispatch only when it makes a large change easier to review;
the invoking assistant remains responsible for the final diff.

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

- Review the classification, scoped diff and check results with the maintainer.
  Leave changes uncommitted unless a commit was explicitly authorized; never
  push or tag without separate authorization. If authorized to commit, keep the
  AGENTS.md row and `cmd/root.go` var in the same commit.
- A personal FUI forum/thread is specific to the original maintainer's setup,
  not a requirement for independent use. If the maintainer provides a
  collaboration channel, send a concise classification and final handoff there;
  otherwise include both in the assistant's response. Do not invent a FUI
  command, thread, or delivery receipt.

## Gotchas

| Gotcha | Problem | Fix |
|--------|---------|-----|
| `gh api ... --jq '.content' \| base64 -d` | Contents API omits `content` for files > 1 MB — CHANGELOG.md crossed 1 MB; 200 OK with an empty body | Raw Accept header form (Phase 1 command) |
| `aic claude` for multi-version gaps | Returns only the newest entry | `gh api` raw form; WebFetch fallback |
| Plain `rg` on a compiled Claude binary | May only say "binary file matches" | `rg -a` |
| Modeling from changelog prose | Prose misstates shapes | Phase 2 binary grep is the only shape source |
| Skipping the classification table | Wastes effort on non-lintable changes | Phase 3 HARD GATE: full-entry table first |
| Returning the full changelog into root context | Token sink | Slice + grep-triage only the gap |
| Partial fetch of the gap | Missing versions can be overlooked | Confirm the previous-version header and inspect targeted context |
| Stale `./cclint` binary | Phantom results | `go build -o cclint .` before smoke |
| Unscoped `go test ./...` | Slow; scope creep | `./internal/cue/ ./internal/lint/` |
| Committing AGENTS.md row without the root.go var | `TestClaudeCodeSurfaceMatchesAGENTSRow` breaks mid-history | Same commit, always |
| Adding a `permissions.*` key to `settings.cue` only | Go-side closed key list in `internal/lint/settings_permissions.go` still rejects it ("unknown key"); per-package tests miss the drift | Edit both sites plus the Supported-keys message in one change; the Phase 7 smoke fixture catches it |
| Editing an AGENTS.md alias | May edit a different file than intended | Edit the physical repo-root AGENTS.md |
| Assuming a clean tree authorizes a commit | Unauthorized history change | Commit only with explicit authorization; preserve pre-existing changes |
| Dispatching small mechanical edits | Adds coordination and drift risk | Prefer direct edits; review any delegated diff |

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
- [ ] Any modeled candidate shapes verified against a suitable installed binary, not changelog prose; unavailable verification reported
- [ ] Classification table covered every gap entry before implementation
- [ ] Code changes include tests; relevant focused checks pass or their limits are reported
- [ ] Relevant `.claude/rules/lint-surface.md` entries and AGENTS.md `claude_code_last_updated` row updated; `cmd.ClaudeCodeSurface` matches when a version bump is justified
- [ ] Scoped diff and focused checks reviewed; any smoke fixture removed
- [ ] Classification, binary-verification limits and handoff reported; changes remain uncommitted unless explicitly authorized
