package scoring

import (
	"strings"
	"testing"
)

func TestNewSkillScorer(t *testing.T) {
	scorer := NewSkillScorer()
	if scorer == nil {
		t.Fatal("NewSkillScorer() returned nil")
	}
}

func TestSkillScorer_Score(t *testing.T) {
	tests := []struct {
		name          string
		frontmatter   map[string]any
		bodyContent   string
		wantTier      string
		wantStructMin int
		wantPractMin  int
		wantCompMin   int
		wantDocMin    int
	}{
		{
			name: "Perfect methodology skill",
			frontmatter: map[string]any{
				"name":        "test-patterns",
				"description": strings.Repeat("Comprehensive skill description that provides detailed information about the methodology and its application in various contexts. ", 2),
			},
			bodyContent: `
## Quick Reference

| User Question | Action |
|---------------|--------|
| How to test? | Read(references/testing.md) |

## Workflow

### Phase 1: Analysis

Analyze requirements

### Phase 2: Implementation

Implement solution

## Anti-Patterns

| Anti-Pattern | Problem | Fix |
|--------------|---------|-----|
| No tests | Bad | Add tests |

## Success Criteria

- [ ] Tests pass
- [ ] Coverage >= 80%

HARD GATE: Must verify tests

References in references/patterns.md

Scoring formula: score = coverage * quality

` + "```\ncode1\n```\n```\ncode2\n```\n```\ncode3\n```" + `
`,
			wantTier:      "A",
			wantStructMin: 35,
			wantPractMin:  30,
			wantCompMin:   8,
			wantDocMin:    8,
		},
		{
			name: "Perfect reference/pattern skill",
			frontmatter: map[string]any{
				"name":        "go-patterns",
				"description": strings.Repeat("Comprehensive pattern library with detailed examples and best practices for Go development. ", 2),
			},
			bodyContent: `
## Quick Reference

| User Question | Action |
|---------------|--------|
| Factory pattern? | See below |

## Patterns

Multiple pattern examples here

## Anti-Patterns

| Anti-Pattern | Problem | Fix |
|--------------|---------|-----|
| Global state | Bad | Dependency injection |

- [ ] Understand pattern
- [ ] Apply correctly

HARD GATE: Must follow best practices

See references/examples.md for more

Scoring formula: quality = correctness * readability

` + "```go\ncode example\n```\n```go\nmore code\n```\n```go\nyet more\n```\n```\nextra1\n```\n```\nextra2\n```\n```\nextra3\n```" + `
`,
			wantTier:      "A",
			wantStructMin: 30,
			wantPractMin:  28, // Semantic routing (10) + anti-patterns table (6) + checkboxes (4) + HARD GATE (4) + references (4)
			wantCompMin:   8,
			wantDocMin:    8,
		},
		{
			name: "Minimal skill - missing features",
			frontmatter: map[string]any{
				"name": "minimal",
			},
			bodyContent: "Basic content",
			wantTier:    "F",
		},
		{
			name: "Skill with semantic routing",
			frontmatter: map[string]any{
				"name":        "test",
				"description": "Test skill",
			},
			bodyContent: `
| User Question | Action |
|---------------|--------|
| How? | Do this |
`,
			wantStructMin: 10, // Has name, Quick Reference
			wantPractMin:  10, // Semantic routing
		},
		{
			name: "Skill with phase-based workflow",
			frontmatter: map[string]any{
				"name":        "test",
				"description": "Test skill",
			},
			bodyContent: `
## Workflow

### Phase 1
### Phase 2
### Phase 3
`,
			wantStructMin: 10,
			wantPractMin:  8, // Phase-based workflow
		},
		{
			name: "Skill with anti-patterns table",
			frontmatter: map[string]any{
				"name":        "test",
				"description": "Test skill",
			},
			bodyContent: `
| Anti-Pattern | Problem | Fix |
|--------------|---------|-----|
| Bad thing | Bad | Good thing |
`,
			wantPractMin: 6, // Anti-patterns table
		},
		{
			name: "Skill with Best Practices fallback for anti-patterns",
			frontmatter: map[string]any{
				"name":        "test",
				"description": "Test skill",
			},
			bodyContent: `
## Best Practices

### Don't

- Don't do this
- Don't do that
`,
			wantStructMin: 10, // Should get anti-patterns points via fallback
		},
		{
			name: "Skill with HARD GATE markers",
			frontmatter: map[string]any{
				"name":        "test",
				"description": "Test skill",
			},
			bodyContent: `
HARD GATE: Must verify
HARD GATE: Must validate
`,
			wantPractMin: 4,
		},
		{
			name: "Skill with checkboxes",
			frontmatter: map[string]any{
				"name":        "test",
				"description": "Test skill",
			},
			bodyContent: `
- [ ] First
- [ ] Second
- [ ] Third
`,
			wantPractMin: 4,
		},
		{
			name: "Skill with references",
			frontmatter: map[string]any{
				"name":        "test",
				"description": "Test skill",
			},
			bodyContent: `
See references/patterns.md
Also references/examples.md
`,
			wantPractMin: 4,
		},
		{
			name: "Skill with scoring formula",
			frontmatter: map[string]any{
				"name":        "test",
				"description": "Test skill",
			},
			bodyContent: `
Score = quality * completeness
Scoring formula: points / maxPoints
`,
			wantPractMin: 4,
		},
		{
			name: "Large skill - over 550 lines",
			frontmatter: map[string]any{
				"name":        "large-skill",
				"description": "Large skill",
			},
			bodyContent: strings.Repeat("Line of content\n", 600),
			wantCompMin:  0,
		},
		{
			name: "Excellent size skill - under 250 lines",
			frontmatter: map[string]any{
				"name":        "concise-skill",
				"description": strings.Repeat("Good description here. ", 5),
			},
			bodyContent: strings.Repeat("Line\n", 200) + "\n```\ncode\n```\n```\nmore\n```\n```\nyet more\n```",
			wantCompMin:  8,
			wantDocMin:   8,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scorer := NewSkillScorer()
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

func TestSkillScorer_MethodologyDetection(t *testing.T) {
	scorer := NewSkillScorer()

	tests := []struct {
		name         string
		bodyContent  string
		isMethodology bool
	}{
		{
			name: "Methodology - has Workflow",
			bodyContent: `
## Workflow
Steps here
`,
			isMethodology: true,
		},
		{
			name: "Methodology - has Phases",
			bodyContent: `
### Phase 1
### Phase 2
`,
			isMethodology: true,
		},
		{
			name: "Reference - no methodology markers",
			bodyContent: `
## Quick Reference
Pattern library
`,
			isMethodology: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			frontmatter := map[string]any{
				"name":        "test",
				"description": "Test skill",
			}
			content := generateFullContent(frontmatter, tt.bodyContent)
			score := scorer.Score(content, frontmatter, tt.bodyContent)

			// Methodology skills should have different section requirements
			// Check if Success Criteria is present in details
			hasSuccessCriteria := false
			for _, detail := range score.Details {
				if detail.Name == "Success Criteria" {
					hasSuccessCriteria = true
					break
				}
			}

			if tt.isMethodology != hasSuccessCriteria {
				t.Errorf("Methodology detection mismatch: isMethodology=%v, hasSuccessCriteria=%v",
					tt.isMethodology, hasSuccessCriteria)
			}
		})
	}
}

func TestSkillScorer_DescriptionQuality(t *testing.T) {
	scorer := NewSkillScorer()

	tests := []struct {
		name       string
		desc       string
		wantPoints int
		wantNote   string
		wantPassed bool
	}{
		{
			name:       "Comprehensive (>=200 chars)",
			desc:       strings.Repeat("a", 200),
			wantPoints: 5,
			wantNote:   "Comprehensive",
			wantPassed: true,
		},
		{
			name:       "Adequate (>=100 chars)",
			desc:       strings.Repeat("a", 100),
			wantPoints: 3,
			wantNote:   "Adequate",
			wantPassed: true,
		},
		{
			name:       "Brief (>0 chars)",
			desc:       "Short",
			wantPoints: 1,
			wantNote:   "Brief",
			wantPassed: false,
		},
		{
			name:       "Missing",
			desc:       "",
			wantPoints: 0,
			wantNote:   "Missing",
			wantPassed: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			frontmatter := map[string]any{
				"name":        "test",
				"description": tt.desc,
			}
			content := generateFullContent(frontmatter, "Body")
			score := scorer.Score(content, frontmatter, "Body")

			var descMetric *Metric
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
				t.Errorf("Points = %d, want %d", descMetric.Points, tt.wantPoints)
			}
			if descMetric.Note != tt.wantNote {
				t.Errorf("Note = %q, want %q", descMetric.Note, tt.wantNote)
			}
			if descMetric.Passed != tt.wantPassed {
				t.Errorf("Passed = %v, want %v", descMetric.Passed, tt.wantPassed)
			}
		})
	}
}

