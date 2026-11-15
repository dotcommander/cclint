# Programmatic Usage

Use cclint internals from code inside this repository (for contributors and tests).

`internal/...` packages are accessible only within the `github.com/dotcommander/cclint` import tree and are not a supported external API surface. For other projects, invoke the `cclint` CLI instead. The snippets below show isolated calls; handle returned errors before using values.

## Core Internal APIs

### Discovery (`internal/discovery`)

Discover and categorize Claude Code components.

```go
import "github.com/dotcommander/cclint/internal/discovery"

// Create discoverer for project root
discoverer := discovery.NewFileDiscovery("/path/to/project", false)

// Discover all component files
files, err := discoverer.DiscoverFiles()

// Detect type from path
fileType, err := discovery.DetectFileType(absPath, rootPath)
```

**Types:**
- `File`: Path, RelPath, Size, Type, Contents
- `FileType`: Agent, Command, Skill, Settings, Context, Plugin, Rule

### Validation (`internal/cue`)

Validate frontmatter against CUE schemas.

```go
import "github.com/dotcommander/cclint/internal/cue"

// Create validator and load embedded schemas before validation.
validator := cue.NewValidator()
if err := validator.LoadSchemas(); err != nil {
    panic(err)
}

// Parse frontmatter
fm, err := cue.ParseFrontmatter(markdownContent)

// Validate by type
agentErrors, err := validator.ValidateAgent(fm.Data)
commandErrors, err := validator.ValidateCommand(fm.Data)
skillErrors, err := validator.ValidateSkill(fm.Data)
// Inspect err and the returned diagnostics for each call.

// Validate file directly; diagnostics can be returned with nil err.
fileErrors, err := validator.ValidateFile(path, content, "agent")
```

**Types:**
- `ValidationError`: File, Message, Severity, Source, Line, Column
- `Frontmatter`: Data, Body

### Lint Context (`internal/lint`)

Orchestrate linting operations.

```go
import "github.com/dotcommander/cclint/internal/lint"

// Empty root uses project.FindProjectRoot("."); nil means no exclusions.
ctx, err := lint.NewLinterContext("", false, false, false, nil)
if err != nil {
    panic(err)
}

// Access components
validator := ctx.Validator
files := ctx.Files
crossValidator := ctx.CrossValidator
```

**Types:**
- `LinterContext`: RootPath, Quiet, Verbose, NoCycleCheck, Validator, Discoverer, Files, CrossValidator

### Scoring (`internal/scoring`)

Calculate quality scores for components.

```go
import "github.com/dotcommander/cclint/internal/scoring"

// Score agent content
scorer := &scoring.AgentScorer{}
quality := scorer.Score(content, frontmatter, body)

// Access score breakdown
score := quality.Overall       // 0-100
tier := quality.Tier           // A-F
structural := quality.Structural    // 0-40
practices := quality.Practices     // 0-40
```

**Types:**
- `QualityScore`: Overall, Tier, Structural, Practices, Composition, Documentation, Details
- `Scorer`: Interface with Score() method

## Example: Validate Single File

```go
package main

import (
    "fmt"
    "os"

    "github.com/dotcommander/cclint/internal/cue"
)

func main() {
    content, err := os.ReadFile("agents/my-agent.md")
    if err != nil {
        panic(err)
    }

    validator := cue.NewValidator()
    if err := validator.LoadSchemas(); err != nil {
        panic(err)
    }

    errors, err := validator.ValidateFile(
        "agents/my-agent.md",
        string(content),
        "agent",
    )

    if err != nil {
        panic(err)
    }

    for _, e := range errors {
        fmt.Printf("%s:%d: %s\n", e.File, e.Line, e.Message)
    }
}
```

## Example: Custom Lint Workflow

```go
package main

import (
    "github.com/dotcommander/cclint/internal/lint"
    "github.com/dotcommander/cclint/internal/discovery"
)

func main() {
    ctx, err := lint.NewLinterContext("/project", false, false, false, nil)
    if err != nil {
        panic(err)
    }

    agents := ctx.FilterFilesByType(discovery.FileTypeAgent)
    for _, agent := range agents {
        diagnostics, err := ctx.Validator.ValidateFile(agent.Path, agent.Contents, "agent")
        if err != nil {
            panic(err)
        }
        _ = diagnostics // Process diagnostics (which can be nonempty with nil err).
    }
}
```

## Example: Quality Scoring

```go
package main

import (
    "fmt"
    "os"

    "github.com/dotcommander/cclint/internal/cue"
    "github.com/dotcommander/cclint/internal/scoring"
)

func main() {
    raw, err := os.ReadFile("agents/my-agent.md")
    if err != nil {
        panic(err)
    }
    content := string(raw)
    fm, err := cue.ParseFrontmatter(content)
    if err != nil {
        panic(err)
    }

    scorer := &scoring.AgentScorer{}
    quality := scorer.Score(content, fm.Data, fm.Body)

    fmt.Printf("Score: %d/100 (%s)\n", quality.Overall, quality.Tier)
    fmt.Printf("Structural: %d/40\n", quality.Structural)
    fmt.Printf("Practices: %d/40\n", quality.Practices)
}
```

## Verification

For code changes, choose a focused check from the repository root; these examples do not require a full test suite:

```bash
go build ./...
go test ./internal/lint/... # or the focused package you changed
```

## Related docs

- Contributor workflow: `docs/change-cclint.md`
- Configuration behavior: `docs/guides/configuration.md`
- Schema contracts: `docs/reference/schemas.md`
