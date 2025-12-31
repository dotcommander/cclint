package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestIsRelevantFile(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		// Agent files
		{"agent in .claude", ".claude/agents/my-agent.md", true},
		{"agent in agents", "agents/my-agent.md", true},
		{"agent nested", "agents/subdir/my-agent.md", true},

		// Command files
		{"command in .claude", ".claude/commands/my-cmd.md", true},
		{"command in commands", "commands/my-cmd.md", true},

		// Skill files
		{"skill file", ".claude/skills/my-skill/SKILL.md", true},
		{"skill file alt", "skills/my-skill/SKILL.md", true},

		// Special files
		{"CLAUDE.md root", "CLAUDE.md", true},
		{"CLAUDE.md in .claude", ".claude/CLAUDE.md", true},
		{"settings.json", ".claude/settings.json", true},
		{"plugin.json", ".claude-plugin/plugin.json", true},

		// Irrelevant files
		{"go source", "main.go", false},
		{"random md", "README.md", false},
		{"random json", "package.json", false},
		{"docs", "docs/guide.md", false},
		{"test file", "test/agent_test.md", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isRelevantFile(tt.path)
			if result != tt.expected {
				t.Errorf("isRelevantFile(%q) = %v, want %v", tt.path, result, tt.expected)
			}
		})
	}
}

func TestFilterRelevantFiles(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()

	// Create test files
	testFiles := map[string]bool{
		".claude/agents/test-agent.md": true,
		"commands/test-cmd.md":         true,
		"CLAUDE.md":                    true,
		"README.md":                    false, // Should be filtered out
		"main.go":                      false, // Should be filtered out
	}

	for path := range testFiles {
		fullPath := filepath.Join(tmpDir, path)
		dir := filepath.Dir(fullPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("failed to create directory %s: %v", dir, err)
		}
		if err := os.WriteFile(fullPath, []byte("test"), 0644); err != nil {
			t.Fatalf("failed to create file %s: %v", fullPath, err)
		}
	}

	// Simulate git diff output
	gitOutput := `.claude/agents/test-agent.md
commands/test-cmd.md
CLAUDE.md
README.md
main.go`

	// Filter files
	filtered, err := filterRelevantFiles(gitOutput, tmpDir)
	if err != nil {
		t.Fatalf("filterRelevantFiles failed: %v", err)
	}

	// Count expected vs actual
	expectedCount := 0
	for _, shouldInclude := range testFiles {
		if shouldInclude {
			expectedCount++
		}
	}

	if len(filtered) != expectedCount {
		t.Errorf("got %d files, want %d files", len(filtered), expectedCount)
	}

	// Verify each filtered file is expected
	for _, absPath := range filtered {
		relPath, err := filepath.Rel(tmpDir, absPath)
		if err != nil {
			t.Errorf("failed to compute relative path: %v", err)
			continue
		}

		shouldInclude, exists := testFiles[relPath]
		if !exists {
			t.Errorf("unexpected file in results: %s", relPath)
		} else if !shouldInclude {
			t.Errorf("file should have been filtered out: %s", relPath)
		}
	}
}

func TestIsGitRepo(t *testing.T) {
	// Test with current directory (should be a git repo)
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current directory: %v", err)
	}

	// Navigate up to find git root
	gitRoot := cwd
	for {
		if _, err := os.Stat(filepath.Join(gitRoot, ".git")); err == nil {
			break
		}
		parent := filepath.Dir(gitRoot)
		if parent == gitRoot {
			t.Skip("not in a git repository, skipping git tests")
			return
		}
		gitRoot = parent
	}

	if !IsGitRepo(gitRoot) {
		t.Error("IsGitRepo should return true for git repository root")
	}

	// Test with non-git directory
	tmpDir := t.TempDir()
	if IsGitRepo(tmpDir) {
		t.Error("IsGitRepo should return false for non-git directory")
	}
}

