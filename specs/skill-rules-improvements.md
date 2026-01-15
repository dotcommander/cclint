# Skill Rules Improvements Specification

**Status:** Draft
**Created:** 2026-01-15
**Author:** Claude Code (spec-create-agent pattern)
**Source:** agentskills.io specification gap analysis

---

## TL;DR

Analysis of agentskills.io specification reveals 10 missing validation rules in cclint's skill linter. Primary gaps: name format validation (hyphen constraints), directory-name matching, and new optional frontmatter fields (license, compatibility, metadata). Alignment to 500-line limit (currently 550) and allowed-tools format validation also needed.

---

## Priority Matrix

| Priority | Rule | Severity | Impact | Effort |
|----------|------|----------|--------|--------|
| P0 | Name cannot start/end with hyphen | error | High - breaks spec compliance | Low |
| P0 | Name cannot have consecutive hyphens | error | High - breaks spec compliance | Low |
| P0 | Name must match parent directory | error | High - discovery failure | Medium |
| P1 | Line limit should be 500 (not 550) | suggestion | Medium - spec alignment | Low |
| P1 | Validate allowed-tools format | warning | Medium - tool restrictions fail | Medium |
| P2 | Support license field validation | suggestion | Low - optional field | Low |
| P2 | Support compatibility field | suggestion | Low - optional field | Low |
| P2 | Support metadata field | suggestion | Low - optional field | Low |
| P2 | Validate directory structure | info | Low - optional best practice | Medium |
| P3 | Progressive disclosure guidance | info | Low - documentation only | Low |

---

## Gap Analysis

### Current vs. Specification

| Feature | cclint Current | agentskills.io Spec | Gap |
|---------|----------------|---------------------|-----|
| **Name validation** | `^[a-z0-9-]+$` | No leading/trailing hyphens, no consecutive hyphens | ❌ Missing hyphen position rules |
| **Directory matching** | Not enforced | Name must match parent directory | ❌ Not validated |
| **Line limit** | 550 lines | 500 lines recommended | ⚠️ Different threshold |
| **Frontmatter: license** | Not supported | Optional field | ❌ Not validated |
| **Frontmatter: compatibility** | Not supported | Optional (max 500 chars) | ❌ Not validated |
| **Frontmatter: metadata** | Not supported | Optional key-value mapping | ❌ Not validated |
| **allowed-tools format** | String validation only | Space-delimited with scope syntax | ⚠️ Format not validated |
| **Progressive disclosure** | Not mentioned | ~100 tokens metadata, <5000 tokens instructions | ℹ️ Guideline only |
| **Optional directories** | Not validated | `scripts/`, `references/`, `assets/` | ℹ️ Optional feature |
| **File references depth** | Not enforced | Keep one level deep | ℹ️ Best practice guidance |

---

## Proposed New Rules

### Rule 048: Name cannot start or end with hyphen

**Severity:** error
**Category:** structural
**Priority:** P0

**Description:**
Skill names must not start or end with a hyphen. This ensures compatibility with directory naming conventions and agent discovery systems.

**Pass Criteria:**
- `name` field does not match `^-` (starts with hyphen)
- `name` field does not match `-$` (ends with hyphen)

**Fail Message:**
`Skill name '{name}' cannot start or end with a hyphen`

**Example violations:**
- `-my-skill` (starts with hyphen)
- `my-skill-` (ends with hyphen)
- `-skill-` (both)

**Implementation:**
```go
// In internal/cli/skills_check.go
if strings.HasPrefix(name, "-") || strings.HasSuffix(name, "-") {
    addIssue(lint.Issue{
        Rule: "048",
        Severity: lint.SeverityError,
        Category: lint.CategoryStructural,
        Message: fmt.Sprintf("Skill name '%s' cannot start or end with a hyphen", name),
        Source: skillPath,
    })
}
```

---

### Rule 049: Name cannot contain consecutive hyphens

**Severity:** error
**Category:** structural
**Priority:** P0

**Description:**
Skill names must not contain consecutive hyphens (e.g., `my--skill`). This ensures clean naming conventions and prevents parsing ambiguities.

