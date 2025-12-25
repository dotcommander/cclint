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
func (s *PluginScorer) Score(content string, frontmatter map[string]interface{}, bodyContent string) QualityScore {
	var details []ScoringMetric

	// === STRUCTURAL (40 points max) ===
	// Required fields (10 points each)
	structural := 0

	// name (10 points)
	hasName := false
	if name, ok := frontmatter["name"].(string); ok && name != "" {
		structural += 10
		hasName = true
	}
	details = append(details, ScoringMetric{
		Category:  "structural",
		Name:      "Has name",
		Points:    boolToInt(hasName) * 10,
		MaxPoints: 10,
		Passed:    hasName,
	})

	// description (10 points)
	hasDescription := false
	if desc, ok := frontmatter["description"].(string); ok && desc != "" {
		structural += 10
		hasDescription = true
	}
	details = append(details, ScoringMetric{
		Category:  "structural",
		Name:      "Has description",
		Points:    boolToInt(hasDescription) * 10,
		MaxPoints: 10,
		Passed:    hasDescription,
	})

	// version (10 points)
	hasVersion := false
	if version, ok := frontmatter["version"].(string); ok && version != "" {
		structural += 10
		hasVersion = true
	}
	details = append(details, ScoringMetric{
		Category:  "structural",
		Name:      "Has version",
		Points:    boolToInt(hasVersion) * 10,
		MaxPoints: 10,
		Passed:    hasVersion,
	})

	// author.name (10 points)
	hasAuthorName := false
	if author, ok := frontmatter["author"].(map[string]interface{}); ok {
		if authorName, ok := author["name"].(string); ok && authorName != "" {
			structural += 10
			hasAuthorName = true
		}
	}
	details = append(details, ScoringMetric{
		Category:  "structural",
		Name:      "Has author.name",
		Points:    boolToInt(hasAuthorName) * 10,
		MaxPoints: 10,
		Passed:    hasAuthorName,
	})

	// === PRACTICES (40 points max) ===
	practices := 0

	// homepage (10 points)
	hasHomepage := false
	if homepage, ok := frontmatter["homepage"].(string); ok && homepage != "" {
		practices += 10
		hasHomepage = true
	}
	details = append(details, ScoringMetric{
		Category:  "practices",
		Name:      "Has homepage",
		Points:    boolToInt(hasHomepage) * 10,
		MaxPoints: 10,
		Passed:    hasHomepage,
	})

	// repository (10 points)
	hasRepository := false
	if repo, ok := frontmatter["repository"].(string); ok && repo != "" {
		practices += 10
		hasRepository = true
	}
	details = append(details, ScoringMetric{
		Category:  "practices",
		Name:      "Has repository",
		Points:    boolToInt(hasRepository) * 10,
		MaxPoints: 10,
		Passed:    hasRepository,
	})

	// license (10 points)
	hasLicense := false
	if license, ok := frontmatter["license"].(string); ok && license != "" {
		practices += 10
		hasLicense = true
	}
	details = append(details, ScoringMetric{
		Category:  "practices",
		Name:      "Has license",
		Points:    boolToInt(hasLicense) * 10,
		MaxPoints: 10,
		Passed:    hasLicense,
	})

	// keywords (10 points)
	hasKeywords := false
	if keywords, ok := frontmatter["keywords"].([]interface{}); ok && len(keywords) > 0 {
		practices += 10
		hasKeywords = true
	}
	details = append(details, ScoringMetric{
		Category:  "practices",
		Name:      "Has keywords",
		Points:    boolToInt(hasKeywords) * 10,
		MaxPoints: 10,
		Passed:    hasKeywords,
	})

	// === COMPOSITION (10 points max) ===
	composition := 0
	var compositionNote string
	var compositionPassed bool

	// File size check
	fileSize := len(content)
	switch {
	case fileSize <= 1000:
		composition = 10
		compositionNote = "Excellent: ≤1KB"
		compositionPassed = true
	case fileSize <= 2000:
		composition = 8
		compositionNote = "Good: ≤2KB"
		compositionPassed = true
	case fileSize <= 5000:
		composition = 6
		compositionNote = "OK: ≤5KB"
		compositionPassed = true
	case fileSize <= 10000:
		composition = 3
		compositionNote = "Large: ≤10KB"
		compositionPassed = false
	default:
		composition = 0
		compositionNote = "Too large: >10KB"
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

	// Description length (5 points)
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
	case descLen >= 20:
		descPoints = 1
		descNote = "Brief"
	default:
		descNote = "Too short"
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

	// README reference (5 points)
	hasReadme := false
	if readme, ok := frontmatter["readme"].(string); ok && readme != "" {
		documentation += 5
		hasReadme = true
	}
	details = append(details, ScoringMetric{
		Category:  "documentation",
		Name:      "Has readme",
		Points:    boolToInt(hasReadme) * 5,
		MaxPoints: 5,
		Passed:    hasReadme,
	})

	return NewQualityScore(structural, practices, composition, documentation, details)
}

// ValidateJSON checks if content is valid JSON
func ValidateJSON(content string) error {
	var data map[string]interface{}
	return json.Unmarshal([]byte(content), &data)
}
