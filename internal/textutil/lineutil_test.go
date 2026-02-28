package textutil

import (
	"testing"
)

func TestFindLineNumber(t *testing.T) {
	content := "line 1\nline 2\ntarget line\nline 4"

	tests := []struct {
		name    string
		pattern string
		want    int
	}{
		{"found", "target", 3},
		{"not found", "missing", 0},
		{"first line", "line 1", 1},
		{"last line", "line 4", 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FindLineNumber(content, tt.pattern)
			if got != tt.want {
				t.Errorf("FindLineNumber(%q) = %d, want %d", tt.pattern, got, tt.want)
			}
		})
	}
}

func TestFindSectionLine(t *testing.T) {
	content := "# Title\n\n## Section1\n\nContent\n\n### Section2\n"

	tests := []struct {
		name        string
		sectionName string
		want        int
	}{
		{"h2 section", "Section1", 3},
		{"h3 section", "Section2", 7},
		{"h1 section", "Title", 1},
		{"not found", "Missing", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FindSectionLine(content, tt.sectionName)
			if got != tt.want {
				t.Errorf("FindSectionLine(%q) = %d, want %d", tt.sectionName, got, tt.want)
			}
		})
	}
}

func TestFindFrontmatterFieldLine(t *testing.T) {
	content := "---\nname: test\ndescription: test desc\n---\nBody"

	tests := []struct {
		name      string
		fieldName string
		want      int
	}{
		{"first field", "name", 2},
		{"second field", "description", 3},
		{"not found", "missing", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FindFrontmatterFieldLine(content, tt.fieldName)
			if got != tt.want {
				t.Errorf("FindFrontmatterFieldLine(%q) = %d, want %d", tt.fieldName, got, tt.want)
			}
		})
	}
}

