# CLI Execution Flows

## Overview

| Metric | Value |
|--------|-------|
| Commands documented | 7/7 |
| Entry points | 1 (main.go → cmd.Execute()) |
| CLI framework | spf13/cobra |
| External dependencies | 5 (cobra, viper, doublestar, lipgloss, cuelang) |

## Entry Point

**File**: `/Users/vampire/go/src/cclint/main.go`

```
main() → cmd.Execute()
```

The main package is minimal - it delegates immediately to the cmd package's Execute() function.

---

## Command Hierarchy

### Root Command: `cclint`

**File**: `/Users/vampire/go/src/cclint/cmd/root.go`

**Entry**: `rootCmd.Run()`

**Mode Detection Logic**:
```
collectFilesToLint(args) → determines mode
├─ len(filesToLint) > 0 → runSingleFileLint()
├─ diffMode || stagedMode → runGitLint()
└─ else → runLint()
```

**Flags**:
- Global: `--root`, `--quiet`, `--verbose`, `--scores`, `--improvements`, `--format`, `--output`, `--fail-on`, `--no-cycle-check`
- Baseline: `--baseline`, `--baseline-create`, `--baseline-path`
- Single-file: `--file`, `--type`
- Git integration: `--diff`, `--staged`

**Configuration Loading**:
```
initConfig() (cobra.OnInitialize)
├─ Searches for .cclintrc.{json,yaml,yml}
├─ Loads with viper.ReadInConfig()
└─ Binds flags to viper
```

---

## Execution Flows

### Flow 1: Full Scan Mode (Default)

**Command**: `cclint` or `cclint agents`

**Path**: `rootCmd.Run() → runLint()`

```
runLint()
├─ config.LoadConfig(rootPath)
│  ├─ Sets defaults (root=~/.claude, format=console)
│  ├─ Reads .cclintrc.{json,yaml,yml}
│  └─ Applies env vars (CCLINT_*)
│
├─ baseline.LoadBaseline() [if --baseline or --baseline-create]
│  └─ Reads .cclintbaseline.json
│
├─ outputters.NewOutputter(cfg)
│
├─ Run all linters in sequence:
│  ├─ cli.LintAgents()
│  ├─ cli.LintCommands()
│  ├─ cli.LintSkills()
│  ├─ cli.LintSettings()
│  ├─ cli.LintContext()
│  └─ cli.LintPlugins()
│
├─ For each linter:
│  ├─ linter(cfg.Root, cfg.Quiet, cfg.Verbose, cfg.NoCycleCheck)
│  ├─ Skip if summary.TotalFiles == 0
│  ├─ Collect issues (if --baseline-create)
│  ├─ cli.FilterResults(summary, baseline) [if --baseline]
│  ├─ outputter.Format(summary, cfg.Format)
│  └─ Accumulate totals
│
├─ baseline.CreateBaseline() [if --baseline-create]
│  ├─ Fingerprints all issues (SHA256 of file+source+normalized message)
│  ├─ Saves to .cclintbaseline.json
│  └─ Exit 0 (accept current state)
│
├─ Print baseline filtering summary [if --baseline]
│
└─ Exit 1 if hasErrors, else Exit 0
```

**External Interactions**:
- File I/O: Read component files from discovered paths
- Config files: `.cclintrc.{json,yaml,yml}`, `.cclintbaseline.json`
- Environment: `CCLINT_*` variables

**Error Flows**:
- Config load failure → stderr + Exit 1
- Linter execution failure → stderr + Exit 1
- Baseline save failure → stderr + Exit 1

---

### Flow 2: Single-File Mode

**Command**: `cclint ./agents/my-agent.md` or `cclint --file agents`

**Path**: `rootCmd.Run() → runSingleFileLint(files)`

```
runSingleFileLint(files)
├─ config.LoadConfig(rootPath)
│
├─ cli.LintFiles(files, rootPath, typeFlag, quiet, verbose)
│  ├─ For each file:
│  │  ├─ discovery.ValidateFilePath(file)
│  │  │  ├─ Check file exists
│  │  │  ├─ Check is file (not directory)
│  │  │  ├─ Check readable
│  │  │  └─ Check not binary
│  │  │
│  │  ├─ discovery.DetectFileType(absPath, rootPath) [if typeFlag empty]
│  │  │  ├─ Compute relative path
│  │  │  ├─ Match against typePatterns (glob patterns)
│  │  │  └─ Fallback to basename matching
│  │  │
│  │  ├─ os.ReadFile(absPath)
│  │  │
│  │  └─ cli.lintSingleFile(file)
│  │     ├─ Create SingleFileLinterContext
│  │     ├─ Select linter by type (agent/command/skill/etc.)
│  │     └─ lintComponent(ctx, linter)
│  │
│  └─ Aggregate results into LintSummary
│
├─ outputter.Format(summary, cfg.Format)
│
└─ Exit 1 if summary.TotalErrors > 0, else Exit 0
```

