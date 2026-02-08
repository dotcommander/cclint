package scoring

import (
	"reflect"
	"testing"
)

func TestScoreRequiredFields(t *testing.T) {
	tests := []struct {
		name         string
		frontmatter  map[string]any
		specs        []FieldSpec
		wantScore    int
		wantMetrics  int
		checkMetrics func([]ScoringMetric) error
	}{
		{
			name: "All fields present",
			frontmatter: map[string]any{
				"name":        "test-agent",
				"description": "Test description",
				"model":       "claude-3-5-sonnet-20241022",
			},
			specs: []FieldSpec{
				{"name", 5},
				{"description", 5},
				{"model", 5},
			},
			wantScore:   15,
			wantMetrics: 3,
		},
		{
			name: "Some fields missing",
			frontmatter: map[string]any{
				"name": "test-agent",
			},
			specs: []FieldSpec{
				{"name", 5},
				{"description", 5},
				{"model", 5},
			},
			wantScore:   5,
			wantMetrics: 3,
		},
		{
			name:        "All fields missing",
			frontmatter: map[string]any{},
			specs: []FieldSpec{
				{"name", 10},
				{"description", 10},
			},
			wantScore:   0,
			wantMetrics: 2,
		},
		{
			name: "Empty specs",
			frontmatter: map[string]any{
				"name": "test",
			},
			specs:       []FieldSpec{},
			wantScore:   0,
			wantMetrics: 0,
		},
		{
			name:        "Nil frontmatter",
			frontmatter: nil,
			specs: []FieldSpec{
				{"name", 5},
			},
			wantScore:   0,
			wantMetrics: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score, metrics := ScoreRequiredFields(tt.frontmatter, tt.specs)

			if score != tt.wantScore {
				t.Errorf("ScoreRequiredFields() score = %d, want %d", score, tt.wantScore)
			}
			if len(metrics) != tt.wantMetrics {
				t.Errorf("ScoreRequiredFields() metrics count = %d, want %d", len(metrics), tt.wantMetrics)
			}

			// Verify all metrics have structural category
			for _, m := range metrics {
				if m.Category != "structural" {
					t.Errorf("Metric category = %q, want %q", m.Category, "structural")
				}
			}
		})
	}
}

func TestScoreSections(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		specs       []SectionSpec
		wantScore   int
		wantMetrics int
	}{
		{
			name: "All sections present",
			content: `
## Foundation
Content here

### Phase 1
Work

## Success Criteria
- [ ] Done
`,
			specs: []SectionSpec{
				{`(?i)## Foundation`, "Foundation section", 5},
				{`(?i)### Phase`, "Phase workflow", 4},
				{`(?i)## Success Criteria`, "Success Criteria", 3},
			},
			wantScore:   12,
			wantMetrics: 3,
		},
		{
			name: "Some sections missing",
			content: `
## Foundation
Content here
`,
			specs: []SectionSpec{
				{`(?i)## Foundation`, "Foundation section", 5},
				{`(?i)### Phase`, "Phase workflow", 4},
				{`(?i)## Success Criteria`, "Success Criteria", 3},
			},
			wantScore:   5,
			wantMetrics: 3,
		},
		{
			name:    "All sections missing",
			content: "No sections here",
			specs: []SectionSpec{
				{`(?i)## Foundation`, "Foundation section", 5},
			},
			wantScore:   0,
			wantMetrics: 1,
		},
		{
			name:        "Empty content",
			content:     "",
			specs:       []SectionSpec{{`(?i)## Test`, "Test", 5}},
			wantScore:   0,
			wantMetrics: 1,
		},
		{
			name:        "Empty specs",
			content:     "## Test",
			specs:       []SectionSpec{},
			wantScore:   0,
			wantMetrics: 0,
		},
		{
			name: "Case insensitive matching",
			content: `
## FOUNDATION
## foundation
## FoUnDaTiOn
`,
			specs: []SectionSpec{
				{`(?i)## Foundation`, "Foundation section", 5},
			},
			wantScore:   5,
			wantMetrics: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score, metrics := ScoreSections(tt.content, tt.specs)

			if score != tt.wantScore {
				t.Errorf("ScoreSections() score = %d, want %d", score, tt.wantScore)
			}
			if len(metrics) != tt.wantMetrics {
				t.Errorf("ScoreSections() metrics count = %d, want %d", len(metrics), tt.wantMetrics)
			}
		})
	}
}

