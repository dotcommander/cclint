package cli

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestExtractImports(t *testing.T) {
	tests := []struct {
		name     string
		contents string
		want     []string
	}{
		{
			name:     "no imports",
			contents: "Just regular markdown content.\nNo imports here.",
			want:     nil,
		},
		{
			name:     "single relative import",
			contents: "@./other.md",
			want:     []string{"./other.md"},
		},
		{
			name:     "single tilde import",
			contents: "@~/docs/rules.md",
			want:     []string{"~/docs/rules.md"},
		},
		{
			name:     "multiple imports",
			contents: "@./a.md\nSome text\n@./b.md\n@./c.md",
			want:     []string{"./a.md", "./b.md", "./c.md"},
		},
		{
			name:     "import in code block skipped",
			contents: "```\n@./should-skip.md\n```\n@./should-include.md",
			want:     []string{"./should-include.md"},
		},
		{
			name:     "duplicate imports deduplicated",
			contents: "@./a.md\n@./a.md\n@./a.md",
			want:     []string{"./a.md"},
		},
		{
			name:     "inline backtick blocks match",
			contents: "See `code` and @./real-import.md",
			want:     nil, // backtick before @ prevents match (code span protection)
		},
		{
			name:     "import after text no backtick",
			contents: "See docs and @./real-import.md",
			want:     []string{"./real-import.md"},
		},
		{
			name:     "parent directory import",
			contents: "@../sibling/file.md",
			want:     []string{"../sibling/file.md"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractImports(tt.contents)

			if len(got) != len(tt.want) {
				t.Errorf("ExtractImports() = %v (len %d), want %v (len %d)",
					got, len(got), tt.want, len(tt.want))
				return
			}

			for i, w := range tt.want {
				if got[i] != w {
					t.Errorf("ExtractImports()[%d] = %q, want %q", i, got[i], w)
				}
			}
		})
	}
}

func TestImportGraphDetectCycles(t *testing.T) {
	tests := []struct {
		name       string
		edges      map[string][]string
		wantCycles int
		wantInMsg  []string // substrings expected in any cycle path
	}{
		{
			name:       "no imports - no cycles",
			edges:      map[string][]string{"/a.md": nil},
			wantCycles: 0,
		},
		{
			name: "linear chain - no cycle",
			edges: map[string][]string{
				"/a.md": {"/b.md"},
				"/b.md": {"/c.md"},
				"/c.md": nil,
			},
			wantCycles: 0,
		},
		{
			name: "simple A->B->A cycle",
			edges: map[string][]string{
				"/a.md": {"/b.md"},
				"/b.md": {"/a.md"},
			},
			wantCycles: 1,
			wantInMsg:  []string{"a.md", "b.md"},
		},
		{
			name: "longer A->B->C->A chain",
			edges: map[string][]string{
				"/a.md": {"/b.md"},
				"/b.md": {"/c.md"},
				"/c.md": {"/a.md"},
			},
			wantCycles: 1,
			wantInMsg:  []string{"a.md", "b.md", "c.md"},
		},
		{
			name: "self-import",
			edges: map[string][]string{
				"/self.md": {"/self.md"},
			},
			wantCycles: 1,
			wantInMsg:  []string{"self.md"},
		},
		{
			name: "diamond - no cycle",
			edges: map[string][]string{
				"/a.md": {"/b.md", "/c.md"},
				"/b.md": {"/d.md"},
				"/c.md": {"/d.md"},
				"/d.md": nil,
			},
			wantCycles: 0,
		},
		{
			name: "disconnected graph with one cycle",
			edges: map[string][]string{
				"/a.md": {"/b.md"},
				"/b.md": nil,
				"/x.md": {"/y.md"},
				"/y.md": {"/x.md"},
			},
			wantCycles: 1,
			wantInMsg:  []string{"x.md", "y.md"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &ImportGraph{edges: tt.edges}
			cycles := g.DetectCycles()

			if len(cycles) != tt.wantCycles {
				t.Errorf("DetectCycles() found %d cycles, want %d", len(cycles), tt.wantCycles)
				for i, c := range cycles {
					t.Logf("  cycle %d: %v", i, c)
				}
				return
			}

			if tt.wantCycles > 0 && len(tt.wantInMsg) > 0 {
				// Flatten all cycle nodes into one string for substring checks
				var allNodes strings.Builder
				for _, cycle := range cycles {
					for _, node := range cycle {
						allNodes.WriteString(filepath.Base(node))
						allNodes.WriteString(" ")
					}
				}
				combined := allNodes.String()

				for _, want := range tt.wantInMsg {
					if !strings.Contains(combined, want) {
						t.Errorf("DetectCycles() cycles missing expected node %q in %q",
							want, combined)
					}
				}
			}
		})
	}
}