**Type Detection Priority**:
1. `--type` flag (explicit override)
2. Glob pattern match: `.claude/agents/**/*.md`, `skills/**/SKILL.md`, etc.
3. Basename match: `SKILL.md`, `CLAUDE.md`, `settings.json`, `plugin.json`
4. Error with actionable message

**Error Flows**:
- File not found → stderr + Exit 2
- Type detection failure → stderr + Exit 2
- Invalid type flag → stderr + Exit 2

---

### Flow 3: Git Integration Mode

**Command**: `cclint --staged` or `cclint --diff`

**Path**: `rootCmd.Run() → runGitLint()`

```
runGitLint()
├─ config.LoadConfig(rootPath)
│
├─ git.IsGitRepo(gitRoot)
│  └─ exec.Command("git", "rev-parse", "--git-dir").Run()
│
├─ [If not git repo] → Fallback to runLint()
│
├─ Get files from git:
│  ├─ [If --staged] git.GetStagedFiles(gitRoot)
│  │  └─ exec.Command("git", "diff", "--name-only", "--staged")
│  │
│  └─ [If --diff] git.GetChangedFiles(gitRoot)
│     ├─ Check if any commits: exec.Command("git", "rev-parse", "HEAD")
│     ├─ [If no commits] exec.Command("git", "ls-files")
│     └─ [Else] exec.Command("git", "diff", "--name-only", "HEAD")
│
├─ filterRelevantFiles(gitOutput, rootPath)
│  ├─ Filter by extension: .md or .json
│  ├─ Filter by path: agents/, commands/, skills/, .claude/
│  └─ Check file exists (skip deletions)
│
├─ [If no files] → Print "No files to lint" + Exit 0
│
├─ cli.LintFiles(files, gitRoot, "", quiet, verbose)
│
├─ outputter.Format(summary, cfg.Format)
│
└─ Exit 1 if summary.TotalErrors > 0, else Exit 0
```

**External Interactions**:
- Git commands: `git rev-parse`, `git diff`, `git ls-files`
- Working directory: Uses git root for relative paths

**Error Flows**:
- Git command failure → stderr + Exit 1
- Not in git repo → Warning + Fallback to full scan

---

### Flow 4: Component-Specific Subcommands

**Commands**: `cclint agents`, `cclint commands`, `cclint skills`, `cclint settings`, `cclint context`, `cclint plugins`

**Files**:
- `/Users/vampire/go/src/cclint/cmd/agents.go`
- `/Users/vampire/go/src/cclint/cmd/commands.go`
- `/Users/vampire/go/src/cclint/cmd/skills.go`
- `/Users/vampire/go/src/cclint/cmd/settings.go`
- `/Users/vampire/go/src/cclint/cmd/context.go`
- `/Users/vampire/go/src/cclint/cmd/plugins.go`

**Path**: `agentsCmd.Run() → runAgentsLint() → runComponentLint("agents", cli.LintAgents)`

```
runComponentLint(linterName, linterFunc)
├─ config.LoadConfig(rootPath)
│
├─ baseline.LoadBaseline() [if --baseline or --baseline-create]
│
├─ linterFunc(cfg.Root, cfg.Quiet, cfg.Verbose, cfg.NoCycleCheck)
│  └─ (See "Component Linter Flow" below)
│
├─ cli.CollectAllIssues(summary) [if --baseline-create]
│
├─ cli.FilterResults(summary, baseline) [if --baseline]
│
├─ outputter.Format(summary, cfg.Format)
│
├─ baseline.CreateBaseline() + SaveBaseline() [if --baseline-create]
│  └─ Exit 0 after baseline creation
│
└─ Print baseline summary [if --baseline]
```

**Linter Functions**:
- `cli.LintAgents()` → NewAgentLinter()
- `cli.LintCommands()` → NewCommandLinter()
- `cli.LintSkills()` → NewSkillLinter()
- `cli.LintSettings()` → NewSettingsLinter()
- `cli.LintContext()` → NewContextLinter()
- `cli.LintPlugins()` → NewPluginLinter()

