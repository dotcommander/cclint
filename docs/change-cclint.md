# Maintain cclint

This guide is for people changing cclint itself. For using the installed linter,
start with [setup](setup.md) or [common tasks](common-tasks.md).

## Develop a change

1. Clone the repository and install the Go toolchain specified in `go.mod`.
2. Locate the behavior in `cmd/` or `internal/`, then update its implementation,
   focused tests, and affected user documentation together. The [rules index](rules/README.md)
   and [schema reference](reference/schemas.md) map validation behavior to its owners.
3. Build and run relevant tests from the repository root. For example:

   ```bash
   go build ./...
   go test ./internal/cue/ ./internal/lint/
   ```

   Pick packages appropriate to the change. If an enclosing `go.work` does not
   include this module, prefix commands with `GOWORK=off`.
4. For CLI changes, build a fresh executable with `go build -o cclint .` before
   testing `./cclint`: an older executable may not reflect the source. Exercise
   the changed behavior with a representative fixture and review the diff.

The [scoring reference](scoring/README.md) and [programmatic usage guide](guides/programmatic-usage.md)
cover those specific interfaces. Running `cclint --scores` lints a target project;
by itself it does not validate changes to this documentation.

## Keep up with Claude Code

cclint models a tracked Claude Code lint surface. Its version is recorded in the
`claude_code_last_updated` row in `AGENTS.md` and in `cmd.ClaudeCodeSurface`;
keep them synchronized. `.claude/rules/lint-surface.md` records the modeled
surface and intentional exclusions.

The repository includes an **assistant-guided** update procedure in
[`skills/updater/SKILL.md`](../skills/updater/SKILL.md). It is not a `cclint`
command or an automatic background service. From a checkout, ask a coding
assistant to read that file and compare the tracked version with newer Claude
Code releases. For a read-only assessment, ask for classification only, with no
edits. For an update, authorize implementation separately or in the request.

The procedure uses the upstream changelog to identify candidates, checks shapes
against an available installed Claude Code binary, and guides changes to schemas,
validators, tests, documentation, and the surface version. It calls for Git,
Go, `rg`, and either GitHub CLI or access to a trusted changelog copy. A binary
older than a candidate cannot verify that candidate; record the gap rather than
guessing. The maintainer should review the classification, evidence, diff, and
focused verification before accepting the update.

The workflow does not automatically publish a release. Commit, push, tag, and
release decisions remain with the maintainer.