func TestFormatImportCycle(t *testing.T) {
	tests := []struct {
		name  string
		cycle []string
		want  string
	}{
		{
			name:  "empty",
			cycle: nil,
			want:  "",
		},
		{
			name:  "self-cycle",
			cycle: []string{"/path/to/self.md", "/path/to/self.md"},
			want:  "self.md -> self.md",
		},
		{
			name:  "two-node cycle",
			cycle: []string{"/a.md", "/b.md", "/a.md"},
			want:  "a.md -> b.md -> a.md",
		},
		{
			name:  "three-node cycle",
			cycle: []string{"/x/a.md", "/y/b.md", "/z/c.md", "/x/a.md"},
			want:  "a.md -> b.md -> c.md -> a.md",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatImportCycle(tt.cycle)
			if got != tt.want {
				t.Errorf("FormatImportCycle() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDetectImportCycles_Integration(t *testing.T) {
	tests := []struct {
		name       string
		files      map[string]string
		wantErrors int
		wantInMsg  string
	}{
		{
			name: "no cycles",
			files: map[string]string{
				"/project/a.md": "@./b.md",
				"/project/b.md": "No imports here",
			},
			wantErrors: 0,
		},
		{
			name: "mutual import cycle",
			files: map[string]string{
				"/project/a.md": "@./b.md",
				"/project/b.md": "@./a.md",
			},
			wantErrors: 1,
			wantInMsg:  "Circular @import detected",
		},
		{
			name: "self-import",
			files: map[string]string{
				"/project/self.md": "@./self.md",
			},
			wantErrors: 1,
			wantInMsg:  "Circular @import detected",
		},
		{
			name: "three-file cycle",
			files: map[string]string{
				"/project/a.md": "@./b.md",
				"/project/b.md": "@./c.md",
				"/project/c.md": "@./a.md",
			},
			wantErrors: 1,
			wantInMsg:  "Circular @import detected",
		},
		{
			name: "no files",
			files: map[string]string{},
			wantErrors: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := DetectImportCycles(tt.files)

			if len(errors) != tt.wantErrors {
				t.Errorf("DetectImportCycles() errors = %d, want %d", len(errors), tt.wantErrors)
				for _, e := range errors {
					t.Logf("  %s: %s", e.Severity, e.Message)
				}
				return
			}

			if tt.wantInMsg != "" && len(errors) > 0 {
				found := false
				for _, e := range errors {
					if strings.Contains(e.Message, tt.wantInMsg) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("DetectImportCycles() expected message containing %q", tt.wantInMsg)
				}
			}
		})
	}
}

func TestResolveImportPath(t *testing.T) {
	tests := []struct {
		name       string
		importPath string
		baseDir    string
		wantSuffix string // suffix of the resolved path
	}{
		{
			name:       "relative path",
			importPath: "./other.md",
			baseDir:    "/project",
			wantSuffix: "/project/other.md",
		},
		{
			name:       "parent directory",
			importPath: "../sibling/file.md",
			baseDir:    "/project/sub",
			wantSuffix: "/project/sibling/file.md",
		},
		{
			name:       "absolute path",
			importPath: "/absolute/path.md",
			baseDir:    "/project",
			wantSuffix: "/absolute/path.md",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := resolveImportPath(tt.importPath, tt.baseDir)
			if got != tt.wantSuffix {
				t.Errorf("resolveImportPath(%q, %q) = %q, want %q",
					tt.importPath, tt.baseDir, got, tt.wantSuffix)
			}
		})
	}
}