func TestGetStagedFiles(t *testing.T) {
	// Skip if not in git repo
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current directory: %v", err)
	}

	if !IsGitRepo(cwd) {
		t.Skip("not in a git repository, skipping git tests")
		return
	}

	// This test verifies the function works without error
	// Actual staged files depend on current git state
	files, err := GetStagedFiles(cwd)
	if err != nil {
		t.Errorf("GetStagedFiles failed: %v", err)
	}

	// Verify all returned files exist
	for _, file := range files {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			t.Errorf("GetStagedFiles returned non-existent file: %s", file)
		}
	}
}

func TestGetChangedFiles(t *testing.T) {
	// Skip if not in git repo
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current directory: %v", err)
	}

	if !IsGitRepo(cwd) {
		t.Skip("not in a git repository, skipping git tests")
		return
	}

	// This test verifies the function works without error
	files, err := GetChangedFiles(cwd)
	if err != nil {
		t.Errorf("GetChangedFiles failed: %v", err)
	}

	// Verify all returned files exist
	for _, file := range files {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			t.Errorf("GetChangedFiles returned non-existent file: %s", file)
		}
	}
}

func TestGitIntegration(t *testing.T) {
	// Create a temporary git repository for testing
	tmpDir := t.TempDir()

	// Initialize git repo
	cmd := exec.Command("git", "init")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Skip("git not available, skipping integration test")
		return
	}

	// Configure git user (required for commits)
	_ = exec.Command("git", "config", "user.email", "test@test.com").Run()
	_ = exec.Command("git", "config", "user.name", "Test User").Run()

	// Create test files
	agentPath := filepath.Join(tmpDir, "agents", "test-agent.md")
	if err := os.MkdirAll(filepath.Dir(agentPath), 0755); err != nil {
		t.Fatalf("failed to create directory: %v", err)
	}
	if err := os.WriteFile(agentPath, []byte("# Test Agent"), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	readmePath := filepath.Join(tmpDir, "README.md")
	if err := os.WriteFile(readmePath, []byte("# README"), 0644); err != nil {
		t.Fatalf("failed to create README: %v", err)
	}

	// Add files to staging
	cmd = exec.Command("git", "add", ".")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("git add failed: %v", err)
	}

	// Get staged files
	staged, err := GetStagedFiles(tmpDir)
	if err != nil {
		t.Fatalf("GetStagedFiles failed: %v", err)
	}

	// Should only include agent file, not README
	if len(staged) != 1 {
		t.Errorf("expected 1 staged file, got %d", len(staged))
	}

	if len(staged) > 0 && !strings.Contains(staged[0], "test-agent.md") {
		t.Errorf("expected test-agent.md in staged files, got %s", staged[0])
	}
}

func TestGetStagedFiles_NonGitRepo(t *testing.T) {
	// Test with non-git directory
	tmpDir := t.TempDir()
	files, err := GetStagedFiles(tmpDir)
	if err != nil {
		t.Errorf("GetStagedFiles should not error for non-git repo: %v", err)
	}
	if len(files) != 0 {
		t.Errorf("expected empty slice for non-git repo, got %d files", len(files))
	}
}

func TestGetChangedFiles_NonGitRepo(t *testing.T) {
	// Test with non-git directory
	tmpDir := t.TempDir()
	files, err := GetChangedFiles(tmpDir)
	if err != nil {
		t.Errorf("GetChangedFiles should not error for non-git repo: %v", err)
	}
	if len(files) != 0 {
		t.Errorf("expected empty slice for non-git repo, got %d files", len(files))
	}
}