**Pass Criteria:**
- `name` field does not match `--`

**Fail Message:**
`Skill name '{name}' contains consecutive hyphens (--) which are not allowed`

**Example violations:**
- `my--skill`
- `data--analysis--tool`
- `pdf--processing`

**Implementation:**
```go
if strings.Contains(name, "--") {
    addIssue(lint.Issue{
        Rule: "049",
        Severity: lint.SeverityError,
        Category: lint.CategoryStructural,
        Message: fmt.Sprintf("Skill name '%s' contains consecutive hyphens (--) which are not allowed", name),
        Source: skillPath,
    })
}
```

---

### Rule 050: Name must match parent directory name

**Severity:** error
**Category:** structural
**Priority:** P0

**Description:**
The skill name in frontmatter must match the parent directory name. This is critical for skill discovery and prevents confusion when multiple skills are present.

**Pass Criteria:**
- `name` field equals parent directory name
- Comparison is case-sensitive

**Fail Message:**
`Skill name '{name}' must match parent directory name '{directory}'`

**Example violations:**
```
# Directory: pdf-processing/
# Frontmatter: name: pdf-tools
❌ Mismatch: "pdf-tools" vs "pdf-processing"

# Directory: data-analysis/
# Frontmatter: name: data-analysis
✅ Match
```

**Implementation:**
```go
// Extract parent directory name
parentDir := filepath.Base(filepath.Dir(skillPath))

if name != parentDir {
    addIssue(lint.Issue{
        Rule: "050",
        Severity: lint.SeverityError,
        Category: lint.CategoryStructural,
        Message: fmt.Sprintf("Skill name '%s' must match parent directory name '%s'", name, parentDir),
        Source: skillPath,
    })
}
```

**Edge case handling:**
- Root-level `SKILL.md` files: Skip validation (no parent directory to match)
- `.claude/skills/SKILL.md`: Skip (not in named directory)
- Symlinks: Follow to real path before extracting directory

---

### Rule 051: Line limit should be 500 lines

**Severity:** suggestion
**Category:** structural
**Priority:** P1

**Description:**
Skills should be kept under 500 lines per agentskills.io specification. Current cclint threshold is 550 (500 + 10% tolerance). Align to spec recommendation.

**Pass Criteria:**
- File has 500 or fewer lines

**Fail Message:**
`Skill is {lines} lines. Best practice per agentskills.io: keep skills under 500 lines - move heavy docs to references/ subdirectory.`

**Implementation:**
- Update Rule 046 threshold from 550 → 500
- Update message to cite agentskills.io specification
- Keep as suggestion (not error) to allow flexibility

---

### Rule 052: Validate allowed-tools format

**Severity:** warning
**Category:** best-practice
**Priority:** P1

**Description:**
When `allowed-tools` field is present, validate it follows space-delimited format with optional scope syntax. Per agentskills.io spec, format is `Bash(git:*) Bash(jq:*) Read`.

**Pass Criteria:**
- If `allowed-tools` is string "*": always valid (wildcard)
- If string: each token matches pattern `^[A-Z][a-zA-Z]+(\([^)]+\))?$`
- Valid tokens: `Read`, `Bash(git:*)`, `Write`, `Bash(jq:*)`

**Fail Message:**
`allowed-tools format should be space-delimited tool names (e.g., 'Bash(git:*) Read Write')`

**Example valid values:**
```yaml
allowed-tools: "*"                           # Wildcard
allowed-tools: "Read Write Edit"             # Simple tools
allowed-tools: "Bash(git:*) Bash(jq:*) Read" # Scoped tools
```

**Example invalid values:**
```yaml
allowed-tools: "read write"                  # Lowercase
allowed-tools: "Read, Write, Edit"           # Comma-delimited
allowed-tools: "['Read', 'Write']"           # JSON array
```

