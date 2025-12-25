# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Run

```bash
# Build and symlink to ~/go/bin
go build -o cclint . && ln -sf $(pwd)/cclint ~/go/bin/cclint

# Run linter (defaults to ~/.claude)
cclint                        # lint all component types
cclint agents                 # lint only agents
cclint commands               # lint only commands
cclint skills                 # lint only skills
cclint plugins                # lint only plugin manifests

# Common flags
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
├── scoring/            # Quality scoring (0-100) with tier grading (A-F)
│   ├── types.go        # QualityScore, ScoringMetric interfaces
│   ├── agent_scorer.go
│   ├── command_scorer.go
│   ├── skill_scorer.go
│   └── plugin_scorer.go
├── cli/                # Lint orchestration per component type
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

**Discovery Patterns**: Searches `.claude/{agents,commands,skills}/**/*.md`, `{agents,commands,skills}/**/*.md`, and `**/.claude-plugin/plugin.json` paths.

## Lint Rules Enforced

| Component | Size Limit | Key Checks |
|-----------|------------|------------|
| Agent | 200 lines | name format, model, triggers, Foundation/Workflow sections, Skill() loading |
| Command | 50 lines | delegation pattern, semantic routing table |
| Skill | 500 lines | SKILL.md filename, Quick Reference table |
| Plugin | 5KB | name format, semver version, author.name, required fields |

## Config

Supports `.cclintrc.json`, `.cclintrc.yaml`, `.cclintrc.yml` in project root. Environment variables with `CCLINT_` prefix also supported.
