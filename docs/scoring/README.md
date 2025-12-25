# Scoring Reference

## Overview

**Total Score**: 0-100 points across four categories

**Categories**:
- **Structural**: Required fields, sections, core architecture
- **Practices**: Best practices, patterns, methodologies
- **Composition**: File size and length guidelines
- **Documentation**: Description quality, examples, structure

**Tier Grades**:
| Grade | Range | Quality |
|-------|-------|---------|
| A | ≥85 | Excellent |
| B | ≥70 | Good |
| C | ≥50 | Acceptable |
| D | ≥30 | Needs improvement |
| F | <30 | Poor |

---

## Agent Scoring

**Total**: 100 points (Structural: 35, Practices: 35, Composition: 10, Documentation: 10)

### Structural (35 points)

| Metric | Points | Criteria |
|--------|--------|----------|
| **Required Frontmatter Fields** | **20** | |
| name | 5 | Non-empty string |
| description | 5 | Non-empty string |
| model | 5 | Non-empty string |
| tools | 5 | Non-empty string or array |
| **Required Sections** | **15** | |
| Foundation section | 5 | Matches `## Foundation` (case-insensitive) |
| Phase workflow | 4 | Contains `### Phase` pattern |
| Success Criteria | 3 | Matches `## Success Criteria` (case-insensitive) |
| Edge Cases | 3 | Matches `## Edge Cases` (case-insensitive) |

### Practices (35 points)

| Metric | Points | Criteria |
|--------|--------|----------|
| Skill: reference | 10 | Detects: `Skill: foo`, `**Skill**: foo`, `Skill(foo)`, `Skills:` |
| Anti-Patterns section | 5 | Contains `## Anti-Patterns` (case-insensitive) |
| Expected Output section | 5 | Contains `## Expected Output` (case-insensitive) |
| HARD GATE markers | 5 | Contains `HARD GATE` (case-insensitive) |
| Third-person description | 5 | Description doesn't start with "I " |
| WHEN triggers in description | 5 | Contains: "PROACTIVELY", "use when", or "when user" |

### Composition (10 points)

| Lines | Points | Note |
|-------|--------|------|
| ≤120 | 10 | Excellent |
| ≤180 | 8 | Good |
| ≤220 | 6 | OK (200±10% tolerance) |
| ≤275 | 3 | Over limit |
| >275 | 0 | Fat agent |

**Tolerance**: ±10% of 200-line guideline = 220 lines acceptable

### Documentation (10 points)

| Metric | Points | Criteria | Note |
|--------|--------|----------|------|
| **Description quality** | **5** | | |
| ≥200 chars | 5 | Comprehensive description | Comprehensive |
| ≥100 chars | 3 | Adequate description | Adequate |
| >0 chars | 1 | Brief description | Brief |
| 0 chars | 0 | Missing description | Missing |
| **Section structure** | **5** | | |
| ≥6 sections | 5 | 6+ `## ` headers | Well-structured |
| ≥4 sections | 3 | 4-5 `## ` headers | Adequate structure |
| ≥2 sections | 1 | 2-3 `## ` headers | Minimal structure |
| <2 sections | 0 | <2 `## ` headers | Poor structure |

---

## Command Scoring

**Total**: 100 points (Structural: 40, Practices: 40, Composition: 10, Documentation: 10)

### Structural (40 points)

| Metric | Points | Criteria |
|--------|--------|----------|
| **Required Frontmatter Fields** | **30** | |
| allowed-tools | 10 | Non-empty field |
| description | 10 | Non-empty string |
| argument-hint | 10 | Non-empty string |
| **Delegation Pattern** | **10** | |
| Task() delegation | 10 | Contains `Task(...)` pattern in body |

### Practices (40 points)

