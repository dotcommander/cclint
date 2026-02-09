package project

import (
	"os"
	"path/filepath"
)

// Info contains information about the detected project.
// Named 'Info' instead of 'ProjectInfo' to avoid stuttering (project.Info vs project.ProjectInfo).
type Info struct {
	Root       string
	IsClaude   bool
	HasGit     bool
	Type       string
	FilesFound []string
}

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

// isProjectRoot determines if a directory is a project root
func isProjectRoot(path string) bool {
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

// Detect detects project information at the given path.
// Named 'Detect' instead of 'DetectProjectInfo' to avoid stuttering.
func Detect(rootPath string) (*Info, error) {
	projectInfo := &Info{
		Root:     rootPath,
		IsClaude: false,
		HasGit:   false,
		Type:     "unknown",
	}

	// Check for specific project markers
	if _, err := os.Stat(filepath.Join(rootPath, ".claude")); err == nil {
		projectInfo.IsClaude = true
		projectInfo.Type = "claude"
	}

	if _, err := os.Stat(filepath.Join(rootPath, ".git")); err == nil {
		projectInfo.HasGit = true
	}

	// Detect project type
	if _, err := os.Stat(filepath.Join(rootPath, "package.json")); err == nil {
		projectInfo.Type = "node"
	}

	if _, err := os.Stat(filepath.Join(rootPath, "go.mod")); err == nil {
		projectInfo.Type = "go"
	}

	// Find relevant files
	projectInfo.FilesFound = findProjectFiles(rootPath)

	return projectInfo, nil
}

// findProjectFiles scans the project for relevant files
func findProjectFiles(rootPath string) []string {
	var files []string

	// This will be enhanced in the file discovery phase
	// For now, just find some basic markers
	if _, err := os.Stat(filepath.Join(rootPath, ".claude")); err == nil {
		files = append(files, ".claude/")
	}

	if _, err := os.Stat(filepath.Join(rootPath, ".git")); err == nil {
		files = append(files, ".git/")
	}

	return files
}
