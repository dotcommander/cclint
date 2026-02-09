package scoring

import (
	"regexp"
	"strings"
)

// antiPatternsSection is the standardized name for anti-patterns sections.
const antiPatternsSection = "Anti-Patterns section"

// SkillScorer scores skill files on a 0-100 scale
type SkillScorer struct{}

// NewSkillScorer creates a new SkillScorer
func NewSkillScorer() *SkillScorer {
	return &SkillScorer{}
}

// Score evaluates a skill and returns a QualityScore
func (s *SkillScorer) Score(content string, frontmatter map[string]any, bodyContent string) QualityScore {
	return computeCombinedScore(content, frontmatter, bodyContent, s)
}

// scoreStructural scores the structural completeness of a skill.
func (s *SkillScorer) scoreStructural(frontmatter map[string]any, bodyContent string) (int, []Metric) {
	var details []Metric

	// Required frontmatter (20 points)
	fieldSpecs := []FieldSpec{
		{"name", 10},
		{"description", 10},
	}
	fieldScore, fieldDetails := ScoreRequiredFields(frontmatter, fieldSpecs)
	details = append(details, fieldDetails...)

	// Required sections (20 points) - different for methodology vs reference skills
	sectionSpecs := s.getSectionSpecs(bodyContent)
	sectionScore, sectionDetails := ScoreSectionsWithFallback(bodyContent, sectionSpecs, s.antiPatternsFallback)
	details = append(details, sectionDetails...)

	return fieldScore + sectionScore, details
}

// getSectionSpecs returns the section specifications based on skill type.
func (s *SkillScorer) getSectionSpecs(bodyContent string) []SectionSpec {
	if IsMethodologySkill(bodyContent) {
		return []SectionSpec{
			{`(?i)## Quick Reference`, "Quick Reference", 8},
			{`(?i)## Workflow`, "Workflow section", 6},
			{`(?i)(## Anti-Patterns?|### Anti-Patterns?|\| Anti-Pattern)`, antiPatternsSection, 4},
			{`(?i)## Success Criteria`, "Success Criteria", 2},
		}
	}
	return []SectionSpec{
		{`(?i)## Quick Reference`, "Quick Reference", 10},
		{`(?i)(## Patterns?|## Templates?|## Examples?)`, "Pattern/Template section", 6},
		{`(?i)(## Anti-Patterns?|### Anti-Patterns?|\| Anti-Pattern)`, antiPatternsSection, 4},
	}
}

// antiPatternsFallback provides fallback detection for Anti-Patterns section.
func (s *SkillScorer) antiPatternsFallback(content string, sectionName string) bool {
	if sectionName == antiPatternsSection {
		hasBestPractices := strings.Contains(content, "## Best Practices")
		hasDont := strings.Contains(strings.ToLower(content), "### don't")
		return hasBestPractices && hasDont
	}
	return false
}

