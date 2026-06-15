package scoring

import (
	"strings"
)

// OutputStyleScorer scores output style files on a 0-100 scale.
type OutputStyleScorer struct{}

// NewOutputStyleScorer creates a new OutputStyleScorer.
func NewOutputStyleScorer() *OutputStyleScorer {
	return &OutputStyleScorer{}
}

// Score evaluates an output style and returns a QualityScore.
func (s *OutputStyleScorer) Score(content string, frontmatter map[string]any, bodyContent string) QualityScore {
	var details []Metric

	// === STRUCTURAL (40 points max) ===
	structural, structuralDetails := s.scoreStructural(content, frontmatter)
	details = append(details, structuralDetails...)

	// === PRACTICES (40 points max) ===
	practices, practiceDetails := s.scorePractices(frontmatter, bodyContent)
	details = append(details, practiceDetails...)

	// === COMPOSITION (10 points max) ===
	composition, compositionMetric := s.scoreComposition(content)
	details = append(details, compositionMetric)

	// === DOCUMENTATION (10 points max) ===
	documentation, docDetails := s.scoreDocumentation(frontmatter, bodyContent)
	details = append(details, docDetails...)

	return NewQualityScore(structural, practices, composition, documentation, details)
}

// scoreStructural scores the structural completeness of an output style.
func (s *OutputStyleScorer) scoreStructural(content string, frontmatter map[string]any) (int, []Metric) {
	var details []Metric
	structural := 0

	// Has frontmatter at all (10 points)
	structural += recordMetric(&details, "structural", "Has frontmatter", strings.HasPrefix(strings.TrimSpace(content), "---"), 10)

	// Has name (15 points)
	structural += recordMetric(&details, "structural", "Has name", hasStringValue(frontmatter, "name"), 15)

	// Has description (15 points)
	structural += recordMetric(&details, "structural", "Has description", hasStringValue(frontmatter, "description"), 15)

	return structural, details
}

// scorePractices scores the best practices adherence of an output style.
func (s *OutputStyleScorer) scorePractices(frontmatter map[string]any, bodyContent string) (int, []Metric) {
	var details []Metric
	practices := 0

	// Has body content (20 points)
	practices += recordMetric(&details, "practices", "Has body content", strings.TrimSpace(bodyContent) != "", 20)

	// Has keep-coding-instructions field (10 points)
	_, hasKeepCoding := frontmatter["keep-coding-instructions"]
	practices += recordMetric(&details, "practices", "Has keep-coding-instructions", hasKeepCoding, 10)

	// Body has meaningful length (10 points)
	bodyLen := len(strings.TrimSpace(bodyContent))
	practices += recordMetric(&details, "practices", "Substantial body content", bodyLen >= 50, 10, bodyLengthNote(bodyLen))

	return practices, details
}

// scoreComposition scores the composition/line count of an output style.
func (s *OutputStyleScorer) scoreComposition(content string) (int, Metric) {
	composition := 0
	lines := strings.Count(content, "\n") + 1
	var compositionNote string
	var compositionPassed bool

	switch {
	case lines <= 50:
		composition = 10
		compositionNote = "Concise: <=50 lines"
		compositionPassed = true
	case lines <= 100:
		composition = 8
		compositionNote = "Good: <=100 lines"
		compositionPassed = true
	case lines <= 200:
		composition = 6
		compositionNote = "OK: <=200 lines"
		compositionPassed = true
	case lines <= 500:
		composition = 3
		compositionNote = "Large: <=500 lines"
		compositionPassed = false
	default:
		composition = 0
		compositionNote = "Too large: >500 lines"
		compositionPassed = false
	}

	return composition, Metric{
		Category:  "composition",
		Name:      "File size",
		Points:    composition,
		MaxPoints: 10,
		Passed:    compositionPassed,
		Note:      compositionNote,
	}
}

// scoreDocumentation scores the documentation quality of an output style.
func (s *OutputStyleScorer) scoreDocumentation(frontmatter map[string]any, bodyContent string) (int, []Metric) {
	var details []Metric
	documentation := 0

	// Description quality (5 points)
	descPoints, descMetric := s.scoreDescriptionQuality(frontmatter)
	documentation += descPoints
	details = append(details, descMetric)

	// Body uses markdown formatting (5 points)
	documentation += recordMetric(&details, "documentation", "Uses markdown formatting", s.hasMarkdownFormatting(bodyContent), 5)

	return documentation, details
}

// scoreDescriptionQuality scores the description based on its length.
func (s *OutputStyleScorer) scoreDescriptionQuality(frontmatter map[string]any) (int, Metric) {
	desc, _ := frontmatter["description"].(string)
	descLen := len(desc)
	descPoints := 0
	var descNote string

	switch {
	case descLen >= 100:
		descPoints = 5
		descNote = descComprehensive
	case descLen >= 50:
		descPoints = 3
		descNote = descAdequate
	case descLen > 0:
		descPoints = 1
		descNote = descBrief
	default:
		descNote = descMissing
	}

	return descPoints, Metric{
		Category:  "documentation",
		Name:      "Description quality",
		Points:    descPoints,
		MaxPoints: 5,
		Passed:    descLen >= 50,
		Note:      descNote,
	}
}

// hasMarkdownFormatting checks if the content uses markdown formatting.
func (s *OutputStyleScorer) hasMarkdownFormatting(bodyContent string) bool {
	return strings.Contains(bodyContent, "#") ||
		strings.Contains(bodyContent, "- ") ||
		strings.Contains(bodyContent, "```")
}

// bodyLengthNote returns a human-readable note for body length.
func bodyLengthNote(length int) string {
	switch {
	case length >= 200:
		return "Rich content"
	case length >= 50:
		return "Adequate content"
	case length > 0:
		return "Minimal content"
	default:
		return "No content"
	}
}
