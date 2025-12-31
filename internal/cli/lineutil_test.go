package cli

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
		data         map[string]interface{}
		filePath     string
		wantErrCount int
	}{
		{
			name:         "no allowed-tools",
			data:         map[string]interface{}{},
			filePath:     "test.md",
			wantErrCount: 0,
		},
		{
			name: "valid tools",
			data: map[string]interface{}{
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
		data          map[string]interface{}
		componentType string
		wantErrCount  int
	}{
		{
			name: "agent with tools",
			data: map[string]interface{}{
				"tools": "Read",
			},
			componentType: "agent",
			wantErrCount:  0,
		},
		{
			name: "command with allowed-tools",
			data: map[string]interface{}{
				"allowed-tools": "Read",
			},
			componentType: "command",
			wantErrCount:  0,
		},
		{
			name: "agent with wrong field",
			data: map[string]interface{}{
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
	data := map[string]interface{}{
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