func TestGetFrontmatterEndLine(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    int
	}{
		{
			name:    "with frontmatter",
			content: "---\nname: test\n---\nBody",
			want:    3,
		},
		{
			name:    "no frontmatter",
			content: "Just body content",
			want:    1,
		},
		{
			name:    "only opening",
			content: "---\nname: test\nno closing",
			want:    1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetFrontmatterEndLine(tt.content)
			if got != tt.want {
				t.Errorf("GetFrontmatterEndLine() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestCountLines(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    int
	}{
		{"single line", "one line", 1},
		{"multiple lines", "line 1\nline 2\nline 3", 3},
		{"empty", "", 1},
		{"trailing newline", "line 1\nline 2\n", 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CountLines(tt.content)
			if got != tt.want {
				t.Errorf("CountLines() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestDetermineSeverity(t *testing.T) {
	tests := []struct {
		name    string
		message string
		want    string
	}{
		{"required field", "Required field 'name' is missing", "high"},
		{"name must", "Name must be lowercase", "high"},
		{"fat agent", "This is a fat agent", "high"},
		{"best practice", "Best practice violation", "medium"},
		{"lines mention", "Over 200 lines limit", "medium"},
		{"foundation", "Missing Foundation section", "medium"},
		{"consider", "Consider adding examples", "low"},
		{"default", "Random message", "low"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DetermineSeverity(tt.message)
			if got != tt.want {
				t.Errorf("DetermineSeverity(%q) = %q, want %q", tt.message, got, tt.want)
			}
		})
	}
}

func TestValidateAllowedTools(t *testing.T) {
	tests := []struct {
		name         string
		data         map[string]any
		filePath     string
		wantErrCount int
	}{
		{
			name:         "no allowed-tools",
			data:         map[string]any{},
			filePath:     "test.md",
			wantErrCount: 0,
		},
		{
			name: "valid tools",
			data: map[string]any{
				"allowed-tools": "Read, Write",
			},
			filePath:     "test.md",
			wantErrCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := ValidateAllowedTools(tt.data, tt.filePath, "")
			if len(errors) < tt.wantErrCount {
				t.Errorf("ValidateAllowedTools() errors = %d, want at least %d", len(errors), tt.wantErrCount)
			}
		})
	}
}

func TestValidateToolFieldName(t *testing.T) {
	tests := []struct {
		name          string
		data          map[string]any
		componentType string
		wantErrCount  int
	}{
		{
			name: "agent with tools",
			data: map[string]any{
				"tools": "Read",
			},
			componentType: "agent",
			wantErrCount:  0,
		},
		{
			name: "command with allowed-tools",
			data: map[string]any{
				"allowed-tools": "Read",
			},
			componentType: "command",
			wantErrCount:  0,
		},
		{
			name: "agent with wrong field",
			data: map[string]any{
				"allowed-tools": "Read",
			},
			componentType: "agent",
			wantErrCount:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := ValidateToolFieldName(tt.data, "test.md", "", tt.componentType)
			errCount := 0
			for _, e := range errors {
				if e.Severity == "error" {
					errCount++
				}
			}
			if errCount < tt.wantErrCount {
				t.Errorf("ValidateToolFieldName() errors = %d, want at least %d", errCount, tt.wantErrCount)
			}
		})
	}
}

func TestGetImprovementFunctions(t *testing.T) {
	data := map[string]any{
		"name":        "test",
		"description": "test",
	}
	content := "test content"

	// Test GetAgentImprovements
	agentImprovements := GetAgentImprovements(content, data)
	if agentImprovements == nil {
		t.Error("GetAgentImprovements() returned nil")
	}

	// Test GetCommandImprovements
	cmdImprovements := GetCommandImprovements(content, data)
	if cmdImprovements == nil {
		t.Error("GetCommandImprovements() returned nil")
	}

	// Test GetSkillImprovements
	skillImprovements := GetSkillImprovements(content, data)
	if skillImprovements == nil {
		t.Error("GetSkillImprovements() returned nil")
	}
}

func TestSeverityConstants(t *testing.T) {
	if SeverityHigh != "high" {
		t.Errorf("SeverityHigh = %q, want %q", SeverityHigh, "high")
	}
	if SeverityMedium != "medium" {
		t.Errorf("SeverityMedium = %q, want %q", SeverityMedium, "medium")
	}
	if SeverityLow != "low" {
		t.Errorf("SeverityLow = %q, want %q", SeverityLow, "low")
	}
}

func TestIsPlaceholderSecret(t *testing.T) {
	tests := []struct {
		name         string
		line         string
		isPlaceholder bool
	}{
		// Should be detected as placeholders
		{"example keyword", `api_key: "your-example-key-here"`, true},
		{"placeholder keyword", `secret: "placeholder-value"`, true},
		{"your-prefix", `api_key: "your-api-key"`, true},
		{"your_prefix", `api_key: "your_api_key"`, true},
		{"angle bracket placeholder", `api_key: "<YOUR_API_KEY>"`, true},
		{"square bracket placeholder", `api_key: "[your-key-here]"`, true},
		{"curly bracket placeholder", `api_key: "{your-key-here}"`, true},
		{"replace keyword", `token: "replace-with-real-token"`, true},
		{"insert keyword", `password: "insert-password"`, true},
		{"changeme keyword", `secret: "changeme"`, true},
		{"change-me keyword", `secret: "change-me-later"`, true},
		{"xxx pattern", `api_key: "sk-xxxxxxxxxxxxxxxxxxxxxx"`, true},
		{"multiple x's", `token: "xxxxxxxxxxxx"`, true},
		{"asterisks", `password: "***hidden***"`, true},
		{"dummy keyword", `api_key: "dummy-key-for-testing"`, true},
		{"sample keyword", `token: "sample-token-value"`, true},
		{"fake keyword", `secret: "fake-secret-value"`, true},
		{"test-key", `api_key: "test-key-12345"`, true},
		{"testkey", `api_key: "testkey12345678"`, true},
		{"test_key", `api_key: "test_key_value"`, true},
		{"my-key", `api_key: "my-key-goes-here"`, true},
		{"mykey", `api_key: "mykey12345"`, true},
		{"ending with here>", `api_key: "put-your-key-here>"`, true},
		{"ending with -here", `token: "api-key-here"`, true},
		{"zeros pattern", `api_key: "sk-0000000000000000000000"`, true},
		{"AWS example access key", "AKIAIOSFODNN7EXAMPLE", true},
		{"AWS example secret key", "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY", true},
		{"angle bracket with key", `<api_key>`, true},
		{"angle bracket with token", `<token>`, true},
		{"angle bracket with secret", `<secret>`, true},
		{"file-value example", `apiKey: "file-value"`, true},
		{"env-value example", `api_key: "env-value"`, true},
		{"some-value example", `secret: "some-value"`, true},
		{"value-here example", `token: "value-here"`, true},
		{"Go comment example", `// config.yaml: apiKey: "real-looking-key-here"`, true},
		{"hash comment example", `# api_key: "actual-secret-value"`, true},

		// Should NOT be detected as placeholders (real-looking secrets)
		{"real-looking API key", `api_key: "sk-AbCdEfGhIjKlMnOpQrStUvWxYz"`, false},
		{"real-looking password", `password: "MyP@ssw0rd!2024"`, false},
		{"real-looking token", `token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9"`, false},
		{"real-looking AWS key", `aws_access_key_id: "AKIAZ7VRSQHFPT9ABCDE"`, false},
		{"GitHub PAT", `ghp_AbCdEfGhIjKlMnOpQrStUvWxYz123456`, false},
		{"random string", `secret: "f8a7b2c3d4e5f6g7h8i9j0"`, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isPlaceholderSecret(tt.line)
			if result != tt.isPlaceholder {
				t.Errorf("isPlaceholderSecret(%q) = %v, want %v", tt.line, result, tt.isPlaceholder)
			}
		})
	}
}


func TestValidateAllowedToolsUnknownTool(t *testing.T) {
	tests := []struct {
		name         string
		data         map[string]any
		wantWarnings int
	}{
		{
			name: "unknown tool",
			data: map[string]any{
				"allowed-tools": "UnknownTool",
			},
			wantWarnings: 1,
		},
		{
			name: "multiple unknown tools",
			data: map[string]any{
				"allowed-tools": "UnknownTool1, UnknownTool2",
			},
			wantWarnings: 2,
		},
		{
			name: "Task with npm prefix",
			data: map[string]any{
				"tools": "Bash(npm:*)",
			},
			wantWarnings: 0,
		},
		{
			name: "Task pattern valid",
			data: map[string]any{
				"allowed-tools": "Task(foo-specialist)",
			},
			wantWarnings: 0,
		},
		{
			name: "valid tools mixed",
			data: map[string]any{
				"allowed-tools": "Read, Write, Edit",
			},
			wantWarnings: 0,
		},
		{
			name: "empty tools",
			data: map[string]any{
				"allowed-tools": "",
			},
			wantWarnings: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			warnings := ValidateAllowedTools(tt.data, "test.md", "---\nallowed-tools: test\n---\n")
			if len(warnings) != tt.wantWarnings {
				t.Errorf("ValidateAllowedTools() warnings = %d, want %d", len(warnings), tt.wantWarnings)
				for _, w := range warnings {
					t.Logf("  Warning: %s", w.Message)
				}
			}
		})
	}
}

func TestValidateToolFieldNameEdgeCases(t *testing.T) {
	tests := []struct {
		name          string
		data          map[string]any
		componentType string
		wantErrors    int
	}{
		{
			name: "command with tools field",
			data: map[string]any{
				"tools": "Read",
			},
			componentType: "command",
			wantErrors:    1,
		},
		{
			name: "skill with tools field",
			data: map[string]any{
				"tools": "Read",
			},
			componentType: "skill",
			wantErrors:    1,
		},
		{
			name:          "no tool fields",
			data:          map[string]any{},
			componentType: "agent",
			wantErrors:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := ValidateToolFieldName(tt.data, "test.md", "---\ntools: test\n---\n", tt.componentType)
			errCount := 0
			for _, e := range errors {
				if e.Severity == "error" {
					errCount++
				}
			}
			if errCount != tt.wantErrors {
				t.Errorf("ValidateToolFieldName() errors = %d, want %d", errCount, tt.wantErrors)
			}
		})
	}
}