| Metric | Points | Criteria |
|--------|--------|----------|
| Success criteria | 15 | Contains "success criteria" (case-insensitive) or checkbox `- [ ]` |
| Task delegation | 15 | Body contains ≥1 `Task(` call |
| Flags documented | 10 | Contains `## Flags` or `--flag` pattern |

### Composition (10 points)

| Lines | Points | Note |
|-------|--------|------|
| ≤30 | 10 | Excellent |
| ≤45 | 8 | Good |
| ≤55 | 6 | OK (50±10% tolerance) |
| ≤65 | 3 | Over limit |
| >65 | 0 | Fat command |

**Tolerance**: ±10% of 50-line guideline = 55 lines acceptable

### Documentation (10 points)

| Metric | Points | Criteria | Note |
|--------|--------|----------|------|
| **Description quality** | **5** | | |
| ≥50 chars | 5 | Clear description | Clear |
| ≥20 chars | 3 | Brief description | Brief |
| >0 chars | 1 | Minimal description | Minimal |
| 0 chars | 0 | Missing description | Missing |
| **Code examples** | **5** | | |
| Has code blocks | 5 | Contains ` ```bash ` or ` ``` ` | Present |
| No code blocks | 0 | No code blocks found | Missing |

---

## Skill Scoring

**Total**: 100 points (Structural: 40, Practices: 40, Composition: 10, Documentation: 10)

### Skill Types

Skills are classified as either **Methodology** or **Reference/Pattern Library**:

- **Methodology skills**: Contain workflow phases (`### Phase \d`)
- **Reference skills**: Pattern tables without workflow phases

Section requirements vary by type.

### Structural (40 points)

| Metric | Points | Criteria |
|--------|--------|----------|
| **Required Frontmatter Fields** | **20** | |
| name | 10 | Non-empty string |
| description | 10 | Non-empty string |
| **Required Sections** | **20** | Varies by skill type |

**Methodology Skills**:
| Section | Points | Pattern |
|---------|--------|---------|
| Quick Reference | 8 | `## Quick Reference` |
| Workflow section | 6 | `## Workflow` |
| Anti-Patterns section | 4 | `## Anti-Patterns` or `### Anti-Patterns` or `\| Anti-Pattern` in table |
| Success Criteria | 2 | `## Success Criteria` |

**Reference/Pattern Library Skills**:
| Section | Points | Pattern |
|---------|--------|---------|
| Quick Reference | 10 | `## Quick Reference` (higher weight for discoverability) |
| Pattern/Template section | 6 | `## Patterns`, `## Templates`, or `## Examples` |
| Anti-Patterns section | 4 | `## Anti-Patterns` or `### Anti-Patterns` or `\| Anti-Pattern` in table |

**Anti-Patterns Fallback**: `## Best Practices` + `### Don't` subsection counts as Anti-Patterns section

### Practices (40 points)

| Metric | Points | Criteria |
|--------|--------|----------|
| Semantic routing table | 10 | Table with `\| User Question \| Action \|` format |
| Phase-based workflow | 8 | Contains `### Phase \d` pattern |
| Anti-patterns table format | 6 | Table with `\| Anti-Pattern \| Problem \| Fix \|` format |
| HARD GATE markers | 4 | Contains `HARD GATE` (case-insensitive) |
| Success criteria checkboxes | 4 | Contains `- [ ]` checkbox |
| References to references/ | 4 | Contains `references/*.md` path |
| Scoring formula | 4 | Contains "score =" or "scoring formula" (case-insensitive) |

### Composition (10 points)

| Lines | Points | Note |
|-------|--------|------|
| ≤250 | 10 | Excellent |
| ≤400 | 8 | Good |
| ≤550 | 6 | OK (500±10% tolerance) |
| ≤660 | 3 | Over limit |
| >660 | 0 | Fat skill |

**Tolerance**: ±10% of 500-line guideline = 550 lines acceptable

### Documentation (10 points)

