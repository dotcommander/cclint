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
	var details []ScoringMetric

	// === STRUCTURAL (40 points max) ===
	structural := 0

	// Has frontmatter at all (10 points)
	hasFrontmatter := strings.HasPrefix(strings.TrimSpace(content), "---")
	if hasFrontmatter {
		structural += 10
	}
	details = append(details, ScoringMetric{
		Category:  "structural",
		Name:      "Has frontmatter",
		Points:    boolToInt(hasFrontmatter) * 10,
		MaxPoints: 10,
		Passed:    hasFrontmatter,
	})

	// Has name (15 points)
	hasName := false
	if name, ok := frontmatter["name"].(string); ok && name != "" {
		structural += 15
		hasName = true
	}
	details = append(details, ScoringMetric{
		Category:  "structural",
		Name:      "Has name",
		Points:    boolToInt(hasName) * 15,
		MaxPoints: 15,
		Passed:    hasName,
	})

	// Has description (15 points)
	hasDescription := false
	if desc, ok := frontmatter["description"].(string); ok && desc != "" {
		structural += 15
		hasDescription = true
	}
	details = append(details, ScoringMetric{
		Category:  "structural",
		Name:      "Has description",
		Points:    boolToInt(hasDescription) * 15,
		MaxPoints: 15,
		Passed:    hasDescription,
	})

	// === PRACTICES (40 points max) ===
	practices := 0

	// Has body content (20 points)
	hasBody := strings.TrimSpace(bodyContent) != ""
	if hasBody {
		practices += 20
	}
	details = append(details, ScoringMetric{
		Category:  "practices",
		Name:      "Has body content",
		Points:    boolToInt(hasBody) * 20,
		MaxPoints: 20,
		Passed:    hasBody,
	})

	// Has keep-coding-instructions field (10 points)
	hasKeepCoding := false
	if _, ok := frontmatter["keep-coding-instructions"]; ok {
		practices += 10
		hasKeepCoding = true
	}
	details = append(details, ScoringMetric{
		Category:  "practices",
		Name:      "Has keep-coding-instructions",
		Points:    boolToInt(hasKeepCoding) * 10,
		MaxPoints: 10,
		Passed:    hasKeepCoding,
	})

	// Body has meaningful length (10 points)
	bodyLen := len(strings.TrimSpace(bodyContent))
	hasSubstantialBody := bodyLen >= 50
	if hasSubstantialBody {
		practices += 10
	}
	details = append(details, ScoringMetric{
		Category:  "practices",
		Name:      "Substantial body content",
		Points:    boolToInt(hasSubstantialBody) * 10,
		MaxPoints: 10,
		Passed:    hasSubstantialBody,
		Note:      bodyLengthNote(bodyLen),
	})

	// === COMPOSITION (10 points max) ===
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

	details = append(details, ScoringMetric{
		Category:  "composition",
		Name:      "File size",
		Points:    composition,
		MaxPoints: 10,
		Passed:    compositionPassed,
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
	case descLen >= 100:
		descPoints = 5
		descNote = "Comprehensive"
	case descLen >= 50:
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
		Passed:    descLen >= 50,
		Note:      descNote,
	})

	// Body uses markdown formatting (5 points)
	hasFormatting := strings.Contains(bodyContent, "#") ||
		strings.Contains(bodyContent, "- ") ||
		strings.Contains(bodyContent, "```")
	if hasFormatting {
		documentation += 5
	}
	details = append(details, ScoringMetric{
		Category:  "documentation",
		Name:      "Uses markdown formatting",
		Points:    boolToInt(hasFormatting) * 5,
		MaxPoints: 5,
		Passed:    hasFormatting,
	})

	return NewQualityScore(structural, practices, composition, documentation, details)
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