func TestScoreSectionsWithFallback(t *testing.T) {
	fallbackFunc := func(content string, sectionName string) bool {
		if sectionName == "Anti-Patterns section" {
			return content == "FALLBACK_MATCH"
		}
		return false
	}

	tests := []struct {
		name        string
		content     string
		specs       []SectionSpec
		fallback    SectionFallbackFunc
		wantScore   int
		wantMetrics int
	}{
		{
			name:    "Primary pattern matches",
			content: "## Anti-Patterns",
			specs: []SectionSpec{
				{`(?i)## Anti-Patterns`, "Anti-Patterns section", 5},
			},
			fallback:    fallbackFunc,
			wantScore:   5,
			wantMetrics: 1,
		},
		{
			name:    "Fallback matches",
			content: "FALLBACK_MATCH",
			specs: []SectionSpec{
				{`(?i)## Anti-Patterns`, "Anti-Patterns section", 5},
			},
			fallback:    fallbackFunc,
			wantScore:   5,
			wantMetrics: 1,
		},
		{
			name:    "Neither matches",
			content: "No match",
			specs: []SectionSpec{
				{`(?i)## Anti-Patterns`, "Anti-Patterns section", 5},
			},
			fallback:    fallbackFunc,
			wantScore:   0,
			wantMetrics: 1,
		},
		{
			name:    "Nil fallback",
			content: "No match",
			specs: []SectionSpec{
				{`(?i)## Test`, "Test section", 5},
			},
			fallback:    nil,
			wantScore:   0,
			wantMetrics: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score, metrics := ScoreSectionsWithFallback(tt.content, tt.specs, tt.fallback)

			if score != tt.wantScore {
				t.Errorf("ScoreSectionsWithFallback() score = %d, want %d", score, tt.wantScore)
			}
			if len(metrics) != tt.wantMetrics {
				t.Errorf("ScoreSectionsWithFallback() metrics count = %d, want %d", len(metrics), tt.wantMetrics)
			}
		})
	}
}

func TestScoreComposition(t *testing.T) {
	thresholds := CompositionThresholds{
		Excellent:     100,
		ExcellentNote: "Excellent: ≤100 lines",
		Good:          150,
		GoodNote:      "Good: ≤150 lines",
		OK:            200,
		OKNote:        "OK: ≤200 lines",
		OverLimit:     250,
		OverLimitNote: "Over limit: >200 lines",
		FatNote:       "Fat: >250 lines",
	}

	tests := []struct {
		name       string
		lines      int
		wantPoints int
		wantNote   string
		wantPassed bool
	}{
		{"Excellent - boundary", 100, 10, "Excellent: ≤100 lines", true},
		{"Excellent - under", 50, 10, "Excellent: ≤100 lines", true},
		{"Good - boundary", 150, 8, "Good: ≤150 lines", true},
		{"Good - mid range", 125, 8, "Good: ≤150 lines", true},
		{"OK - boundary", 200, 6, "OK: ≤200 lines", true},
		{"OK - mid range", 175, 6, "OK: ≤200 lines", true},
		{"Over limit - boundary", 250, 3, "Over limit: >200 lines", false},
		{"Over limit - mid range", 225, 3, "Over limit: >200 lines", false},
		{"Fat - above limit", 300, 0, "Fat: >250 lines", false},
		{"Fat - way over", 500, 0, "Fat: >250 lines", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			points, metric := ScoreComposition(tt.lines, thresholds)

			if points != tt.wantPoints {
				t.Errorf("ScoreComposition() points = %d, want %d", points, tt.wantPoints)
			}
			if metric.Points != tt.wantPoints {
				t.Errorf("metric.Points = %d, want %d", metric.Points, tt.wantPoints)
			}
			if metric.MaxPoints != 10 {
				t.Errorf("metric.MaxPoints = %d, want 10", metric.MaxPoints)
			}
			if metric.Note != tt.wantNote {
				t.Errorf("metric.Note = %q, want %q", metric.Note, tt.wantNote)
			}
			if metric.Passed != tt.wantPassed {
				t.Errorf("metric.Passed = %v, want %v", metric.Passed, tt.wantPassed)
			}
			if metric.Category != "composition" {
				t.Errorf("metric.Category = %q, want %q", metric.Category, "composition")
			}
			if metric.Name != "Line count" {
				t.Errorf("metric.Name = %q, want %q", metric.Name, "Line count")
			}
		})
	}
}

func TestBoolToInt(t *testing.T) {
	tests := []struct {
		name string
		b    bool
		want int
	}{
		{"True converts to 1", true, 1},
		{"False converts to 0", false, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := boolToInt(tt.b)
			if got != tt.want {
				t.Errorf("boolToInt(%v) = %d, want %d", tt.b, got, tt.want)
			}
		})
	}
}