func TestSkillScorer_CodeExamples(t *testing.T) {
	scorer := NewSkillScorer()

	tests := []struct {
		name       string
		body       string
		wantPoints int
		wantNote   string
		wantPassed bool
	}{
		{
			name:       "Rich examples (>=6 backtick marks)",
			body:       "```\na\n```\n```\nb\n```\n```\nc\n```", // 6 backtick marks
			wantPoints: 5,
			wantNote:   "Rich examples",
			wantPassed: true,
		},
		{
			name:       "Adequate examples (3-5 backtick marks)",
			body:       "```\na\n```\n```\nb\n```", // 4 backtick marks
			wantPoints: 3,
			wantNote:   "Adequate examples",
			wantPassed: true,
		},
		{
			name:       "Few examples (1-2 backtick marks)",
			body:       "```\ncode\n```", // 2 backtick marks
			wantPoints: 1,
			wantNote:   "Few examples",
			wantPassed: false,
		},
		{
			name:       "No examples",
			body:       "No code blocks",
			wantPoints: 0,
			wantNote:   "No examples",
			wantPassed: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			frontmatter := map[string]any{
				"name":        "test",
				"description": "Test skill",
			}
			content := generateFullContent(frontmatter, tt.body)
			score := scorer.Score(content, frontmatter, tt.body)

			var codeMetric *Metric
			for i := range score.Details {
				if score.Details[i].Name == "Code examples" {
					codeMetric = &score.Details[i]
					break
				}
			}

			if codeMetric == nil {
				t.Fatal("Code examples metric not found")
			}

			if codeMetric.Points != tt.wantPoints {
				t.Errorf("Points = %d, want %d", codeMetric.Points, tt.wantPoints)
			}
			if codeMetric.Note != tt.wantNote {
				t.Errorf("Note = %q, want %q", codeMetric.Note, tt.wantNote)
			}
			if codeMetric.Passed != tt.wantPassed {
				t.Errorf("Passed = %v, want %v", codeMetric.Passed, tt.wantPassed)
			}
		})
	}
}

