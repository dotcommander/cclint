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
func (s *CommandScorer) Score(content string, frontmatter map[string]interface{}, bodyContent string) QualityScore {
	var details []ScoringMetric
	lines := strings.Count(content, "\n") + 1

	// === STRUCTURAL (40 points max) ===
	structural := 0

	// Required frontmatter fields (10 points each, 30 total)
	requiredFields := []struct {
		name   string
		points int
	}{
		{"allowed-tools", 10},
		{"description", 10},
		{"argument-hint", 10},
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
	practices := 0

	// Quick Reference section (10 points)
	hasQuickRef, _ := regexp.MatchString(`(?i)## Quick Reference`, bodyContent)
	if hasQuickRef {
		practices += 10
	}
	details = append(details, ScoringMetric{
		Category:  "practices",
		Name:      "Quick Reference section",
		Points:    boolToInt(hasQuickRef) * 10,
		MaxPoints: 10,
		Passed:    hasQuickRef,
	})

	// Semantic routing table (10 points)
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

	// Usage section (10 points)
	hasUsage, _ := regexp.MatchString(`(?i)## Usage`, bodyContent)
	if hasUsage {
		practices += 10
	}
	details = append(details, ScoringMetric{
		Category:  "practices",
		Name:      "Usage section",
		Points:    boolToInt(hasUsage) * 10,
		MaxPoints: 10,
		Passed:    hasUsage,
	})

	// Success criteria (5 points)
	hasSuccessCriteria, _ := regexp.MatchString(`(?i)Success criteria`, bodyContent)
	if hasSuccessCriteria {
		practices += 5
	}
	details = append(details, ScoringMetric{
		Category:  "practices",
		Name:      "Success criteria",
		Points:    boolToInt(hasSuccessCriteria) * 5,
		MaxPoints: 5,
		Passed:    hasSuccessCriteria,
	})

	// Single Task delegation (5 points) - penalize multiple
	taskCount := strings.Count(bodyContent, "Task(")
	singleTask := taskCount == 1
	if singleTask {
		practices += 5
	}
	details = append(details, ScoringMetric{
		Category:  "practices",
		Name:      "Single Task delegation",
		Points:    boolToInt(singleTask) * 5,
		MaxPoints: 5,
		Passed:    singleTask,
		Note:      pluralize(taskCount, "Task() call"),
	})

	// === COMPOSITION (10 points max) ===
	composition := 0
	var compositionNote string

	switch {
	case lines <= 30:
		composition = 10
		compositionNote = "Excellent: ≤30 lines"
	case lines <= 40:
		composition = 8
		compositionNote = "Good: ≤40 lines"
	case lines <= 50:
		composition = 6
		compositionNote = "OK: ≤50 lines"
	case lines <= 60:
		composition = 3
		compositionNote = "Over limit: >50 lines"
	default:
		composition = 0
		compositionNote = "Fat command: >60 lines"
	}
	details = append(details, ScoringMetric{
		Category:  "composition",
		Name:      "Line count",
		Points:    composition,
		MaxPoints: 10,
		Passed:    lines <= 50,
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
