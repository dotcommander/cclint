package scoring

import (
	"regexp"
)

// FieldSpec defines a required field with its point value
type FieldSpec struct {
	Name   string
	Points int
}

// SectionSpec defines a section pattern with its display name and point value
type SectionSpec struct {
	Pattern string
	Name    string
	Points  int
}

// CompositionThresholds defines the line count thresholds for composition scoring.
// Each threshold maps to a specific score and note.
type CompositionThresholds struct {
	Excellent     int    // Lines at or below for 10 points
	ExcellentNote string // e.g., "Excellent: <=100 lines"
	Good          int    // Lines at or below for 8 points
	GoodNote      string
	OK            int // Lines at or below for 6 points
	OKNote        string
	OverLimit     int // Lines at or below for 3 points
	OverLimitNote string
	FatNote       string // Note for lines above OverLimit
}

// ScoreRequiredFields scores the presence of required frontmatter fields.
// It follows the Open/Closed Principle - new fields can be added via the specs slice.
func ScoreRequiredFields(frontmatter map[string]any, specs []FieldSpec) (int, []ScoringMetric) {
	var total int
	var details []ScoringMetric

	for _, field := range specs {
		_, exists := frontmatter[field.Name]
		points := 0
		if exists {
			points = field.Points
		}
		total += points
		details = append(details, ScoringMetric{
			Category:  "structural",
			Name:      "Has " + field.Name,
			Points:    points,
			MaxPoints: field.Points,
			Passed:    exists,
		})
	}

	return total, details
}

// ScoreSections scores the presence of required sections in content.
// It follows the Open/Closed Principle - new sections can be added via the specs slice.
func ScoreSections(content string, specs []SectionSpec) (int, []ScoringMetric) {
	var total int
	var details []ScoringMetric

	for _, sec := range specs {
		matched, _ := regexp.MatchString(sec.Pattern, content)
		points := 0
		if matched {
			points = sec.Points
		}
		total += points
		details = append(details, ScoringMetric{
			Category:  "structural",
			Name:      sec.Name,
			Points:    points,
			MaxPoints: sec.Points,
			Passed:    matched,
		})
	}

	return total, details
}

// ScoreSectionsWithFallback scores sections with a custom fallback check function.
// This allows for complex matching logic like checking for alternative patterns.
type SectionFallbackFunc func(content string, sectionName string) bool

// ScoreSectionsWithFallback scores sections with an optional fallback check.
func ScoreSectionsWithFallback(content string, specs []SectionSpec, fallback SectionFallbackFunc) (int, []ScoringMetric) {
	var total int
	var details []ScoringMetric

	for _, sec := range specs {
		matched, _ := regexp.MatchString(sec.Pattern, content)
		// Try fallback if primary pattern didn't match
		if !matched && fallback != nil {
			matched = fallback(content, sec.Name)
		}
		points := 0
		if matched {
			points = sec.Points
		}
		total += points
		details = append(details, ScoringMetric{
			Category:  "structural",
			Name:      sec.Name,
			Points:    points,
			MaxPoints: sec.Points,
			Passed:    matched,
		})
	}

	return total, details
}

// ScoreComposition scores the component based on line count.
// It follows the Open/Closed Principle - thresholds can be customized per component type.
func ScoreComposition(lines int, thresholds CompositionThresholds) (int, ScoringMetric) {
	var points int
	var note string
	var passed bool

	switch {
	case lines <= thresholds.Excellent:
		points = 10
		note = thresholds.ExcellentNote
		passed = true
	case lines <= thresholds.Good:
		points = 8
		note = thresholds.GoodNote
		passed = true
	case lines <= thresholds.OK:
		points = 6
		note = thresholds.OKNote
		passed = true
	case lines <= thresholds.OverLimit:
		points = 3
		note = thresholds.OverLimitNote
		passed = false
	default:
		points = 0
		note = thresholds.FatNote
		passed = false
	}

	return points, ScoringMetric{
		Category:  "composition",
		Name:      "Line count",
		Points:    points,
		MaxPoints: 10,
		Passed:    passed,
		Note:      note,
	}
}

// boolToInt converts a boolean to 0 or 1
func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// IsMethodologySkill returns true if skill has workflow/phase patterns.
// Methodology skills require Success Criteria; reference/pattern skills don't.
func IsMethodologySkill(content string) bool {
	patterns := []string{
		`(?i)## Workflow`,
		`(?i)### Phase \d`,
		`(?i)## Algorithm`,
		`(?i)## Process`,
		`(?i)### Step \d`,
	}
	for _, pattern := range patterns {
		if matched, _ := regexp.MatchString(pattern, content); matched {
			return true
		}
	}
	return false
}
