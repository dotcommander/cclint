package scoring

import (
	"strings"
	"testing"
)

func TestNewAgentScorer(t *testing.T) {
	scorer := NewAgentScorer()
	if scorer == nil {
		t.Fatal("NewAgentScorer() returned nil")
	}
}

func TestAgentScorer_Score(t *testing.T) {
	tests := []struct {
		name          string
		frontmatter   map[string]interface{}
		bodyContent   string
		wantTier      string
		wantStructMin int
		wantPractMin  int
		wantCompMin   int
		wantDocMin    int
	}{
		{
			name: "Perfect agent",
			frontmatter: map[string]interface{}{
				"name":        "test-agent",
				"description": "PROACTIVELY handles test cases when user needs testing. This is a comprehensive description that is at least 200 characters long to ensure we get full points for description quality. It provides detailed information about the agent's capabilities and use cases.",
				"model":       "claude-3-5-sonnet-20241022",
				"tools":       []string{"Read", "Write"},
			},
			bodyContent: `## Foundation

Skill: test-patterns
Skill: another-skill

### Phase 1: Analysis

HARD GATE: Must complete analysis

### Phase 2: Execution

Work here

## Success Criteria

- [ ] Tests pass
- [ ] Coverage > 80%

## Edge Cases

Handle these cases

## Anti-Patterns

Don't do this

## Expected Output

Results here
`,
			wantTier:      "A",
			wantStructMin: 35,
			wantPractMin:  30,
			wantCompMin:   8,
			wantDocMin:    8,
		},
		{
			name: "Minimal agent - missing most features",
			frontmatter: map[string]interface{}{
				"name": "minimal",
			},
			bodyContent: "Basic content",
			wantTier:    "F",
		},
		{
			name: "Agent with Skill() format",
			frontmatter: map[string]interface{}{
				"name":        "test",
				"description": "Test agent",
				"model":       "claude-3-5-sonnet-20241022",
				"tools":       []string{"Read"},
			},
			bodyContent: `
Skill("test-patterns")
Skill('another-pattern')
`,
			wantStructMin: 20,
			wantPractMin:  10, // Skill reference
		},
		{
			name: "Agent with **Skill**: format",
			frontmatter: map[string]interface{}{
				"name":        "test",
				"description": "Test agent",
				"model":       "claude-3-5-sonnet-20241022",
				"tools":       []string{"Read"},
			},
			bodyContent: `
**Skill**: test-patterns
`,
			wantStructMin: 20,
			wantPractMin:  10,
		},
		{
			name: "Agent with Skills: list format",
			frontmatter: map[string]interface{}{
				"name":        "test",
				"description": "Test agent",
				"model":       "claude-3-5-sonnet-20241022",
				"tools":       []string{"Read"},
			},
			bodyContent: `
Skills:
- pattern-one
- pattern-two
`,
			wantStructMin: 20,
			wantPractMin:  10,
		},
		{
			name: "Third-person description",
			frontmatter: map[string]interface{}{
				"name":        "test",
				"description": "Handles testing tasks effectively",
				"model":       "claude-3-5-sonnet-20241022",
				"tools":       []string{"Read"},
			},
			bodyContent: "Content",
			wantPractMin: 5, // Third-person description
		},
		{
			name: "First-person description (fails third-person check)",
			frontmatter: map[string]interface{}{
				"name":        "test",
				"description": "I handle testing tasks",
				"model":       "claude-3-5-sonnet-20241022",
				"tools":       []string{"Read"},
			},
			bodyContent: "Content",
			// Should not get third-person points
		},
		{
			name: "WHEN triggers in description",
			frontmatter: map[string]interface{}{
				"name":        "test",
				"description": "Use when user needs testing assistance",
				"model":       "claude-3-5-sonnet-20241022",
				"tools":       []string{"Read"},
			},
			bodyContent: "Content",
			wantPractMin: 10, // Third-person (5) + WHEN trigger (5)
		},
		{
			name: "PROACTIVELY in description",
			frontmatter: map[string]interface{}{
				"name":        "test",
				"description": "PROACTIVELY runs tests when code changes",
				"model":       "claude-3-5-sonnet-20241022",
				"tools":       []string{"Read"},
			},
			bodyContent: "Content",
			wantPractMin: 10,
		},
		{
			name: "Large agent - over 220 lines",
			frontmatter: map[string]interface{}{
				"name":        "large-agent",
				"description": "Large agent",
				"model":       "claude-3-5-sonnet-20241022",
				"tools":       []string{"Read"},
			},
			bodyContent: strings.Repeat("Line of content\n", 250),
			wantCompMin:  0, // Over limit
		},
		{
			name: "Well-structured agent - many sections",
			frontmatter: map[string]interface{}{
				"name":        "test",
				"description": "Test agent with comprehensive description that is at least 100 characters long to get adequate points",
				"model":       "claude-3-5-sonnet-20241022",
				"tools":       []string{"Read"},
			},
			bodyContent: `
## Foundation
## Workflow
## Success Criteria
## Edge Cases
## Anti-Patterns
## Expected Output
`,
			wantDocMin: 8, // Description (3) + Structure (5)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scorer := NewAgentScorer()
			content := generateFullContent(tt.frontmatter, tt.bodyContent)
			score := scorer.Score(content, tt.frontmatter, tt.bodyContent)

			if tt.wantTier != "" && score.Tier != tt.wantTier {
				t.Errorf("Score() tier = %q, want %q (overall: %d)", score.Tier, tt.wantTier, score.Overall)
			}
			if tt.wantStructMin > 0 && score.Structural < tt.wantStructMin {
				t.Errorf("Score() structural = %d, want >= %d", score.Structural, tt.wantStructMin)
			}
			if tt.wantPractMin > 0 && score.Practices < tt.wantPractMin {
				t.Errorf("Score() practices = %d, want >= %d", score.Practices, tt.wantPractMin)
			}
			if tt.wantCompMin > 0 && score.Composition < tt.wantCompMin {
				t.Errorf("Score() composition = %d, want >= %d", score.Composition, tt.wantCompMin)
			}
			if tt.wantDocMin > 0 && score.Documentation < tt.wantDocMin {
				t.Errorf("Score() documentation = %d, want >= %d", score.Documentation, tt.wantDocMin)
			}

			// Verify overall score is sum of categories
			expectedOverall := score.Structural + score.Practices + score.Composition + score.Documentation
			if score.Overall != expectedOverall {
				t.Errorf("Overall score = %d, want %d (sum of categories)", score.Overall, expectedOverall)
			}

			// Verify details are present
			if len(score.Details) == 0 {
				t.Error("Score() should have details")
			}
		})
	}
}