func TestSkillScorer_CompositionScoring(t *testing.T) {
	scorer := NewSkillScorer()

	// Note: lines counted as strings.Count(content, "\n") + 1
	tests := []struct {
		name       string
		repeats    int // number of "Line\n" to add
		wantPoints int
		wantPassed bool
	}{
		{"Excellent - 200 lines", 199, 10, true},
		{"Good - under 400", 350, 8, true},
		{"Good boundary - 399", 398, 8, true},
		{"OK - 500 lines", 499, 6, true},
		{"OK boundary - 549", 548, 6, true},
		{"Over limit - 600", 599, 3, false},
		{"Fat skill - 800", 799, 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			frontmatter := map[string]any{
				"name":        "test",
				"description": "Test skill",
			}
			bodyContent := strings.Repeat("Line\n", tt.repeats)
			content := generateFullContent(frontmatter, bodyContent)
			score := scorer.Score(content, frontmatter, bodyContent)

			if score.Composition != tt.wantPoints {
				t.Errorf("Composition score = %d, want %d", score.Composition, tt.wantPoints)
			}

			var compMetric *Metric
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
				t.Errorf("Passed = %v, want %v", compMetric.Passed, tt.wantPassed)
			}
		})
	}
}

func TestSkillScorer_SemanticRouting(t *testing.T) {
	scorer := NewSkillScorer()

	tests := []struct {
		name       string
		body       string
		wantPoints bool
	}{
		{
			name: "Has semantic routing table",
			body: `
| User Question | Action |
|---------------|--------|
| How? | Do this |
`,
			wantPoints: true,
		},
		{
			name: "Different column names",
			body: `
| User Question | Action |
| What? | Read file |
`,
			wantPoints: true,
		},
		{
			name:       "No semantic routing",
			body:       "Just content",
			wantPoints: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			frontmatter := map[string]any{
				"name":        "test",
				"description": "Test skill",
			}
			content := generateFullContent(frontmatter, tt.body)
			score := scorer.Score(content, frontmatter, tt.body)

			var routingMetric *Metric
			for i := range score.Details {
				if score.Details[i].Name == "Semantic routing table" {
					routingMetric = &score.Details[i]
					break
				}
			}

			if routingMetric == nil {
				t.Fatal("Semantic routing metric not found")
			}

			if routingMetric.Passed != tt.wantPoints {
				t.Errorf("Semantic routing passed = %v, want %v", routingMetric.Passed, tt.wantPoints)
			}
		})
	}
}

