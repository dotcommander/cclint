package scoring

import (
	"regexp"
	"strings"
)

// CommandScorer scores command files on a 0-100 scale
type CommandScorer struct{}

// NewCommandScorer creates a new CommandScorer
func NewCommandScorer() *CommandScorer {
	return &CommandScorer{}
}

// Score evaluates a command and returns a QualityScore
func (s *CommandScorer) Score(content string, frontmatter map[string]any, bodyContent string) QualityScore {
	var details []ScoringMetric
	lines := strings.Count(content, "\n") + 1

	// === STRUCTURAL (40 points max) ===
	// Required frontmatter fields (10 points each, 30 total)
	fieldSpecs := []FieldSpec{
		{"allowed-tools", 10},
		{"description", 10},
		{"argument-hint", 10},
	}
	structural, fieldDetails := ScoreRequiredFields(frontmatter, fieldSpecs)
	details = append(details, fieldDetails...)

	// Task delegation (10 points)
	hasTaskDelegation, _ := regexp.MatchString(`Task\([^)]+\)`, bodyContent)
	if hasTaskDelegation {
		structural += 10
	}
	details = append(details, ScoringMetric{
		Category:  "structural",
		Name:      "Task() delegation",
		Points:    boolToInt(hasTaskDelegation) * 10,
		MaxPoints: 10,
		Passed:    hasTaskDelegation,
	})

	// === PRACTICES (40 points max) ===
	// Thin command pattern: Commands delegate to agents, not contain methodology.
	// Quick Reference and Semantic routing belong in SKILLS, not commands.
	practices := 0

	// Success criteria (15 points) - important for validation
	hasSuccessCriteria, _ := regexp.MatchString(`(?i)Success criteria|^\s*- \[ \]`, bodyContent)
	if hasSuccessCriteria {
		practices += 15
	}
	details = append(details, ScoringMetric{
		Category:  "practices",
		Name:      "Success criteria",
		Points:    boolToInt(hasSuccessCriteria) * 15,
		MaxPoints: 15,
		Passed:    hasSuccessCriteria,
	})

	// Task delegation present (15 points) - core thin command pattern
	taskCount := strings.Count(bodyContent, "Task(")
	hasTask := taskCount >= 1
	if hasTask {
		practices += 15
	}
	details = append(details, ScoringMetric{
		Category:  "practices",
		Name:      "Task delegation",
		Points:    boolToInt(hasTask) * 15,
		MaxPoints: 15,
		Passed:    hasTask,
		Note:      pluralize(taskCount, "Task() call"),
	})

	// Flags/options documented (10 points) - helps Claude understand inputs
	hasFlags, _ := regexp.MatchString(`(?i)## Flags|--\w+`, bodyContent)
	if hasFlags {
		practices += 10
	}
	details = append(details, ScoringMetric{
		Category:  "practices",
		Name:      "Flags documented",
		Points:    boolToInt(hasFlags) * 10,
		MaxPoints: 10,
		Passed:    hasFlags,
	})

	// === COMPOSITION (10 points max) ===
	// ±10% tolerance: 50 base -> 55 OK threshold
	commandThresholds := CompositionThresholds{
		Excellent:     30,
		ExcellentNote: "Excellent: ≤30 lines",
		Good:          45,
		GoodNote:      "Good: ≤45 lines",
		OK:            55,
		OKNote:        "OK: ≤55 lines (50±10%)",
		OverLimit:     65,
		OverLimitNote: "Over limit: >55 lines",
		FatNote:       "Fat command: >65 lines",
	}
	composition, compositionMetric := ScoreComposition(lines, commandThresholds)
	details = append(details, compositionMetric)

	// === DOCUMENTATION (10 points max) ===
	documentation := 0

	// Description quality (5 points)
	desc, _ := frontmatter["description"].(string)
	descLen := len(desc)
	descPoints := 0
	var descNote string
	switch {
	case descLen >= 50:
		descPoints = 5
		descNote = "Clear"
	case descLen >= 20:
		descPoints = 3
		descNote = "Brief"
	case descLen > 0:
		descPoints = 1
		descNote = "Minimal"
	default:
		descNote = "Missing"
	}
	documentation += descPoints
	details = append(details, ScoringMetric{
		Category:  "documentation",
		Name:      "Description quality",
		Points:    descPoints,
		MaxPoints: 5,
		Passed:    descLen >= 20,
		Note:      descNote,
	})

	// Code examples (5 points)
	hasCodeExamples := strings.Contains(bodyContent, "```bash") || strings.Contains(bodyContent, "```")
	if hasCodeExamples {
		documentation += 5
	}
	details = append(details, ScoringMetric{
		Category:  "documentation",
		Name:      "Code examples",
		Points:    boolToInt(hasCodeExamples) * 5,
		MaxPoints: 5,
		Passed:    hasCodeExamples,
	})

	return NewQualityScore(structural, practices, composition, documentation, details)
}

func pluralize(count int, singular string) string {
	if count == 1 {
		return "1 " + singular
	}
	return strings.Replace(singular, "call", "calls", 1)
}
