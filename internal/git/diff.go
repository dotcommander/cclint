package git

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// defaultGitTimeout bounds every git probe invoked by this package. Pre-commit
// style workflows must fail promptly rather than hang on a stalled git index
// or unreachable upstream. Overridable from tests via SetGitTimeout.
var defaultGitTimeout = 10 * time.Second

// SetGitTimeout overrides the deadline applied to git probes. Intended for
// tests that need to force a fast timeout or extend the default. Returns the
// previous value so callers can restore it via defer.
func SetGitTimeout(d time.Duration) time.Duration {
	prev := defaultGitTimeout
	defaultGitTimeout = d
	return prev
}

// gitCommand returns an exec.Cmd whose lifetime is bounded by defaultGitTimeout
// and whose working directory is rootPath. Centralizes deadline + cwd wiring so
// every git probe in this package picks up the same policy.
func gitCommand(rootPath string, args ...string) (*exec.Cmd, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultGitTimeout)
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = rootPath
	return cmd, cancel
}

// gitTimeoutError annotates an exec error with the configured deadline when
// the context was cancelled, so callers get an actionable message instead of
// the opaque "signal: killed".
func gitTimeoutError(op string, err error, output []byte) error {
	if errors.Is(err, context.DeadlineExceeded) {
		return fmt.Errorf("git %s timed out after %s", op, defaultGitTimeout)
	}
	if len(output) > 0 {
		return fmt.Errorf("git %s failed: %w: %s", op, err, output)
	}
	return fmt.Errorf("git %s failed: %w", op, err)
}

// GetStagedFiles returns absolute paths of files in git staging area.
// Only returns files with extensions matching Claude Code components (.md, .json).
// Returns empty slice if not in a git repository.
func GetStagedFiles(rootPath string) ([]string, error) {
	if !IsGitRepo(rootPath) {
		return []string{}, nil
	}

	// Get staged files relative to git root
	cmd, cancel := gitCommand(rootPath, "diff", "--name-only", "--staged")
	defer cancel()
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, gitTimeoutError("diff --staged", err, output)
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
	checkCmd, cancelCheck := gitCommand(rootPath, "rev-parse", "HEAD")
	checkErr := checkCmd.Run()
	cancelCheck()
	if checkErr != nil {
		if errors.Is(checkErr, context.DeadlineExceeded) {
			return nil, gitTimeoutError("rev-parse HEAD", checkErr, nil)
		}
		// No commits yet - show all tracked and untracked files.
		cmd, cancel := gitCommand(rootPath, "ls-files", "--cached", "--others", "--exclude-standard")
		defer cancel()
		output, err := cmd.CombinedOutput()
		if err != nil {
			return nil, gitTimeoutError("ls-files", err, output)
		}
		return filterRelevantFiles(string(output), rootPath)
	}

	// Get all changed files (staged + unstaged) relative to git root
	cmd, cancel := gitCommand(rootPath, "diff", "--name-only", "HEAD")
	defer cancel()
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, gitTimeoutError("diff HEAD", err, output)
	}

	untracked, err := getUntrackedFiles(rootPath)
	if err != nil {
		return nil, err
	}

	return filterRelevantFiles(combineGitOutputs(string(output), untracked), rootPath)
}

// IsGitRepo checks if the given directory is within a git repository.
func IsGitRepo(rootPath string) bool {
	cmd, cancel := gitCommand(rootPath, "rev-parse", "--git-dir")
	defer cancel()
	cmd.Stderr = nil // Suppress error output
	err := cmd.Run()
	return err == nil
}

func getUntrackedFiles(rootPath string) (string, error) {
	cmd, cancel := gitCommand(rootPath, "ls-files", "--others", "--exclude-standard")
	defer cancel()
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", gitTimeoutError("ls-files --others", err, output)
	}
	return string(output), nil
}

func combineGitOutputs(outputs ...string) string {
	seen := make(map[string]bool)
	var combined []string
	for _, output := range outputs {
		for _, line := range strings.Split(strings.TrimSpace(output), "\n") {
			line = strings.TrimSpace(line)
			if line == "" || seen[line] {
				continue
			}
			seen[line] = true
			combined = append(combined, line)
		}
	}
	return strings.Join(combined, "\n")
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
