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

// AgentScorer scores agent files on a 0-100 scale
type AgentScorer struct{}

// NewAgentScorer creates a new AgentScorer
func NewAgentScorer() *AgentScorer {
	return &AgentScorer{}
}

// Score evaluates an agent and returns a QualityScore
func (s *AgentScorer) Score(content string, frontmatter map[string]any, bodyContent string) QualityScore {
	var details []Metric
	lines := strings.Count(content, "\n") + 1

	// === STRUCTURAL (35 points max) ===
	structural, structuralDetails := s.scoreStructural(frontmatter, bodyContent)
	details = append(details, structuralDetails...)

	// === PRACTICES (35 points max) ===
	practices, practiceDetails := s.scorePractices(frontmatter, bodyContent)
	details = append(details, practiceDetails...)

	// === COMPOSITION (10 points max) ===
	composition, compositionMetric := s.scoreComposition(lines)
	details = append(details, compositionMetric)

	// === DOCUMENTATION (10 points max) ===
	documentation, docDetails := s.scoreDocumentation(frontmatter, bodyContent)
	details = append(details, docDetails...)

	return NewQualityScore(structural, practices, composition, documentation, details)
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
func (s *AgentScorer) scorePractices(frontmatter map[string]any, bodyContent string) (int, []Metric) {
	var details []Metric
	practices := 0

	// Skill reference (10 points)
	hasSkillRef := s.hasSkillReference(bodyContent)
	if hasSkillRef {
		practices += 10
	}
	details = append(details, Metric{
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
	details = append(details, Metric{
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
	details = append(details, Metric{
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
	details = append(details, Metric{
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
	details = append(details, Metric{
		Category:  "practices",
		Name:      "Third-person description",
		Points:    boolToInt(isThirdPerson && len(desc) > 0) * 5,
		MaxPoints: 5,
		Passed:    isThirdPerson && len(desc) > 0,
	})

	// PROACTIVELY/WHEN in description (5 points)
	hasProactiveTriggers := s.hasProactiveTriggers(desc)
	if hasProactiveTriggers {
		practices += 5
	}
	details = append(details, Metric{
		Category:  "practices",
		Name:      "WHEN triggers in description",
		Points:    boolToInt(hasProactiveTriggers) * 5,
		MaxPoints: 5,
		Passed:    hasProactiveTriggers,
	})

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
