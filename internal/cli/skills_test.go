package cli

import (
	"strings"
	"testing"
)

func TestValidateSkillBestPractices(t *testing.T) {
	tests := []struct {
		name         string
		filePath     string
		contents     string
		fmData       map[string]interface{}
		wantContains []string
	}{
		{
			name:     "first person description",
			filePath: "skills/test/SKILL.md",
			contents: "---\ndescription: I analyze code\n---\n",
			fmData: map[string]interface{}{
				"description": "I analyze code",
			},
			wantContains: []string{"third person"},
		},
		{
			name:     "addresses user",
			filePath: "skills/test/SKILL.md",
			contents: "---\ndescription: You should use this\n---\n",
			fmData: map[string]interface{}{
				"description": "You should use this",
			},
			wantContains: []string{"address the user"},
		},
		{
			name:     "short description",
			filePath: "skills/test/SKILL.md",
			contents: "---\ndescription: Short\n---\n",
			fmData: map[string]interface{}{
				"description": "Short",
			},
			wantContains: []string{"50+"},
		},
		{
			name:     "missing trigger phrases",
			filePath: "skills/test/SKILL.md",
			contents: "---\ndescription: This is a skill that does things without triggers\n---\n",
			fmData: map[string]interface{}{
				"description": "This is a skill that does things without triggers",
			},
			wantContains: []string{"trigger phrases"},
		},
		{
			name:     "invalid semver",
			filePath: "skills/test/SKILL.md",
			contents: "---\nversion: 1.0\n---\n",
			fmData: map[string]interface{}{
				"version": "1.0",
			},
			wantContains: []string{"semver"},
		},
		{
			name:         "missing anti-patterns",
			filePath:     "skills/test/SKILL.md",
			contents:     "---\nname: test\n---\nContent without anti-patterns",
			fmData:       map[string]interface{}{"name": "test"},
			wantContains: []string{"Anti-Patterns"},
		},
		{
			name:         "missing examples",
			filePath:     "skills/test/SKILL.md",
			contents:     "---\nname: test\n---\nContent without examples",
			fmData:       map[string]interface{}{"name": "test"},
			wantContains: []string{"Examples"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			suggestions := validateSkillBestPractices(tt.filePath, tt.contents, tt.fmData)

			for _, want := range tt.wantContains {
				found := false
				for _, sugg := range suggestions {
					if strings.Contains(sugg.Message, want) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("validateSkillBestPractices() should contain suggestion about %q", want)
					for _, s := range suggestions {
						t.Logf("  Got: %s", s.Message)
					}
				}
			}
		})
	}
}

func TestExtractSkillName(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		filePath string
		want     string
	}{
		{
			name:     "extract from heading",
			content:  "# Test\n\nContent",
			filePath: "skills/test-skill/SKILL.md",
			want:     "test",
		},
		{
			name:     "strip Agent prefix",
			content:  "# Agent Test\n\nContent",
			filePath: "skills/test/SKILL.md",
			want:     "test",
		},
		{
			name:     "strip Command prefix",
			content:  "# Command Test\n\nContent",
			filePath: "skills/test/SKILL.md",
			want:     "test",
		},
		{
			name:     "strip Skill prefix",
			content:  "# Skill Test\n\nContent",
			filePath: "skills/test/SKILL.md",
			want:     "test",
		},
		{
			name:     "strip Patterns suffix",
			content:  "# Test Patterns\n\nContent",
			filePath: "skills/test/SKILL.md",
			want:     "test",
		},
		{
			name:     "fallback to filename",
			content:  "No heading",
			filePath: "skills/fallback-name/SKILL.md",
			want:     "SKILL.md",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractSkillName(tt.content, tt.filePath)
			if got != tt.want {
				t.Errorf("extractSkillName() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestKnownSkillFields(t *testing.T) {
	expected := []string{"name", "description", "allowed-tools", "model"}
	for _, field := range expected {
		if !knownSkillFields[field] {
			t.Errorf("knownSkillFields missing expected field: %s", field)
		}
	}
}
