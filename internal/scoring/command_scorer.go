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
	return computeCombinedScore(content, frontmatter, bodyContent, s)
}

// scoreStructural evaluates required frontmatter and task delegation (40 points max).
func (s *CommandScorer) scoreStructural(frontmatter map[string]any, body string) (int, []Metric) {
	fieldSpecs := []FieldSpec{
		{"allowed-tools", 10},
		{"description", 10},
		{"argument-hint", 10},
	}
	points, details := ScoreRequiredFields(frontmatter, fieldSpecs)

	hasTaskDelegation, _ := regexp.MatchString(`Task\([^)]+\)`, body)
	points += recordMetric(&details, "structural", "Task() delegation", hasTaskDelegation, 10)

	return points, details
}

// scorePractices evaluates thin-command patterns (40 points max).
func (s *CommandScorer) scorePractices(body string) (int, []Metric) {
	var details []Metric
	points := 0

	hasSuccessCriteria, _ := regexp.MatchString(`(?i)Success criteria|^\s*- \[ \]`, body)
	points += recordMetric(&details, "practices", "Success criteria", hasSuccessCriteria, 15)

	taskCount := strings.Count(body, "Task(")
	hasTask := taskCount >= 1
	points += recordMetric(&details, "practices", "Task delegation", hasTask, 15, pluralize(taskCount, "Task() call"))

	hasFlags, _ := regexp.MatchString(`(?i)## Flags|--\w+`, body)
	points += recordMetric(&details, "practices", "Flags documented", hasFlags, 10)

	return points, details
}

// scoreComposition evaluates file length against command thresholds (10 points max).
func (s *CommandScorer) scoreComposition(lines int) (int, Metric) {
	thresholds := CompositionThresholds{
		Excellent: 30, ExcellentNote: "Excellent: ≤30 lines",
		Good: 45, GoodNote: "Good: ≤45 lines",
		OK: 55, OKNote: "OK: ≤55 lines (50±10%)",
		OverLimit: 65, OverLimitNote: "Over limit: >55 lines",
		FatNote: "Fat command: >65 lines",
	}
	return ScoreComposition(lines, thresholds)
}

// scoreDocumentation evaluates description quality and examples (10 points max).
func (s *CommandScorer) scoreDocumentation(frontmatter map[string]any, body string) (int, []Metric) {
	var details []Metric
	points := 0

	desc, _ := frontmatter["description"].(string)
	descPoints, descNote := scoreDescriptionQuality(len(desc))
	points += descPoints
	details = append(details, Metric{
		Category: "documentation", Name: "Description quality",
		Points: descPoints, MaxPoints: 5, Passed: len(desc) >= 20, Note: descNote,
	})

	hasCodeExamples := strings.Contains(body, "```bash") || strings.Contains(body, "```")
	points += recordMetric(&details, "documentation", "Code examples", hasCodeExamples, 5)

	return points, details
}

// scoreDescriptionQuality returns points and note for a description of the given length.
func scoreDescriptionQuality(length int) (int, string) {
	switch {
	case length >= 50:
		return 5, descClear
	case length >= 20:
		return 3, descBrief
	case length > 0:
		return 1, descMinimal
	default:
		return 0, descMissing
	}
}

func pluralize(count int, singular string) string {
	if count == 1 {
		return "1 " + singular
	}
	return strings.Replace(singular, "call", "calls", 1)
}
