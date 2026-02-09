package scoring

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestNewPluginScorer(t *testing.T) {
	scorer := NewPluginScorer()
	if scorer == nil {
		t.Fatal("NewPluginScorer() returned nil")
	}
}

func TestPluginScorer_Score(t *testing.T) {
	tests := []struct {
		name          string
		jsonData      map[string]any
		wantTier      string
		wantStructMin int
		wantPractMin  int
		wantCompMin   int
		wantDocMin    int
	}{
		{
			name: "Perfect plugin",
			jsonData: map[string]any{
				"name":        "test-plugin",
				"description": strings.Repeat("Comprehensive plugin description with detailed information about functionality. ", 2),
				"version":     "1.0.0",
				"author": map[string]any{
					"name": "Test Author",
				},
				"homepage":   "https://example.com",
				"repository": "https://github.com/user/repo",
				"license":    "MIT",
				"keywords":   []any{"test", "plugin"},
				"readme":     "README.md",
			},
			wantTier:      "A",
			wantStructMin: 40,
			wantPractMin:  40,
			wantCompMin:   8,
			wantDocMin:    10,
		},
		{
			name: "Minimal plugin - only required fields",
			jsonData: map[string]any{
				"name":        "minimal",
				"description": "Short description",
				"version":     "1.0.0",
				"author": map[string]any{
					"name": "Author",
				},
			},
			wantTier:      "C", // 40 (structural) + 0 (practices) + ~6-8 (comp) + 0 (doc) = ~50
			wantStructMin: 40,
			wantPractMin:  0,
		},
		{
			name: "Missing required fields",
			jsonData: map[string]any{
				"name": "incomplete",
			},
			wantTier: "F",
		},
		{
			name: "Plugin with homepage only",
			jsonData: map[string]any{
				"name":        "test",
				"description": "Test plugin",
				"version":     "1.0.0",
				"author": map[string]any{
					"name": "Test Author",
				},
				"homepage": "https://example.com",
			},
			wantStructMin: 40,
			wantPractMin:  10,
		},
		{
			name: "Plugin with repository only",
			jsonData: map[string]any{
				"name":        "test",
				"description": "Test plugin",
				"version":     "1.0.0",
				"author": map[string]any{
					"name": "Test Author",
				},
				"repository": "https://github.com/user/repo",
			},
			wantStructMin: 40,
			wantPractMin:  10,
		},
		{
			name: "Plugin with license only",
			jsonData: map[string]any{
				"name":        "test",
				"description": "Test plugin",
				"version":     "1.0.0",
				"author": map[string]any{
					"name": "Test Author",
				},
				"license": "Apache-2.0",
			},
			wantStructMin: 40,
			wantPractMin:  10,
		},
		{
			name: "Plugin with keywords only",
			jsonData: map[string]any{
				"name":        "test",
				"description": "Test plugin",
				"version":     "1.0.0",
				"author": map[string]any{
					"name": "Test Author",
				},
				"keywords": []any{"testing", "automation"},
			},
			wantStructMin: 40,
			wantPractMin:  10,
		},
		{
			name: "Large plugin - over 10KB",
			jsonData: map[string]any{
				"name":        "large",
				"description": strings.Repeat("x", 15000),
				"version":     "1.0.0",
				"author": map[string]any{
					"name": "Author",
				},
			},
			wantCompMin: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scorer := NewPluginScorer()
			jsonBytes, err := json.Marshal(tt.jsonData)
			if err != nil {
				t.Fatalf("Failed to marshal JSON: %v", err)
			}
			content := string(jsonBytes)
			score := scorer.Score(content, tt.jsonData, "")

			if tt.wantTier != "" && score.Tier != tt.wantTier {
				t.Errorf("Score() tier = %q, want %q (overall: %d)", score.Tier, tt.wantTier, score.Overall)
			}
			if tt.wantStructMin > 0 && score.Structural < tt.wantStructMin {
				t.Errorf("Score() structural = %d, want >= %d", score.Structural, tt.wantStructMin)
			}
			if tt.wantPractMin > 0 && score.Practices < tt.wantPractMin {
				t.Errorf("Score() practices = %d, want >= %d", score.Practices, tt.wantPractMin)
			}
			if tt.wantCompMin >= 0 && score.Composition < tt.wantCompMin {
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

func TestPluginScorer_StructuralFields(t *testing.T) {
	scorer := NewPluginScorer()

	tests := []struct {
		name       string
		jsonData   map[string]any
		wantPoints int
	}{
		{
			name: "All structural fields present",
			jsonData: map[string]any{
				"name":        "test",
				"description": "Test",
				"version":     "1.0.0",
				"author": map[string]any{
					"name": "Author",
				},
			},
			wantPoints: 40,
		},
		{
			name: "Missing author.name",
			jsonData: map[string]any{
				"name":        "test",
				"description": "Test",
				"version":     "1.0.0",
			},
			wantPoints: 30,
		},
		{
			name: "Author without name",
			jsonData: map[string]any{
				"name":        "test",
				"description": "Test",
				"version":     "1.0.0",
				"author": map[string]any{
					"email": "author@example.com",
				},
			},
			wantPoints: 30,
		},
		{
			name: "Missing version",
			jsonData: map[string]any{
				"name":        "test",
				"description": "Test",
				"author": map[string]any{
					"name": "Author",
				},
			},
			wantPoints: 30,
		},
		{
			name: "Missing description",
			jsonData: map[string]any{
				"name":    "test",
				"version": "1.0.0",
				"author": map[string]any{
					"name": "Author",
				},
			},
			wantPoints: 30,
		},
		{
			name: "Missing name",
			jsonData: map[string]any{
				"description": "Test",
				"version":     "1.0.0",
				"author": map[string]any{
					"name": "Author",
				},
			},
			wantPoints: 30,
		},
		{
			name:       "All fields missing",
			jsonData:   map[string]any{},
			wantPoints: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonBytes, _ := json.Marshal(tt.jsonData)
			content := string(jsonBytes)
			score := scorer.Score(content, tt.jsonData, "")

			if score.Structural != tt.wantPoints {
				t.Errorf("Structural score = %d, want %d", score.Structural, tt.wantPoints)
			}
		})
	}
}

func TestPluginScorer_PracticesFields(t *testing.T) {
	scorer := NewPluginScorer()

	baseData := map[string]any{
		"name":        "test",
		"description": "Test",
		"version":     "1.0.0",
		"author": map[string]any{
			"name": "Author",
		},
	}

	tests := []struct {
		name       string
		extraData  map[string]any
		wantPoints int
	}{
		{
			name: "All practices fields present",
			extraData: map[string]any{
				"homepage":   "https://example.com",
				"repository": "https://github.com/user/repo",
				"license":    "MIT",
				"keywords":   []any{"test"},
			},
			wantPoints: 40,
		},
		{
			name: "Only homepage",
			extraData: map[string]any{
				"homepage": "https://example.com",
			},
			wantPoints: 10,
		},
		{
			name: "Only repository",
			extraData: map[string]any{
				"repository": "https://github.com/user/repo",
			},
			wantPoints: 10,
		},
		{
			name: "Only license",
			extraData: map[string]any{
				"license": "Apache-2.0",
			},
			wantPoints: 10,
		},
		{
			name: "Only keywords",
			extraData: map[string]any{
				"keywords": []any{"test", "plugin"},
			},
			wantPoints: 10,
		},
		{
			name: "Empty keywords array (doesn't count)",
			extraData: map[string]any{
				"keywords": []any{},
			},
			wantPoints: 0,
		},
		{
			name:       "No practices fields",
			extraData:  map[string]any{},
			wantPoints: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonData := make(map[string]any)
			for k, v := range baseData {
				jsonData[k] = v
			}
			for k, v := range tt.extraData {
				jsonData[k] = v
			}

			jsonBytes, _ := json.Marshal(jsonData)
			content := string(jsonBytes)
			score := scorer.Score(content, jsonData, "")

			if score.Practices != tt.wantPoints {
				t.Errorf("Practices score = %d, want %d", score.Practices, tt.wantPoints)
			}
		})
	}
}

func TestPluginScorer_CompositionScoring(t *testing.T) {
	scorer := NewPluginScorer()

	tests := []struct {
		name       string
		size       int
		wantPoints int
		wantNote   string
		wantPassed bool
	}{
		{"Excellent - 500 bytes", 500, 10, "Excellent: ≤1KB", true},
		{"Good - 1500 bytes", 1500, 8, "Good: ≤2KB", true},
		{"OK - 3000 bytes", 3000, 6, "OK: ≤5KB", true},
		{"Large - 7000 bytes", 7000, 3, "Large: ≤10KB", false},
		{"Too large - 15000 bytes", 15000, 0, "Too large: >10KB", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create JSON with specific size
			desc := strings.Repeat("x", tt.size-200) // Account for other fields
			jsonData := map[string]any{
				"name":        "test",
				"description": desc,
				"version":     "1.0.0",
				"author": map[string]any{
					"name": "Author",
				},
			}

			jsonBytes, _ := json.Marshal(jsonData)
			content := string(jsonBytes)
			score := scorer.Score(content, jsonData, "")

			if score.Composition != tt.wantPoints {
				t.Errorf("Composition score = %d, want %d (size: %d)", score.Composition, tt.wantPoints, len(content))
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

			if compMetric.Note != tt.wantNote {
				t.Errorf("Note = %q, want %q", compMetric.Note, tt.wantNote)
			}
			if compMetric.Passed != tt.wantPassed {
				t.Errorf("Passed = %v, want %v", compMetric.Passed, tt.wantPassed)
			}
		})
	}
}

func TestPluginScorer_DescriptionQuality(t *testing.T) {
	scorer := NewPluginScorer()

	tests := []struct {
		name       string
		desc       string
		wantPoints int
		wantNote   string
		wantPassed bool
	}{
		{
			name:       "Comprehensive (>=100 chars)",
			desc:       strings.Repeat("a", 100),
			wantPoints: 5,
			wantNote:   "Comprehensive",
			wantPassed: true,
		},
		{
			name:       "Adequate (>=50 chars)",
			desc:       strings.Repeat("a", 50),
			wantPoints: 3,
			wantNote:   "Adequate",
			wantPassed: true,
		},
		{
			name:       "Brief (>=20 chars)",
			desc:       strings.Repeat("a", 20),
			wantPoints: 1,
			wantNote:   "Brief",
			wantPassed: false,
		},
		{
			name:       "Too short (<20 chars)",
			desc:       "Short",
			wantPoints: 0,
			wantNote:   "Too short",
			wantPassed: false,
		},
		{
			name:       "Empty",
			desc:       "",
			wantPoints: 0,
			wantNote:   "Too short",
			wantPassed: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonData := map[string]any{
				"name":        "test",
				"description": tt.desc,
				"version":     "1.0.0",
				"author": map[string]any{
					"name": "Author",
				},
			}

			jsonBytes, _ := json.Marshal(jsonData)
			content := string(jsonBytes)
			score := scorer.Score(content, jsonData, "")

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

func TestPluginScorer_ReadmeField(t *testing.T) {
	scorer := NewPluginScorer()

	tests := []struct {
		name       string
		readme     any
		wantPoints bool
	}{
		{"Has readme", "README.md", true},
		{"Has readme with path", "docs/README.md", true},
		{"Empty readme", "", false},
		{"Missing readme", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonData := map[string]any{
				"name":        "test",
				"description": strings.Repeat("Good description here. ", 3),
				"version":     "1.0.0",
				"author": map[string]any{
					"name": "Author",
				},
			}
			if tt.readme != nil {
				jsonData["readme"] = tt.readme
			}

			jsonBytes, _ := json.Marshal(jsonData)
			content := string(jsonBytes)
			score := scorer.Score(content, jsonData, "")

			var readmeMetric *Metric
			for i := range score.Details {
				if score.Details[i].Name == "Has readme" {
					readmeMetric = &score.Details[i]
					break
				}
			}

			if readmeMetric == nil {
				t.Fatal("Has readme metric not found")
			}

			if readmeMetric.Passed != tt.wantPoints {
				t.Errorf("Readme passed = %v, want %v", readmeMetric.Passed, tt.wantPoints)
			}
		})
	}
}

func TestValidateJSON(t *testing.T) {
	tests := []struct {
		name    string
		content string
		wantErr bool
	}{
		{
			name:    "Valid JSON",
			content: `{"name": "test", "version": "1.0.0"}`,
			wantErr: false,
		},
		{
			name:    "Invalid JSON - missing quote",
			content: `{"name: "test"}`,
			wantErr: true,
		},
		{
			name:    "Invalid JSON - trailing comma",
			content: `{"name": "test",}`,
			wantErr: true,
		},
		{
			name:    "Empty string",
			content: "",
			wantErr: true,
		},
		{
			name:    "Not JSON",
			content: "This is not JSON",
			wantErr: true,
		},
		{
			name:    "Valid empty object",
			content: `{}`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateJSON(tt.content)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateJSON() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPluginScorer_CategoryBreakdown(t *testing.T) {
	scorer := NewPluginScorer()

	jsonData := map[string]any{
		"name":        "test-plugin",
		"description": strings.Repeat("Comprehensive description. ", 5),
		"version":     "1.0.0",
		"author": map[string]any{
			"name": "Test Author",
		},
		"homepage":   "https://example.com",
		"repository": "https://github.com/user/repo",
		"license":    "MIT",
		"keywords":   []any{"test", "plugin"},
		"readme":     "README.md",
	}

	jsonBytes, _ := json.Marshal(jsonData)
	content := string(jsonBytes)
	score := scorer.Score(content, jsonData, "")

	// Verify each category has expected number of metrics
	categoryCounts := make(map[string]int)
	for _, detail := range score.Details {
		categoryCounts[detail.Category]++
	}

	expectedCounts := map[string]int{
		"structural":    4, // name, description, version, author.name
		"practices":     4, // homepage, repository, license, keywords
		"composition":   1, // file size
		"documentation": 2, // description quality, readme
	}

	for category, expectedCount := range expectedCounts {
		if count := categoryCounts[category]; count != expectedCount {
			t.Errorf("Category %q has %d metrics, want %d", category, count, expectedCount)
		}
	}

	// Verify all metrics sum correctly
	totalFromMetrics := 0
	for _, detail := range score.Details {
		totalFromMetrics += detail.Points
	}

	if totalFromMetrics != score.Overall {
		t.Errorf("Sum of metric points (%d) != overall score (%d)", totalFromMetrics, score.Overall)
	}
}