---

### Flow 5: Format Command

**Command**: `cclint fmt` or `cclint fmt --write`

**File**: `/Users/vampire/go/src/cclint/cmd/fmt.go`

**Path**: `fmtCmd.Run() → runFmt(args)`

```
runFmt(args)
├─ config.LoadConfig(rootPath)
│
├─ collectFilesToFormat(args, cfg.Root)
│  ├─ [If --file flag] → Use explicit files
│  ├─ [If path args] → Walk directories or use files directly
│  ├─ [If component type arg] → discoverFilesByType(rootPath, type)
│  └─ [Else] → discoverAllFiles(rootPath)
│
├─ For each file:
│  ├─ discovery.ValidateFilePath(filePath)
│  ├─ discovery.DetectFileType(absPath, cfg.Root) [if typeFlag empty]
│  ├─ os.ReadFile(absPath)
│  ├─ format.NewComponentFormatter(fileType.String())
│  ├─ formatter.Format(content)
│  │  ├─ Normalize frontmatter field order
│  │  ├─ Ensure blank line after frontmatter
│  │  ├─ Trim trailing whitespace
│  │  └─ Ensure file ends with single newline
│  │
│  └─ [If content changed]:
│     ├─ [If --check] → Track + print "needs formatting"
│     ├─ [If --diff] → format.Diff(original, formatted)
│     ├─ [If --write] → os.WriteFile(absPath, formatted, 0644)
│     └─ [Else] → Print to stdout
│
├─ Print summary [if not quiet]
│
└─ [If --check && needsFormatting] → Exit 1
```

**Output Modes**:
- Default: Print formatted content to stdout
- `--write`: Overwrite files in place
- `--diff`: Show unified diff
- `--check`: Exit 1 if formatting needed (CI mode)

**Error Flows**:
- No files to format → stderr + Exit 1
- File read error → Skip file + warning
- Type detection failure → Skip file + warning

---

### Flow 6: Summary Command

**Command**: `cclint summary`

**File**: `/Users/vampire/go/src/cclint/cmd/summary.go`

**Path**: `summaryCmd.Run() → runSummary()`

```
runSummary()
├─ config.LoadConfig(rootPath)
│
├─ Run all linters (quiet=true):
│  ├─ cli.LintAgents() → aggregateResults()
│  ├─ cli.LintCommands() → aggregateResults()
│  └─ cli.LintSkills() → aggregateResults()
│
├─ aggregateResults() (for each linter):
│  ├─ Accumulate component counts
│  ├─ Count tier distribution (A/B/C/D/F)
│  ├─ Build lowest-scoring component list
│  └─ Categorize issues into buckets
│
├─ Sort lowest-scoring components by score
│
└─ printSummaryReport()
   ├─ Component counts by type
   ├─ Quality distribution (tier percentages + bars)
   ├─ Top 5 issues (with counts)
   └─ Lowest 5 scoring components
```

**Issue Categorization**:
- Oversized component
- Missing Foundation/Workflow/Anti-Patterns sections
- Missing semantic routing
- Embedded methodology
- Missing triggers/PROACTIVELY pattern
- Missing model specification

**Output**: Styled box-drawn report using lipgloss (always console format)

---

## Component Linter Flow

**File**: `/Users/vampire/go/src/cclint/internal/cli/generic_linter.go`

All component-specific linters follow this unified pipeline:

```
lintBatch(ctx, linter)
├─ ctx.FilterFilesByType(linter.FileType())
│  └─ discovery.NewFileDiscovery(rootPath, verbose)
│     ├─ DiscoverFiles() using doublestar glob patterns
│     │  ├─ .claude/agents/**/*.md
│     │  ├─ .claude/commands/**/*.md
│     │  ├─ .claude/skills/**/SKILL.md
│     │  ├─ .claude/settings.json
│     │  ├─ .claude/CLAUDE.md
│     │  └─ **/.claude-plugin/plugin.json
│     └─ Read file contents
│
├─ For each file: lintBatchFile(ctx, file, linter)
│  │
│  ├─ [Optional] linter.PreValidate() (ISP interface)
│  │  └─ Filename checks, empty content checks
│  │
│  ├─ linter.ParseContent(contents)
│  │  ├─ [If .md] frontend.ParseYAMLFrontmatter()
│  │  └─ [If .json] json.Unmarshal()
│  │
│  ├─ linter.ValidateCUE(validator, data)
│  │  └─ cue.Validator with embedded schemas
│  │     ├─ internal/cue/schemas/agent.cue
│  │     ├─ internal/cue/schemas/command.cue
│  │     ├─ internal/cue/schemas/settings.cue
│  │     └─ internal/cue/schemas/plugin.cue
│  │
│  ├─ linter.ValidateSpecific(data, filePath, contents)
│  │  └─ Component-specific validation:
│  │     ├─ [Agents] validateAgentSpecific()
│  │     │  ├─ Required fields: name, description
│  │     │  ├─ Name format: lowercase alphanumeric + hyphens
│  │     │  ├─ Color validation
│  │     │  ├─ Tool field naming (tools: not allowed-tools:)
│  │     │  └─ Reserved words check
│  │     │
│  │     ├─ [Commands] validateCommandSpecific()
│  │     │  ├─ Name format (if present)
│  │     │  └─ Tool field naming (allowed-tools: not tools:)
│  │     │
│  │     ├─ [Skills] validateSkillSpecific()
│  │     │  └─ SKILL.md filename requirement
│  │     │
│  │     └─ [Settings/Context/Plugins] Type-specific rules
│  │
│  ├─ [Optional] linter.ValidateBestPractices() (ISP interface)
│  │  ├─ XML tag detection in descriptions
│  │  ├─ Size limits (200 lines for agents, 50 for commands, 500 for skills)
│  │  ├─ Bloat section detection (Quick Reference, Usage, etc.)
│  │  ├─ Inline methodology detection (scoring formulas, patterns)
│  │  ├─ Missing model specification
│  │  ├─ Skill loading pattern check
│  │  ├─ PROACTIVELY pattern in description
│  │  └─ permissionMode for editing tools
│  │
│  ├─ [Optional] linter.ValidateCrossFile() (ISP interface)
│  │  └─ crossValidator.ValidateSkillReferences()
│  │     ├─ Find "Skill: skill-name" patterns
│  │     ├─ Verify skill exists in discovered files
│  │     └─ Report orphaned skills
│  │
│  ├─ detectSecrets(contents, filePath)
│  │  └─ Warn on potential API keys/secrets
│  │
│  ├─ [Optional] linter.Score() (ISP interface)
│  │  └─ scoring.{Agent|Command|Skill|Plugin}Scorer.Score()
│  │     ├─ Structural score (0-40): sections, frontmatter fields
│  │     ├─ Practices score (0-40): size, patterns, methodology
│  │     ├─ Composition score (0-10): delegation, reusability
│  │     ├─ Documentation score (0-10): clarity, completeness
│  │     └─ Overall = sum, Tier = A≥85, B≥70, C≥50, D≥30, F<30
│  │
│  ├─ [Optional] linter.GetImprovements() (ISP interface)
│  │  └─ Point-valued recommendations (e.g., "+5: Add model specification")
│  │
│  ├─ [Optional] linter.PostProcess() (ISP interface)
│  │
│  └─ Return LintResult
│     ├─ File, Type, Success
│     ├─ Errors, Warnings, Suggestions
│     ├─ Quality (score + tier)
│     └─ Improvements
│
├─ [Optional] linter.PostProcessBatch() (ISP interface)
│  └─ [Agents] Circular dependency detection
│     ├─ Build Task() delegation graph
│     ├─ Tarjan's algorithm for cycle detection
│     └─ Inject cycle errors into results
│
└─ Return LintSummary
   ├─ TotalFiles, SuccessfulFiles, FailedFiles
   ├─ TotalErrors, TotalWarnings, TotalSuggestions
   └─ Results[]
```

**Interface Segregation Principle**:
- Core interface: `ComponentLinter` (Type, FileType, ParseContent, ValidateCUE, ValidateSpecific)
- Optional interfaces:
  - `PreValidator`: Pre-validation checks
  - `BestPracticeValidator`: Best practice checks
  - `CrossFileValidatable`: Cross-file reference validation
  - `Scorable`: Quality scoring
  - `Improvable`: Improvement recommendations
  - `PostProcessable`: Result post-processing
  - `BatchPostProcessor`: Batch-level post-processing