**Implementation:**
```go
// In internal/cli/skills_check.go
if allowedTools, ok := frontmatter["allowed-tools"].(string); ok && allowedTools != "*" {
    tokens := strings.Fields(allowedTools)
    toolPattern := regexp.MustCompile(`^[A-Z][a-zA-Z]+(\([^)]+\))?$`)

    for _, token := range tokens {
        if !toolPattern.MatchString(token) {
            addIssue(lint.Issue{
                Rule: "052",
                Severity: lint.SeverityWarning,
                Category: lint.CategoryBestPractice,
                Message: "allowed-tools format should be space-delimited tool names (e.g., 'Bash(git:*) Read Write')",
                Source: skillPath,
            })
            break
        }
    }
}
```

---

### Rule 053: License field validation

**Severity:** suggestion
**Category:** documentation
**Priority:** P2

**Description:**
The optional `license` field should contain a valid SPDX license identifier or reference to a bundled license file. This helps users understand usage rights.

**Pass Criteria:**
- If `license` field present: non-empty string
- Optional: validate against SPDX license list

**Fail Message:**
`license field should contain SPDX identifier (e.g., 'MIT', 'Apache-2.0') or license file reference`

**Example valid values:**
```yaml
license: MIT
license: Apache-2.0
license: Proprietary
license: See LICENSE.txt for complete terms
```

**Implementation (Phase 1 - basic):**
```go
if license, ok := frontmatter["license"].(string); ok {
    if strings.TrimSpace(license) == "" {
        addIssue(lint.Issue{
            Rule: "053",
            Severity: lint.SeveritySuggestion,
            Category: lint.CategoryDocumentation,
            Message: "license field is empty - provide SPDX identifier or license file reference",
            Source: skillPath,
        })
    }
}
```

**Implementation (Phase 2 - SPDX validation):**
- Add SPDX license list to `internal/validation/licenses.go`
- Validate against known identifiers
- Allow file references: `LICENSE.txt`, `LICENSE.md`, etc.

---

### Rule 054: Compatibility field length

**Severity:** warning
**Category:** documentation
**Priority:** P2

**Description:**
Per agentskills.io spec, the optional `compatibility` field must not exceed 500 characters. This field describes environment requirements (product compatibility, system packages, network access).

**Pass Criteria:**
- If `compatibility` field present: ≤500 characters

**Fail Message:**
`compatibility field is {length} chars (max 500 per agentskills.io spec)`

**Example valid values:**
```yaml
compatibility: Designed for Claude Code
compatibility: Requires git, docker, jq, and internet access
compatibility: Compatible with Claude Code, Cursor, and other Anthropic-compatible agents
```

**Implementation:**
```go
if compat, ok := frontmatter["compatibility"].(string); ok {
    if len(compat) > 500 {
        addIssue(lint.Issue{
            Rule: "054",
            Severity: lint.SeverityWarning,
            Category: lint.CategoryDocumentation,
            Message: fmt.Sprintf("compatibility field is %d chars (max 500 per agentskills.io spec)", len(compat)),
            Source: skillPath,
        })
    }
}
```

---

### Rule 055: Metadata field structure

**Severity:** info
**Category:** documentation
**Priority:** P2

**Description:**
The optional `metadata` field should be a key-value mapping of strings. This provides extensibility for additional properties not covered by standard fields.

**Pass Criteria:**
- If `metadata` field present: must be map[string]interface{}
- All values should be primitives (string, number, boolean)

**Fail Message:**
`metadata field should be key-value mapping (e.g., metadata:\n  author: example-org\n  version: "1.0")`

**Example valid values:**
```yaml
metadata:
  author: example-org
  version: "1.0"
  category: data-processing
```

**Implementation:**
```go
if metadata, ok := frontmatter["metadata"]; ok {
    metaMap, isMap := metadata.(map[string]interface{})
    if !isMap {
        addIssue(lint.Issue{
            Rule: "055",
            Severity: lint.SeverityInfo,
            Category: lint.CategoryDocumentation,
            Message: "metadata field should be key-value mapping",
            Source: skillPath,
        })
    } else {
        // Optional: validate value types are primitives
        for key, val := range metaMap {
            switch val.(type) {
            case string, int, int64, float64, bool:
                // Valid primitive types
            default:
                addIssue(lint.Issue{
                    Rule: "055",
                    Severity: lint.SeverityInfo,
                    Category: lint.CategoryDocumentation,
                    Message: fmt.Sprintf("metadata.%s should be primitive value (string, number, or boolean)", key),
                    Source: skillPath,
                })
            }
        }
    }
}
```