func TestAgentScorer_SkillReferenceDetection(t *testing.T) {
	tests := []struct {
		name        string
		bodyContent string
		wantPoints  bool
	}{
		{"Skill: format", "Skill: test-pattern", true},
		{"**Skill**: format", "**Skill**: test-pattern", true},
		{"Skill() format", "Skill(test-pattern)", true},
		{"Skill() with quotes", `Skill("test-pattern")`, true},
		{"Skill() with single quotes", `Skill('test-pattern')`, true},
		{"Skills: list", "Skills:\n- pattern", true},
		{"No skill reference", "No skills here", false},
		{"Partial match", "Skilled worker", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scorer := NewAgentScorer()
			frontmatter := map[string]interface{}{
				"name":        "test",
				"description": "Test",
				"model":       "claude-3-5-sonnet-20241022",
				"tools":       []string{"Read"},
			}
			content := generateFullContent(frontmatter, tt.bodyContent)
			score := scorer.Score(content, frontmatter, tt.bodyContent)

			hasSkillPoints := false
			for _, detail := range score.Details {
				if detail.Name == "Skill: reference" && detail.Points > 0 {
					hasSkillPoints = true
					break
				}
			}

			if hasSkillPoints != tt.wantPoints {
				t.Errorf("Skill reference detected = %v, want %v", hasSkillPoints, tt.wantPoints)
			}
		})
	}
}

func TestAgentScorer_DescriptionQuality(t *testing.T) {
	scorer := NewAgentScorer()

	tests := []struct {
		name       string
		desc       string
		wantPoints int
		wantNote   string
	}{
		{
			name:       "Comprehensive description (>=200 chars)",
			desc:       strings.Repeat("a", 200),
			wantPoints: 5,
			wantNote:   "Comprehensive",
		},
		{
			name:       "Adequate description (>=100 chars)",
			desc:       strings.Repeat("a", 100),
			wantPoints: 3,
			wantNote:   "Adequate",
		},
		{
			name:       "Brief description (>0 chars)",
			desc:       "Short",
			wantPoints: 1,
			wantNote:   "Brief",
		},
		{
			name:       "Missing description",
			desc:       "",
			wantPoints: 0,
			wantNote:   "Missing",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			frontmatter := map[string]interface{}{
				"name":        "test",
				"description": tt.desc,
				"model":       "claude-3-5-sonnet-20241022",
				"tools":       []string{"Read"},
			}
			content := generateFullContent(frontmatter, "Body content")
			score := scorer.Score(content, frontmatter, "Body content")

			var descMetric *ScoringMetric
			for i := range score.Details {
				if score.Details[i].Name == "Description quality" {
					descMetric = &score.Details[i]
					break
				}
			}

			if descMetric == nil {
				t.Fatal("Description quality metric not found")
			}

			if descMetric.Points != tt.wantPoints {
				t.Errorf("Description points = %d, want %d", descMetric.Points, tt.wantPoints)
			}
			if descMetric.Note != tt.wantNote {
				t.Errorf("Description note = %q, want %q", descMetric.Note, tt.wantNote)
			}
		})
	}
}

