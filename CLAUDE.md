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
cclint agents commands        # lint multiple types
cclint skills                 # lint only skills
cclint settings               # lint only settings.json files
cclint rules                  # lint only rule files
cclint context                # lint only CLAUDE.md files
cclint plugins                # lint only plugin manifests
cclint output-styles          # lint only output style files
cclint summary                # show quality summary
cclint fmt                    # format component files
cclint fmt --write            # format and write in place

#-File and directory mode
cclint ./agents/my-agent.md            # lint one file (auto-detect type)
cclint path/to/file.md                 # lint by path
cclint a.md b.md c.md                  # lint multiple files
cclint ./commands/                     # lint all files in a directory
cclint ./command/                      # singular dir names auto-detected
cclint --type agent ./custom/file.md   # override type detection

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
├── types/              # Shared types: ValidationError, constants
├── textutil/           # Shared text utilities: KnownTools, frontmatter parser
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
├── lint/               # Lint orchestration per component type
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

**Location**: `internal/lint/crossfile/crossfile.go`

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

**Exclude patterns**: Use `exclude` to skip files from linting. Patterns use doublestar glob matching against relative paths from the root.

```json
{
  "exclude": [
    "commands/skill-optimize.md",
    "plugins/cache/**",
    "plugins/marketplaces/**"
  ]
}
```

## Validation Surface

What cclint validates, for the `/updater` workflow. Source files in parentheses.

### Hook Events — 28 total (`internal/lint/settings.go`)

`PreToolUse`, `PermissionRequest`, `PostToolUse`, `PostToolUseFailure`, `Notification`, `UserPromptSubmit`, `Stop`, `StopFailure`, `SubagentStart`, `SubagentStop`, `PreCompact`, `SessionStart`, `SessionEnd`, `TeammateIdle`, `TaskCompleted`, `TaskCreated`, `ConfigChange`, `WorktreeCreate`, `WorktreeRemove`, `InstructionsLoaded`, `PostCompact`, `Elicitation`, `ElicitationResult`, `CwdChanged`, `FileChanged`, `PermissionDenied`, `Setup`, `MessageDisplay` (v2.1.152, transforms/hides assistant message text)

Component hooks (agents/skills) only: `PreToolUse`, `PostToolUse`, `Stop`

### KnownTools (`internal/textutil/lineutil.go`)

`Read`, `Write`, `Edit`, `MultiEdit`, `Bash`, `Grep`, `Glob`, `LS`, `Task`, `NotebookEdit`, `WebFetch`, `WebSearch`, `TodoWrite`, `AskUserQuestion`, `TaskCreate`, `TaskUpdate`, `TaskList`, `TaskGet`, `TaskStop`, `Skill`, `LSP`, `KillShell`, `TaskOutput`, `SendMessage`, `Monitor`, `RemoteTrigger`, `EnterPlanMode`, `ExitPlanMode`, `EnterWorktree`, `ExitWorktree`, `CronCreate`, `CronDelete`, `CronList`, `Workflow`, `ScheduleWakeup`, `PushNotification`, `REPL`, `*`

### CUE Schema Fields

**Agent** (`internal/cue/schemas/agent.cue`): name (required), description (required), model, color, tools, disallowedTools, permissionMode, maxTurns, effort, skills, hooks, memory, mcpServers, isolation, background, initialPrompt. Open struct (`...`).

