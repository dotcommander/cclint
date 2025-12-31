package scoring

import (
	"strings"
	"testing"
)

func TestNewCommandScorer(t *testing.T) {
	scorer := NewCommandScorer()
	if scorer == nil {
		t.Fatal("NewCommandScorer() returned nil")
	}
}

func TestCommandScorer_Score(t *testing.T) {
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
			name: "Perfect command",
			frontmatter: map[string]interface{}{
				"allowed-tools":  []string{"Read", "Write", "Task"},
				"description":    "Comprehensive command that does many things effectively",
				"argument-hint":  "file-path",
			},
			bodyContent: `
Task(agent-type: "general-purpose", prompt: "Do work")

## Flags

--dry-run: Preview changes
--verbose: Show details

## Success Criteria

- [ ] Task completes
- [ ] Output is correct

## Examples

` + "```bash" + `
command --flag value
` + "```" + `
`,
			wantTier:      "A",
			wantStructMin: 35,
			wantPractMin:  35,
			wantCompMin:   8,
			wantDocMin:    8,
		},
		{
			name: "Minimal command - missing features",
			frontmatter: map[string]interface{}{
				"allowed-tools": []string{"Task"},
			},
			bodyContent: "Basic content",
			wantTier:    "F",
		},
		{
			name: "Command with Task delegation",
			frontmatter: map[string]interface{}{
				"allowed-tools":  []string{"Task"},
				"description":    "Test command",
				"argument-hint":  "input",
			},
			bodyContent: `
Task(agent-type: "test-specialist", prompt: "Run tests")
Task(agent-type: "report-generator", prompt: "Generate report")
`,
			wantStructMin: 40, // All required fields + Task delegation
			wantPractMin:  15, // Task delegation present
		},
		{
			name: "Command with success criteria",
			frontmatter: map[string]interface{}{
				"allowed-tools":  []string{"Task"},
				"description":    "Test command",
				"argument-hint":  "input",
			},
			bodyContent: `
## Success Criteria

- [ ] Tests pass
- [ ] Coverage >= 80%
- [ ] No errors

Task(agent-type: "test", prompt: "test")
`,
			wantPractMin: 30, // Success criteria (15) + Task delegation (15)
		},
		{
			name: "Command with checkboxes format",
			frontmatter: map[string]interface{}{
				"allowed-tools":  []string{"Task"},
				"description":    "Test",
				"argument-hint":  "input",
			},
			bodyContent: `
- [ ] First criteria
- [ ] Second criteria

Task(test)
`,
			wantPractMin: 15, // Success criteria detected via checkbox
		},
		{
			name: "Command with flags documentation",
			frontmatter: map[string]interface{}{
				"allowed-tools":  []string{"Task"},
				"description":    "Test",
				"argument-hint":  "input",
			},
			bodyContent: `
## Flags

--dry-run: Preview mode
--force: Skip confirmations

Task(test)
`,
			wantPractMin: 25, // Task (15) + Flags (10)
		},
		{
			name: "Command with inline flags",
			frontmatter: map[string]interface{}{
				"allowed-tools":  []string{"Task"},
				"description":    "Test",
				"argument-hint":  "input",
			},
			bodyContent: `
Use --verbose for detailed output.
Use --quiet to suppress output.

Task(test)
`,
			wantPractMin: 25, // Task (15) + Flags (10)
		},
		{
			name: "Large command - over 55 lines",
			frontmatter: map[string]interface{}{
				"allowed-tools":  []string{"Task"},
				"description":    "Large command",
				"argument-hint":  "input",
			},
			bodyContent: strings.Repeat("Line of content\n", 60),
			wantCompMin:  0, // Over limit
		},
		{
			name: "Small command - under 30 lines",
			frontmatter: map[string]interface{}{
				"allowed-tools":  []string{"Task"},
				"description":    "Concise command with good description length here",
				"argument-hint":  "input",
			},
			bodyContent: "Task(test)\n\n" + strings.Repeat("Line\n", 5),
			wantCompMin:  8,
			wantDocMin:   3, // Description >= 50 chars gives 5 points, no code examples = 3 total
		},
		{
			name: "Command with code examples",
			frontmatter: map[string]interface{}{
				"allowed-tools":  []string{"Task"},
				"description":    "Test command with examples and good description",
				"argument-hint":  "input",
			},
			bodyContent: "```bash\ncommand arg\n```\n\nTask(test)",
			wantDocMin:   8, // Description (3 for <50 chars) + Code examples (5)
		},
		{
			name: "Command with generic code block",
			frontmatter: map[string]interface{}{
				"allowed-tools":  []string{"Task"},
				"description":    "Test command with great description goes here now",
				"argument-hint":  "input",
			},
			bodyContent: "```\ncode here\n```\n\nTask(test)",
			wantDocMin:   8, // Description (3) + Code examples (5)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scorer := NewCommandScorer()
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

func TestCommandScorer_TaskDelegationDetection(t *testing.T) {
	tests := []struct {
		name        string
		bodyContent string
		wantStruct  bool
		wantPract   bool
	}{
		{"Has Task() call", "Task(agent-type: 'test', prompt: 'test')", true, true},
		{"Has Task() with content", "Task(test)", true, true},
		{"Has multiple Task() calls", "Task(a)\nTask(b)\nTask(c)", true, true},
		{"No Task() call", "No delegation here", false, false},
		{"Partial match", "Tasker() not a match", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scorer := NewCommandScorer()
			frontmatter := map[string]interface{}{
				"allowed-tools":  []string{"Task"},
				"description":    "Test",
				"argument-hint":  "input",
			}
			content := generateFullContent(frontmatter, tt.bodyContent)
			score := scorer.Score(content, frontmatter, tt.bodyContent)

			// Check structural Task delegation metric
			var structMetric *ScoringMetric
			var practMetric *ScoringMetric
			for i := range score.Details {
				if score.Details[i].Name == "Task() delegation" && score.Details[i].Category == "structural" {
					structMetric = &score.Details[i]
				}
				if score.Details[i].Name == "Task delegation" && score.Details[i].Category == "practices" {
					practMetric = &score.Details[i]
				}
			}

			if structMetric == nil {
				t.Fatal("Structural Task() delegation metric not found")
			}
			if practMetric == nil {
				t.Fatal("Practices Task delegation metric not found")
			}

			if structMetric.Passed != tt.wantStruct {
				t.Errorf("Structural Task delegation = %v, want %v", structMetric.Passed, tt.wantStruct)
			}
			if practMetric.Passed != tt.wantPract {
				t.Errorf("Practices Task delegation = %v, want %v", practMetric.Passed, tt.wantPract)
			}
		})
	}
}