---

### Rule 056: Optional directory structure

**Severity:** info
**Category:** documentation
**Priority:** P2

**Description:**
Per agentskills.io spec, skills can include optional directories: `scripts/`, `references/`, `assets/`. Detect and validate these follow conventions.

**Pass Criteria:**
- Check for existence of optional directories
- Provide info-level suggestions for organization

**Informational Messages:**
- `Found scripts/ directory - ensure scripts are executable and self-contained`
- `Found references/ directory - good practice for detailed documentation`
- `Found assets/ directory - ensure templates and static resources are versioned`

**Implementation:**
```go
skillDir := filepath.Dir(skillPath)

// Check for optional directories
optionalDirs := []string{"scripts", "references", "assets"}
for _, dir := range optionalDirs {
    dirPath := filepath.Join(skillDir, dir)
    if stat, err := os.Stat(dirPath); err == nil && stat.IsDir() {
        var message string
        switch dir {
        case "scripts":
            message = "Found scripts/ directory - ensure scripts are executable and self-contained"
        case "references":
            message = "Found references/ directory - good practice for detailed documentation"
        case "assets":
            message = "Found assets/ directory - ensure templates and resources are versioned"
        }

        addIssue(lint.Issue{
            Rule: "056",
            Severity: lint.SeverityInfo,
            Category: lint.CategoryDocumentation,
            Message: message,
            Source: skillPath,
        })
    }
}
```

**Note:** This is informational only - directories are optional and their presence/absence should not affect lint pass/fail status.

---

## Schema Updates

### CUE Schema Changes

**File:** `internal/cue/schemas/skill.cue`

```diff
 #Skill: {
 	// Required fields
-	name: string & =~("^[a-z0-9-]+$") & strings.MaxRunes(64)
+	name: string & =~("^[a-z0-9]+(-[a-z0-9]+)*$") & strings.MaxRunes(64)  // no leading/trailing/consecutive hyphens
 	description: string & !="" & strings.MaxRunes(1024)

 	// Optional Claude Code fields
 	"allowed-tools"?: "*" | string | [...#KnownTool]
 	model?: #Model
 	context?: "fork"
 	agent?: string
 	"user-invocable"?: bool
 	hooks?: #SkillHooks
+
+	// Optional agentskills.io fields
+	license?: string                                              // SPDX identifier or license file reference
+	compatibility?: string & strings.MaxRunes(500)                // environment requirements (max 500 chars)
+	metadata?: {[string]: string | number | bool}                 // arbitrary key-value mapping

 	// Allow additional fields
 	...
 }
```

**Key changes:**
1. **Name regex tightened:** `^[a-z0-9]+(-[a-z0-9]+)*$` prevents leading/trailing/consecutive hyphens
2. **New optional fields:** `license`, `compatibility`, `metadata`
3. **Compatibility constraint:** Max 500 characters per spec
4. **Metadata constraint:** Primitive values only (string, number, bool)

---

## Implementation Notes

### Validation Order

Maintain rule execution order:
1. **Structural** (035-036, 048-050): File naming, emptiness, name format
2. **Schema** (CUE validation): Frontmatter structure
3. **Best Practice** (038-044, 052): Reserved words, description quality, version format, allowed-tools
4. **Documentation** (045, 047, 053-056): Sections, optional fields
5. **Size** (046/051): Line limits

### Backward Compatibility

**Breaking changes:**
- Rule 048-050: New errors may fail existing skills
- Rule 051: Changed threshold (550 → 500) may increase suggestions

**Migration path:**
1. Release as warnings first (v1.x.0)
2. Document in changelog
3. Provide grace period (2-3 releases)
4. Promote to errors (v2.0.0)

**Flag consideration:**
```bash
cclint --spec-version agentskills.io  # Enable strict agentskills.io compliance
cclint --spec-version claude-code     # Current behavior (default)
```

