package cli

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dotcommander/cclint/internal/cue"
	"github.com/dotcommander/cclint/internal/discovery"
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

// isInGitignore checks if a file pattern is covered by .gitignore
func isInGitignore(gitignorePath, filename string) bool {
	file, err := os.Open(gitignorePath)
	if err != nil {
		return false
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip comments and empty lines
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Check various patterns that would match CLAUDE.local.md
		patterns := []string{
			filename,                 // Exact match
			"*.local.md",             // Wildcard for .local.md files
			"CLAUDE.local.*",         // Wildcard for CLAUDE.local.*
			"*.local.*",              // Wildcard for any .local. files
			"/" + filename,           // Root-only match
			"**/" + filename,         // Any directory match
			".claude/" + filename,    // Inside .claude directory
			"**/*.local.md",          // Any .local.md in any directory
		}

		for _, pattern := range patterns {
			if line == pattern {
				return true
			}
		}

		// Check if line is a negation that would un-ignore it
		if strings.HasPrefix(line, "!") {
			negated := strings.TrimPrefix(line, "!")
			if negated == filename {
				return false
			}
		}
	}

	return false
}

// CheckCombinedMemorySize checks if total CLAUDE.md + rules size exceeds threshold
func CheckCombinedMemorySize(rootPath string, files []discovery.File) []cue.ValidationError {
	var errors []cue.ValidationError

	var totalSize int64

	// Sum up context files (CLAUDE.md)
	for _, f := range files {
		if f.Type == discovery.FileTypeContext {
			totalSize += int64(len(f.Contents))
		}
	}

	// Sum up rule files
	for _, f := range files {
		if f.Type == discovery.FileTypeRule {
			totalSize += int64(len(f.Contents))
		}
	}

	// Also check for CLAUDE.local.md which might not be in discovery
	localMdPath := filepath.Join(rootPath, "CLAUDE.local.md")
	if info, err := os.Stat(localMdPath); err == nil {
		totalSize += info.Size()
	}

	// Check threshold (20KB)
	thresholdBytes := int64(CombinedMemorySizeWarningKB * 1024)
	if totalSize > thresholdBytes {
		errors = append(errors, cue.ValidationError{
			File:     ".claude/",
			Message:  formatSizeWarning(totalSize, thresholdBytes),
			Severity: "warning",
			Source:   cue.SourceCClintObserve,
		})
	}

	return errors
}

func formatSizeWarning(actual, threshold int64) string {
	actualKB := float64(actual) / 1024
	thresholdKB := float64(threshold) / 1024
	return fmt.Sprintf(
		"Combined memory size (CLAUDE.md + rules) is %.1fKB, exceeds %.0fKB threshold. "+
			"Large memory files increase token usage and may slow down Claude Code startup. "+
			"Consider: moving detailed docs to @imports, splitting into conditional rules with paths: frontmatter.",
		actualKB, thresholdKB)
}