func TestCommandScorer_DescriptionQuality(t *testing.T) {
	scorer := NewCommandScorer()

	tests := []struct {
		name       string
		desc       string
		wantPoints int
		wantNote   string
		wantPassed bool
	}{
		{
			name:       "Clear description (>=50 chars)",
			desc:       strings.Repeat("a", 50),
			wantPoints: 5,
			wantNote:   "Clear",
			wantPassed: true,
		},
		{
			name:       "Brief description (>=20 chars)",
			desc:       strings.Repeat("a", 20),
			wantPoints: 3,
			wantNote:   "Brief",
			wantPassed: true,
		},
		{
			name:       "Minimal description (>0 chars)",
			desc:       "Short",
			wantPoints: 1,
			wantNote:   "Minimal",
			wantPassed: false,
		},
		{
			name:       "Missing description",
			desc:       "",
			wantPoints: 0,
			wantNote:   "Missing",
			wantPassed: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			frontmatter := map[string]interface{}{
				"allowed-tools":  []string{"Task"},
				"description":    tt.desc,
				"argument-hint":  "input",
			}
			content := generateFullContent(frontmatter, "Task(test)")
			score := scorer.Score(content, frontmatter, "Task(test)")

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

func TestCommandScorer_CompositionScoring(t *testing.T) {
	scorer := NewCommandScorer()

	// Note: lines counted as strings.Count(content, "\n") + 1
	// Body has "Task(test)\n" (1 newline) + N repeats of "Line\n" (N newlines)
	// Total lines = 1 + N + 1 = N + 2
	// To get exactly T lines, need T - 2 repeats
	tests := []struct {
		name       string
		repeats    int // number of "Line\n" to add
		wantPoints int
		wantPassed bool
	}{
		{"Excellent - 25 lines", 20, 10, true},   // 22 lines counted, ≤30
		{"Good - under 45", 40, 8, true},         // 42 lines counted, ≤45
		{"Good boundary - 45", 43, 8, true},      // 45 lines counted, =45
		{"OK - 50 lines", 48, 6, true},           // 50 lines counted, ≤55
		{"OK boundary - 54", 52, 6, true},        // 54 lines counted, ≤55
		{"Over limit - 60", 58, 3, false},        // 60 lines counted
		{"Fat command - 100", 98, 0, false},      // 100 lines counted
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			frontmatter := map[string]interface{}{
				"allowed-tools":  []string{"Task"},
				"description":    "Test command",
				"argument-hint":  "input",
			}
			bodyContent := "Task(test)\n" + strings.Repeat("Line\n", tt.repeats)
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
				t.Errorf("Passed = %v, want %v", compMetric.Passed, tt.wantPassed)
			}
		})
	}
}

