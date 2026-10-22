package cmd

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestVersionLineIncludesSurface(t *testing.T) {
	t.Parallel()
	got := versionLine()
	if !strings.Contains(got, "cclint version "+Version) || !strings.Contains(got, "(claude-code surface: "+ClaudeCodeSurface+")") {
		t.Fatalf("versionLine() = %q, want cclint version prefix and claude-code surface %s", got, ClaudeCodeSurface)
	}
}

func TestClaudeCodeSurfaceMatchesAGENTSRow(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile(filepath.Join("..", "AGENTS.md"))
	if err != nil {
		t.Fatalf("read AGENTS.md: %v", err)
	}
	row := regexp.MustCompile(`claude_code_last_updated\s*\|\s*(v[0-9.]+)`).FindSubmatch(data)
	if row == nil {
		t.Fatal("claude_code_last_updated row not found in AGENTS.md")
	}
	if string(row[1]) != ClaudeCodeSurface {
		t.Fatalf("cmd.ClaudeCodeSurface = %s but AGENTS.md claude_code_last_updated = %s; bump both together", ClaudeCodeSurface, row[1])
	}
}