| Metric | Points | Criteria | Note |
|--------|--------|----------|------|
| **Description quality** | **5** | | |
| ≥200 chars | 5 | Comprehensive description | Comprehensive |
| ≥100 chars | 3 | Adequate description | Adequate |
| >0 chars | 1 | Brief description | Brief |
| 0 chars | 0 | Missing description | Missing |
| **Code examples** | **5** | | |
| ≥6 code blocks | 5 | ≥6 ` ``` ` blocks | Rich examples |
| ≥3 code blocks | 3 | 3-5 ` ``` ` blocks | Adequate examples |
| ≥1 code block | 1 | 1-2 ` ``` ` blocks | Few examples |
| 0 code blocks | 0 | No ` ``` ` blocks | No examples |

---

## Plugin Scoring

**Total**: 100 points (Structural: 40, Practices: 40, Composition: 10, Documentation: 10)

**Note**: Plugins are JSON manifests. Frontmatter = parsed JSON data. Body content is unused.

### Structural (40 points)

| Metric | Points | Criteria |
|--------|--------|----------|
| Has name | 10 | Non-empty `name` string |
| Has description | 10 | Non-empty `description` string |
| Has version | 10 | Non-empty `version` string |
| Has author.name | 10 | Non-empty `author.name` string |

### Practices (40 points)

| Metric | Points | Criteria |
|--------|--------|----------|
| Has homepage | 10 | Non-empty `homepage` string |
| Has repository | 10 | Non-empty `repository` string |
| Has license | 10 | Non-empty `license` string |
| Has keywords | 10 | Non-empty `keywords` array with ≥1 element |

### Composition (10 points)

| File Size | Points | Note |
|-----------|--------|------|
| ≤1KB | 10 | Excellent |
| ≤2KB | 8 | Good |
| ≤5KB | 6 | OK |
| ≤10KB | 3 | Large |
| >10KB | 0 | Too large |

### Documentation (10 points)

| Metric | Points | Criteria | Note |
|--------|--------|----------|------|
| **Description quality** | **5** | | |
| ≥100 chars | 5 | Comprehensive description | Comprehensive |
| ≥50 chars | 3 | Adequate description | Adequate |
| ≥20 chars | 1 | Brief description | Brief |
| <20 chars | 0 | Too short | Too short |
| **README reference** | **5** | | |
| Has readme | 5 | Non-empty `readme` field | Present |
| No readme | 0 | Missing `readme` field | Missing |

---

## Implementation Notes

### Category Weights

All component types use the same category structure but with different maximum points:

| Component | Structural | Practices | Composition | Documentation |
|-----------|------------|-----------|-------------|---------------|
| Agent | 35 | 35 | 10 | 10 |
| Command | 40 | 40 | 10 | 10 |
| Skill | 40 | 40 | 10 | 10 |
| Plugin | 40 | 40 | 10 | 10 |

### Composition Tolerance

All line-based composition scoring uses ±10% tolerance:
- **Agent**: 200 base → 220 acceptable (200 × 1.1)
- **Command**: 50 base → 55 acceptable (50 × 1.1)
- **Skill**: 500 base → 550 acceptable (500 × 1.1)

### Pattern Matching

Most structural checks use case-insensitive regex:
- Section headers: `(?i)## Foundation`
- Keywords: `(?i)HARD GATE`
- Tables: `\|.*User Question.*\|.*Action.*\|`

### Scoring Helpers

Common scoring utilities:
- `ScoreRequiredFields()`: Validates frontmatter fields
- `ScoreSections()`: Matches regex patterns for sections
- `ScoreComposition()`: Applies thresholds with tolerance
- `NewQualityScore()`: Creates final 0-100 score with tier grade

### Source Attribution

Scoring thresholds are configurable per project via `.cclintrc.json`:

```json
{
  "scoring": {
    "agent": {
      "composition": {
        "excellent": 120,
        "good": 180,
        "ok": 220
      }
    }
  }
}
```

Default values are shown in this reference.
