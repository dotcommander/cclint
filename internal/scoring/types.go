package scoring

// QualityScore represents the overall quality score for a component
type QualityScore struct {
	Overall       int              `json:"overall"`       // 0-100 total score
	Tier          string           `json:"tier"`          // A, B, C, D, F
	Structural    int              `json:"structural"`    // 0-40 structural completeness
	Practices     int              `json:"practices"`     // 0-40 best practices adherence
	Composition   int              `json:"composition"`   // 0-10 size/complexity
	Documentation int              `json:"documentation"` // 0-10 documentation quality
	Details       []ScoringMetric  `json:"details"`       // Detailed breakdown
}

// ScoringMetric represents a single scoring criterion
type ScoringMetric struct {
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
func NewQualityScore(structural, practices, composition, documentation int, details []ScoringMetric) QualityScore {
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
	Score(content string, frontmatter map[string]interface{}, bodyContent string) QualityScore
}
