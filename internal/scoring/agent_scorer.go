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

	// === STRUCTURAL (40 points max) ===
	structural := 0

	// Required fields (5 points each, 25 total)
	requiredFields := []struct {
		name   string
		points int
	}{
		{"name", 5},
		{"description", 5},
		{"model", 5},
		{"tools", 5},
		{"triggers", 5},
	}

	for _, field := range requiredFields {
		_, exists := frontmatter[field.name]
		points := 0
		if exists {
			points = field.points
		}
		structural += points
		details = append(details, ScoringMetric{
			Category:  "structural",
			Name:      "Has " + field.name,
			Points:    points,
			MaxPoints: field.points,
			Passed:    exists,
		})
	}

	// Required sections (15 points total)
	sections := []struct {
		pattern string
		name    string
		points  int
	}{
		{`(?i)## Foundation`, "Foundation section", 5},
		{`(?i)### Phase`, "Phase workflow", 4},
		{`(?i)## Success Criteria`, "Success Criteria", 3},
		{`(?i)## Edge Cases`, "Edge Cases", 3},
	}

	for _, sec := range sections {
		matched, _ := regexp.MatchString(sec.pattern, bodyContent)
		points := 0
		if matched {
			points = sec.points
		}
		structural += points
		details = append(details, ScoringMetric{
			Category:  "structural",
			Name:      sec.name,
			Points:    points,
			MaxPoints: sec.points,
			Passed:    matched,
		})
	}

	// === PRACTICES (40 points max) ===
	practices := 0

	// context_isolation (5 points)
	hasContextIsolation := strings.Contains(content, "context_isolation: true")
	if hasContextIsolation {
		practices += 5
	}
	details = append(details, ScoringMetric{
		Category:  "practices",
		Name:      "context_isolation: true",
		Points:    boolToInt(hasContextIsolation) * 5,
		MaxPoints: 5,
		Passed:    hasContextIsolation,
	})

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
	composition := 0
	var compositionNote string

	switch {
	case lines <= 100:
		composition = 10
		compositionNote = "Excellent: ≤100 lines"
	case lines <= 150:
		composition = 8
		compositionNote = "Good: ≤150 lines"
	case lines <= 200:
		composition = 6
		compositionNote = "OK: ≤200 lines"
	case lines <= 250:
		composition = 3
		compositionNote = "Over limit: >200 lines"
	default:
		composition = 0
		compositionNote = "Fat agent: >250 lines"
	}
	details = append(details, ScoringMetric{
		Category:  "composition",
		Name:      "Line count",
		Points:    composition,
		MaxPoints: 10,
		Passed:    lines <= 200,
		Note:      compositionNote,
	})

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

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
