package scoring

import (
	"regexp"
	"strings"
)

// Description quality rating constants to avoid goconst warnings.
const (
	descClear         = "Clear"
	descComprehensive = "Comprehensive"
	descAdequate      = "Adequate"
	descBrief         = "Brief"
	descMinimal       = "Minimal"
	descMissing       = "Missing"
)

// AgentScorer scores agent files on a 0-100 scale.
// Implements ScorerComponent for use with computeCombinedScore.
type AgentScorer struct {
	frontmatter map[string]any // set by Score before delegation
}

// NewAgentScorer creates a new AgentScorer
func NewAgentScorer() *AgentScorer {
	return &AgentScorer{}
}

// Score evaluates an agent and returns a QualityScore.
// Delegates to computeCombinedScore via the ScorerComponent interface.
func (s *AgentScorer) Score(content string, frontmatter map[string]any, bodyContent string) QualityScore {
	s.frontmatter = frontmatter
	return computeCombinedScore(content, frontmatter, bodyContent, s)
}

// scoreStructural scores the structural completeness of an agent.
func (s *AgentScorer) scoreStructural(frontmatter map[string]any, bodyContent string) (int, []Metric) {
	var details []Metric

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

	return fieldScore + sectionScore, details
}

// scorePractices scores the best practices adherence of an agent.
func (s *AgentScorer) scorePractices(bodyContent string) (int, []Metric) {
	frontmatter := s.frontmatter
	var details []Metric
	practices := 0

	add := func(name string, passed bool, points int) {
		if passed {
			practices += points
		}
		details = append(details, Metric{
			Category:  "practices",
			Name:      name,
			Points:    boolToInt(passed) * points,
			MaxPoints: points,
			Passed:    passed,
		})
	}

	// Skill reference (10 points)
	add("Skill: reference", s.hasSkillReference(bodyContent), 10)

	// Anti-Patterns section (5 points)
	hasAntiPatterns, _ := regexp.MatchString(`(?i)## Anti-Patterns`, bodyContent)
	add("Anti-Patterns section", hasAntiPatterns, 5)

	// Expected Output section (5 points)
	hasExpectedOutput, _ := regexp.MatchString(`(?i)## Expected Output`, bodyContent)
	add("Expected Output section", hasExpectedOutput, 5)

	// HARD GATE markers (5 points)
	hasHardGates, _ := regexp.MatchString(`(?i)HARD GATE`, bodyContent)
	add("HARD GATE markers", hasHardGates, 5)

	// Third-person description (5 points)
	desc, _ := frontmatter["description"].(string)
	isThirdPerson := !strings.HasPrefix(strings.TrimSpace(desc), "I ") && len(desc) > 0
	add("Third-person description", isThirdPerson, 5)

	// PROACTIVELY/WHEN in description (5 points)
	add("WHEN triggers in description", s.hasProactiveTriggers(desc), 5)

	return practices, details
}

// hasSkillReference checks if the content contains skill references.
func (s *AgentScorer) hasSkillReference(bodyContent string) bool {
	skillPatterns := []string{
		`(?i)Skill:\s*\S+`,
		`(?i)\*\*Skill\*\*:\s*\S+`,
		`(?i)Skill\(\s*["']?[a-z0-9-]+`,
		`(?i)Skills:\s*\n`,
	}
	for _, pattern := range skillPatterns {
		if matched, _ := regexp.MatchString(pattern, bodyContent); matched {
			return true
		}
	}
	return false
}

// hasProactiveTriggers checks if the description contains proactive trigger keywords.
func (s *AgentScorer) hasProactiveTriggers(desc string) bool {
	return strings.Contains(strings.ToUpper(desc), "PROACTIVELY") ||
		strings.Contains(strings.ToLower(desc), "use when") ||
		strings.Contains(strings.ToLower(desc), "when user")
}

// scoreComposition scores the composition/line count of an agent.
func (s *AgentScorer) scoreComposition(lines int) (int, Metric) {
	agentThresholds := CompositionThresholds{
		Excellent:     120,
		ExcellentNote: "Excellent: ≤120 lines",
		Good:          180,
		GoodNote:      "Good: ≤180 lines",
		OK:            220,
		OKNote:        "OK: ≤220 lines (200±10%)",
		OverLimit:     275,
		OverLimitNote: "Over limit: >220 lines",
		FatNote:       "Fat agent: >275 lines",
	}
	return ScoreComposition(lines, agentThresholds)
}

// scoreDocumentation scores the documentation quality of an agent.
func (s *AgentScorer) scoreDocumentation(frontmatter map[string]any, bodyContent string) (int, []Metric) {
	var details []Metric
	documentation := 0
	desc, _ := frontmatter["description"].(string)

	// Description length (5 points)
	descPoints, descMetric := s.scoreDescriptionLength(desc)
	documentation += descPoints
	details = append(details, descMetric)

	// Clear section headers (5 points)
	sectionCount := strings.Count(bodyContent, "## ")
	headerPoints, headerMetric := s.scoreSectionStructure(sectionCount)
	documentation += headerPoints
	details = append(details, headerMetric)

	return documentation, details
}

// scoreDescriptionLength scores the description based on its length.
func (s *AgentScorer) scoreDescriptionLength(desc string) (int, Metric) {
	descLen := len(desc)
	descPoints := 0
	var descNote string

	switch {
	case descLen >= 200:
		descPoints = 5
		descNote = descComprehensive
	case descLen >= 100:
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
		Passed:    descLen >= 100,
		Note:      descNote,
	}
}

// scoreSectionStructure scores the section structure based on header count.
func (s *AgentScorer) scoreSectionStructure(sectionCount int) (int, Metric) {
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

	return headerPoints, Metric{
		Category:  "documentation",
		Name:      "Section structure",
		Points:    headerPoints,
		MaxPoints: 5,
		Passed:    sectionCount >= 4,
		Note:      headerNote,
	}
}
