package scoring

import (
	"encoding/json"
)

// PluginScorer scores plugin manifest files on a 0-100 scale
type PluginScorer struct{}

// NewPluginScorer creates a new PluginScorer
func NewPluginScorer() *PluginScorer {
	return &PluginScorer{}
}

// Score evaluates a plugin manifest and returns a QualityScore
// For plugins, frontmatter is the parsed JSON data, bodyContent is unused
func (s *PluginScorer) Score(content string, frontmatter map[string]any, bodyContent string) QualityScore {
	var details []Metric

	// === STRUCTURAL (40 points max) ===
	structural, structuralDetails := s.scoreStructural(frontmatter)
	details = append(details, structuralDetails...)

	// === PRACTICES (40 points max) ===
	practices, practiceDetails := s.scorePractices(frontmatter)
	details = append(details, practiceDetails...)

	// === COMPOSITION (10 points max) ===
	composition, compositionMetric := s.scoreComposition(content)
	details = append(details, compositionMetric)

	// === DOCUMENTATION (10 points max) ===
	documentation, docDetails := s.scoreDocumentation(frontmatter)
	details = append(details, docDetails...)

	return NewQualityScore(structural, practices, composition, documentation, details)
}

// scoreStructural evaluates required fields for structure (40 points max).
func (s *PluginScorer) scoreStructural(frontmatter map[string]any) (int, []Metric) {
	var details []Metric
	structural := 0

	// name (10 points)
	structural += recordMetric(&details, "structural", "Has name", hasStringValue(frontmatter, "name"), 10)

	// description (10 points)
	structural += recordMetric(&details, "structural", "Has description", hasStringValue(frontmatter, "description"), 10)

	// version (10 points)
	structural += recordMetric(&details, "structural", "Has version", hasStringValue(frontmatter, "version"), 10)

	// author.name (10 points)
	hasAuthorName := false
	if author, ok := frontmatter["author"].(map[string]any); ok {
		if authorName, ok := author["name"].(string); ok && authorName != "" {
			hasAuthorName = true
		}
	}
	structural += recordMetric(&details, "structural", "Has author.name", hasAuthorName, 10)

	return structural, details
}

// scorePractices evaluates best practices (40 points max).
func (s *PluginScorer) scorePractices(frontmatter map[string]any) (int, []Metric) {
	var details []Metric
	practices := 0

	// homepage (10 points)
	practices += recordMetric(&details, "practices", "Has homepage", hasStringValue(frontmatter, "homepage"), 10)

	// repository (10 points)
	practices += recordMetric(&details, "practices", "Has repository", hasStringValue(frontmatter, "repository"), 10)

	// license (10 points)
	practices += recordMetric(&details, "practices", "Has license", hasStringValue(frontmatter, "license"), 10)

	// keywords (10 points)
	hasKeywords := false
	if keywords, ok := frontmatter["keywords"].([]any); ok && len(keywords) > 0 {
		hasKeywords = true
	}
	practices += recordMetric(&details, "practices", "Has keywords", hasKeywords, 10)

	return practices, details
}

// scoreComposition evaluates file size (10 points max).
func (s *PluginScorer) scoreComposition(content string) (int, Metric) {
	fileSize := len(content)
	var points int
	var note string
	var passed bool

	switch {
	case fileSize <= 1000:
		points, note, passed = 10, "Excellent: ≤1KB", true //nolint:gosec // False positive - this is a size rating string, not a credential
	case fileSize <= 2000:
		points, note, passed = 8, "Good: ≤2KB", true //nolint:gosec // False positive - this is a size rating string, not a credential
	case fileSize <= 5000:
		points, note, passed = 6, "OK: ≤5KB", true //nolint:gosec // False positive - this is a size rating string, not a credential
	case fileSize <= 10000:
		points, note, passed = 3, "Large: ≤10KB", false //nolint:gosec // False positive - this is a size rating string, not a credential
	default:
		points, note, passed = 0, "Too large: >10KB", false //nolint:gosec // False positive - this is a size rating string, not a credential
	}

	return points, Metric{Category: "composition", Name: "File size", Points: points, MaxPoints: 10, Passed: passed, Note: note}
}

// scoreDocumentation evaluates documentation quality (10 points max).
func (s *PluginScorer) scoreDocumentation(frontmatter map[string]any) (int, []Metric) {
	var details []Metric
	documentation := 0

	// Description length (5 points)
	desc, _ := frontmatter["description"].(string)
	descLen := len(desc)
	descPoints, descNote := s.scoreDescriptionLength(descLen)
	documentation += descPoints
	details = append(details, Metric{Category: "documentation", Name: "Description quality", Points: descPoints, MaxPoints: 5, Passed: descLen >= 50, Note: descNote})

	// README reference (5 points)
	documentation += recordMetric(&details, "documentation", "Has readme", hasStringValue(frontmatter, "readme"), 5)

	return documentation, details
}

// scoreDescriptionLength returns points and note based on description length.
func (s *PluginScorer) scoreDescriptionLength(length int) (int, string) {
	switch {
	case length >= 100:
		return 5, "Comprehensive"
	case length >= 50:
		return 3, "Adequate"
	case length >= 20:
		return 1, "Brief"
	default:
		return 0, "Too short"
	}
}

// ValidateJSON checks if content is valid JSON
func ValidateJSON(content string) error {
	var data map[string]any
	return json.Unmarshal([]byte(content), &data)
}
