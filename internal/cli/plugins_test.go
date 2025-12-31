package cli

import (
	"testing"
)

func TestLintPlugins(t *testing.T) {
	// Test with empty directory
	summary, err := LintPlugins("testdata/empty", false, false, true)
	if err != nil {
		t.Fatalf("LintPlugins() error = %v", err)
	}
	if summary == nil {
		t.Fatal("LintPlugins() returned nil summary")
	}
}

func TestValidatePluginSpecific(t *testing.T) {
	tests := []struct {
		name           string
		data           map[string]interface{}
		filePath       string
		contents       string
		wantMinErrors  int // Minimum errors expected (not counting suggestions)
	}{
		{
			name: "valid plugin",
			data: map[string]interface{}{
				"name":        "test-plugin",
				"description": "A comprehensive test plugin for validation purposes with detailed description",
				"version":     "1.0.0",
				"author": map[string]interface{}{
					"name": "Test Author",
				},
				"homepage":   "https://example.com",
				"repository": "https://github.com/test/test",
				"license":    "MIT",
				"keywords":   []interface{}{"test"},
			},
			filePath:      "plugin.json",
			contents:      `{"name":"test-plugin"}`,
			wantMinErrors: 0,
		},
		{
			name:          "missing name",
			data:          map[string]interface{}{},
			filePath:      "plugin.json",
			contents:      `{}`,
			wantMinErrors: 3, // name, description, author (errors only, not suggestions)
		},
		{
			name: "reserved word name",
			data: map[string]interface{}{
				"name":        "claude",
				"description": "A comprehensive test plugin for validation purposes with detailed description",
				"author": map[string]interface{}{
					"name": "Test",
				},
				"homepage":   "https://example.com",
				"repository": "https://github.com/test/test",
				"license":    "MIT",
				"keywords":   []interface{}{"test"},
			},
			filePath:      "plugin.json",
			contents:      `{"name":"claude"}`,
			wantMinErrors: 1,
		},
		{
			name: "name too long",
			data: map[string]interface{}{
				"name":        "this-is-a-very-long-plugin-name-that-exceeds-the-sixty-four-character-limit",
				"description": "A comprehensive test plugin for validation purposes with detailed description",
				"author": map[string]interface{}{
					"name": "Test",
				},
				"homepage":   "https://example.com",
				"repository": "https://github.com/test/test",
				"license":    "MIT",
				"keywords":   []interface{}{"test"},
			},
			filePath:      "plugin.json",
			contents:      `{"name":"this-is-a-very-long-plugin-name-that-exceeds-the-sixty-four-character-limit"}`,
			wantMinErrors: 1,
		},
		{
			name: "description too long",
			data: map[string]interface{}{
				"name":        "test",
				"description": string(make([]byte, 1025)),
				"author": map[string]interface{}{
					"name": "Test",
				},
				"homepage":   "https://example.com",
				"repository": "https://github.com/test/test",
				"license":    "MIT",
				"keywords":   []interface{}{"test"},
			},
			filePath:      "plugin.json",
			contents:      `{"name":"test","description":"..."}`,
			wantMinErrors: 1,
		},
		{
			name: "invalid semver",
			data: map[string]interface{}{
				"name":        "test",
				"description": "A comprehensive test plugin for validation purposes with detailed description",
				"version":     "1.0",
				"author": map[string]interface{}{
					"name": "Test",
				},
				"homepage":   "https://example.com",
				"repository": "https://github.com/test/test",
				"license":    "MIT",
				"keywords":   []interface{}{"test"},
			},
			filePath:      "plugin.json",
			contents:      `{"version":"1.0"}`,
			wantMinErrors: 1, // semver warning
		},
		{
			name: "missing author.name",
			data: map[string]interface{}{
				"name":        "test",
				"description": "A comprehensive test plugin for validation purposes with detailed description",
				"author":      map[string]interface{}{},
				"homepage":    "https://example.com",
				"repository":  "https://github.com/test/test",
				"license":     "MIT",
				"keywords":    []interface{}{"test"},
			},
			filePath:      "plugin.json",
			contents:      `{"author":{}}`,
			wantMinErrors: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			allIssues := validatePluginSpecific(tt.data, tt.filePath, tt.contents)

			// Count only errors and warnings (not suggestions)
			errorCount := 0
			for _, issue := range allIssues {
				if issue.Severity == "error" || issue.Severity == "warning" {
					errorCount++
				}
			}

			if errorCount < tt.wantMinErrors {
				t.Errorf("validatePluginSpecific() error count = %d, want at least %d", errorCount, tt.wantMinErrors)
				for _, err := range allIssues {
					t.Logf("  - %s: %s", err.Severity, err.Message)
				}
			}
		})
	}
}

