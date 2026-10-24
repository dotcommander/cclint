package discovery

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectFileTypeRetainedPaths(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	tests := []struct {
		path string
		want FileType
	}{
		{".claude/rules/core.md", FileTypeRule},
		{".claude/rules/nested/quality.md", FileTypeRule},
		{".claude/agents/test.md", FileTypeAgent},
		{"agents/nested/test.md", FileTypeAgent},
		{".claude/commands/test.md", FileTypeCommand},
		{"commands/test.md", FileTypeCommand},
		{".claude/settings.json", FileTypeSettings},
		{"config/settings.json", FileTypeSettings},
		{".claude/CLAUDE.md", FileTypeContext},
		{"CLAUDE.md", FileTypeContext},
		{".claude/claude.md", FileTypeContext},
		{".claude-plugin/plugin.json", FileTypePlugin},
		{"pkg/.claude-plugin/plugin.json", FileTypePlugin},
		{"random/file.md", FileTypeUnknown},
		{"data.json", FileTypeUnknown},
		{"README.md", FileTypeUnknown},
		{"rules/go.md", FileTypeRule},
	}
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			t.Parallel()
			got, err := DetectFileType(filepath.Join(root, filepath.FromSlash(tt.path)), root)
			if got != tt.want || (err != nil) != (tt.want == FileTypeUnknown) {
				t.Errorf("DetectFileType(%q) = %v, %v; want %v", tt.path, got, err, tt.want)
			}
		})
	}
}

func TestProcessMatchContents(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		contents string
		exists   bool
	}{
		{name: "content", contents: "test file content\nwith multiple lines", exists: true},
		{name: "empty", exists: true},
		{name: "missing"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			if err := os.Mkdir(filepath.Join(root, "agents"), 0755); err != nil {
				t.Fatal(err)
			}
			relPath := "agents/test.md"
			fullPath := filepath.Join(root, filepath.FromSlash(relPath))
			if tt.exists {
				if err := os.WriteFile(fullPath, []byte(tt.contents), 0644); err != nil {
					t.Fatal(err)
				}
			}
			file, found := NewFileDiscovery(root, false).processMatch(relPath, FileTypeAgent)
			if found != tt.exists {
				t.Fatalf("found=%v, want %v", found, tt.exists)
			}
			if found && (file.Contents != tt.contents || file.Type != FileTypeAgent || file.RelPath != relPath || file.Path != fullPath || file.Size != int64(len(tt.contents))) {
				t.Errorf("discovered file fields changed: %+v", file)
			}
		})
	}
}