**Component-Specific Linters**:
- `/Users/vampire/go/src/cclint/internal/cli/agent_linter.go` → AgentLinter
- `/Users/vampire/go/src/cclint/internal/cli/command_linter.go` → CommandLinter
- `/Users/vampire/go/src/cclint/internal/cli/skill_linter.go` → SkillLinter
- `/Users/vampire/go/src/cclint/internal/cli/settings_linter.go` → SettingsLinter
- `/Users/vampire/go/src/cclint/internal/cli/context_linter.go` → ContextLinter
- `/Users/vampire/go/src/cclint/internal/cli/plugin_linter.go` → PluginLinter

---

## Output Flow

**File**: `/Users/vampire/go/src/cclint/internal/outputters/outputter.go`

```
outputter.Format(summary, format)
├─ [If format == "console"]
│  └─ output.FormatConsole(summary, cfg)
│     ├─ For each result:
│     │  ├─ Print file path + status (✓ or ✗)
│     │  ├─ Print errors (red)
│     │  ├─ Print warnings (yellow)
│     │  ├─ Print suggestions (cyan)
│     │  ├─ [If --scores] Print quality score + tier
│     │  └─ [If --improvements] Print improvement recommendations
│     │
│     └─ Print summary: X/Y passed, N errors, M suggestions
│
├─ [If format == "json"]
│  └─ output.FormatJSON(summary, cfg)
│     ├─ json.MarshalIndent(summary)
│     └─ [If -o -] Write to stdout, else os.WriteFile()
│
└─ [If format == "markdown"]
   └─ output.FormatMarkdown(summary, cfg)
      ├─ Render markdown tables
      └─ [If -o -] Write to stdout, else os.WriteFile()
```

**Output Formatters**:
- `/Users/vampire/go/src/cclint/internal/output/console.go` → FormatConsole (lipgloss styling)
- `/Users/vampire/go/src/cclint/internal/output/json.go` → FormatJSON
- `/Users/vampire/go/src/cclint/internal/output/markdown.go` → FormatMarkdown

---

## Key Dependencies

### External Dependencies

| Package | Purpose | Used By |
|---------|---------|---------|
| `spf13/cobra` | CLI framework | cmd/root.go, all subcommands |
| `spf13/viper` | Configuration management | cmd/root.go, internal/config/config.go |
| `bmatcuk/doublestar/v4` | Glob pattern matching | internal/discovery/discovery.go |
| `charmbracelet/lipgloss` | Terminal styling | internal/output/console.go, cmd/summary.go |
| `cuelang.org/go` | Schema validation | internal/cue/validator.go |

### Internal Dependencies

| Package | Purpose |
|---------|---------|
| `internal/discovery` | File discovery with glob patterns |
| `internal/frontend` | YAML frontmatter parsing |
| `internal/cue` | CUE schema validation |
| `internal/scoring` | Quality score calculation |
| `internal/cli` | Generic linting pipeline + component-specific linters |
| `internal/baseline` | Baseline fingerprinting + filtering |
| `internal/config` | Configuration loading (viper) |
| `internal/git` | Git integration (diff, staged) |
| `internal/output` | Output formatters (console, json, markdown) |
| `internal/format` | Component formatting |

---

## Verification Status

### Command Coverage

| Command | Path Verified | Deps Verified |
|---------|---------------|---------------|
| `cclint` (default) | ✓ | ✓ |
| `cclint agents` | ✓ | ✓ |
| `cclint commands` | ✓ | ✓ |
| `cclint skills` | ✓ | ✓ |
| `cclint settings` | ✓ | ✓ |
| `cclint context` | ✓ | ✓ |
| `cclint plugins` | ✓ | ✓ |
| `cclint fmt` | ✓ | ✓ |
| `cclint summary` | ✓ | ✓ |
| `cclint --staged` | ✓ | ✓ |
| `cclint --diff` | ✓ | ✓ |
| `cclint file.md` | ✓ | ✓ |

### Execution Path Verification

**Sample 1: `cclint agents`**
- Entry: main.go:6 → cmd.Execute()
- Command: cmd/agents.go:27 → runAgentsLint()
- Generic: cmd/lint.go:24 → runComponentLint("agents", cli.LintAgents)
- Linter: internal/cli/agents.go:41 → LintAgents() → lintBatch()
- Pipeline: internal/cli/generic_linter.go:219 → lintBatch() → lintBatchFile()
- Validation: internal/cli/agents.go:62 → validateAgentSpecific()
- Output: internal/outputters/outputter.go → Format() → console/json/markdown
- ✓ **Path Verified**