func TestValidatePluginBestPractices(t *testing.T) {
	tests := []struct {
		name               string
		data               map[string]interface{}
		wantSuggestionCount int
	}{
		{
			name: "complete plugin",
			data: map[string]interface{}{
				"name":        "test",
				"description": "A comprehensive test plugin with detailed description that exceeds fifty characters",
				"homepage":    "https://example.com",
				"repository":  "https://github.com/user/repo",
				"license":     "MIT",
				"keywords":    []interface{}{"test", "plugin"},
			},
			wantSuggestionCount: 0,
		},
		{
			name: "minimal plugin",
			data: map[string]interface{}{
				"name":        "test",
				"description": "Short",
			},
			wantSuggestionCount: 5, // homepage, repository, license, keywords, short description
		},
		{
			name: "short description",
			data: map[string]interface{}{
				"name":        "test",
				"description": "Short description",
				"homepage":    "https://example.com",
				"repository":  "https://github.com/user/repo",
				"license":     "MIT",
				"keywords":    []interface{}{"test"},
			},
			wantSuggestionCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			suggestions := validatePluginBestPractices("plugin.json", "{}", tt.data)
			if len(suggestions) != tt.wantSuggestionCount {
				t.Errorf("validatePluginBestPractices() suggestion count = %d, want %d", len(suggestions), tt.wantSuggestionCount)
				for _, sugg := range suggestions {
					t.Logf("  - %s", sugg.Message)
				}
			}
		})
	}
}

func TestGetPluginImprovements(t *testing.T) {
	tests := []struct {
		name     string
		data     map[string]interface{}
		wantRecs int
	}{
		{
			name: "complete plugin",
			data: map[string]interface{}{
				"name":        "test",
				"description": "A comprehensive description that is long enough to meet requirements",
				"homepage":    "https://example.com",
				"repository":  "https://github.com/user/repo",
				"license":     "MIT",
				"keywords":    []interface{}{"test"},
				"readme":      "README.md",
			},
			wantRecs: 0,
		},
		{
			name: "minimal plugin",
			data: map[string]interface{}{
				"name":        "test",
				"description": "Short",
			},
			wantRecs: 6, // homepage, repository, license, keywords, readme, short description
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recs := GetPluginImprovements("{}", tt.data)
			if len(recs) != tt.wantRecs {
				t.Errorf("GetPluginImprovements() recommendation count = %d, want %d", len(recs), tt.wantRecs)
			}
		})
	}
}

func TestFindJSONFieldLine(t *testing.T) {
	content := `{
  "name": "test-plugin",
  "version": "1.0.0",
  "description": "A test plugin"
}`

	tests := []struct {
		field    string
		wantLine int
	}{
		{"name", 2},
		{"version", 3},
		{"description", 4},
		{"nonexistent", 0},
	}

	for _, tt := range tests {
		t.Run(tt.field, func(t *testing.T) {
			line := FindJSONFieldLine(content, tt.field)
			if line != tt.wantLine {
				t.Errorf("FindJSONFieldLine(%q) = %d, want %d", tt.field, line, tt.wantLine)
			}
		})
	}
}
