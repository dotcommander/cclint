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
	changes, err := GetStagedChanges(rootPath)
	return changes.Files, err
}

// FileChanges contains lintable changed files and deletion metadata that can
// require project-wide validation even when no changed file remains to lint.
type FileChanges struct {
	Files        []string
	DeletedFiles []string
}

// HasRelevantDeletion reports whether the selected Git tree deletes at least
// one Claude Code component.
func (c FileChanges) HasRelevantDeletion() bool { return len(c.DeletedFiles) > 0 }

// GetStagedChanges returns staged lintable files and whether a relevant
// component was deleted.
func GetStagedChanges(rootPath string) (FileChanges, error) {
	if !IsGitRepo(rootPath) {
		return FileChanges{}, nil
	}
	var err error
	rootPath, err = RepositoryRoot(rootPath)
	if err != nil {
		return FileChanges{}, err
	}

	// Get staged files relative to git root
	cmd, cancel := gitCommand(rootPath, "diff", "--name-status", "-z", "--no-renames", "--staged")
	defer cancel()
	output, err := cmd.CombinedOutput()
	if err != nil {
		return FileChanges{}, gitTimeoutError("diff --staged", err, output)
	}

	return parseNameStatusChanges(output, rootPath)
}

// GetChangedFiles returns absolute paths of all uncommitted changes (staged + unstaged).
// Only returns files with extensions matching Claude Code components (.md, .json).
// Returns empty slice if not in a git repository.
func GetChangedFiles(rootPath string) ([]string, error) {
	changes, err := GetChangedFileChanges(rootPath)
	return changes.Files, err
}

// GetChangedFileChanges returns all uncommitted lintable files and whether a
// relevant tracked component was deleted.
func GetChangedFileChanges(rootPath string) (FileChanges, error) {
	if !IsGitRepo(rootPath) {
		return FileChanges{}, nil
	}
	var err error
	rootPath, err = RepositoryRoot(rootPath)
	if err != nil {
		return FileChanges{}, err
	}

	// Check if there are any commits
	checkCmd, cancelCheck := gitCommand(rootPath, "rev-parse", "HEAD")
	checkErr := checkCmd.Run()
	cancelCheck()
	if checkErr != nil {
		if errors.Is(checkErr, context.DeadlineExceeded) {
			return FileChanges{}, gitTimeoutError("rev-parse HEAD", checkErr, nil)
		}
		// No commits yet: preserve staged addition status against the empty tree,
		// then add untracked files separately. Worktree absence is not deletion.
		cmd, cancel := gitCommand(rootPath, "diff", "--cached", "--name-status", "-z", "--no-renames")
		defer cancel()
		output, err := cmd.CombinedOutput()
		if err != nil {
			return FileChanges{}, gitTimeoutError("diff --cached", err, output)
		}
		changes, err := parseNameStatusChanges(output, rootPath)
		if err != nil {
			return FileChanges{}, err
		}
		untracked, err := getUntrackedFiles(rootPath)
		if err != nil {
			return FileChanges{}, err
		}
		untrackedFiles, err := filterRelevantFiles(untracked, rootPath)
		if err != nil {
			return FileChanges{}, err
		}
		changes.Files = append(changes.Files, untrackedFiles...)
		return changes, nil
	}

	// Get all changed files (staged + unstaged) relative to git root
	cmd, cancel := gitCommand(rootPath, "diff", "--name-status", "-z", "--no-renames", "HEAD")
	defer cancel()
	output, err := cmd.CombinedOutput()
	if err != nil {
		return FileChanges{}, gitTimeoutError("diff HEAD", err, output)
	}

	untracked, err := getUntrackedFiles(rootPath)
	if err != nil {
		return FileChanges{}, err
	}

	changes, err := parseNameStatusChanges(output, rootPath)
	if err != nil {
		return FileChanges{}, err
	}
	untrackedFiles, err := filterRelevantFiles(untracked, rootPath)
	if err != nil {
		return FileChanges{}, err
	}
	changes.Files = append(changes.Files, untrackedFiles...)
	return changes, nil
}

// RepositoryRoot resolves the working-tree top level from any descendant.
func RepositoryRoot(path string) (string, error) {
	cmd, cancel := gitCommand(path, "rev-parse", "--show-toplevel")
	defer cancel()
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", gitTimeoutError("rev-parse --show-toplevel", err, output)
	}
	return filepath.Clean(strings.TrimSpace(string(output))), nil
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
	changes, err := filterRelevantChanges(gitOutput, rootPath)
	return changes.Files, err
}

func filterRelevantChanges(gitOutput, rootPath string) (FileChanges, error) {
	var files []string
	var deletedFiles []string
	lines := strings.Split(strings.TrimSpace(gitOutput), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if !isRelevantFile(line) {
			continue
		}

		// Convert to absolute path
		absPath := filepath.Join(rootPath, line)

		// Check if file exists (git reports deletions too)
		if _, err := os.Stat(absPath); os.IsNotExist(err) {
			deletedFiles = append(deletedFiles, filepath.ToSlash(line))
			continue
		}

		files = append(files, absPath)
	}

	return FileChanges{Files: files, DeletedFiles: deletedFiles}, nil
}

func parseNameStatusChanges(output []byte, rootPath string) (FileChanges, error) {
	var changes FileChanges
	fields := strings.Split(strings.TrimSuffix(string(output), "\x00"), "\x00")
	if len(fields) == 1 && fields[0] == "" {
		return changes, nil
	}
	if len(fields)%2 != 0 {
		return FileChanges{}, fmt.Errorf("malformed git name-status output")
	}
	for i := 0; i < len(fields); i += 2 {
		status, relPath := fields[i], fields[i+1]
		if status == "" || relPath == "" {
			return FileChanges{}, fmt.Errorf("malformed git name-status entry")
		}
		if !isRelevantFile(relPath) {
			continue
		}
		if status[0] == 'D' {
			changes.DeletedFiles = append(changes.DeletedFiles, filepath.ToSlash(relPath))
			continue
		}
		absPath := filepath.Join(rootPath, relPath)
		if _, err := os.Stat(absPath); err == nil {
			changes.Files = append(changes.Files, absPath)
		} else if !os.IsNotExist(err) {
			return FileChanges{}, fmt.Errorf("stat changed file %s: %w", relPath, err)
		}
	}
	return changes, nil
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
		case "agents", "commands", "skills", "rules", "output-styles", ".claude", ".claude-plugin":
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