**Sample 2: `cclint --staged`**
- Entry: main.go:6 → cmd.Execute()
- Root: cmd/root.go:112 → runGitLint()
- Git: internal/git/diff.go:14 → GetStagedFiles()
- Exec: os/exec.Command("git", "diff", "--name-only", "--staged")
- Filter: internal/git/diff.go:78 → filterRelevantFiles()
- Lint: internal/cli/singlefile.go → LintFiles()
- ✓ **Path Verified**

**Sample 3: `cclint ./agents/foo.md`**
- Entry: main.go:6 → cmd.Execute()
- Root: cmd/root.go:106 → runSingleFileLint(filesToLint)
- Files: cmd/root.go:322 → collectFilesToLint(args)
- Validation: internal/discovery/discovery.go:147 → ValidateFilePath()
- Type: internal/discovery/discovery.go:70 → DetectFileType()
- Pipeline: internal/cli/generic_linter.go:95 → lintComponent()
- ✓ **Path Verified**

### Dependency Validation

| Dependency | Import Path | Used In | Verified |
|------------|-------------|---------|----------|
| cobra | github.com/spf13/cobra | cmd/*.go | ✓ |
| viper | github.com/spf13/viper | cmd/root.go, internal/config/ | ✓ |
| doublestar | github.com/bmatcuk/doublestar/v4 | internal/discovery/ | ✓ |
| lipgloss | github.com/charmbracelet/lipgloss | internal/output/, cmd/summary.go | ✓ |
| cuelang | cuelang.org/go | internal/cue/ | ✓ |

### Error Flow Coverage

| Error Type | Handler | Exit Code |
|------------|---------|-----------|
| Config load failure | cmd/root.go:196 | 1 |
| Linter execution failure | cmd/root.go:242 | 1 |
| Baseline save failure | cmd/lint.go:79 | 1 |
| File not found | cmd/root.go:109, 352 | 2 |
| Invalid type flag | internal/discovery/discovery.go:70 | 2 |
| Git command failure | cmd/root.go:425 | 1 |
| Lint errors found | cmd/root.go:306, 380, 453 | 1 |
| Format check failed | cmd/fmt.go:218 | 1 |

---

## Architecture Notes

### Design Patterns

**Interface Segregation Principle (ISP)**:
- Core `ComponentLinter` interface is minimal (5 methods)
- Optional capabilities via separate interfaces (PreValidator, BestPracticeValidator, etc.)
- Linters implement only the interfaces they need

**Template Method Pattern**:
- `lintBatch()` and `lintComponent()` are generic orchestrators
- Component-specific logic injected via ComponentLinter interface
- Consistent pipeline across all component types

**Strategy Pattern**:
- Different output formatters (console, json, markdown)
- Different linter implementations (agent, command, skill, etc.)
- Different scorers (AgentScorer, CommandScorer, etc.)

**Factory Pattern**:
- `NewAgentLinter()`, `NewCommandLinter()`, etc. construct linters
- `NewComponentFormatter()` constructs formatters by type

### File Discovery

**Glob Patterns** (via doublestar):
- `.claude/agents/**/*.md` → Agent files
- `.claude/commands/**/*.md` → Command files
- `.claude/skills/**/SKILL.md` → Skill files (requires exact filename)
- `.claude/settings.json` → Settings file
- `.claude/CLAUDE.md` → Context file
- `**/.claude-plugin/plugin.json` → Plugin manifests

**Fallback Matching**:
- Basename matching for files outside standard directories
- `SKILL.md`, `CLAUDE.md`, `settings.json`, `plugin.json`

### Validation Pipeline

**Stages**:
1. Pre-validation (filename, empty content)
2. Content parsing (YAML frontmatter or JSON)
3. CUE schema validation (embedded schemas)
4. Component-specific validation (field format, naming)
5. Best practice checks (size, patterns, structure)
6. Cross-file validation (skill references, cycles)
7. Secrets detection
8. Quality scoring (0-100 + tier)
9. Improvement recommendations

**Severity Levels**:
- `error`: Blocks build (exit 1)
- `warning`: Does not block build
- `suggestion`: Best practice recommendations
- `info`: Informational messages

### Baseline System

**Purpose**: Gradual adoption - accept current state, fail only on new issues

**Fingerprinting**:
```
SHA256(file + source + normalizedMessage)
```
- Line numbers ignored (stable across code shifts)
- Message normalization: Remove dynamic parts (numbers, file paths)

**Workflow**:
1. `cclint --baseline-create` → Generate `.cclintbaseline.json`
2. Commit baseline to version control
3. `cclint --baseline` in CI → Filter known issues
4. Only new issues fail the build

**Filtering**:
- `cli.FilterResults()` matches issue fingerprints
- Removes baseline issues from results
- Prints summary: "N baseline issues ignored (X errors, Y suggestions)"

---

## Environment Variables

| Variable | Purpose | Default |
|----------|---------|---------|
| `CCLINT_ROOT` | Project root directory | `~/.claude` |
| `CCLINT_FORMAT` | Output format | `console` |
| `CCLINT_QUIET` | Suppress non-essential output | `false` |
| `CCLINT_VERBOSE` | Enable verbose output | `false` |
| `CCLINT_FAIL_ON` | Fail build on level | `error` |

---

## Exit Codes

| Code | Meaning | Triggered By |
|------|---------|--------------|
| 0 | Success (no errors) | All validations pass |
| 0 | Baseline created | `--baseline-create` (always exit 0) |
| 1 | Lint errors found | Validation failures |
| 1 | Git command failure | `git diff`, `git rev-parse` errors |
| 1 | Linter execution failure | Internal errors |
| 1 | Format check failed | `cclint fmt --check` with unformatted files |
| 2 | Invocation error | File not found, invalid type, invalid args |

---

## Notable Implementation Details

### Go Regex Quirk (Cross-File Validation)

**File**: `/Users/vampire/go/src/cclint/internal/cli/crossfile.go`

```go
// Pattern: ^[^*\n]*\bSkill:\s*([a-z0-9][a-z0-9-]*)
// Key: [^*\n]* prevents matching across newlines
```

**Issue**: Character classes like `[^*]` match newlines by default in Go
**Solution**: Use `[^*\n]` to exclude newlines and prevent greedy cross-line matching

### CLI stdout Convention

**File**: `/Users/vampire/go/src/cclint/cmd/fmt.go`

When accepting `-o`/`--output` flags, always check `if outputFile == "-"` before `os.WriteFile` to write to stdout instead of creating a literal file named `-`.

**Implementation**:
```go
// internal/output/json.go, markdown.go
if cfg.Output == "-" {
    fmt.Print(output)
} else {
    os.WriteFile(cfg.Output, []byte(output), 0644)
}
```

### Symlink Build Pattern

**File**: `/Users/vampire/go/src/cclint/CLAUDE.md`

```bash
go build -o cclint . && ln -sf $(pwd)/cclint ~/go/bin/cclint
```

Symlink instead of copy allows in-place rebuilds without updating `~/go/bin`.

---

## Cross-Cutting Concerns

### Logging

- Verbose mode: File processing details via `ctx.LogProcessed()`
- Quiet mode: Suppress non-essential output
- Errors always go to stderr

### Error Handling

- Config errors: `fmt.Errorf("error loading configuration: %w", err)`
- Linter errors: Wrapped and propagated up to root
- File errors: Descriptive messages with absolute paths

### Performance

- File discovery: Single pass with doublestar glob
- Parallel linting: `config.Concurrency` (default 10)
- Baseline filtering: O(n) fingerprint matching

### Testing

- Unit tests: `internal/baseline/baseline_test.go`, `internal/format/formatter_test.go`, `internal/cli/singlefile_test.go`, `internal/git/diff_test.go`
- Test discovery: `go test ./...`

---

## Future Extension Points

### Adding New Component Types

1. Create linter in `internal/cli/{type}_linter.go`
2. Implement `ComponentLinter` interface
3. Add optional interfaces (Scorable, BestPracticeValidator, etc.)
4. Add subcommand in `cmd/{type}.go`
5. Register in root command's `runLint()` linters array
6. Add glob pattern to `internal/discovery/discovery.go`

### Adding New Output Formats

1. Create formatter in `internal/output/{format}.go`
2. Implement `Format{Type}()` function
3. Add to `outputters.NewOutputter()` switch
4. Update config validation in `internal/config/config.go`

### Adding New Validators

1. Define validator interface in `internal/cli/generic_linter.go`
2. Implement in component-specific linter
3. Call in `lintComponent()` or `lintBatchFile()` pipeline
4. No changes needed to orchestration layer (ISP)
