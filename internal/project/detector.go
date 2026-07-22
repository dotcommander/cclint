package project

import (
	"os"
	"path/filepath"
)

// FindProjectRoot searches for a project root starting from the given path
// and climbing up the directory tree if needed.
func FindProjectRoot(startPath string) (string, error) {
	// Convert to absolute path
	absPath, err := filepath.Abs(startPath)
	if err != nil {
		return "", err
	}

	// If no path provided, start from current directory
	if absPath == "." {
		absPath, _ = os.Getwd()
	}

	// Climb up the directory tree
	currentDir := absPath
	for {
		// Check if current directory is a project root
		if isProjectRoot(currentDir) {
			return currentDir, nil
		}

		// Go up one level
		parent := filepath.Dir(currentDir)
		if parent == currentDir {
			// Reached filesystem root
			break
		}
		currentDir = parent
	}

	// Default to current directory if no project root found
	return absPath, nil
}

// isProjectRoot determines if a directory is a project root.
//
// Claude Code plugins (directories containing .claude-plugin/plugin.json) are
// checked first so FindProjectRoot stops at the innermost plugin directory
// before climbing out to a surrounding git repo. This allows cclint to resolve
// component paths correctly when a plugin is nested inside a larger repo, e.g.
// plugins/dc/{agents,commands,skills}/... inside a repo with .git at the top.
func isProjectRoot(path string) bool {
	// Check for .claude-plugin/plugin.json (Claude Code plugin root).
	// Checked before .git so plugin dirs nested inside a git repo are found first.
	if _, err := os.Stat(filepath.Join(path, ".claude-plugin", "plugin.json")); err == nil {
		return true
	}

	// Check for .git directory
	if _, err := os.Stat(filepath.Join(path, ".git")); err == nil {
		return true
	}

	// Check for .claude directory
	if _, err := os.Stat(filepath.Join(path, ".claude")); err == nil {
		return true
	}

	// Check for package.json (for JS/TS projects)
	if _, err := os.Stat(filepath.Join(path, "package.json")); err == nil {
		return true
	}

	// Check for go.mod (for Go projects)
	if _, err := os.Stat(filepath.Join(path, "go.mod")); err == nil {
		return true
	}

	return false
}
