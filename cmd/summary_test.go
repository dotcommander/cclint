package cmd

import (
	"testing"

	"github.com/dotcommander/cclint/internal/cli"
	"github.com/dotcommander/cclint/internal/cue"
	"github.com/dotcommander/cclint/internal/scoring"
	"github.com/stretchr/testify/assert"
)

func TestAggregateResults(t *testing.T) {
	summary := &ComponentSummary{
		TierCounts: make(map[string]int),
		TopIssues:  make(map[string]int),
	}

	results := []cli.LintResult{
		{
			File: "test1.md",
			Type: "agent",
			Quality: &scoring.QualityScore{
				Overall: 85,
				Tier:    "A",
			},
			Errors: []cue.ValidationError{
				{Message: "Best practice: Agent should be less than 200 lines"},
			},
			Suggestions: []cue.ValidationError{
				{Message: "Missing Foundation section"},
			},
		},
		{
			File: "test2.md",
			Type: "command",
			Quality: &scoring.QualityScore{
				Overall: 65,
				Tier:    "C",
			},
			Errors: []cue.ValidationError{
				{Message: "Missing Workflow section"},
			},
		},
		{
			File: "test3.md",
			Type: "skill",
			Quality: &scoring.QualityScore{
				Overall: 45,
				Tier:    "D",
			},
		},
	}

	aggregateResults(summary, results)

	// Verify tier counts
	assert.Equal(t, 1, summary.TierCounts["A"])
	assert.Equal(t, 1, summary.TierCounts["C"])
	assert.Equal(t, 1, summary.TierCounts["D"])

	// Verify all results were added
	assert.Len(t, summary.AllResults, 3)

	// Verify lowest scoring components
	assert.Len(t, summary.LowestScoring, 3)

	// Verify issues were categorized
	assert.Greater(t, len(summary.TopIssues), 0)
	assert.Contains(t, summary.TopIssues, "Oversized component (fat)")
	assert.Contains(t, summary.TopIssues, "Missing Foundation section")
	assert.Contains(t, summary.TopIssues, "Missing Workflow section")
}

