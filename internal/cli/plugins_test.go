package cli

import (
	"os"
	"path/filepath"
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
		name          string
		data          map[string]interface{}
		filePath      string
		contents      string
		wantMinErrors int // Minimum errors expected (not counting suggestions)
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
		{
			name: "valid relative paths",
			data: map[string]interface{}{
				"name":        "test-plugin",
				"description": "A comprehensive test plugin for validation purposes with detailed description",
				"version":     "1.0.0",
				"author":      map[string]interface{}{"name": "Test Author"},
				"homepage":    "https://example.com",
				"repository":  "https://github.com/test/test",
				"license":     "MIT",
				"keywords":    []interface{}{"test"},
				"commands":    []interface{}{"./commands/greet.md"},
				"agents":      []interface{}{"./agents/helper.md"},
			},
			filePath:      "plugin.json",
			contents:      `{"name":"test-plugin","commands":["./commands/greet.md"]}`,
			wantMinErrors: 0,
		},
		{
			name: "absolute path in commands",
			data: map[string]interface{}{
				"name":        "test-plugin",
				"description": "A comprehensive test plugin for validation purposes with detailed description",
				"version":     "1.0.0",
				"author":      map[string]interface{}{"name": "Test Author"},
				"homepage":    "https://example.com",
				"repository":  "https://github.com/test/test",
				"license":     "MIT",
				"keywords":    []interface{}{"test"},
				"commands":    []interface{}{"/etc/commands/greet.md"},
			},
			filePath:      "plugin.json",
			contents:      `{"name":"test-plugin","commands":["/etc/commands/greet.md"]}`,
			wantMinErrors: 1,
		},
		{
			name: "path traversal in skills",
			data: map[string]interface{}{
				"name":        "test-plugin",
				"description": "A comprehensive test plugin for validation purposes with detailed description",
				"version":     "1.0.0",
				"author":      map[string]interface{}{"name": "Test Author"},
				"homepage":    "https://example.com",
				"repository":  "https://github.com/test/test",
				"license":     "MIT",
				"keywords":    []interface{}{"test"},
				"skills":      []interface{}{"./skills/../../../etc/passwd"},
			},
			filePath:      "plugin.json",
			contents:      `{"name":"test-plugin","skills":["./skills/../../../etc/passwd"]}`,
			wantMinErrors: 1, // warning counts as error/warning
		},
		{
			name: "outputStyles field recognized",
			data: map[string]interface{}{
				"name":         "test-plugin",
				"description":  "A comprehensive test plugin for validation purposes with detailed description",
				"version":      "1.0.0",
				"author":       map[string]interface{}{"name": "Test Author"},
				"homepage":     "https://example.com",
				"repository":   "https://github.com/test/test",
				"license":      "MIT",
				"keywords":     []interface{}{"test"},
				"outputStyles": []interface{}{"./styles/compact.json"},
			},
			filePath:      "plugin.json",
			contents:      `{"name":"test-plugin","outputStyles":["./styles/compact.json"]}`,
			wantMinErrors: 0,
		},
		{
			name: "lspServers field recognized",
			data: map[string]interface{}{
				"name":        "test-plugin",
				"description": "A comprehensive test plugin for validation purposes with detailed description",
				"version":     "1.0.0",
				"author":      map[string]interface{}{"name": "Test Author"},
				"homepage":    "https://example.com",
				"repository":  "https://github.com/test/test",
				"license":     "MIT",
				"keywords":    []interface{}{"test"},
				"lspServers":  []interface{}{"./lsp/gopls.json"},
			},
			filePath:      "plugin.json",
			contents:      `{"name":"test-plugin","lspServers":["./lsp/gopls.json"]}`,
			wantMinErrors: 0,
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
		name                string
		data                map[string]interface{}
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

func TestValidatePluginPaths(t *testing.T) {
	tests := []struct {
		name         string
		data         map[string]interface{}
		wantErrors   int
		wantWarnings int
	}{
		{
			name: "all relative paths",
			data: map[string]interface{}{
				"commands": []interface{}{"./commands/greet.md", "./commands/help.md"},
				"agents":   []interface{}{"./agents/helper.md"},
			},
			wantErrors:   0,
			wantWarnings: 0,
		},
		{
			name: "absolute path error",
			data: map[string]interface{}{
				"commands": []interface{}{"/usr/local/commands/greet.md"},
			},
			wantErrors:   1,
			wantWarnings: 0,
		},
		{
			name: "path traversal warning",
			data: map[string]interface{}{
				"skills": []interface{}{"./skills/../../secret.md"},
			},
			wantErrors:   0,
			wantWarnings: 1,
		},
		{
			name: "absolute path with traversal triggers both",
			data: map[string]interface{}{
				"hooks": []interface{}{"/etc/../passwd"},
			},
			wantErrors:   1,
			wantWarnings: 1,
		},
		{
			name: "object values checked",
			data: map[string]interface{}{
				"mcpServers": map[string]interface{}{
					"server1": "/opt/mcp/server",
				},
			},
			wantErrors:   1,
			wantWarnings: 0,
		},
		{
			name: "no path fields present",
			data: map[string]interface{}{
				"name": "test-plugin",
			},
			wantErrors:   0,
			wantWarnings: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issues := validatePluginPaths(tt.data, "plugin.json", "{}")
			errorCount := 0
			warningCount := 0
			for _, issue := range issues {
				switch issue.Severity {
				case "error":
					errorCount++
				case "warning":
					warningCount++
				}
			}
			if errorCount != tt.wantErrors {
				t.Errorf("errors = %d, want %d", errorCount, tt.wantErrors)
				for _, issue := range issues {
					t.Logf("  - %s: %s", issue.Severity, issue.Message)
				}
			}
			if warningCount != tt.wantWarnings {
				t.Errorf("warnings = %d, want %d", warningCount, tt.wantWarnings)
				for _, issue := range issues {
					t.Logf("  - %s: %s", issue.Severity, issue.Message)
				}
			}
		})
	}
}

func TestExtractPaths(t *testing.T) {
	tests := []struct {
		name      string
		value     interface{}
		wantCount int
	}{
		{
			name:      "string value",
			value:     "./commands/greet.md",
			wantCount: 1,
		},
		{
			name:      "string array",
			value:     []interface{}{"./a.md", "./b.md"},
			wantCount: 2,
		},
		{
			name: "object array",
			value: []interface{}{
				map[string]interface{}{"path": "./commands/greet.md", "name": "greet"},
			},
			wantCount: 2, // extracts all string values from the object
		},
		{
			name: "map value",
			value: map[string]interface{}{
				"server1": "./mcp/server1.json",
				"server2": "./mcp/server2.json",
			},
			wantCount: 2,
		},
		{
			name:      "nil value",
			value:     nil,
			wantCount: 0,
		},
		{
			name:      "non-path type",
			value:     42,
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			paths := extractPaths(tt.value)
			if len(paths) != tt.wantCount {
				t.Errorf("extractPaths() count = %d, want %d; paths = %v", len(paths), tt.wantCount, paths)
			}
		})
	}
}

