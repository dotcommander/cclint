package lint

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/dotcommander/cclint/internal/discovery"
)

func TestCheckClaudeLocalGitignore(t *testing.T) {
	tests := []struct {
		name             string
		createLocal      bool
		gitignoreContent string
		wantWarnings     int
	}{
		{
			name:         "no CLAUDE.local.md",
			createLocal:  false,
			wantWarnings: 0,
		},
		{
			name:         "CLAUDE.local.md with no .gitignore",
			createLocal:  true,
			wantWarnings: 1,
		},
		{
			name:             "CLAUDE.local.md in .gitignore",
			createLocal:      true,
			gitignoreContent: "CLAUDE.local.md\n",
			wantWarnings:     0,
		},
		{
			name:             "wildcard pattern",
			createLocal:      true,
			gitignoreContent: "*.local.md\n",
			wantWarnings:     0,
		},
		{
			name:             "CLAUDE.local.md not in .gitignore",
			createLocal:      true,
			gitignoreContent: "node_modules/\n",
			wantWarnings:     1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			if tt.createLocal {
				localPath := filepath.Join(tmpDir, "CLAUDE.local.md")
				if err := os.WriteFile(localPath, []byte("test"), 0644); err != nil {
					t.Fatal(err)
				}
			}

			if tt.gitignoreContent != "" {
				gitignorePath := filepath.Join(tmpDir, ".gitignore")
				if err := os.WriteFile(gitignorePath, []byte(tt.gitignoreContent), 0644); err != nil {
					t.Fatal(err)
				}
			}

			errors := CheckClaudeLocalGitignore(tmpDir)

			warnCount := 0
			for _, e := range errors {
				if e.Severity == "warning" {
					warnCount++
				}
			}

			if warnCount != tt.wantWarnings {
				t.Errorf("CheckClaudeLocalGitignore() warnings = %d, want %d", warnCount, tt.wantWarnings)
				for _, e := range errors {
					t.Logf("  %s: %s", e.Severity, e.Message)
				}
			}
		})
	}
}

func TestIsInGitignore(t *testing.T) {
	tests := []struct {
		name             string
		gitignoreContent string
		filename         string
		want             bool
	}{
		{
			name:             "exact match",
			gitignoreContent: "CLAUDE.local.md\n",
			filename:         "CLAUDE.local.md",
			want:             true,
		},
		{
			name:             "wildcard match",
			gitignoreContent: "*.local.md\n",
			filename:         "CLAUDE.local.md",
			want:             true,
		},
		{
			name:             "CLAUDE.local.* pattern",
			gitignoreContent: "CLAUDE.local.*\n",
			filename:         "CLAUDE.local.md",
			want:             true,
		},
		{
			name:             "comment line",
			gitignoreContent: "# CLAUDE.local.md\n",
			filename:         "CLAUDE.local.md",
			want:             false,
		},
		{
			name:             "negation",
			gitignoreContent: "*.md\n!CLAUDE.local.md\n",
			filename:         "CLAUDE.local.md",
			want:             false,
		},
		{
			name:             "not matched",
			gitignoreContent: "node_modules/\n",
			filename:         "CLAUDE.local.md",
			want:             false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			gitignorePath := filepath.Join(tmpDir, ".gitignore")

			if err := os.WriteFile(gitignorePath, []byte(tt.gitignoreContent), 0644); err != nil {
				t.Fatal(err)
			}

			got := isInGitignore(gitignorePath, tt.filename)
			if got != tt.want {
				t.Errorf("isInGitignore() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCheckCombinedMemorySize(t *testing.T) {
	tests := []struct {
		name         string
		files        []discovery.File
		wantWarnings int
	}{
		{
			name: "under threshold",
			files: []discovery.File{
				{Type: discovery.FileTypeContext, Contents: "Small content"},
			},
			wantWarnings: 0,
		},
		{
			name: "over threshold",
			files: []discovery.File{
				{Type: discovery.FileTypeContext, Contents: string(make([]byte, 25*1024))},
			},
			wantWarnings: 1,
		},
		{
			name: "conditional rules excluded",
			files: []discovery.File{
				{Type: discovery.FileTypeContext, Contents: "Small content"},
				{Type: discovery.FileTypeRule, Contents: "---\npaths: \"**/*.go\"\n---\nRule content"},
			},
			wantWarnings: 0,
		},
		{
			name: "always-loaded rules counted",
			files: []discovery.File{
				{Type: discovery.FileTypeContext, Contents: string(make([]byte, 15*1024))},
				{Type: discovery.FileTypeRule, Contents: string(make([]byte, 10*1024))},
			},
			wantWarnings: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			errors := CheckCombinedMemorySize(tmpDir, tt.files)

			warnCount := 0
			for _, e := range errors {
				if e.Severity == "warning" {
					warnCount++
				}
			}

			if warnCount != tt.wantWarnings {
				t.Errorf("CheckCombinedMemorySize() warnings = %d, want %d", warnCount, tt.wantWarnings)
			}
		})
	}
}

func TestRuleHasPathsConstraint(t *testing.T) {
	tests := []struct {
		name     string
		contents string
		want     bool
	}{
		{
			name:     "has paths",
			contents: "---\npaths: \"**/*.go\"\n---\nContent",
			want:     true,
		},
		{
			name:     "no paths",
			contents: "---\nname: test\n---\nContent",
			want:     false,
		},
		{
			name:     "no frontmatter",
			contents: "Just content",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ruleHasPathsConstraint(tt.contents)
			if got != tt.want {
				t.Errorf("ruleHasPathsConstraint() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFormatSizeWarning(t *testing.T) {
	tests := []struct {
		name            string
		alwaysLoaded    int64
		conditional     int64
		threshold       int64
		wantContains    []string
		wantNotContains []string
	}{
		{
			name:         "basic warning",
			alwaysLoaded: 25 * 1024,
			conditional:  0,
			threshold:    20 * 1024,
			wantContains: []string{"25.0KB", "20KB", "threshold"},
		},
		{
			name:         "with conditional",
			alwaysLoaded: 25 * 1024,
			conditional:  10 * 1024,
			threshold:    20 * 1024,
			wantContains: []string{"25.0KB", "10.0KB", "conditional"},
		},
		{
			name:         "no conditional",
			alwaysLoaded: 25 * 1024,
			conditional:  0,
			threshold:    20 * 1024,
			wantContains: []string{"25.0KB", "20KB"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := formatSizeWarning(tt.alwaysLoaded, tt.conditional, tt.threshold)

			for _, want := range tt.wantContains {
				if !containsSubstring(msg, want) {
					t.Errorf("formatSizeWarning() should contain %q\nGot: %s", want, msg)
				}
			}

			for _, notWant := range tt.wantNotContains {
				if containsSubstring(msg, notWant) {
					t.Errorf("formatSizeWarning() should NOT contain %q\nGot: %s", notWant, msg)
				}
			}
		})
	}
}

func TestCombinedMemorySizeWarningKB(t *testing.T) {
	if CombinedMemorySizeWarningKB != 20 {
		t.Errorf("CombinedMemorySizeWarningKB = %d, want 20", CombinedMemorySizeWarningKB)
	}
}