**Command** (`internal/cue/schemas/command.cue`): name, description, allowed-tools, argument-hint, model, effort, disable-model-invocation, hooks, disallowed-tools (v2.1.152+, string or array of #KnownTool — removes tools from the model while the command is active). Open struct.

**Skill** (`internal/cue/schemas/skill.cue`): name (required), description (required, max 1536 chars — raised from 250 in v2.1.105), argument-hint, disable-model-invocation, user-invocable, allowed-tools, model, effort, context, agent, hooks, license, compatibility, metadata, disallowed-tools (v2.1.152+, string or array of #KnownTool — removes tools from the model while the skill is active). Open struct.

**Settings** (`internal/cue/schemas/settings.cue`): hooks, language, respectGitignore, plansDirectory, spinnerTipsOverride, enabledPlugins, extraKnownMarketplaces, strictKnownMarketplaces, blockedMarketplaces, disableAllHooks, autoMemoryDirectory, worktree, feedbackSurveyRate, sandbox, disableDeepLinkRegistration, cleanupPeriodDays, allowedChannelPlugins, showThinkingSummaries, disableSkillShellExecution, forceRemoteSettingsRefresh, refreshInterval, allowManagedHooksOnly, allowManagedDomainsOnly, allowManagedReadPathsOnly, channelsEnabled, skillOverrides, wslInheritsWindowsSettings, prUrlTemplate, tui, autoScrollEnabled, parentSettingsBehavior, autoMode. Open struct. Sandbox supports nested network.allowedDomains and network.deniedDomains (v2.1.113+). strictKnownMarketplaces/blockedMarketplaces are arrays of #MarketplaceSource entries (managed-settings gates, checked BEFORE download). wslInheritsWindowsSettings (v2.1.118+, bool, policy key). prUrlTemplate (v2.1.119+, string). mcpServers.*.alwaysLoad (v2.1.121+, bool, MCP server config — currently unmodeled open struct, passes silently). allowManagedDomainsOnly / allowManagedReadPathsOnly (v2.1.126+, bool, managed-settings policy keys gating sandbox.network.allowedDomains and sandbox.allowedReadPaths from non-managed sources). channelsEnabled (v2.1.128+, bool, managed-settings policy gating the `--channels` flow under console/API-key authentication). skillOverrides (v2.1.129+, map[string]string of skill name to visibility level — "on", "name-only", "user-invocable-only", or "off"). worktree.baseRef (v2.1.133+, enum: "fresh" | "head") chooses whether new worktrees branch from origin/<default> ("fresh", default) or local HEAD ("head"). sandbox.bwrapPath / sandbox.socatPath (v2.1.133+, strings, Linux/WSL only) point to custom bubblewrap and socat binaries via managed settings. parentSettingsBehavior (v2.1.133+, enum: "first-wins" | "merge", admin tier) controls whether SDK managedSettings (parent tier) merges into the policy chain — default "first-wins" preserves prior behavior. autoMode (v2.1.136+, open struct) carries managed auto-mode classifier config — autoMode.hard_deny ([...string]) is a list of permission-rule patterns that block unconditionally in auto mode regardless of user intent or allow exceptions. worktree.bgIsolation (v2.1.143+, enum: "none") opts background sessions out of EnterWorktree and edits the working copy directly. allowAllClaudeAiMcps (v2.1.149+, bool, enterprise managed setting that loads claude.ai cloud MCP connectors alongside managed-mcp.json). pluginSuggestionMarketplaces (v2.1.152+, array of #MarketplaceSource entries, managed setting allowlisting org marketplaces whose plugins may be suggested via context-aware tips). allowedMcpServers / deniedMcpServers (managed-settings MCP server allow/deny policies, modeled as arrays of strings — entry shape inferred). #MarketplaceSource gained a skipLfs field (v2.1.153+, bool, skips Git LFS download during github/git clone/update).

**Plugin** (`internal/lint/plugins.go` `knownPluginFields`): $schema (v2.1.120+), name (required), description (required), version, author, homepage, repository, license, keywords, readme, commands, agents, skills, hooks, mcpServers, outputStyles, lspServers, monitors, themes, experimental (v2.1.129+), defaultEnabled (v2.1.154+, bool — set false to disable plugin by default; enable via /plugin or claude plugin enable). Unknown fields emit a suggestion. In v2.1.129+, top-level themes/monitors are deprecated — prefer experimental.themes / experimental.monitors. cclint emits a suggestion when found at top level.

### CUE #KnownTool Union (shared across agent/command/skill schemas)

`Read`, `Write`, `Edit`, `MultiEdit`, `Bash`, `Grep`, `Glob`, `LS`, `Task`, `NotebookEdit`, `WebFetch`, `WebSearch`, `TodoWrite`, `BashOutput`, `KillBash`, `ExitPlanMode`, `AskUserQuestion`, `Agent`, `TaskCreate`, `TaskUpdate`, `TaskList`, `TaskGet`, `TaskStop`, `EnterPlanMode`, `EnterWorktree`, `ExitWorktree`, `KillShell`, `TaskOutput`, `LSP`, `Skill`, `DBClient`, `SendMessage`, `Monitor`, `RemoteTrigger`, `CronCreate`, `CronDelete`, `CronList`, `Workflow`, `ScheduleWakeup`, `PushNotification`, `REPL`

Note: CUE `#KnownTool` and Go `KnownTools` map are maintained separately and may diverge.

### Hook Handler Fields per Schema

**Settings** `#HookCommand`: type (`command`|`prompt`|`agent`|`http`|`mcp_tool`), command, args, prompt, async, timeout, once, continueOnBlock, statusMessage, url, headers, allowedEnvVars, if. Closed struct. `args` (v2.1.139+, [...string]) is the exec-form alternative to `command` — when type is `"command"`, either `command` (shell form) or `args` (exec form, no shell) satisfies. `continueOnBlock` (v2.1.139+, bool) is the PostToolUse-only option to feed the hook's rejection reason back to Claude and continue the turn.

**Agent** `#AgentHookCommand`: type (`command` only), command, args, timeout, once, continueOnBlock, if. Closed struct.

**Skill** `#SkillHookCommand`: type (`command` only), command, args, timeout, once, continueOnBlock, if. Closed struct.

**Command** `#CommandHookCommand`: type (`command` only), command, args, timeout, once, continueOnBlock, if. Closed struct.

### Valid Enum Values

**Colors**: red, blue, green, yellow, purple, orange, pink, cyan, gray, magenta, white

**Models**: sonnet, opus, haiku, sonnet[1m], opusplan, inherit, `claude-*` pattern (e.g. claude-opus-4-5) — all schemas

**Permission modes**: default, acceptEdits, delegate, dontAsk, bypassPermissions, plan

**Memory scopes**: user, project, local

**Marketplace sources**: github, git, git-subdir, url, npm, file, directory, hostPattern, settings

## Operational Context

- Documentation now has a role-based first-hop route: `README.md` -> `docs/README.md` -> `docs/setup.md` / `docs/connect-assistant.md` / `docs/change-cclint.md`.
- First-hop docs are flat and plain-language (`setup`, `common-tasks`, `connect-assistant`, `change-cclint`) while deep reference material remains under `docs/guides/`, `docs/reference/`, and `docs/rules/`.
- Rules documentation is canonicalized under `docs/rules/*.md`; do not recreate `docs/RULES.md`.
- `docs/` is instruction-only: historical snapshots and docs test specs were removed (`docs/ANTHROPIC_REQUIREMENTS.md`, `docs/tests/*.md`) and should stay outside `docs/`.
- v2.1.156 is a `next`/pre-release (npm `next` tag; `latest` remains v2.1.154, v2.1.155 skipped) with no published changelog. Its lint surface was derived deterministically from the bundled `sdk-tools.d.ts` tool definitions — which surfaced `ScheduleWakeup`, `PushNotification` (present since v2.1.152), and `REPL` as KnownTools that cclint was missing.

| Key | Value |
|-----|-------|
| claude_code_last_updated | v2.1.156 |
| valid_agent_colors | red, blue, green, yellow, purple, orange, pink, cyan, gray, magenta, white (11 total) |
| command_allowed_tools | Task, Agent, Skill, AskUserQuestion (delegation tools) — other tools are warnings |
| body_tool_mismatch_threshold | 8+ declared tools = general-purpose agent, check suppressed |
| knowntools_location | `internal/textutil/lineutil.go` exported `KnownTools` var (shared by tool validation and body scanning) |
| stale_binary_trap | Always `go build -o cclint .` before running `./cclint` — stale binary causes phantom results |
| golangci_lint_version | v2 config format required (`version: "2"` in `.golangci.yml`) |
| aic_cli | `aic claude` (Homebrew) fetches Claude Code changelog — primary method for `/updater`, fallback to `gh api` |