func TestIsMethodologySkill(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    bool
	}{
		{
			name: "Has Workflow section",
			content: `
## Workflow

Steps here
`,
			want: true,
		},
		{
			name: "Has Phase pattern",
			content: `
### Phase 1
### Phase 2
`,
			want: true,
		},
		{
			name: "Has Algorithm section",
			content: `
## Algorithm

Logic here
`,
			want: true,
		},
		{
			name: "Has Process section",
			content: `
## Process

Steps
`,
			want: true,
		},
		{
			name: "Has Step pattern",
			content: `
### Step 1
### Step 2
`,
			want: true,
		},
		{
			name: "Reference skill - no methodology markers",
			content: `
## Quick Reference

| Pattern | Use Case |
|---------|----------|
| Foo     | Bar      |
`,
			want: false,
		},
		{
			name:    "Empty content",
			content: "",
			want:    false,
		},
		{
			name: "Case insensitive - WORKFLOW",
			content: `
## WORKFLOW
`,
			want: true,
		},
		{
			name: "Case insensitive - workflow",
			content: `
## workflow
`,
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsMethodologySkill(tt.content)
			if got != tt.want {
				t.Errorf("IsMethodologySkill() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFieldSpec(t *testing.T) {
	spec := FieldSpec{
		Name:   "test-field",
		Points: 10,
	}

	if spec.Name != "test-field" {
		t.Errorf("Name = %q, want %q", spec.Name, "test-field")
	}
	if spec.Points != 10 {
		t.Errorf("Points = %d, want %d", spec.Points, 10)
	}
}

func TestSectionSpec(t *testing.T) {
	spec := SectionSpec{
		Pattern: `(?i)## Test`,
		Name:    "Test section",
		Points:  5,
	}

	if spec.Pattern != `(?i)## Test` {
		t.Errorf("Pattern = %q, want %q", spec.Pattern, `(?i)## Test`)
	}
	if spec.Name != "Test section" {
		t.Errorf("Name = %q, want %q", spec.Name, "Test section")
	}
	if spec.Points != 5 {
		t.Errorf("Points = %d, want %d", spec.Points, 5)
	}
}

func TestCompositionThresholds(t *testing.T) {
	thresholds := CompositionThresholds{
		Excellent:     100,
		ExcellentNote: "Excellent",
		Good:          200,
		GoodNote:      "Good",
		OK:            300,
		OKNote:        "OK",
		OverLimit:     400,
		OverLimitNote: "Over",
		FatNote:       "Fat",
	}

	if thresholds.Excellent != 100 {
		t.Errorf("Excellent = %d, want %d", thresholds.Excellent, 100)
	}
	if thresholds.Good != 200 {
		t.Errorf("Good = %d, want %d", thresholds.Good, 200)
	}
	if thresholds.OK != 300 {
		t.Errorf("OK = %d, want %d", thresholds.OK, 300)
	}
	if thresholds.OverLimit != 400 {
		t.Errorf("OverLimit = %d, want %d", thresholds.OverLimit, 400)
	}
}

func TestSectionFallbackFunc(t *testing.T) {
	// Test that the type exists and can be used
	var fn SectionFallbackFunc = func(content string, sectionName string) bool {
		return content == "test" && sectionName == "test"
	}

	if fn == nil {
		t.Error("SectionFallbackFunc should not be nil")
	}

	// Test invocation
	if !fn("test", "test") {
		t.Error("SectionFallbackFunc(\"test\", \"test\") should return true")
	}
	if fn("other", "test") {
		t.Error("SectionFallbackFunc(\"other\", \"test\") should return false")
	}
}

func TestScoreRequiredFieldsNilSafety(t *testing.T) {
	// Test with nil frontmatter
	score, metrics := ScoreRequiredFields(nil, []FieldSpec{
		{"name", 5},
		{"description", 5},
	})

	if score != 0 {
		t.Errorf("Score with nil frontmatter = %d, want 0", score)
	}
	if len(metrics) != 2 {
		t.Errorf("Metrics count = %d, want 2", len(metrics))
	}
	for _, m := range metrics {
		if m.Passed {
			t.Errorf("Metric %q should not pass with nil frontmatter", m.Name)
		}
	}
}

func TestScoreSectionsEdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		content string
		specs   []SectionSpec
		want    int
	}{
		{
			name:    "Multiple matches of same pattern",
			content: "## Test\n## Test\n## Test",
			specs: []SectionSpec{
				{`## Test`, "Test", 5},
			},
			want: 5, // Should only count once
		},
		{
			name:    "Regex special characters",
			content: "## Test (with parens)",
			specs: []SectionSpec{
				{`## Test \(with parens\)`, "Test", 5},
			},
			want: 5,
		},
		{
			name:    "Multiline content",
			content: "Line 1\n## Section\nLine 3",
			specs: []SectionSpec{
				{`## Section`, "Section", 5},
			},
			want: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score, _ := ScoreSections(tt.content, tt.specs)
			if score != tt.want {
				t.Errorf("ScoreSections() = %d, want %d", score, tt.want)
			}
		})
	}
}

func TestScoreCompositionMetricStructure(t *testing.T) {
	thresholds := CompositionThresholds{
		Excellent:     100,
		ExcellentNote: "Excellent",
		Good:          200,
		GoodNote:      "Good",
		OK:            300,
		OKNote:        "OK",
		OverLimit:     400,
		OverLimitNote: "Over",
		FatNote:       "Fat",
	}

	_, metric := ScoreComposition(50, thresholds)

	// Verify metric structure
	wantMetric := ScoringMetric{
		Category:  "composition",
		Name:      "Line count",
		Points:    10,
		MaxPoints: 10,
		Passed:    true,
		Note:      "Excellent",
	}

	if !reflect.DeepEqual(metric, wantMetric) {
		t.Errorf("ScoreComposition() metric = %+v, want %+v", metric, wantMetric)
	}
}
