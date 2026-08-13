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

### Hook Events — 29 total (`internal/lint/settings.go`)

`PreToolUse`, `PermissionRequest`, `PostToolUse`, `PostToolUseFailure`, `Notification`, `UserPromptSubmit`, `Stop`, `StopFailure`, `SubagentStart`, `SubagentStop`, `PreCompact`, `SessionStart`, `SessionEnd`, `TeammateIdle`, `TaskCompleted`, `TaskCreated`, `ConfigChange`, `WorktreeCreate`, `WorktreeRemove`, `InstructionsLoaded`, `PostCompact`, `Elicitation`, `ElicitationResult`, `CwdChanged`, `FileChanged`, `DirectoryAdded`, `PermissionDenied`, `Setup`, `MessageDisplay` (v2.1.152, transforms/hides assistant message text)

Component hooks (agents/skills) only: `PreToolUse`, `PostToolUse`, `Stop`, `SubagentStop`, `UserPromptSubmit`, `PreCompact`, `PostCompact`, `SessionStart`, `SessionEnd`

### KnownTools (`internal/textutil/lineutil.go`)

`Read`, `Write`, `Edit`, `MultiEdit`, `Bash`, `Grep`, `Glob`, `LS`, `Task`, `Agent`, `NotebookEdit`, `WebFetch`, `WebSearch`, `TodoWrite`, `AskUserQuestion`, `TaskCreate`, `TaskUpdate`, `TaskList`, `TaskGet`, `TaskStop`, `Skill`, `LSP`, `KillShell`, `TaskOutput`, `SendMessage`, `Monitor`, `RemoteTrigger`, `EnterPlanMode`, `ExitPlanMode`, `EnterWorktree`, `ExitWorktree`, `CronCreate`, `CronDelete`, `CronList`, `Workflow`, `ScheduleWakeup`, `PushNotification`, `REPL`, `EndConversation`, `*`

### CUE Schema Fields

**Agent** (`internal/cue/schemas/agent.cue`): name (required), description (required), model, color, tools, disallowedTools, permissionMode, maxTurns, effort, skills, hooks, memory, mcpServers, isolation, background, initialPrompt, requiredMcpServers (v2.1.156, [...string] — agent only runs when these MCP servers are connected), criticalSystemReminder_EXPERIMENTAL (v2.1.156, string — experimental reminder re-injected as a system message). Open struct (`...`). requiredMcpServers + criticalSystemReminder_EXPERIMENTAL are also in `internal/lint/agents.go` `knownAgentFields` (fixes a v2.1.156 false "unknown frontmatter field" suggestion).