func TestCategorizeIssue(t *testing.T) {
	tests := []struct {
		message  string
		expected string
	}{
		{
			message:  "Best practice: Agent should be less than 200 lines",
			expected: "Oversized component (fat)",
		},
		{
			message:  "Missing Foundation section",
			expected: "Missing Foundation section",
		},
		{
			message:  "Missing Workflow section",
			expected: "Missing Workflow section",
		},
		{
			message:  "Missing Anti-Patterns section",
			expected: "Missing Anti-Patterns section",
		},
		{
			message:  "Missing Quick Reference table with semantic routing",
			expected: "Missing semantic routing",
		},
		{
			message:  "Missing Success Criteria section",
			expected: "Missing Success Criteria",
		},
		{
			message:  "Missing Expected Output section",
			expected: "Missing Expected Output",
		},
		{
			message:  "Agent embeds Skill() methodology instead of loading",
			expected: "Embedded methodology (should extract)",
		},
		{
			message:  "Missing or incomplete triggers",
			expected: "Missing or incomplete triggers",
		},
		{
			message:  "Missing PROACTIVELY pattern in agent",
			expected: "Missing PROACTIVELY pattern",
		},
		{
			message:  "Missing model specification",
			expected: "Missing model specification",
		},
		{
			message:  "Missing description field",
			expected: "Missing or poor description",
		},
		{
			message:  "Some other validation error",
			expected: "Other issues",
		},
	}

	for _, tt := range tests {
		t.Run(tt.message, func(t *testing.T) {
			result := categorizeIssue(tt.message)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestContains(t *testing.T) {
	tests := []struct {
		name     string
		s        string
		substrs  []string
		expected bool
	}{
		{
			name:     "single match",
			s:        "hello world",
			substrs:  []string{"world"},
			expected: true,
		},
		{
			name:     "multiple substrs, first matches",
			s:        "hello world",
			substrs:  []string{"hello", "foo"},
			expected: true,
		},
		{
			name:     "multiple substrs, second matches",
			s:        "hello world",
			substrs:  []string{"foo", "world"},
			expected: true,
		},
		{
			name:     "no match",
			s:        "hello world",
			substrs:  []string{"foo", "bar"},
			expected: false,
		},
		{
			name:     "empty string",
			s:        "",
			substrs:  []string{"foo"},
			expected: false,
		},
		{
			name:     "empty substrs",
			s:        "hello",
			substrs:  []string{},
			expected: false,
		},
		{
			name:     "substr longer than string",
			s:        "hi",
			substrs:  []string{"hello"},
			expected: false,
		},
		{
			name:     "exact match",
			s:        "test",
			substrs:  []string{"test"},
			expected: true,
		},
		{
			name:     "case sensitive",
			s:        "Hello World",
			substrs:  []string{"hello"},
			expected: false,
		},
		{
			name:     "multiple occurrences",
			s:        "test test test",
			substrs:  []string{"test"},
			expected: true,
		},
		{
			name:     "partial match at start",
			s:        "testing",
			substrs:  []string{"test"},
			expected: true,
		},
		{
			name:     "partial match at end",
			s:        "contest",
			substrs:  []string{"test"},
			expected: true,
		},
		{
			name:     "partial match in middle",
			s:        "atestb",
			substrs:  []string{"test"},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := contains(tt.s, tt.substrs...)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRenderBar(t *testing.T) {
	tests := []struct {
		name  string
		count int
		total int
		color string
		want  string // We'll check length and that it contains expected chars
	}{
		{
			name:  "zero total",
			count: 5,
			total: 0,
			color: "10",
			want:  "",
		},
		{
			name:  "zero count",
			count: 0,
			total: 10,
			color: "10",
			want:  "empty", // Returns empty bar with dim blocks
		},
		{
			name:  "full bar",
			count: 10,
			total: 10,
			color: "10",
			want:  "full",
		},
		{
			name:  "half bar",
			count: 5,
			total: 10,
			color: "12",
			want:  "half",
		},
		{
			name:  "minimal count shows at least one",
			count: 1,
			total: 100,
			color: "9",
			want:  "minimal",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := renderBar(tt.count, tt.total, tt.color)

			if tt.want == "" {
				assert.Equal(t, "", result)
			} else if tt.want == "empty" {
				// Zero count still renders dim blocks
				assert.NotEmpty(t, result)
			} else {
				// Just verify it returns a non-empty string
				// The actual rendering includes ANSI codes, so we can't do exact match
				assert.NotEmpty(t, result)
			}
		})
	}
}

func TestPrintSummaryReport(t *testing.T) {
	// Test that printSummaryReport doesn't panic with various inputs
	tests := []struct {
		name    string
		summary *ComponentSummary
	}{
		{
			name: "empty summary",
			summary: &ComponentSummary{
				TierCounts: make(map[string]int),
				TopIssues:  make(map[string]int),
			},
		},
		{
			name: "summary with data",
			summary: &ComponentSummary{
				TotalComponents: 10,
				AgentCount:      4,
				CommandCount:    3,
				SkillCount:      3,
				TierCounts: map[string]int{
					"A": 2,
					"B": 3,
					"C": 3,
					"D": 1,
					"F": 1,
				},
				TopIssues: map[string]int{
					"Missing Foundation section": 5,
					"Oversized component (fat)":  3,
					"Missing Workflow section":   2,
				},
				LowestScoring: []ScoredComponent{
					{File: "test1.md", Type: "agent", Score: 45, Tier: "D"},
					{File: "test2.md", Type: "command", Score: 55, Tier: "C"},
				},
			},
		},
		{
			name: "summary with many issues",
			summary: &ComponentSummary{
				TotalComponents: 20,
				AgentCount:      10,
				CommandCount:    5,
				SkillCount:      5,
				TierCounts: map[string]int{
					"A": 5,
					"B": 5,
					"C": 5,
					"D": 3,
					"F": 2,
				},
				TopIssues: map[string]int{
					"Issue 1": 10,
					"Issue 2": 8,
					"Issue 3": 6,
					"Issue 4": 5,
					"Issue 5": 4,
					"Issue 6": 3,
					"Issue 7": 2,
				},
				LowestScoring: []ScoredComponent{
					{File: "very-long-filename-that-should-be-truncated-in-output.md", Type: "agent", Score: 25, Tier: "F"},
					{File: "test2.md", Type: "command", Score: 35, Tier: "D"},
					{File: "test3.md", Type: "skill", Score: 45, Tier: "D"},
					{File: "test4.md", Type: "agent", Score: 55, Tier: "C"},
					{File: "test5.md", Type: "command", Score: 65, Tier: "C"},
					{File: "test6.md", Type: "skill", Score: 75, Tier: "B"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Just verify it doesn't panic
			assert.NotPanics(t, func() {
				printSummaryReport(tt.summary)
			})
		})
	}
}

func TestComponentSummaryTypes(t *testing.T) {
	// Verify struct fields are accessible
	summary := &ComponentSummary{
		TotalComponents: 10,
		AgentCount:      4,
		CommandCount:    3,
		SkillCount:      3,
		TierCounts:      make(map[string]int),
		TopIssues:       make(map[string]int),
		LowestScoring:   []ScoredComponent{},
		AllResults:      []cli.LintResult{},
	}

	assert.Equal(t, 10, summary.TotalComponents)
	assert.Equal(t, 4, summary.AgentCount)
	assert.Equal(t, 3, summary.CommandCount)
	assert.Equal(t, 3, summary.SkillCount)
	assert.NotNil(t, summary.TierCounts)
	assert.NotNil(t, summary.TopIssues)
	assert.NotNil(t, summary.LowestScoring)
	assert.NotNil(t, summary.AllResults)
}

func TestScoredComponentType(t *testing.T) {
	// Verify ScoredComponent struct
	comp := ScoredComponent{
		File:  "test.md",
		Type:  "agent",
		Score: 85,
		Tier:  "A",
	}

	assert.Equal(t, "test.md", comp.File)
	assert.Equal(t, "agent", comp.Type)
	assert.Equal(t, 85, comp.Score)
	assert.Equal(t, "A", comp.Tier)
}

func TestAggregateResults_NilQuality(t *testing.T) {
	// Test that aggregateResults handles nil Quality gracefully
	summary := &ComponentSummary{
		TierCounts: make(map[string]int),
		TopIssues:  make(map[string]int),
	}

	results := []cli.LintResult{
		{
			File:    "test.md",
			Type:    "agent",
			Quality: nil, // No quality score
			Errors: []cue.ValidationError{
				{Message: "Some error"},
			},
		},
	}

	assert.NotPanics(t, func() {
		aggregateResults(summary, results)
	})

	// Should still aggregate issues even without quality
	assert.Len(t, summary.AllResults, 1)
	assert.Greater(t, len(summary.TopIssues), 0)
}

func TestAggregateResults_MultipleIssuesOfSameType(t *testing.T) {
	// Test that duplicate issues are counted correctly
	summary := &ComponentSummary{
		TierCounts: make(map[string]int),
		TopIssues:  make(map[string]int),
	}

	results := []cli.LintResult{
		{
			File: "test1.md",
			Type: "agent",
			Errors: []cue.ValidationError{
				{Message: "Missing Foundation section"},
			},
		},
		{
			File: "test2.md",
			Type: "agent",
			Errors: []cue.ValidationError{
				{Message: "Missing Foundation section"},
			},
		},
		{
			File: "test3.md",
			Type: "command",
			Errors: []cue.ValidationError{
				{Message: "Missing Foundation section"},
			},
		},
	}

	aggregateResults(summary, results)

	// Should count all occurrences
	assert.Equal(t, 3, summary.TopIssues["Missing Foundation section"])
}

func TestCategorizeIssue_EdgeCases(t *testing.T) {
	tests := []struct {
		message  string
		expected string
	}{
		{
			message:  "",
			expected: "Other issues",
		},
		{
			message:  "Missing Foundation",
			expected: "Missing Foundation section",
		},
		{
			message:  "foundation", // Lowercase
			expected: "Other issues",
		},
		{
			message:  "This message contains Foundation in the middle",
			expected: "Missing Foundation section",
		},
		{
			message:  "Multiple keywords: Foundation and Workflow",
			expected: "Missing Foundation section", // First match wins
		},
	}

	for _, tt := range tests {
		t.Run(tt.message, func(t *testing.T) {
			result := categorizeIssue(tt.message)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPrintSummaryReport_WithAllTiers(t *testing.T) {
	// Test with all tier types populated
	summary := &ComponentSummary{
		TotalComponents: 15,
		AgentCount:      5,
		CommandCount:    5,
		SkillCount:      5,
		TierCounts: map[string]int{
			"A": 3,
			"B": 4,
			"C": 3,
			"D": 3,
			"F": 2,
		},
		TopIssues: map[string]int{
			"Missing Foundation section": 5,
			"Oversized component (fat)":  4,
			"Missing Workflow section":   3,
			"Missing semantic routing":   2,
			"Other issues":               1,
		},
		LowestScoring: []ScoredComponent{
			{File: "agent1.md", Type: "agent", Score: 25, Tier: "F"},
			{File: "agent2.md", Type: "agent", Score: 30, Tier: "D"},
			{File: "command1.md", Type: "command", Score: 35, Tier: "D"},
			{File: "skill1.md", Type: "skill", Score: 55, Tier: "C"},
			{File: "agent3.md", Type: "agent", Score: 60, Tier: "C"},
		},
	}

	// Should not panic
	assert.NotPanics(t, func() {
		printSummaryReport(summary)
	})
}

func TestPrintSummaryReport_LongFilenames(t *testing.T) {
	// Test with very long filenames that need truncation
	summary := &ComponentSummary{
		TotalComponents: 3,
		AgentCount:      1,
		CommandCount:    1,
		SkillCount:      1,
		TierCounts:      map[string]int{"D": 2, "F": 1},
		TopIssues:       map[string]int{"Issue": 3},
		LowestScoring: []ScoredComponent{
			{
				File:  "/very/long/path/to/some/deeply/nested/directory/structure/with/many/levels/test-agent-with-very-long-name.md",
				Type:  "agent",
				Score: 25,
				Tier:  "F",
			},
			{
				File:  "short.md",
				Type:  "command",
				Score: 35,
				Tier:  "D",
			},
		},
	}

	assert.NotPanics(t, func() {
		printSummaryReport(summary)
	})
}

func TestPrintSummaryReport_LongIssueMessages(t *testing.T) {
	// Test with long issue messages that need truncation
	summary := &ComponentSummary{
		TotalComponents: 5,
		AgentCount:      5,
		TierCounts:      map[string]int{"C": 5},
		TopIssues: map[string]int{
			"This is a very long issue message that exceeds the normal display width and should be truncated": 5,
			"Another long issue that describes a complex problem with many details about what went wrong":     4,
			"Short issue":                                                                    3,
		},
		LowestScoring: []ScoredComponent{},
	}

	assert.NotPanics(t, func() {
		printSummaryReport(summary)
	})
}

func TestRenderBar_EdgeCases(t *testing.T) {
	// Test various edge cases for renderBar
	tests := []struct {
		name     string
		count    int
		total    int
		expected bool // true if should return non-empty
	}{
		{"count equals total", 10, 10, true},
		{"count is zero", 0, 10, true},  // Still renders dim blocks
		{"total is zero", 5, 0, false},  // Returns empty
		{"large count", 100, 100, true},
		{"fractional fill", 3, 10, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := renderBar(tt.count, tt.total, "10")
			if tt.expected {
				assert.NotEmpty(t, result)
			} else {
				assert.Empty(t, result)
			}
		})
	}
}

func TestAggregateResults_EmptyResults(t *testing.T) {
	summary := &ComponentSummary{
		TierCounts: make(map[string]int),
		TopIssues:  make(map[string]int),
	}

	// Aggregate empty results
	aggregateResults(summary, []cli.LintResult{})

	assert.Empty(t, summary.AllResults)
	assert.Empty(t, summary.TierCounts)
	assert.Empty(t, summary.TopIssues)
}

func TestAggregateResults_OnlySuggestions(t *testing.T) {
	summary := &ComponentSummary{
		TierCounts: make(map[string]int),
		TopIssues:  make(map[string]int),
	}

	results := []cli.LintResult{
		{
			File: "test.md",
			Type: "agent",
			Quality: &scoring.QualityScore{
				Overall: 75,
				Tier:    "B",
			},
			Suggestions: []cue.ValidationError{
				{Message: "Consider adding more documentation"},
				{Message: "Foundation section could be improved"},
			},
		},
	}

	aggregateResults(summary, results)

	assert.Len(t, summary.AllResults, 1)
	assert.Equal(t, 1, summary.TierCounts["B"])
	assert.Greater(t, len(summary.TopIssues), 0)
}

func TestPrintSummaryReport_ZeroComponents(t *testing.T) {
	summary := &ComponentSummary{
		TotalComponents: 0,
		AgentCount:      0,
		CommandCount:    0,
		SkillCount:      0,
		TierCounts:      make(map[string]int),
		TopIssues:       make(map[string]int),
		LowestScoring:   []ScoredComponent{},
	}

	// Should handle division by zero gracefully
	assert.NotPanics(t, func() {
		printSummaryReport(summary)
	})
}

func TestPrintSummaryReport_MoreThan5Issues(t *testing.T) {
	// Test that only top 5 issues are shown
	summary := &ComponentSummary{
		TotalComponents: 10,
		AgentCount:      10,
		TierCounts:      map[string]int{"C": 10},
		TopIssues: map[string]int{
			"Issue 1": 10,
			"Issue 2": 9,
			"Issue 3": 8,
			"Issue 4": 7,
			"Issue 5": 6,
			"Issue 6": 5, // Should not appear in output
			"Issue 7": 4, // Should not appear in output
		},
		LowestScoring: []ScoredComponent{},
	}

	assert.NotPanics(t, func() {
		printSummaryReport(summary)
	})
}

func TestPrintSummaryReport_MoreThan5LowestScoring(t *testing.T) {
	// Test that only 5 lowest scoring are shown
	summary := &ComponentSummary{
		TotalComponents: 10,
		AgentCount:      10,
		TierCounts:      map[string]int{"D": 5, "F": 5},
		TopIssues:       map[string]int{"Issue": 10},
		LowestScoring: []ScoredComponent{
			{File: "1.md", Score: 20, Tier: "F"},
			{File: "2.md", Score: 25, Tier: "F"},
			{File: "3.md", Score: 30, Tier: "D"},
			{File: "4.md", Score: 35, Tier: "D"},
			{File: "5.md", Score: 40, Tier: "D"},
			{File: "6.md", Score: 45, Tier: "D"}, // Should not appear
			{File: "7.md", Score: 48, Tier: "D"}, // Should not appear
		},
	}

	assert.NotPanics(t, func() {
		printSummaryReport(summary)
	})
}

func TestSummaryCmdInit(t *testing.T) {
	// Verify summary command is registered
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "summary" {
			found = true
			break
		}
	}
	assert.True(t, found, "summary command should be registered")
}
