package scoring

import (
	"regexp"
	"strings"
)

// AgentScorer scores agent files on a 0-100 scale
type AgentScorer struct{}

// NewAgentScorer creates a new AgentScorer
func NewAgentScorer() *AgentScorer {
	return &AgentScorer{}
}

// Score evaluates an agent and returns a QualityScore
func (s *AgentScorer) Score(content string, frontmatter map[string]interface{}, bodyContent string) QualityScore {
	var details []ScoringMetric
	lines := strings.Count(content, "\n") + 1

	// === STRUCTURAL (35 points max) ===
	// Required fields (5 points each, 20 total)
	fieldSpecs := []FieldSpec{
		{"name", 5},
		{"description", 5},
		{"model", 5},
		{"tools", 5},
	}
	fieldScore, fieldDetails := ScoreRequiredFields(frontmatter, fieldSpecs)
	details = append(details, fieldDetails...)

	// Required sections (15 points total)
	sectionSpecs := []SectionSpec{
		{`(?i)## Foundation`, "Foundation section", 5},
		{`(?i)### Phase`, "Phase workflow", 4},
		{`(?i)## Success Criteria`, "Success Criteria", 3},
		{`(?i)## Edge Cases`, "Edge Cases", 3},
	}
	sectionScore, sectionDetails := ScoreSections(bodyContent, sectionSpecs)
	details = append(details, sectionDetails...)

	structural := fieldScore + sectionScore

	// === PRACTICES (35 points max) ===
	practices := 0

	// Skill reference (10 points)
	hasSkillRef, _ := regexp.MatchString(`(?i)Skill:\s*\S+`, bodyContent)
	if hasSkillRef {
		practices += 10
	}
	details = append(details, ScoringMetric{
		Category:  "practices",
		Name:      "Skill: reference",
		Points:    boolToInt(hasSkillRef) * 10,
		MaxPoints: 10,
		Passed:    hasSkillRef,
	})

	// Anti-Patterns section (5 points)
	hasAntiPatterns, _ := regexp.MatchString(`(?i)## Anti-Patterns`, bodyContent)
	if hasAntiPatterns {
		practices += 5
	}
	details = append(details, ScoringMetric{
		Category:  "practices",
		Name:      "Anti-Patterns section",
		Points:    boolToInt(hasAntiPatterns) * 5,
		MaxPoints: 5,
		Passed:    hasAntiPatterns,
	})

	// Expected Output section (5 points)
	hasExpectedOutput, _ := regexp.MatchString(`(?i)## Expected Output`, bodyContent)
	if hasExpectedOutput {
		practices += 5
	}
	details = append(details, ScoringMetric{
		Category:  "practices",
		Name:      "Expected Output section",
		Points:    boolToInt(hasExpectedOutput) * 5,
		MaxPoints: 5,
		Passed:    hasExpectedOutput,
	})

	// HARD GATE markers (5 points)
	hasHardGates, _ := regexp.MatchString(`(?i)HARD GATE`, bodyContent)
	if hasHardGates {
		practices += 5
	}
	details = append(details, ScoringMetric{
		Category:  "practices",
		Name:      "HARD GATE markers",
		Points:    boolToInt(hasHardGates) * 5,
		MaxPoints: 5,
		Passed:    hasHardGates,
	})

	// Third-person description (5 points)
	desc, _ := frontmatter["description"].(string)
	isThirdPerson := !strings.HasPrefix(strings.TrimSpace(desc), "I ")
	if isThirdPerson && len(desc) > 0 {
		practices += 5
	}
	details = append(details, ScoringMetric{
		Category:  "practices",
		Name:      "Third-person description",
		Points:    boolToInt(isThirdPerson && len(desc) > 0) * 5,
		MaxPoints: 5,
		Passed:    isThirdPerson && len(desc) > 0,
	})

	// PROACTIVELY/WHEN in description (5 points)
	hasProactiveTriggers := strings.Contains(strings.ToUpper(desc), "PROACTIVELY") ||
		strings.Contains(strings.ToLower(desc), "use when") ||
		strings.Contains(strings.ToLower(desc), "when user")
	if hasProactiveTriggers {
		practices += 5
	}
	details = append(details, ScoringMetric{
		Category:  "practices",
		Name:      "WHEN triggers in description",
		Points:    boolToInt(hasProactiveTriggers) * 5,
		MaxPoints: 5,
		Passed:    hasProactiveTriggers,
	})

	// === COMPOSITION (10 points max) ===
	agentThresholds := CompositionThresholds{
		Excellent:     100,
		ExcellentNote: "Excellent: ≤100 lines",
		Good:          150,
		GoodNote:      "Good: ≤150 lines",
		OK:            200,
		OKNote:        "OK: ≤200 lines",
		OverLimit:     250,
		OverLimitNote: "Over limit: >200 lines",
		FatNote:       "Fat agent: >250 lines",
	}
	composition, compositionMetric := ScoreComposition(lines, agentThresholds)
	details = append(details, compositionMetric)

	// === DOCUMENTATION (10 points max) ===
	documentation := 0

	// Description length (5 points)
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
	documentation += descPoints
	details = append(details, ScoringMetric{
		Category:  "documentation",
		Name:      "Description quality",
		Points:    descPoints,
		MaxPoints: 5,
		Passed:    descLen >= 100,
		Note:      descNote,
	})

	// Clear section headers (5 points)
	sectionCount := strings.Count(bodyContent, "## ")
	headerPoints := 0
	var headerNote string
	switch {
	case sectionCount >= 6:
		headerPoints = 5
		headerNote = "Well-structured"
	case sectionCount >= 4:
		headerPoints = 3
		headerNote = "Adequate structure"
	case sectionCount >= 2:
		headerPoints = 1
		headerNote = "Minimal structure"
	default:
		headerNote = "Poor structure"
	}
	documentation += headerPoints
	details = append(details, ScoringMetric{
		Category:  "documentation",
		Name:      "Section structure",
		Points:    headerPoints,
		MaxPoints: 5,
		Passed:    sectionCount >= 4,
		Note:      headerNote,
	})

	return NewQualityScore(structural, practices, composition, documentation, details)
}