**Command** (`internal/cue/schemas/command.cue`): name, description, allowed-tools, argument-hint, model, effort, disable-model-invocation, hooks, disallowed-tools (v2.1.152+, string or array of #KnownTool — removes tools from the model while the command is active). Open struct.

**Skill** (`internal/cue/schemas/skill.cue`): name (required), description (required, max 1536 chars — raised from 250 in v2.1.105), argument-hint, disable-model-invocation, user-invocable, allowed-tools, model, effort, context, background (v2.1.218+, bool), agent, hooks, license, compatibility, metadata, disallowed-tools (v2.1.152+, string or array of #KnownTool — removes tools from the model while the skill is active), version, cross_verify, triggers (legacy compatibility fields). Open struct.

**Settings** (`internal/cue/schemas/settings.cue`): hooks, language, respectGitignore, plansDirectory, spinnerTipsOverride, enabledPlugins, extraKnownMarketplaces, strictKnownMarketplaces, blockedMarketplaces, disableAllHooks, autoMemoryDirectory, worktree, feedbackSurveyRate, sandbox, crossSessionInbound, dialogExpiry, disableDeepLinkRegistration, cleanupPeriodDays, allowedChannelPlugins, showThinkingSummaries, disableSkillShellExecution, forceRemoteSettingsRefresh, refreshInterval, allowManagedHooksOnly, allowManagedDomainsOnly, allowManagedReadPathsOnly, channelsEnabled, skillOverrides, wslInheritsWindowsSettings, prUrlTemplate, tui, autoScrollEnabled, parentSettingsBehavior, autoMode, fallbackModel, requiredMinimumVersion, requiredMaximumVersion, disableBundledSkills. Open struct. Sandbox supports nested network.allowedDomains, network.deniedDomains, and network.strictAllowlist (v2.1.219+), plus boolean sub-keys sandbox.allowAppleEvents (v2.1.181+, macOS-only opt-in letting sandboxed commands send Apple Events) and sandbox.allowUnsandboxedCommands (v2.1.181+, policy-tier gate disabling the dangerouslyDisableSandbox escape hatch). sandbox.credentials (v2.1.187+, object blocking sandboxed commands from reading credential files / secret env vars — files: [...{path, mode:"deny"|"mask", extract?, onExtractNoMatch?, decode?:"jwt", maskClaims?:[...string]}], envVars: [...{name, mode:"deny"|"mask", extract?, onExtractNoMatch?, decode?:"jwt", maskClaims?:[...string], injectHosts?}], allowPlaintextInject: bool, awsPairs: [...{accessKeyIdVar, secretAccessKeyVar, sessionTokenVar?}], sigv4: {streaming?, presigned?, sigv4a?: "deny"|"passthrough"}). strictKnownMarketplaces/blockedMarketplaces are arrays of #MarketplaceSource entries (managed-settings gates, checked BEFORE download). wslInheritsWindowsSettings (v2.1.118+, bool, policy key). prUrlTemplate (v2.1.119+, string). mcpServers.*.alwaysLoad (v2.1.121+, bool, MCP server config — currently unmodeled open struct, passes silently). allowManagedDomainsOnly / allowManagedReadPathsOnly (v2.1.126+, bool, managed-settings policy keys gating sandbox.network.allowedDomains and sandbox.allowedReadPaths from non-managed sources). channelsEnabled (v2.1.128+, bool, managed-settings policy gating the `--channels` flow under console/API-key authentication). skillOverrides (v2.1.129+, map[string]string of skill name to visibility level — "on", "name-only", "user-invocable-only", or "off"). worktree.baseRef (v2.1.133+, enum: "fresh" | "head") chooses whether new worktrees branch from origin/<default> ("fresh", default) or local HEAD ("head"). sandbox.bwrapPath / sandbox.socatPath (v2.1.133+, strings, Linux/WSL only) point to custom bubblewrap and socat binaries via managed settings. parentSettingsBehavior (v2.1.133+, enum: "first-wins" | "merge", admin tier) controls whether SDK managedSettings (parent tier) merges into the policy chain — default "first-wins" preserves prior behavior. autoMode (v2.1.136+, open struct) carries managed auto-mode classifier config — autoMode.hard_deny ([...string]) is a list of permission-rule patterns that block unconditionally in auto mode regardless of user intent or allow exceptions. worktree.bgIsolation (v2.1.143+, enum: "none") opts background sessions out of EnterWorktree and edits the working copy directly. allowAllClaudeAiMcps (v2.1.149+, bool, enterprise managed setting that loads claude.ai cloud MCP connectors alongside managed-mcp.json). pluginSuggestionMarketplaces (v2.1.152+, array of #MarketplaceSource entries, managed setting allowlisting org marketplaces whose plugins may be suggested via context-aware tips). allowedMcpServers / deniedMcpServers (managed-settings MCP server allow/deny policies, modeled as arrays of strings — entry shape inferred). #MarketplaceSource gained a skipLfs field (v2.1.153+, bool, skips Git LFS download during github/git clone/update). autoMemoryEnabled / autoDreamEnabled (v2.1.156, bool — enable auto-memory for the project / enable background memory consolidation; modeled near autoMemoryDirectory). autoMode gained allow / soft_deny / deny / environment ([...string]) alongside the existing hard_deny — allow/soft_deny/deny tune the classifier's default disposition and environment carries free-form context about the user's setup; the struct keeps its trailing `...`. autoMode.classifyAllShell (v2.1.193+, bool) routes every Bash/PowerShell command through the auto-mode classifier while auto mode is active. The hook handler `#HookCommand` gained asyncRewake (v2.1.156, bool — when an async hook exits with code 2 it wakes the model and blocks the turn), rewakeMessage (string — custom prefix for the rewake system-reminder), and rewakeSummary (string — short summary shown to the user); these sit beside the existing `async` field and are settings-only (the agent/skill/command hook structs do not carry `async`, so they do not carry the rewake trio either). agent (v2.1.157, string — default subagent for dispatched sessions, overridable via --agent). requiredMinimumVersion / requiredMaximumVersion (v2.1.163+, string, semver — managed-settings-only version gates; Claude Code refuses to start outside the allowed range). fallbackModel (v2.1.166+, [...string] — fallback models tried in order when the primary is overloaded or unavailable; "default" expands to the default model; effective list capped at 3; CLI --fallback-model takes precedence). disableBundledSkills (v2.1.169+, bool — hides bundled skills, workflows, and built-in slash commands; env equivalent CLAUDE_CODE_DISABLE_BUNDLED_SKILLS). wheelScrollAccelerationEnabled (v2.1.174+, bool — ramp mouse-wheel scroll speed during fast scrolls, fullscreen mode only; general TUI setting). availableModels ([...string] — model allowlist of family aliases, version prefixes, or full model IDs; user/project/managed tiers; modeled v2.1.176 backfill). enforceAvailableModels (v2.1.175+, bool, managed/policy tier — when true and availableModels is non-empty, the Default model selection is also constrained to the allowlist). footerLinksRegexes (v2.1.176+, array of {type, pattern, url, label?} objects — extra clickable footer badges shown when a regex matches turn output; user or managed settings; modeled permissively as an array of open structs, NOT [...string]). Plus 44 long-standing top-level keys backfilled v2.1.177 (deterministically type-verified against the binary zod schema): auth/credential helpers apiKeyHelper, awsAuthRefresh, awsCredentialExport, gcpAuthRefresh, proxyAuthHelper, otelHeadersHelper (string); model, advisorModel, outputStyle (string), modelOverrides ({...} map); additionalDirectories, enabledMcpjsonServers, disabledMcpjsonServers, companyAnnouncements ([...string]); defaultMode/editorMode/preferredNotifChannel/teammateMode (enums modeled permissively as string — binary members not cleanly extractable); statusLine, subagentStatusLine, attribution ({...} open structs); autoCompactWindow (number); autoUpdatesChannel ("latest"|"stable"|"rc"), forceLoginMethod ("claudeai"|"console"|"gateway") (strict enums); includeCoAuthoredBy (bool, deprecated → attribution), autoCompactEnabled, fileCheckpointingEnabled, alwaysThinkingEnabled, verbose, hideVimModeIndicator, syntaxHighlightingDisabled, todoFeatureEnabled, spinnerTipsEnabled, promptSuggestionEnabled, remoteControlAtStartup, disableRemoteControl, skipDangerousModePermissionPrompt, skipAutoPermissionPrompt, skipWebFetchPreflight, enableAllProjectMcpServers, allowManagedMcpServersOnly, allowManagedPermissionRulesOnly, enableWorkflows, disableWorkflows (bool). Plus 26 more UI/UX/session keys (grep-verified v2.1.177): TUI toggles terminalProgressBarEnabled, showMessageTimestamps, showTurnDuration, terminalTitleFromRename, prefersReducedMotion, showClearContextOnPlanAccept, switchModelsOnFlag (bool) and defaultShell ("bash"|"powershell"); fast/auto/workflow toggles fastMode, fastModePerSessionOptIn, autoSubmit, useAutoModeDuringPlan, ultracode, workflowKeywordTriggerEnabled, respondToBashCommands (v2.1.186+, bool — keep context-only behavior after an !-bash command), workflowSizeGuideline (v2.1.202+, enum small|medium|large|unrestricted — dynamic-workflow agent-count target) (bool/enum); notification toggles awaySummaryEnabled, agentPushNotifEnabled, inputNeededNotifEnabled, voiceEnabled (bool); voice, breakReminder, quietHours ({...} open structs), breakThresholdMinutes, skillListingBudgetFraction, skillListingMaxDescChars (number); symlinkDirectories, sparsePaths ([...string]); emojiCompletionEnabled (v2.1.217, bool — emoji shortcode autocomplete in the prompt input), axScreenReader (v2.1.208, bool — opt-in plain-text rendering for screen-reader users), vimInsertModeRemaps (v2.1.208, {...} open map — maps two-key insert-mode sequences like jj→Escape in vim mode; binary shape v.record(string,unknown)), disableAutoMode (v2.1.207, top-level enum — only accepted value is the literal "disable", turns auto mode off; distinct from the autoMode.* classifier sub-struct), sandbox.filesystem.disabled (v2.1.216, bool sub-key of a new sandbox.filesystem object — skip filesystem isolation while keeping network/seccomp).

**Plugin** (`internal/lint/plugins.go` `knownPluginFields`): $schema (v2.1.120+), name (required), description (required), version, author, homepage, repository, license, keywords, readme, commands, agents, skills, hooks, mcpServers, outputStyles, lspServers, monitors, themes, experimental (v2.1.129+), defaultEnabled (v2.1.154+, bool — set false to disable plugin by default; enable via /plugin or claude plugin enable). Unknown fields emit a suggestion. In v2.1.129+, top-level themes/monitors are deprecated — prefer experimental.themes / experimental.monitors. cclint emits a suggestion when found at top level.

### CUE #KnownTool Union (shared across agent/command/skill schemas)

`Read`, `Write`, `Edit`, `MultiEdit`, `Bash`, `Grep`, `Glob`, `LS`, `Task`, `NotebookEdit`, `WebFetch`, `WebSearch`, `TodoWrite`, `BashOutput`, `KillBash`, `ExitPlanMode`, `AskUserQuestion`, `Agent`, `TaskCreate`, `TaskUpdate`, `TaskList`, `TaskGet`, `TaskStop`, `EnterPlanMode`, `EnterWorktree`, `ExitWorktree`, `KillShell`, `TaskOutput`, `LSP`, `Skill`, `DBClient`, `SendMessage`, `Monitor`, `RemoteTrigger`, `CronCreate`, `CronDelete`, `CronList`, `Workflow`, `ScheduleWakeup`, `PushNotification`, `REPL`, `EndConversation`

Note: the CUE `#KnownTool` union is GENERATED from Go's `textutil.KnownTools` at validator init (`internal/cue/knowntool.go` `knownToolUnionCUE()`, injected per-schema in `validator.go`) — it is no longer hand-maintained in the `.cue` files. The union = `KnownTools` minus the `*` wildcard, plus the schema-only extras `BashOutput`, `KillBash`, `DBClient` (the `schemaOnlyTools` var). To change the union, edit `KnownTools` or `schemaOnlyTools` in Go; `TestKnownToolUnionMatchesSource` guards the mapping. Single source of truth — Go and CUE cannot diverge.

### Intentionally Not Modeled (scope guardrail)

Surfaced by the v2.1.156 source-read mining pass but deliberately left out of cclint's schemas:

- `omitClaudeMd` — a built-in agent property / SDK option (e.g. `agentType:"Explore",...,omitClaudeMd:!0`), not a user-authored frontmatter key. Not modeled.
- `claudeMdProcessingMode` — absent from the v2.1.156 binary (zero hits). Not modeled.
- `yoloClassifier` — does not exist in the binary; a blog-side naming error. Not modeled.
- Hook stdout response fields (`updatedInput`, `permissionDecision`, `permissionDecisionReason`, `additionalContext`, `watchPaths`, `initialUserMessage`, `updatedMCPToolOutput`, `decision`, etc.) — these are a runtime hook-script stdout contract, not static configuration. They are outside cclint's static-config lint scope.

### Hook Handler Fields per Schema

**Settings** `#HookCommand`: type (`command`|`prompt`|`agent`|`http`|`mcp_tool`), command, args, prompt, async, timeout, once, continueOnBlock, statusMessage, url, headers, allowedEnvVars, if. Closed struct. `args` (v2.1.139+, [...string]) is the exec-form alternative to `command` — when type is `"command"`, either `command` (shell form) or `args` (exec form, no shell) satisfies. `continueOnBlock` (v2.1.139+, bool) is the PostToolUse-only option to feed the hook's rejection reason back to Claude and continue the turn.

**Agent** `#AgentHookCommand`: type (`command` only), command, args, timeout, once, continueOnBlock, if. Closed struct.

**Skill** `#SkillHookCommand`: type (`command` only), command, args, timeout, once, continueOnBlock, if. Closed struct.

**Command** `#CommandHookCommand`: type (`command` only), command, args, timeout, once, continueOnBlock, if. Closed struct.

### Valid Enum Values

**Colors**: red, blue, green, yellow, purple, orange, pink, cyan, gray, magenta, white

**Models**: sonnet, opus, haiku, fable, best, sonnet[1m], opus[1m], fable[1m], opusplan, inherit, claude-opus-5, `claude-*` pattern with optional bracket suffix (e.g. claude-opus-4-5, claude-fable-5[1m]) — all schemas. The CUE `#Model` union is GENERATED from Go (`internal/cue/model.go` `modelUnionCUE()`: `modelAliases` slice + `claudeModelRegexCUE`), injected per-schema in `validator.go` alongside `#KnownTool` — the four hand copies in agent/command/skill/claude_md.cue were removed. NOTE: Go's `validModelPattern` (`internal/lint/agent_fields.go`) is a SEPARATE, still-independent regex used for the agent-only model warning — it is not yet sourced from `modelAliases`, so it can drift from the CUE union (flagged follow-up).

**Permission modes**: default, acceptEdits, delegate, dontAsk, bypassPermissions, plan

**Memory scopes**: user, project, local

**Marketplace sources**: github, git, git-subdir, url, npm, file, directory, hostPattern, settings, archive, command, pathPattern, skills-dir

## Operational Context

- Documentation now has a role-based first-hop route: `README.md` -> `docs/README.md` -> `docs/setup.md` / `docs/connect-assistant.md` / `docs/change-cclint.md`.
- First-hop docs are flat and plain-language (`setup`, `common-tasks`, `connect-assistant`, `change-cclint`) while deep reference material remains under `docs/guides/`, `docs/reference/`, and `docs/rules/`.
- Rules documentation is canonicalized under `docs/rules/*.md`; do not recreate `docs/RULES.md`.
- `docs/` is instruction-only: historical snapshots and docs test specs were removed (`docs/ANTHROPIC_REQUIREMENTS.md`, `docs/tests/*.md`) and should stay outside `docs/`.
- v2.1.156 swept: pre-release (`next` tag, no published changelog; surface derived from bundled `sdk-tools.d.ts`). Modeled KnownTools `ScheduleWakeup`/`PushNotification`/`REPL`; agent fields `requiredMcpServers`, `criticalSystemReminder_EXPERIMENTAL`; command/skill `disallowed-tools`; hook event `MessageDisplay`; settings `dialogExpiry`.
- v2.1.157–v2.1.160 swept: modeled settings `agent` (v2.1.157). `ultracode` + `workflowKeywordTriggerEnabled` (keyword renamed workflow→ultracode in v2.1.160) both now modeled. 158/159/160 auto-mode/internal/security/UX bugfix — no other surface.
- v2.1.161–v2.1.173 swept (2.1.164/171 skipped): modeled `requiredMinimumVersion`/`requiredMaximumVersion` (v2.1.163), `fallbackModel` (v2.1.166), `disableBundledSkills` (v2.1.169); Fable-5 aliases `fable`/`best`/`fable[1m]`/`opus[1m]` + `[1m]`-suffixed `claude-*` IDs (v2.1.170/173).
- v2.1.174–v2.1.176 swept: modeled `wheelScrollAccelerationEnabled` (v2.1.174), `enforceAvailableModels` + backfilled `availableModels` (v2.1.175), `footerLinksRegexes` (v2.1.176). Installed binary was v2.1.177 (ahead of published changelog).
- v2.1.177–v2.1.178 swept (2026-06-15): v2.1.177 had no changelog entry. v2.1.178 added `Tool(param:value)` permission-rule syntax + `mcp__` spec handling — NO lint surface (permission rules never schema-checked; disallowed fields accept bare string). Closed two false-positive tripwires: `permissions.ask` accepted; `mcp__`/`Tool()` tokens in array-form tool fields.
- v2.1.179 swept (2026-06-16): bugfix-only, no lint surface.
- v2.1.180–v2.1.181 swept (2026-06-17): modeled `sandbox.allowAppleEvents` + `sandbox.allowUnsandboxedCommands` (v2.1.181). v2.1.180 absent header.
- v2.1.182–v2.1.183 swept (2026-06-18): bugfix-only, no lint surface.
- v2.1.184–v2.1.185 swept (2026-06-20): UI/runtime-only (stream-stall hint rewording), no lint surface.
- v2.1.186–v2.1.204 swept (2026-07-08): modeled `respondToBashCommands` (v2.1.186), `sandbox.credentials` (v2.1.187), `autoMode.classifyAllShell` (v2.1.193), `workflowSizeGuideline` (v2.1.202). Bugfix-only: 2.1.190/197/201.
- v2.1.205–v2.1.217 swept (2026-07-22): modeled settings `disableAutoMode`, `axScreenReader`, `vimInsertModeRemaps`, `emojiCompletionEnabled`, `sandbox.filesystem.disabled`; KnownTool `EndConversation` (v2.1.214). Bugfix/UI/env-only: 2.1.205/206/209/210/211/212/215.
- v2.1.218–v2.1.226 swept (2026-08-08): modeled skill `background` + case-insensitive bool literals (v2.1.218); `claude-opus-5`, `sandbox.network.strictAllowlist`, hook event `DirectoryAdded` (v2.1.219); credential `mode:"mask"` (v2.1.221); `archive`/`crossSessionInbound`/`dialogExpiry`/structured credential masking/`awsPairs`/SigV4 policy enums (v2.1.224). Bugfix-only: 2.1.220/226.
- v2.1.227–v2.1.228 swept (2026-08-11): bugfix-only, no lint surface.
- v2.1.229–v2.1.231 swept (2026-08-13): modeled marketplace sources `command` (v2.1.229 changelog), `pathPattern`/`skills-dir` (reconciled from v2.1.231 binary) in `#MarketplaceSource`. v2.1.230 absent header; v2.1.231 OAuth-runtime-only. Dropped: server-supplied self-hosted-runner hooks (runner-internal, no settings key). Follow-up: v2.1.231 splits marketplace (`Sln`) vs plugin (`_3c`) source unions in the binary; cclint models a conservative superset (added only, kept `git-subdir`/`archive`) pending reliable disambiguation.
- CUE `#KnownTool` and `#Model` unions are generated from Go (`KnownTools`/`schemaOnlyTools` and `modelAliases`/`claudeModelRegexCUE`) at validator init — see Validation Surface for detail. Open unification follow-ups: Go `validModelPattern` (`agent_fields.go`) duplicates the model alias set; `#MarketplaceSource` has no Go source; per-component hook-command defs (`#Agent/#Command/#SkillHookCommand`) are byte-identical bodies under distinct names.
|-----|-------|
| claude_code_last_updated | v2.1.231 (binary installed: v2.1.231) |
| valid_agent_colors | red, blue, green, yellow, purple, orange, pink, cyan, gray, magenta, white (11 total) |
| command_allowed_tools | Task, Agent, Skill, AskUserQuestion (delegation tools) — other tools are warnings |
| body_tool_mismatch_threshold | 8+ declared tools = general-purpose agent, check suppressed |
| knowntools_location | `internal/textutil/lineutil.go` exported `KnownTools` var (shared by tool validation and body scanning) |
| stale_binary_trap | Always `go build -o cclint .` before running `./cclint` — stale binary causes phantom results |
| golangci_lint_version | v2 config format required (`version: "2"` in `.golangci.yml`) |
| updater_changelog_fetch | `gh api repos/anthropics/claude-code/contents/CHANGELOG.md` with base64 decoding is primary for `/updater`; `aic claude` is fallback only |