### Testing Requirements

**Unit tests needed:**
```
internal/cli/skills_check_test.go:
  - TestRule048_NameStartsWithHyphen
  - TestRule048_NameEndsWithHyphen
  - TestRule049_ConsecutiveHyphens
  - TestRule050_DirectoryMismatch
  - TestRule051_LineLimitUpdated
  - TestRule052_AllowedToolsFormat
  - TestRule053_LicenseField
  - TestRule054_CompatibilityLength
  - TestRule055_MetadataStructure
  - TestRule056_OptionalDirectories
```

**Integration tests needed:**
```
testdata/skills/invalid-hyphen-start/SKILL.md
testdata/skills/invalid-hyphen-end/SKILL.md
testdata/skills/invalid-consecutive-hyphens/SKILL.md
testdata/skills/mismatched-directory/SKILL.md
testdata/skills/valid-optional-fields/SKILL.md
```

### Documentation Updates

**Files to update:**
1. `docs/rules/skills.md` - Add Rules 048-056
2. `CHANGELOG.md` - Document breaking changes
3. `README.md` - Update examples if needed
4. `internal/cli/skills_check.go` - Inline comments
5. `docs/agentskills-io-compliance.md` - NEW: Spec alignment guide

---

## Progressive Disclosure Guidance

Per agentskills.io spec, skills should follow progressive disclosure pattern:

| Stage | Size | Content |
|-------|------|---------|
| Metadata | ~100 tokens | `name` + `description` (loaded at startup) |
| Instructions | <5000 tokens | `SKILL.md` body (loaded when activated) |
| Resources | As needed | `scripts/`, `references/`, `assets/` (on-demand) |

**Recommendations (info-level, not enforced):**
- Keep `SKILL.md` under 500 lines
- Move detailed docs to `references/`
- Keep file references one level deep
- Use relative paths from skill root

**Not implemented as rules** (guidance only):
- Token counting (requires LLM tokenizer)
- File reference depth traversal (complex graph analysis)
- Resource loading metrics (runtime only)

---

## Open Questions

1. **Directory matching edge cases:**
   - Should we skip validation for `.claude/skills/SKILL.md` (no named parent)?
   - How to handle symlinks?
   - Should we allow namespace prefixes (e.g., `org-skill-name/` directory with `name: skill-name`)?

2. **SPDX license validation:**
   - Maintain full SPDX list (400+ licenses) or common subset?
   - Allow custom license file references unconditionally?
   - Validate referenced license files exist?

3. **Allowed-tools scope syntax:**
   - Full validation of scope patterns (e.g., `git:*`, `jq:*`)?
   - Known tool list completeness (is `#KnownTool` exhaustive)?
   - How to handle future tool additions?

4. **Backward compatibility strategy:**
   - Warnings-first rollout timeline?
   - Feature flag for strict agentskills.io mode?
   - How to communicate breaking changes to users?

5. **Optional directory validation:**
   - Should we validate script executability?
   - Check for common issues (e.g., missing shebangs)?
   - Validate referenced files exist?

---

## References

- [agentskills.io Specification](https://agentskills.io/specification)
- [agentskills GitHub Repository](https://github.com/agentskills/agentskills)
- [Claude Code Skills Documentation](https://code.claude.com/docs/en/skills)
- [cclint Current Rules](../docs/rules/skills.md)
- [CUE Schema](../internal/cue/schemas/skill.cue)

---

## Next Steps

1. **Review & Approve:** Validate spec with stakeholders
2. **Prioritize:** Confirm P0/P1/P2 priority assignments
3. **Implementation:**
   - Update CUE schema (Rules 048-055)
   - Add Go validation logic (Rules 048-056)
   - Write unit tests for each rule
   - Create integration test fixtures
4. **Documentation:**
   - Update `docs/rules/skills.md`
   - Create migration guide
   - Update CHANGELOG
5. **Release Strategy:**
   - v1.x.0: Rules as warnings
   - v2.0.0: Promote to errors after grace period

**Estimated effort:** 2-3 days for P0/P1 rules, 1 day for P2/P3 documentation and guidance.