func TestGetChangedFiles_NoCommits(t *testing.T) {
	// Create a temporary git repository with no commits
	tmpDir := t.TempDir()

	// Initialize git repo
	cmd := exec.Command("git", "init")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Skip("git not available, skipping integration test")
		return
	}

	// Configure git
	configCmd := exec.Command("git", "config", "user.email", "test@test.com")
	configCmd.Dir = tmpDir
	_ = configCmd.Run()
	configCmd = exec.Command("git", "config", "user.name", "Test User")
	configCmd.Dir = tmpDir
	_ = configCmd.Run()

	// Create and add files (but don't commit)
	agentPath := filepath.Join(tmpDir, "agents", "test-agent.md")
	if err := os.MkdirAll(filepath.Dir(agentPath), 0755); err != nil {
		t.Fatalf("failed to create directory: %v", err)
	}
	if err := os.WriteFile(agentPath, []byte("# Test Agent"), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	cmdPath := filepath.Join(tmpDir, "commands", "test-cmd.md")
	if err := os.MkdirAll(filepath.Dir(cmdPath), 0755); err != nil {
		t.Fatalf("failed to create directory: %v", err)
	}
	if err := os.WriteFile(cmdPath, []byte("# Test Command"), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	readmePath := filepath.Join(tmpDir, "README.md")
	if err := os.WriteFile(readmePath, []byte("# README"), 0644); err != nil {
		t.Fatalf("failed to create README: %v", err)
	}

	// Add files to git (but don't commit)
	addCmd := exec.Command("git", "add", ".")
	addCmd.Dir = tmpDir
	if err := addCmd.Run(); err != nil {
		t.Fatalf("git add failed: %v", err)
	}

	// Get changed files (should use git ls-files since no commits exist)
	files, err := GetChangedFiles(tmpDir)
	if err != nil {
		t.Fatalf("GetChangedFiles failed: %v", err)
	}

	// Should return tracked files, filtered to relevant ones (agent and command, not README)
	if len(files) != 2 {
		t.Errorf("expected 2 files, got %d", len(files))
	}

	// Verify returned files are relevant
	foundAgent := false
	foundCmd := false
	for _, file := range files {
		if strings.Contains(file, "test-agent.md") {
			foundAgent = true
		}
		if strings.Contains(file, "test-cmd.md") {
			foundCmd = true
		}
		if strings.Contains(file, "README.md") {
			t.Errorf("README.md should be filtered out, but was included")
		}
	}

	if !foundAgent {
		t.Error("expected test-agent.md in changed files")
	}
	if !foundCmd {
		t.Error("expected test-cmd.md in changed files")
	}
}

func TestGetChangedFiles_WithCommits(t *testing.T) {
	// Create a temporary git repository with commits
	tmpDir := t.TempDir()

	// Initialize git repo
	cmd := exec.Command("git", "init")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Skip("git not available, skipping integration test")
		return
	}

	// Configure git
	configCmd := exec.Command("git", "config", "user.email", "test@test.com")
	configCmd.Dir = tmpDir
	_ = configCmd.Run()
	configCmd = exec.Command("git", "config", "user.name", "Test User")
	configCmd.Dir = tmpDir
	_ = configCmd.Run()

	// Create initial file and commit
	initialPath := filepath.Join(tmpDir, ".gitkeep")
	if err := os.WriteFile(initialPath, []byte(""), 0644); err != nil {
		t.Fatalf("failed to create initial file: %v", err)
	}

	addCmd := exec.Command("git", "add", ".")
	addCmd.Dir = tmpDir
	if err := addCmd.Run(); err != nil {
		t.Fatalf("git add failed: %v", err)
	}

	commitCmd := exec.Command("git", "commit", "-m", "Initial commit")
	commitCmd.Dir = tmpDir
	if err := commitCmd.Run(); err != nil {
		t.Fatalf("git commit failed: %v", err)
	}

	// Create new files (not committed)
	agentPath := filepath.Join(tmpDir, "agents", "test-agent.md")
	if err := os.MkdirAll(filepath.Dir(agentPath), 0755); err != nil {
		t.Fatalf("failed to create directory: %v", err)
	}
	if err := os.WriteFile(agentPath, []byte("# Test Agent"), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	readmePath := filepath.Join(tmpDir, "README.md")
	if err := os.WriteFile(readmePath, []byte("# README"), 0644); err != nil {
		t.Fatalf("failed to create README: %v", err)
	}

	// Add to git (staged but not committed)
	addCmd = exec.Command("git", "add", ".")
	addCmd.Dir = tmpDir
	if err := addCmd.Run(); err != nil {
		t.Fatalf("git add failed: %v", err)
	}

	// Get changed files (should use git diff HEAD)
	files, err := GetChangedFiles(tmpDir)
	if err != nil {
		t.Fatalf("GetChangedFiles failed: %v", err)
	}

	// Should only include agent file, not README
	if len(files) != 1 {
		t.Errorf("expected 1 file, got %d", len(files))
	}

	if len(files) > 0 && !strings.Contains(files[0], "test-agent.md") {
		t.Errorf("expected test-agent.md in changed files, got %s", files[0])
	}
}

func TestIsRelevantFile_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		// Case sensitivity tests
		{"SKILL.md uppercase", "skills/my-skill/SKILL.MD", true},
		{"skill.md lowercase", "skills/my-skill/skill.md", true},
		{"CLAUDE.MD uppercase", "CLAUDE.MD", true},
		{"claude.md lowercase", "claude.md", true},

		// Non-.md/.json extensions
		{"txt file in agents", "agents/test.txt", false},
		{"yaml file in .claude", ".claude/config.yaml", false},
		{"no extension", "agents/file", false},

		// Edge path patterns
		{"nested .claude", "foo/.claude/agents/test.md", true},
		{"nested agents", "foo/agents/test.md", true},
		{"plugin.json in root", "plugin.json", true},
		// plugin.json matches special filename regardless of location
		{"plugin.json in subdirectory", "src/plugin.json", true},

		// Windows-style paths don't work (filepath.ToSlash not used for splitting)
		{"windows path agents", "agents\\test-agent.md", false},
		{"windows path .claude", ".claude\\commands\\test.md", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isRelevantFile(tt.path)
			if result != tt.expected {
				t.Errorf("isRelevantFile(%q) = %v, want %v", tt.path, result, tt.expected)
			}
		})
	}
}

