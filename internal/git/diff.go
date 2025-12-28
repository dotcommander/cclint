package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// GetStagedFiles returns absolute paths of files in git staging area.
// Only returns files with extensions matching Claude Code components (.md, .json).
// Returns empty slice if not in a git repository.
func GetStagedFiles(rootPath string) ([]string, error) {
	if !IsGitRepo(rootPath) {
		return []string{}, nil
	}

	// Get staged files relative to git root
	cmd := exec.Command("git", "diff", "--name-only", "--staged")
	cmd.Dir = rootPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("git diff --staged failed: %w: %s", err, output)
	}

	return filterRelevantFiles(string(output), rootPath)
}

// GetChangedFiles returns absolute paths of all uncommitted changes (staged + unstaged).
// Only returns files with extensions matching Claude Code components (.md, .json).
// Returns empty slice if not in a git repository.
func GetChangedFiles(rootPath string) ([]string, error) {
	if !IsGitRepo(rootPath) {
		return []string{}, nil
	}

	// Check if there are any commits
	checkCmd := exec.Command("git", "rev-parse", "HEAD")
	checkCmd.Dir = rootPath
	if err := checkCmd.Run(); err != nil {
		// No commits yet - show all tracked files
		cmd := exec.Command("git", "ls-files")
		cmd.Dir = rootPath
		output, err := cmd.CombinedOutput()
		if err != nil {
			return nil, fmt.Errorf("git ls-files failed: %w: %s", err, output)
		}
		return filterRelevantFiles(string(output), rootPath)
	}

	// Get all changed files (staged + unstaged) relative to git root
	cmd := exec.Command("git", "diff", "--name-only", "HEAD")
	cmd.Dir = rootPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("git diff HEAD failed: %w: %s", err, output)
	}

	return filterRelevantFiles(string(output), rootPath)
}

// IsGitRepo checks if the given directory is within a git repository.
func IsGitRepo(rootPath string) bool {
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	cmd.Dir = rootPath
	cmd.Stderr = nil // Suppress error output
	err := cmd.Run()
	return err == nil
}

// filterRelevantFiles filters git diff output to only include Claude Code component files.
// Filters by:
//   - Extension: .md or .json
//   - Path patterns: agents/, commands/, skills/, .claude/, or specific filenames
//
// Returns absolute paths.
func filterRelevantFiles(gitOutput, rootPath string) ([]string, error) {
	var files []string
	lines := strings.Split(strings.TrimSpace(gitOutput), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Convert to absolute path
		absPath := filepath.Join(rootPath, line)

		// Check if file exists (git reports deletions too)
		if _, err := os.Stat(absPath); os.IsNotExist(err) {
			continue
		}

		// Filter by extension and path
		if !isRelevantFile(line) {
			continue
		}

		files = append(files, absPath)
	}

	return files, nil
}

// isRelevantFile checks if a file path is relevant for Claude Code linting.
// Matches files in standard directories or with special filenames.
func isRelevantFile(relPath string) bool {
	lowerPath := strings.ToLower(relPath)

	// Extension check
	ext := filepath.Ext(lowerPath)
	if ext != ".md" && ext != ".json" {
		return false
	}

	// Path-based filtering
	pathComponents := strings.Split(filepath.ToSlash(relPath), "/")

	for _, component := range pathComponents {
		switch component {
		case "agents", "commands", "skills", ".claude", ".claude-plugin":
			return true
		}
	}

	// Special filenames (case-insensitive)
	basename := filepath.Base(relPath)
	switch {
	case strings.EqualFold(basename, "SKILL.md"):
		return true
	case strings.EqualFold(basename, "CLAUDE.md"):
		return true
	case basename == "settings.json":
		return true
	case basename == "plugin.json":
		return true
	}

	return false
}
