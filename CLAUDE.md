# CLAUDE.md

This file provides guidance to Claude Code when working with code in this repository.

## Build & Run

```bash
#-Build and symlink to ~/go/bin
go build -o cclint . && ln -sf $(pwd)/cclint ~/go/bin/cclint

#-Build with version override
go build -ldflags "-X github.com/dotcommander/cclint/cmd.Version=1.0.0" -o cclint .

#-Version
cclint --version              # show version (defaults to "dev")
cclint -V                     # short form

#-Run linter (defaults to ~/.claude)
cclint                        # lint all component types
cclint agents                 # lint only agents
cclint commands               # lint only commands
cclint skills                 # lint only skills
cclint plugins                # lint only plugin manifests

#-Single-file mode
cclint ./agents/my-agent.md            # lint one file (auto-detect type)
cclint path/to/file.md                 # lint by path
cclint a.md b.md c.md                  # lint multiple files
cclint --type agent ./custom/file.md   # override type detection
cclint --file agents                   # lint file literally named "agents"

#-Git integration (pre-commit hooks)
cclint --staged                        # lint only staged files
cclint --diff                          # lint all uncommitted changes

#-Baseline mode (gradual adoption)
cclint --baseline-create                  # create baseline from current issues
cclint --baseline                         # lint with baseline filtering (only new issues fail)
cclint --baseline-path custom.json        # use custom baseline file path

#-Common flags
cclint --root /path/to/project agents    # specify project root
cclint --scores agents                    # show quality scores (0-100)
cclint --improvements agents              # show improvement recommendations
cclint --format json --output report.json # JSON output for CI
cclint --quiet                            # errors only, no suggestions
cclint --verbose                          # detailed processing info
```

## Testing

```bash
go test ./...                             # run all tests
go test ./internal/scoring/...            # test specific package
```

## Architecture

```
cmd/                    # Cobra commands (root, agents, commands, skills, etc.)
internal/
├── discovery/          # File discovery using doublestar glob patterns
├── frontend/           # YAML frontmatter parsing
├── cue/
│   ├── validator.go    # CUE-based schema validation
│   └── schemas/        # Embedded CUE schemas (agent, command, settings, claude_md)
├── baseline/           # Baseline support for gradual adoption
│   ├── baseline.go     # Baseline creation, loading, filtering
│   └── baseline_test.go
├── scoring/            # Quality scoring (0-100) with tier grading (A-F)
│   ├── types.go        # QualityScore, ScoringMetric interfaces
│   ├── agent_scorer.go
│   ├── command_scorer.go
│   ├── skill_scorer.go
│   └── plugin_scorer.go
├── cli/                # Lint orchestration per component type
│   ├── crossfile/         # Cross-file validation
│   │   ├── crossfile.go   # Reference validation, orphan detection
│   │   ├── triggers.go    # Ghost trigger detection in reference files
│   │   ├── refs.go        # Reference extraction helpers
│   │   └── graph.go       # Cycle detection
│   └── baseline_filter.go  # Baseline filtering logic
├── output/             # Formatters (console, json, markdown)
├── outputters/         # Output coordination
├── config/             # Viper-based config (.cclintrc.json/.yaml)
└── project/            # Project root detection
schemas/                # Root-level CUE schemas (duplicated in internal/cue/schemas)
```

## Key Patterns

**Validation Pipeline**: Discovery → Frontmatter Parse → CUE Schema Validation → Go-based Best Practice Checks → Scoring → Output

**CUE Schemas**: Embedded via `//go:embed` in `internal/cue/validator.go`. Validate frontmatter structure for agents, commands, settings.

**Quality Scoring**: Four categories (structural 0-40, practices 0-40, composition 0-10, documentation 0-10) summed to 0-100. Tier grades: A≥85, B≥70, C≥50, D≥30, F<30.

**Discovery Patterns**: Searches `.claude/{agents,commands,skills}/**/*.md`, `{agents,commands,skills}/**/*.md`, and `**/.claude-plugin/plugin.json` paths. Lowercase `skill.md` files are also discovered and flagged as naming errors.