func TestFilterRelevantFiles_EmptyInput(t *testing.T) {
	tmpDir := t.TempDir()

	// Test with empty git output
	filtered, err := filterRelevantFiles("", tmpDir)
	if err != nil {
		t.Fatalf("filterRelevantFiles failed: %v", err)
	}

	if len(filtered) != 0 {
		t.Errorf("expected 0 files for empty input, got %d", len(filtered))
	}
}

func TestFilterRelevantFiles_DeletedFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Git output includes a deleted file
	gitOutput := "agents/deleted-agent.md\ncommands/existing-cmd.md"

	// Only create the existing file
	cmdPath := filepath.Join(tmpDir, "commands", "existing-cmd.md")
	if err := os.MkdirAll(filepath.Dir(cmdPath), 0755); err != nil {
		t.Fatalf("failed to create directory: %v", err)
	}
	if err := os.WriteFile(cmdPath, []byte("# Test"), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	// Filter files
	filtered, err := filterRelevantFiles(gitOutput, tmpDir)
	if err != nil {
		t.Fatalf("filterRelevantFiles failed: %v", err)
	}

	// Should only include existing file
	if len(filtered) != 1 {
		t.Errorf("expected 1 file (deleted file should be skipped), got %d", len(filtered))
	}

	if len(filtered) > 0 && !strings.Contains(filtered[0], "existing-cmd.md") {
		t.Errorf("expected existing-cmd.md, got %s", filtered[0])
	}
}

func TestFilterRelevantFiles_WhitespaceHandling(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test file
	agentPath := filepath.Join(tmpDir, "agents", "test.md")
	if err := os.MkdirAll(filepath.Dir(agentPath), 0755); err != nil {
		t.Fatalf("failed to create directory: %v", err)
	}
	if err := os.WriteFile(agentPath, []byte("# Test"), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	// Git output with extra whitespace
	gitOutput := "\n  agents/test.md  \n\n  \n"

	// Filter files
	filtered, err := filterRelevantFiles(gitOutput, tmpDir)
	if err != nil {
		t.Fatalf("filterRelevantFiles failed: %v", err)
	}

	if len(filtered) != 1 {
		t.Errorf("expected 1 file, got %d", len(filtered))
	}
}