func TestKnownPluginFields(t *testing.T) {
	expected := []string{
		"name", "description", "version", "author", "homepage", "repository",
		"license", "keywords", "readme", "commands", "agents", "skills",
		"hooks", "mcpServers", "outputStyles", "lspServers",
	}
	for _, field := range expected {
		if !knownPluginFields[field] {
			t.Errorf("knownPluginFields missing expected field: %s", field)
		}
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

func TestIsGlobPattern(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		{"./commands/greet.md", false},
		{"./commands/*.md", true},
		{"./agents/helper-?.md", true},
		{"./skills/[abc].md", true},
		{"./hooks/{pre,post}.md", true},
		{"README.md", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got := isGlobPattern(tt.path)
			if got != tt.want {
				t.Errorf("isGlobPattern(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

func TestValidatePluginPathsExist(t *testing.T) {
	// Helper to create a temp directory with a .claude-plugin structure.
	// fileRelPaths are paths relative to the plugin directory.
	setupPluginDir := func(t *testing.T, fileRelPaths []string) string {
		t.Helper()
		root := t.TempDir()
		pluginDir := filepath.Join(root, ".claude-plugin")
		if err := os.MkdirAll(pluginDir, 0o755); err != nil {
			t.Fatalf("failed to create plugin dir: %v", err)
		}
		for _, rel := range fileRelPaths {
			abs := filepath.Join(pluginDir, rel)
			dir := filepath.Dir(abs)
			if err := os.MkdirAll(dir, 0o755); err != nil {
				t.Fatalf("failed to create dir %s: %v", dir, err)
			}
			if err := os.WriteFile(abs, []byte("content"), 0o644); err != nil {
				t.Fatalf("failed to create file %s: %v", abs, err)
			}
		}
		return root
	}

	tests := []struct {
		name         string
		files        []string // files to create relative to plugin dir
		data         map[string]interface{}
		wantWarnings int
		emptyRoot    bool // use empty rootPath to disable validation
	}{
		{
			name:  "all paths exist",
			files: []string{"./commands/greet.md", "./agents/helper.md", "README.md"},
			data: map[string]interface{}{
				"commands": []interface{}{"./commands/greet.md"},
				"agents":   []interface{}{"./agents/helper.md"},
				"readme":   "README.md",
			},
			wantWarnings: 0,
		},
		{
			name:  "missing command file",
			files: []string{"./agents/helper.md"},
			data: map[string]interface{}{
				"commands": []interface{}{"./commands/missing.md"},
				"agents":   []interface{}{"./agents/helper.md"},
			},
			wantWarnings: 1,
		},
		{
			name:  "missing readme file",
			files: []string{},
			data: map[string]interface{}{
				"readme": "README.md",
			},
			wantWarnings: 1,
		},
		{
			name:  "multiple missing paths",
			files: []string{},
			data: map[string]interface{}{
				"commands": []interface{}{"./commands/a.md", "./commands/b.md"},
				"agents":   []interface{}{"./agents/helper.md"},
			},
			wantWarnings: 3,
		},
		{
			name:  "glob patterns skipped",
			files: []string{},
			data: map[string]interface{}{
				"commands": []interface{}{"./commands/*.md"},
				"skills":   []interface{}{"./skills/[a-z]*.md"},
			},
			wantWarnings: 0,
		},
		{
			name:  "absolute paths skipped",
			files: []string{},
			data: map[string]interface{}{
				"commands": []interface{}{"/etc/commands/greet.md"},
			},
			wantWarnings: 0,
		},
		{
			name:  "traversal paths skipped",
			files: []string{},
			data: map[string]interface{}{
				"skills": []interface{}{"./skills/../../etc/passwd"},
			},
			wantWarnings: 0,
		},
		{
			name:  "empty rootPath disables validation",
			files: []string{},
			data: map[string]interface{}{
				"commands": []interface{}{"./commands/missing.md"},
			},
			wantWarnings: 0,
			emptyRoot:    true,
		},
		{
			name:  "no path fields present",
			files: []string{},
			data: map[string]interface{}{
				"name": "test-plugin",
			},
			wantWarnings: 0,
		},
		{
			name:  "map-style mcpServers with missing paths",
			files: []string{},
			data: map[string]interface{}{
				"mcpServers": map[string]interface{}{
					"server1": "./mcp/server1.json",
				},
			},
			wantWarnings: 1,
		},
		{
			name:  "map-style mcpServers with existing paths",
			files: []string{"./mcp/server1.json"},
			data: map[string]interface{}{
				"mcpServers": map[string]interface{}{
					"server1": "./mcp/server1.json",
				},
			},
			wantWarnings: 0,
		},
		{
			name:  "URLs in path fields skipped",
			files: []string{},
			data: map[string]interface{}{
				"hooks": []interface{}{"http://example.com/hook", "https://example.com/hook"},
			},
			wantWarnings: 0,
		},
		{
			name:  "mixed existing and missing paths",
			files: []string{"./commands/exists.md"},
			data: map[string]interface{}{
				"commands": []interface{}{"./commands/exists.md", "./commands/missing.md"},
			},
			wantWarnings: 1,
		},
		{
			name:  "directory path exists",
			files: []string{"./skills/my-skill/SKILL.md"},
			data: map[string]interface{}{
				"skills": []interface{}{"./skills/my-skill"},
			},
			wantWarnings: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var rootPath string
			if tt.emptyRoot {
				rootPath = ""
			} else {
				rootPath = setupPluginDir(t, tt.files)
			}

			filePath := ".claude-plugin/plugin.json"
			contents := `{"commands": []}`

			warnings := validatePluginPathsExist(tt.data, rootPath, filePath, contents)
			if len(warnings) != tt.wantWarnings {
				t.Errorf("warnings = %d, want %d", len(warnings), tt.wantWarnings)
				for _, w := range warnings {
					t.Logf("  - %s: %s", w.Severity, w.Message)
				}
			}

			// Verify all returned issues are warnings (not errors)
			for _, w := range warnings {
				if w.Severity != "warning" {
					t.Errorf("expected severity 'warning', got %q for: %s", w.Severity, w.Message)
				}
			}
		})
	}
}