## Lint Rules Enforced

| Component | Size Limit | Key Checks |
|-----------|------------|------------|
| Agent | 200 lines | name format, model, triggers, Foundation/Workflow sections, Skill() loading, body-tool mismatch (suppressed for 8+ tools) |
| Command | 50 lines | delegation pattern, semantic routing table, tool allowlist (Task/Skill/AskUserQuestion only), Skill-without-Task detection |
| Skill | 500 lines | SKILL.md filename, Quick Reference table, examples detection (recognizes references/ pointers) |
| Plugin | 5KB | name format, semver version, author.name, required fields |

## Cross-File Validation

**Location**: `internal/cli/crossfile.go`

Detects skill references using `findSkillReferences()`:

```go
// Pattern: ^[^*\n]*\bSkill:\s*([a-z0-9][a-z0-9-]*)
// Key: [^*\n]* prevents matching across newlines
```

**Go Regex Quirk**: Character classes like `[^*]` match newlines by default in Go. Use `[^*\n]` to exclude newlines and prevent greedy cross-line matching.

**Orphan Detection**: `FindOrphanedSkills()` builds a reference graph and reports skills with zero incoming edges as info-level suggestions.

**Ghost Trigger Detection**: `ValidateTriggerMaps()` scans `skills/*/references/*.md` for trigger routing tables and validates that referenced skills/agents exist.

See: `docs/cross-file-validation.md`

## Baseline Support (Gradual Adoption)

Baseline allows teams to adopt cclint incrementally by accepting the current state and only failing on new issues.

**Workflow**:
1. Run `cclint --baseline-create` to snapshot all current issues into `.cclintbaseline.json`
2. Commit the baseline file to version control
3. Run `cclint --baseline` in CI/CD - only new issues will fail the build
4. Fix issues incrementally, update baseline as needed

**Fingerprinting**: Issues are fingerprinted using SHA256 hash of (file + source + normalized message pattern). Line numbers are ignored so issues remain stable when code shifts.

**Example**:
```bash
#-Legacy project with 100 existing issues
cclint agents
#-0/70 passed, 100 errors

#-Create baseline to accept current state
cclint --baseline-create
#-Baseline created: .cclintbaseline.json (100 issues)

#-Now only new issues fail
cclint --baseline agents
#-✓ All passed
#-100 baseline issues ignored (100 errors, 0 suggestions)

#-New issue added
cclint --baseline agents
#-69/70 passed, 1 error (new issue not in baseline)
```

## Config

Supports `.cclintrc.json`, `.cclintrc.yaml`, `.cclintrc.yml` in project root. Environment variables with `CCLINT_` prefix also supported.

## Operational Context

- Documentation now has a role-based first-hop route: `README.md` -> `docs/README.md` -> `docs/setup.md` / `docs/connect-assistant.md` / `docs/change-cclint.md`.
- First-hop docs are flat and plain-language (`setup`, `common-tasks`, `connect-assistant`, `change-cclint`) while deep reference material remains under `docs/guides/`, `docs/reference/`, and `docs/rules/`.
- Rules documentation is canonicalized under `docs/rules/*.md`; do not recreate `docs/RULES.md`.
- `docs/` is instruction-only: historical snapshots and docs test specs were removed (`docs/ANTHROPIC_REQUIREMENTS.md`, `docs/tests/*.md`) and should stay outside `docs/`.

| Key | Value |
|-----|-------|
| claude_code_last_updated | v2.1.50 |
| valid_agent_colors | red, blue, green, yellow, purple, orange, pink, cyan, gray, magenta, white (11 total) |
| command_allowed_tools | Task, Skill, AskUserQuestion only — other tools are errors |
| body_tool_mismatch_threshold | 8+ declared tools = general-purpose agent, check suppressed |
| knowntools_location | `internal/cli/lineutil.go` exported `KnownTools` var (shared by tool validation and body scanning) |
| stale_binary_trap | Always `go build -o cclint .` before running `./cclint` — stale binary causes phantom results |
| golangci_lint_version | v2 config format required (`version: "2"` in `.golangci.yml`) |
