package cli

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dotcommander/cclint/internal/cue"
	"github.com/dotcommander/cclint/internal/discovery"
	"github.com/dotcommander/cclint/internal/frontend"
)

const (
	// CombinedMemorySizeWarningKB is the threshold for warning about combined memory size
	CombinedMemorySizeWarningKB = 20
)

// CheckClaudeLocalGitignore checks if CLAUDE.local.md exists and is in .gitignore
func CheckClaudeLocalGitignore(rootPath string) []cue.ValidationError {
	var errors []cue.ValidationError

	localMdPath := filepath.Join(rootPath, "CLAUDE.local.md")
	if _, err := os.Stat(localMdPath); os.IsNotExist(err) {
		// CLAUDE.local.md doesn't exist, nothing to check
		return nil
	}

	// Check if it's in .gitignore
	gitignorePath := filepath.Join(rootPath, ".gitignore")
	if _, err := os.Stat(gitignorePath); os.IsNotExist(err) {
		// No .gitignore exists
		errors = append(errors, cue.ValidationError{
			File:     "CLAUDE.local.md",
			Message:  "CLAUDE.local.md exists but no .gitignore found - this file should not be committed to version control",
			Severity: "warning",
			Source:   cue.SourceAnthropicDocs,
		})
		return errors
	}

	// Parse .gitignore and check if CLAUDE.local.md is covered
	if !isInGitignore(gitignorePath, "CLAUDE.local.md") {
		errors = append(errors, cue.ValidationError{
			File:     "CLAUDE.local.md",
			Message:  "CLAUDE.local.md exists but is not in .gitignore - add 'CLAUDE.local.md' to .gitignore to prevent committing personal preferences",
			Severity: "warning",
			Source:   cue.SourceAnthropicDocs,
		})
	}

	return errors
}

// isInGitignore checks if a filename is covered by any pattern in .gitignore.
// Evaluates each gitignore line as a glob pattern against the filename using
// filepath.Match. Handles negation patterns (!) and leading slash stripping.
func isInGitignore(gitignorePath, filename string) bool {
	file, err := os.Open(gitignorePath)
	if err != nil {
		return false
	}
	defer file.Close()

	matched := false
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip comments and empty lines
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Handle negation patterns — un-ignores a previously matched file
		if strings.HasPrefix(line, "!") {
			pattern := strings.TrimPrefix(line, "!")
			pattern = strings.TrimPrefix(pattern, "/")
			if matchGitignorePattern(pattern, filename) {
				matched = false
			}
			continue
		}

		// Strip leading slash (root-anchored patterns)
		pattern := strings.TrimPrefix(line, "/")

		if matchGitignorePattern(pattern, filename) {
			matched = true
		}
	}

	return matched
}

// matchGitignorePattern evaluates a gitignore-style pattern against a filename.
// Supports filepath.Match globs and exact string match as fallback.
func matchGitignorePattern(pattern, filename string) bool {
	// Try filepath.Match (handles *, ?, [])
	if ok, _ := filepath.Match(pattern, filename); ok {
		return true
	}
	// Exact string match (for patterns like "CLAUDE.local.md")
	return pattern == filename
}

// CheckCombinedMemorySize checks if total CLAUDE.md + always-loaded rules size exceeds threshold.
// Rules with paths: frontmatter are conditionally loaded and don't count toward the threshold.
func CheckCombinedMemorySize(rootPath string, files []discovery.File) []cue.ValidationError {
	var errors []cue.ValidationError

	var alwaysLoadedSize int64
	var conditionalSize int64

	// Sum up context files (CLAUDE.md) - always loaded
	for _, f := range files {
		if f.Type == discovery.FileTypeContext {
			alwaysLoadedSize += int64(len(f.Contents))
		}
	}

	// Sum up rule files, separating always-loaded from conditional
	for _, f := range files {
		if f.Type == discovery.FileTypeRule {
			if ruleHasPathsConstraint(f.Contents) {
				conditionalSize += int64(len(f.Contents))
			} else {
				alwaysLoadedSize += int64(len(f.Contents))
			}
		}
	}

	// Also check for CLAUDE.local.md which might not be in discovery
	localMdPath := filepath.Join(rootPath, "CLAUDE.local.md")
	if info, err := os.Stat(localMdPath); err == nil {
		alwaysLoadedSize += info.Size()
	}

	// Check threshold (20KB) - only for always-loaded content
	thresholdBytes := int64(CombinedMemorySizeWarningKB * 1024)
	if alwaysLoadedSize > thresholdBytes {
		errors = append(errors, cue.ValidationError{
			File:     ".claude/",
			Message:  formatSizeWarning(alwaysLoadedSize, conditionalSize, thresholdBytes),
			Severity: "warning",
			Source:   cue.SourceCClintObserve,
		})
	}

	return errors
}

// ruleHasPathsConstraint checks if a rule file has paths: frontmatter,
// meaning it's only loaded conditionally when working on matching files.
func ruleHasPathsConstraint(contents string) bool {
	fm, err := frontend.ParseYAMLFrontmatter(contents)
	if err != nil {
		return false
	}
	_, hasPaths := fm.Data["paths"]
	return hasPaths
}

func formatSizeWarning(alwaysLoaded, conditional, threshold int64) string {
	alwaysKB := float64(alwaysLoaded) / 1024
	thresholdKB := float64(threshold) / 1024

	msg := fmt.Sprintf(
		"Always-loaded memory (CLAUDE.md + global rules) is %.1fKB, exceeds %.0fKB threshold. "+
			"Large memory files increase token usage and may slow down Claude Code startup. "+
			"Consider: moving detailed docs to @imports, or adding paths: frontmatter to make rules conditional.",
		alwaysKB, thresholdKB)

	if conditional > 0 {
		conditionalKB := float64(conditional) / 1024
		msg += fmt.Sprintf(" (%.1fKB in conditional rules excluded from threshold)", conditionalKB)
	}

	return msg
}
