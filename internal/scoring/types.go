package scoring

import "strings"

// QualityScore represents the overall quality score for a component
type QualityScore struct {
	Overall       int      `json:"overall"`       // 0-100 total score
	Tier          string   `json:"tier"`          // A, B, C, D, F
	Structural    int      `json:"structural"`    // 0-40 structural completeness
	Practices     int      `json:"practices"`     // 0-40 best practices adherence
	Composition   int      `json:"composition"`   // 0-10 size/complexity
	Documentation int      `json:"documentation"` // 0-10 documentation quality
	Details       []Metric `json:"details"`       // Detailed breakdown
}

// Metric represents a single scoring criterion
type Metric struct {
	Category  string `json:"category"`   // structural, practices, composition, documentation
	Name      string `json:"name"`       // Human-readable name
	Points    int    `json:"points"`     // Actual points earned
	MaxPoints int    `json:"max_points"` // Maximum possible points
	Passed    bool   `json:"passed"`     // Whether this check passed
	Note      string `json:"note"`       // Optional note/reason
}

// Tier returns the quality tier based on score
func TierFromScore(score int) string {
	switch {
	case score >= 85:
		return "A"
	case score >= 70:
		return "B"
	case score >= 50:
		return "C"
	case score >= 30:
		return "D"
	default:
		return "F"
	}
}

// NewQualityScore creates a new QualityScore and calculates the overall score
func NewQualityScore(structural, practices, composition, documentation int, details []Metric) QualityScore {
	overall := structural + practices + composition + documentation
	return QualityScore{
		Overall:       overall,
		Tier:          TierFromScore(overall),
		Structural:    structural,
		Practices:     practices,
		Composition:   composition,
		Documentation: documentation,
		Details:       details,
	}
}

// Scorer is the interface for component scorers
type Scorer interface {
	Score(content string, frontmatter map[string]any, bodyContent string) QualityScore
}

// ScorerComponent defines the component-specific scoring methods that each scorer
// type (Command, Skill, Agent, Plugin) implements. This allows the shared Score
// method to delegate to type-specific logic.
type ScorerComponent interface {
	// scoreStructural evaluates required frontmatter and structural elements (40 points max).
	scoreStructural(frontmatter map[string]any, body string) (int, []Metric)
	// scorePractices evaluates best practices adherence (40 points max).
	scorePractices(body string) (int, []Metric)
	// scoreComposition evaluates file length against thresholds (10 points max).
	scoreComposition(lines int) (int, Metric)
	// scoreDocumentation evaluates description quality and examples (10 points max).
	scoreDocumentation(frontmatter map[string]any, body string) (int, []Metric)
}

// computeCombinedScore combines all scoring components into a final QualityScore.
// This is the shared implementation for all scorer types.
func computeCombinedScore(
	content string,
	frontmatter map[string]any,
	bodyContent string,
	component ScorerComponent,
) QualityScore {
	var details []Metric
	lines := strings.Count(content, "\n") + 1

	structural, structDetails := component.scoreStructural(frontmatter, bodyContent)
	details = append(details, structDetails...)

	practices, practDetails := component.scorePractices(bodyContent)
	details = append(details, practDetails...)

	composition, compMetric := component.scoreComposition(lines)
	details = append(details, compMetric)

	documentation, docDetails := component.scoreDocumentation(frontmatter, bodyContent)
	details = append(details, docDetails...)

	return NewQualityScore(structural, practices, composition, documentation, details)
}