func TestPluralize(t *testing.T) {
	tests := []struct {
		name     string
		count    int
		singular string
		want     string
	}{
		{"Single call", 1, "Task() call", "1 Task() call"},
		{"Multiple calls", 3, "Task() call", "Task() calls"},
		{"Zero calls", 0, "Task() call", "Task() calls"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := pluralize(tt.count, tt.singular)
			if got != tt.want {
				t.Errorf("pluralize(%d, %q) = %q, want %q", tt.count, tt.singular, got, tt.want)
			}
		})
	}
}

func TestCommandScorer_SuccessCriteriaFormats(t *testing.T) {
	scorer := NewCommandScorer()

	tests := []struct {
		name        string
		bodyContent string
		wantPoints  bool
	}{
		{
			name: "Success criteria header",
			bodyContent: `
## Success Criteria
All tests pass
Task(test)
`,
			wantPoints: true,
		},
		{
			name: "Case insensitive",
			bodyContent: `
## SUCCESS CRITERIA
Task(test)
`,
			wantPoints: true,
		},
		{
			name: "Checkbox format",
			bodyContent: `
- [ ] First check
- [ ] Second check
Task(test)
`,
			wantPoints: true,
		},
		{
			name: "No success criteria",
			bodyContent: `
Just some content
Task(test)
`,
			wantPoints: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			frontmatter := map[string]interface{}{
				"allowed-tools":  []string{"Task"},
				"description":    "Test",
				"argument-hint":  "input",
			}
			content := generateFullContent(frontmatter, tt.bodyContent)
			score := scorer.Score(content, frontmatter, tt.bodyContent)

			var successMetric *ScoringMetric
			for i := range score.Details {
				if score.Details[i].Name == "Success criteria" {
					successMetric = &score.Details[i]
					break
				}
			}

			if successMetric == nil {
				t.Fatal("Success criteria metric not found")
			}

			if successMetric.Passed != tt.wantPoints {
				t.Errorf("Success criteria passed = %v, want %v", successMetric.Passed, tt.wantPoints)
			}
		})
	}
}

func TestCommandScorer_CodeExamples(t *testing.T) {
	scorer := NewCommandScorer()

	tests := []struct {
		name        string
		bodyContent string
		wantPoints  bool
	}{
		{
			name:        "Bash code block",
			bodyContent: "```bash\ncommand\n```\nTask(test)",
			wantPoints:  true,
		},
		{
			name:        "Generic code block",
			bodyContent: "```\ncode\n```\nTask(test)",
			wantPoints:  true,
		},
		{
			name:        "Multiple code blocks",
			bodyContent: "```bash\nfirst\n```\n```\nsecond\n```\nTask(test)",
			wantPoints:  true,
		},
		{
			name:        "No code blocks",
			bodyContent: "Just text\nTask(test)",
			wantPoints:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			frontmatter := map[string]interface{}{
				"allowed-tools":  []string{"Task"},
				"description":    "Test command with good description",
				"argument-hint":  "input",
			}
			content := generateFullContent(frontmatter, tt.bodyContent)
			score := scorer.Score(content, frontmatter, tt.bodyContent)

			var codeMetric *ScoringMetric
			for i := range score.Details {
				if score.Details[i].Name == "Code examples" {
					codeMetric = &score.Details[i]
					break
				}
			}

			if codeMetric == nil {
				t.Fatal("Code examples metric not found")
			}

			if codeMetric.Passed != tt.wantPoints {
				t.Errorf("Code examples passed = %v, want %v", codeMetric.Passed, tt.wantPoints)
			}
		})
	}
}