func TestSkillScorer_AntPatternsFallback(t *testing.T) {
	scorer := NewSkillScorer()

	tests := []struct {
		name       string
		body       string
		wantPoints bool
	}{
		{
			name: "Standard anti-patterns section",
			body: `
## Anti-Patterns
Bad things
`,
			wantPoints: true,
		},
		{
			name: "Anti-patterns table",
			body: `
| Anti-Pattern | Problem | Fix |
|--------------|---------|-----|
| Bad | Why | Good |
`,
			wantPoints: true,
		},
		{
			name: "Best Practices with Don't subsection (fallback)",
			body: `
## Best Practices

### Don't

- Don't do this
`,
			wantPoints: true,
		},
		{
			name: "Best Practices without Don't subsection",
			body: `
## Best Practices

### Do

- Do this
`,
			wantPoints: false,
		},
		{
			name:       "No anti-patterns",
			body:       "Just content",
			wantPoints: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			frontmatter := map[string]any{
				"name":        "test",
				"description": "Test skill",
			}
			content := generateFullContent(frontmatter, tt.body)
			score := scorer.Score(content, frontmatter, tt.body)

			var antiPatternMetric *Metric
			for i := range score.Details {
				if score.Details[i].Name == "Anti-Patterns section" {
					antiPatternMetric = &score.Details[i]
					break
				}
			}

			if antiPatternMetric == nil {
				t.Fatal("Anti-Patterns section metric not found")
			}

			if antiPatternMetric.Passed != tt.wantPoints {
				t.Errorf("Anti-patterns passed = %v, want %v", antiPatternMetric.Passed, tt.wantPoints)
			}
		})
	}
}

func TestSkillScorer_AllPracticesMetrics(t *testing.T) {
	scorer := NewSkillScorer()

	frontmatter := map[string]any{
		"name":        "comprehensive-skill",
		"description": strings.Repeat("Comprehensive description. ", 10),
	}

	bodyContent := `
## Quick Reference

| User Question | Action |
|---------------|--------|
| How? | Do this |

## Workflow

### Phase 1
First step

### Phase 2
Second step

## Anti-Patterns

| Anti-Pattern | Problem | Fix |
|--------------|---------|-----|
| Bad | Why | Good |

## Success Criteria

- [ ] First
- [ ] Second

HARD GATE: Must verify

See references/patterns.md

Score = quality * completeness

` + "```\ncode\n```\n```\nmore\n```\n```\nyet more\n```"

	content := generateFullContent(frontmatter, bodyContent)
	score := scorer.Score(content, frontmatter, bodyContent)

	expectedMetrics := map[string]bool{
		"Semantic routing table":       true,
		"Phase-based workflow":         true,
		"Anti-patterns table format":   true,
		"HARD GATE markers":            true,
		"Success criteria checkboxes":  true,
		"References to references/":    true,
		"Scoring formula":              true,
	}

	for metricName, shouldPass := range expectedMetrics {
		found := false
		for _, detail := range score.Details {
			if detail.Name == metricName {
				found = true
				if detail.Passed != shouldPass {
					t.Errorf("Metric %q: passed = %v, want %v", metricName, detail.Passed, shouldPass)
				}
				if detail.Category != "practices" {
					t.Errorf("Metric %q: category = %q, want %q", metricName, detail.Category, "practices")
				}
				break
			}
		}
		if !found {
			t.Errorf("Metric %q not found in details", metricName)
		}
	}
}