func TestAgentScorer_SectionStructure(t *testing.T) {
	scorer := NewAgentScorer()

	tests := []struct {
		name         string
		bodyContent  string
		wantPoints   int
		wantNote     string
		wantPassed   bool
	}{
		{
			name:        "Well-structured (>=6 sections)",
			bodyContent: "## S1\n## S2\n## S3\n## S4\n## S5\n## S6",
			wantPoints:  5,
			wantNote:    "Well-structured",
			wantPassed:  true,
		},
		{
			name:        "Adequate structure (>=4 sections)",
			bodyContent: "## S1\n## S2\n## S3\n## S4",
			wantPoints:  3,
			wantNote:    "Adequate structure",
			wantPassed:  true,
		},
		{
			name:        "Minimal structure (>=2 sections)",
			bodyContent: "## S1\n## S2",
			wantPoints:  1,
			wantNote:    "Minimal structure",
			wantPassed:  false,
		},
		{
			name:        "Poor structure (<2 sections)",
			bodyContent: "## S1",
			wantPoints:  0,
			wantNote:    "Poor structure",
			wantPassed:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			frontmatter := map[string]interface{}{
				"name":        "test",
				"description": "Test agent",
				"model":       "claude-3-5-sonnet-20241022",
				"tools":       []string{"Read"},
			}
			content := generateFullContent(frontmatter, tt.bodyContent)
			score := scorer.Score(content, frontmatter, tt.bodyContent)

			var structMetric *ScoringMetric
			for i := range score.Details {
				if score.Details[i].Name == "Section structure" {
					structMetric = &score.Details[i]
					break
				}
			}

			if structMetric == nil {
				t.Fatal("Section structure metric not found")
			}

			if structMetric.Points != tt.wantPoints {
				t.Errorf("Points = %d, want %d", structMetric.Points, tt.wantPoints)
			}
			if structMetric.Note != tt.wantNote {
				t.Errorf("Note = %q, want %q", structMetric.Note, tt.wantNote)
			}
			if structMetric.Passed != tt.wantPassed {
				t.Errorf("Passed = %v, want %v", structMetric.Passed, tt.wantPassed)
			}
		})
	}
}

func TestAgentScorer_CompositionScoring(t *testing.T) {
	scorer := NewAgentScorer()

	// Note: lines are counted as strings.Count(content, "\n") + 1
	// So N repeats of "Line\n" creates N newlines = N+1 lines counted
	// To hit threshold exactly, use threshold - 1 repeats
	tests := []struct {
		name       string
		lines      int // number of "Line\n" repeats
		wantPoints int
		wantPassed bool
	}{
		{"Excellent - under 120 lines", 100, 10, true},  // 101 lines counted, ≤120
		{"Good - under 180 lines", 170, 8, true},        // 171 lines counted, ≤180
		{"Good boundary - 179 lines", 179, 8, true},     // 180 lines counted, =180
		{"OK - 200 lines counted", 199, 6, true},        // 200 lines counted, ≤220
		{"OK boundary - 219 lines", 219, 6, true},       // 220 lines counted, =220
		{"Over limit - 250 lines", 250, 3, false},       // 251 lines counted, ≤275
		{"Fat agent - 300 lines", 300, 0, false},        // 301 lines counted, >275
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			frontmatter := map[string]interface{}{
				"name":        "test",
				"description": "Test agent",
				"model":       "claude-3-5-sonnet-20241022",
				"tools":       []string{"Read"},
			}
			bodyContent := strings.Repeat("Line\n", tt.lines)
			content := generateFullContent(frontmatter, bodyContent)
			score := scorer.Score(content, frontmatter, bodyContent)

			if score.Composition != tt.wantPoints {
				t.Errorf("Composition score = %d, want %d", score.Composition, tt.wantPoints)
			}

			var compMetric *ScoringMetric
			for i := range score.Details {
				if score.Details[i].Category == "composition" {
					compMetric = &score.Details[i]
					break
				}
			}

			if compMetric == nil {
				t.Fatal("Composition metric not found")
			}

			if compMetric.Passed != tt.wantPassed {
				t.Errorf("Composition passed = %v, want %v", compMetric.Passed, tt.wantPassed)
			}
		})
	}
}

// Helper function to generate full content with frontmatter
func generateFullContent(frontmatter map[string]interface{}, body string) string {
	// Simple content generation - just return body since we're passing frontmatter separately
	return body
}
