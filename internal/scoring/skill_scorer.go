package scoring

import (
	"regexp"
	"strings"
)

// SkillScorer scores skill files on a 0-100 scale
type SkillScorer struct{}

// NewSkillScorer creates a new SkillScorer
func NewSkillScorer() *SkillScorer {
	return &SkillScorer{}
}

// Score evaluates a skill and returns a QualityScore
func (s *SkillScorer) Score(content string, frontmatter map[string]interface{}, bodyContent string) QualityScore {
	var details []ScoringMetric
	lines := strings.Count(content, "\n") + 1

	// === STRUCTURAL (40 points max) ===
	structural := 0

	// Required frontmatter (20 points)
	requiredFields := []struct {
		name   string
		points int
	}{
		{"name", 10},
		{"description", 10},
	}

	for _, field := range requiredFields {
		_, exists := frontmatter[field.name]
		points := 0
		if exists {
			points = field.points
		}
		structural += points
		details = append(details, ScoringMetric{
			Category:  "structural",
			Name:      "Has " + field.name,
			Points:    points,
			MaxPoints: field.points,
			Passed:    exists,
		})
	}

	// Required sections (20 points)
	sections := []struct {
		pattern string
		name    string
		points  int
	}{
		{`(?i)## Quick Reference`, "Quick Reference", 8},
		{`(?i)## Workflow`, "Workflow section", 6},
		{`(?i)## Anti-Patterns`, "Anti-Patterns section", 4},
		{`(?i)## Success Criteria`, "Success Criteria", 2},
	}

	for _, sec := range sections {
		matched, _ := regexp.MatchString(sec.pattern, bodyContent)
		points := 0
		if matched {
			points = sec.points
		}
		structural += points
		details = append(details, ScoringMetric{
			Category:  "structural",
			Name:      sec.name,
			Points:    points,
			MaxPoints: sec.points,
			Passed:    matched,
		})
	}

	// === PRACTICES (40 points max) ===
	practices := 0

	// Semantic routing table format (10 points)
	hasSemanticRouting, _ := regexp.MatchString(`\|.*User Question.*\|.*Action.*\|`, bodyContent)
	if hasSemanticRouting {
		practices += 10
	}
	details = append(details, ScoringMetric{
		Category:  "practices",
		Name:      "Semantic routing table",
		Points:    boolToInt(hasSemanticRouting) * 10,
		MaxPoints: 10,
		Passed:    hasSemanticRouting,
	})

	// Phase-based workflow (8 points)
	hasPhases, _ := regexp.MatchString(`(?i)### Phase \d`, bodyContent)
	if hasPhases {
		practices += 8
	}
	details = append(details, ScoringMetric{
		Category:  "practices",
		Name:      "Phase-based workflow",
		Points:    boolToInt(hasPhases) * 8,
		MaxPoints: 8,
		Passed:    hasPhases,
	})

	// Anti-patterns as table (6 points)
	hasAntiPatternTable, _ := regexp.MatchString(`\|.*Anti-Pattern.*\|.*Problem.*\|.*Fix.*\|`, bodyContent)
	if hasAntiPatternTable {
		practices += 6
	}
	details = append(details, ScoringMetric{
		Category:  "practices",
		Name:      "Anti-patterns table format",
		Points:    boolToInt(hasAntiPatternTable) * 6,
		MaxPoints: 6,
		Passed:    hasAntiPatternTable,
	})

	// HARD GATE markers (4 points)
	hasHardGates, _ := regexp.MatchString(`(?i)HARD GATE`, bodyContent)
	if hasHardGates {
		practices += 4
	}
	details = append(details, ScoringMetric{
		Category:  "practices",
		Name:      "HARD GATE markers",
		Points:    boolToInt(hasHardGates) * 4,
		MaxPoints: 4,
		Passed:    hasHardGates,
	})

	// Success criteria checkboxes (4 points)
	hasCheckboxes := strings.Contains(bodyContent, "- [ ]")
	if hasCheckboxes {
		practices += 4
	}
	details = append(details, ScoringMetric{
		Category:  "practices",
		Name:      "Success criteria checkboxes",
		Points:    boolToInt(hasCheckboxes) * 4,
		MaxPoints: 4,
		Passed:    hasCheckboxes,
	})

	// References to references/ directory (4 points)
	hasReferences, _ := regexp.MatchString(`references/\w+\.md`, bodyContent)
	if hasReferences {
		practices += 4
	}
	details = append(details, ScoringMetric{
		Category:  "practices",
		Name:      "References to references/",
		Points:    boolToInt(hasReferences) * 4,
		MaxPoints: 4,
		Passed:    hasReferences,
	})

	// Scoring formula (4 points)
	hasScoringFormula, _ := regexp.MatchString(`(?i)(score\s*=|scoring formula)`, bodyContent)
	if hasScoringFormula {
		practices += 4
	}
	details = append(details, ScoringMetric{
		Category:  "practices",
		Name:      "Scoring formula",
		Points:    boolToInt(hasScoringFormula) * 4,
		MaxPoints: 4,
		Passed:    hasScoringFormula,
	})

	// === COMPOSITION (10 points max) ===
	composition := 0
	var compositionNote string

	switch {
	case lines <= 200:
		composition = 10
		compositionNote = "Excellent: ≤200 lines"
	case lines <= 350:
		composition = 8
		compositionNote = "Good: ≤350 lines"
	case lines <= 500:
		composition = 6
		compositionNote = "OK: ≤500 lines"
	case lines <= 600:
		composition = 3
		compositionNote = "Over limit: >500 lines"
	default:
		composition = 0
		compositionNote = "Fat skill: >600 lines"
	}
	details = append(details, ScoringMetric{
		Category:  "composition",
		Name:      "Line count",
		Points:    composition,
		MaxPoints: 10,
		Passed:    lines <= 500,
		Note:      compositionNote,
	})

	// === DOCUMENTATION (10 points max) ===
	documentation := 0

	// Description quality (5 points)
	desc, _ := frontmatter["description"].(string)
	descLen := len(desc)
	descPoints := 0
	var descNote string
	switch {
	case descLen >= 200:
		descPoints = 5
		descNote = "Comprehensive"
	case descLen >= 100:
		descPoints = 3
		descNote = "Adequate"
	case descLen > 0:
		descPoints = 1
		descNote = "Brief"
	default:
		descNote = "Missing"
	}
	documentation += descPoints
	details = append(details, ScoringMetric{
		Category:  "documentation",
		Name:      "Description quality",
		Points:    descPoints,
		MaxPoints: 5,
		Passed:    descLen >= 100,
		Note:      descNote,
	})

	// Code examples (5 points)
	codeBlockCount := strings.Count(bodyContent, "```")
	codePoints := 0
	var codeNote string
	switch {
	case codeBlockCount >= 6:
		codePoints = 5
		codeNote = "Rich examples"
	case codeBlockCount >= 3:
		codePoints = 3
		codeNote = "Adequate examples"
	case codeBlockCount >= 1:
		codePoints = 1
		codeNote = "Few examples"
	default:
		codeNote = "No examples"
	}
	documentation += codePoints
	details = append(details, ScoringMetric{
		Category:  "documentation",
		Name:      "Code examples",
		Points:    codePoints,
		MaxPoints: 5,
		Passed:    codeBlockCount >= 3,
		Note:      codeNote,
	})

	return NewQualityScore(structural, practices, composition, documentation, details)
}