// scorePractices scores the best practices adherence of a skill.
func (s *SkillScorer) scorePractices(bodyContent string) (int, []Metric) {
	var details []Metric
	practices := 0

	// Semantic routing table format (10 points)
	practices += addPracticeMetric(practiceMetricCheck{
		details:  &details,
		name:     "Semantic routing table",
		pattern:  "\\|.*User Question.*\\|.*Action.*\\|",
		points:   10,
		content:  bodyContent,
		useRegex: true,
	})

	// Phase-based workflow (8 points)
	practices += addPracticeMetric(practiceMetricCheck{
		details:  &details,
		name:     "Phase-based workflow",
		pattern:  "(?i)### Phase \\d",
		points:   8,
		content:  bodyContent,
		useRegex: true,
	})

	// Anti-patterns as table (6 points)
	practices += addPracticeMetric(practiceMetricCheck{
		details:  &details,
		name:     "Anti-patterns table format",
		pattern:  "\\|.*Anti-Pattern.*\\|.*Problem.*\\|.*Fix.*\\|",
		points:   6,
		content:  bodyContent,
		useRegex: true,
	})

	// HARD GATE markers (4 points)
	practices += addPracticeMetric(practiceMetricCheck{
		details:  &details,
		name:     "HARD GATE markers",
		pattern:  "(?i)HARD GATE",
		points:   4,
		content:  bodyContent,
		useRegex: true,
	})

	// Success criteria checkboxes (4 points) - string contains check
	hasCheckboxes := strings.Contains(bodyContent, "- [ ]")
	if hasCheckboxes {
		practices += 4
	}
	details = append(details, Metric{
		Category: "practices", Name: "Success criteria checkboxes",
		Points: boolToInt(hasCheckboxes) * 4, MaxPoints: 4, Passed: hasCheckboxes,
	})

	// References to references/ directory (4 points)
	practices += addPracticeMetric(practiceMetricCheck{
		details:  &details,
		name:     "References to references/",
		pattern:  "references/\\w+\\.md",
		points:   4,
		content:  bodyContent,
		useRegex: true,
	})

	// Scoring formula (4 points)
	practices += addPracticeMetric(practiceMetricCheck{
		details:  &details,
		name:     "Scoring formula",
		pattern:  "(?i)(score\\s*=|scoring formula)",
		points:   4,
		content:  bodyContent,
		useRegex: true,
	})

	return practices, details
}

// practiceMetricCheck encapsulates parameters for a practice metric check.
type practiceMetricCheck struct {
	details  *[]Metric
	name     string
	pattern  string
	points   int
	content  string
	useRegex bool
}

// addPracticeMetric checks for a pattern and adds a metric if present.
// Returns the points awarded.
func addPracticeMetric(check practiceMetricCheck) int {
	hasMatch := false
	if check.useRegex {
		hasMatch, _ = regexp.MatchString(check.pattern, check.content)
	} else {
		hasMatch = strings.Contains(check.content, check.pattern)
	}

	*check.details = append(*check.details, Metric{
		Category:  "practices",
		Name:      check.name,
		Points:    boolToInt(hasMatch) * check.points,
		MaxPoints: check.points,
		Passed:    hasMatch,
	})

	if hasMatch {
		return check.points
	}
	return 0
}

// scoreComposition scores the composition/line count of a skill.
func (s *SkillScorer) scoreComposition(lines int) (int, Metric) {
	skillThresholds := CompositionThresholds{
		Excellent:     250,
		ExcellentNote: "Excellent: ≤250 lines",
		Good:          400,
		GoodNote:      "Good: ≤400 lines",
		OK:            550,
		OKNote:        "OK: ≤550 lines (500±10%)",
		OverLimit:     660,
		OverLimitNote: "Over limit: >550 lines",
		FatNote:       "Fat skill: >660 lines",
	}
	return ScoreComposition(lines, skillThresholds)
}

// scoreDocumentation scores the documentation quality of a skill.
func (s *SkillScorer) scoreDocumentation(frontmatter map[string]any, bodyContent string) (int, []Metric) {
	var details []Metric
	documentation := 0

	// Description quality (5 points)
	descPoints, descMetric := s.scoreDescriptionQuality(frontmatter)
	documentation += descPoints
	details = append(details, descMetric)

	// Code examples (5 points)
	codePoints, codeMetric := s.scoreCodeExamples(bodyContent)
	documentation += codePoints
	details = append(details, codeMetric)

	return documentation, details
}

// scoreDescriptionQuality scores the description based on its length.
func (s *SkillScorer) scoreDescriptionQuality(frontmatter map[string]any) (int, Metric) {
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

	return descPoints, Metric{
		Category:  "documentation",
		Name:      "Description quality",
		Points:    descPoints,
		MaxPoints: 5,
		Passed:    descLen >= 100,
		Note:      descNote,
	}
}

// scoreCodeExamples scores the code examples based on code block count.
func (s *SkillScorer) scoreCodeExamples(bodyContent string) (int, Metric) {
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

	return codePoints, Metric{
		Category:  "documentation",
		Name:      "Code examples",
		Points:    codePoints,
		MaxPoints: 5,
		Passed:    codeBlockCount >= 3,
		Note:      codeNote,
	}
}
