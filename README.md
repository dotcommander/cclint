# cclint

Linter for Claude Code components (agents, commands, skills).

Validates structure, checks frontmatter schemas, suggests improvements based on patterns I've found useful. Your mileage may vary.

## Install

```bash
git clone https://github.com/dotcommander/cclint.git && cd cclint
go build -o cclint .
ln -sf $(pwd)/cclint ~/go/bin/cclint
```

## Usage

```bash
cclint                  # check everything
cclint agents           # just agents
cclint commands         # just commands
cclint skills           # just skills

cclint --root ~/project agents          # specific directory
cclint --format json --output report.json   # JSON for CI
cclint --quiet                          # errors only
cclint --scores                         # show quality scores
```

## What It Checks

| Component | Errors | Suggestions |
|-----------|--------|-------------|
| Agents | Missing name/description, invalid name format, bad color values | Line count >200, missing model/triggers, missing sections |
| Commands | Missing required fields, invalid format | Line count >50, no routing table |
| Skills | Wrong filename (must be SKILL.md), empty content | Line count >500, no Quick Reference table |

## Output

```
💡 agents/my-agent.md
    💡 Agent is 251 lines. Best practice: keep under 200 lines.
    💡 Missing 'model' field.

✗ commands/broken.md
    ✗ Required field 'name' is missing.

70/70 passed, 0 errors, 12 suggestions
```

- 💡 Suggestion - won't fail the build
- ✗ Error - exit code 1

## Customization

Fork it. The rules are my preferences, not universal truths.

**Schemas:** `internal/cue/schemas/*.cue`

**Line limits:** Search for `lines > 200` in `internal/cli/*.go`

**Add/remove checks:** Edit validation functions in `internal/cli/*.go`

Rebuild after changes: `go build -o cclint .`

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | No errors (suggestions don't count) |
| 1 | Errors found |

## License

MIT
