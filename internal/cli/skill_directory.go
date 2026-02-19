package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/dotcommander/cclint/internal/cue"
)

// ValidateSkillDirectory checks the skill's directory structure per agentskills.io spec.
// This includes scripts/, references/, and assets/ subdirectories.
func ValidateSkillDirectory(skillPath, contents string) []cue.ValidationError {
	var issues []cue.ValidationError

	skillDir := filepath.Dir(skillPath)

	// Check scripts/ directory
	scriptsDir := filepath.Join(skillDir, "scripts")
	if stat, err := os.Stat(scriptsDir); err == nil && stat.IsDir() {
		issues = append(issues, validateScriptsDirectory(scriptsDir, skillPath)...)
	}

	// Check for absolute paths in markdown links
	issues = append(issues, validateRelativePaths(contents, skillPath)...)

	// Check reference chain depth (markdown links to references that link to other references)
	issues = append(issues, validateReferenceDepth(contents, skillDir, skillPath)...)

	return issues
}

// validateScriptsDirectory checks scripts for shebangs and basic structure.
func validateScriptsDirectory(scriptsDir, skillPath string) []cue.ValidationError {
	var issues []cue.ValidationError

	entries, err := os.ReadDir(scriptsDir)
	if err != nil {
		return issues
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		scriptPath := filepath.Join(scriptsDir, entry.Name())
		relPath := filepath.Join("scripts", entry.Name())

		// Check for shebang in script files
		content, err := os.ReadFile(scriptPath)
		if err != nil {
			continue
		}

		// Skip non-script files (images, data files, etc.)
		ext := strings.ToLower(filepath.Ext(entry.Name()))
		scriptExts := map[string]bool{
			".sh": true, ".bash": true, ".zsh": true,
			".py": true, ".python": true,
			".js": true, ".ts": true, ".mjs": true,
			".rb": true, ".pl": true, ".php": true,
		}

		// If it's a known script extension or has no extension, check for shebang
		if scriptExts[ext] || ext == "" {
			if len(content) > 0 && !strings.HasPrefix(string(content), "#!") {
				issues = append(issues, cue.ValidationError{
					File:     skillPath,
					Message:  fmt.Sprintf("Script '%s' missing shebang (e.g., #!/usr/bin/env python3)", relPath),
					Severity: "suggestion",
					Source:   cue.SourceAgentSkillsIO,
				})
			}
		}

		// Check for executable permission on Unix-like systems
		info, err := entry.Info()
		if err == nil {
			mode := info.Mode()
			if mode&0111 == 0 && scriptExts[ext] {
				issues = append(issues, cue.ValidationError{
					File:     skillPath,
					Message:  fmt.Sprintf("Script '%s' is not executable (chmod +x)", relPath),
					Severity: "suggestion",
					Source:   cue.SourceAgentSkillsIO,
				})
			}
		}
	}

	return issues
}

// validateRelativePaths checks for absolute paths in markdown links.
func validateRelativePaths(contents, skillPath string) []cue.ValidationError {
	var issues []cue.ValidationError

	// Match markdown links: [text](path) or [text](path "title")
	linkPattern := regexp.MustCompile(`\[([^\]]*)\]\(([^)\s]+)(?:\s+"[^"]*")?\)`)
	matches := linkPattern.FindAllStringSubmatch(contents, -1)

	for _, match := range matches {
		if len(match) < 3 {
			continue
		}
		linkPath := match[2]

		// Skip URLs
		if strings.HasPrefix(linkPath, "http://") || strings.HasPrefix(linkPath, "https://") {
			continue
		}

		// Check for absolute paths
		if strings.HasPrefix(linkPath, "/") || strings.HasPrefix(linkPath, "~") {
			line := findSubstringLine(contents, match[0])
			issues = append(issues, cue.ValidationError{
				File:     skillPath,
				Message:  fmt.Sprintf("Use relative path instead of absolute: '%s'", linkPath),
				Severity: "warning",
				Source:   cue.SourceAgentSkillsIO,
				Line:     line,
			})
		}
	}

	return issues
}

// validateReferenceDepth checks for nested reference chains >1 level deep.
func validateReferenceDepth(contents string, skillDir, skillPath string) []cue.ValidationError {
	var issues []cue.ValidationError

	// Extract markdown links to local files
	linkPattern := regexp.MustCompile(`\[([^\]]*)\]\(([^)\s]+)(?:\s+"[^"]*")?\)`)
	matches := linkPattern.FindAllStringSubmatch(contents, -1)

	for _, match := range matches {
		if len(match) < 3 {
			continue
		}
		linkPath := match[2]

		if !isLocalReferenceLink(linkPath) {
			continue
		}

		// Check if it's a reference to references/ directory
		if !strings.Contains(linkPath, "references/") {
			continue
		}

		// Check for nested references in the linked file
		if nestedIssues := checkNestedReferences(linkPath, skillDir, skillPath, linkPattern); len(nestedIssues) > 0 {
			issues = append(issues, nestedIssues...)
		}
	}

	return issues
}

// isLocalReferenceLink checks if a link is a local reference (not URL or anchor).
func isLocalReferenceLink(linkPath string) bool {
	return !(strings.HasPrefix(linkPath, "http") || strings.HasPrefix(linkPath, "#"))
}

// isSubdirIndexFile returns true if linkPath points to a file inside a subdirectory
// of references/ (e.g. references/work-tasks/work-tasks.md). Such files are
// subdirectory index files whose own links are co-located siblings, not chains.
func isSubdirIndexFile(linkPath string) bool {
	// Normalize to forward slashes for consistent splitting
	normalized := filepath.ToSlash(linkPath)
	// Strip leading references/ prefix
	after, found := strings.CutPrefix(normalized, "references/")
	if !found {
		return false
	}
	// If there is still a "/" in the remainder, the file lives in a subdir
	return strings.Contains(after, "/")
}

// checkNestedReferences checks if a referenced file contains links to other references/.
// Returns validation errors for each nested reference chain found.
func checkNestedReferences(linkPath, skillDir, skillPath string, linkPattern *regexp.Regexp) []cue.ValidationError {
	var issues []cue.ValidationError

	// Files inside references/<subdir>/ are subdirectory index files. Their own
	// links are co-located siblings and should not be flagged as reference chains.
	if isSubdirIndexFile(linkPath) {
		return issues
	}

	// Resolve the full path
	fullPath := filepath.Join(skillDir, linkPath)
	if _, err := os.Stat(fullPath); err != nil {
		return issues
	}

	// Read the referenced file
	refContent, err := os.ReadFile(fullPath)
	if err != nil {
		return issues
	}

	// Check for nested references/
	nestedMatches := linkPattern.FindAllStringSubmatch(string(refContent), -1)
	for _, nested := range nestedMatches {
		if len(nested) < 3 {
			continue
		}
		nestedPath := nested[2]

		if !isLocalReferenceLink(nestedPath) {
			continue
		}

		// Check if it links to another references/ file
		if hasNestedReferenceLink(nestedPath) {
			issues = append(issues, cue.ValidationError{
				File:     skillPath,
				Message:  fmt.Sprintf("Reference chain detected: SKILL.md → %s → %s (keep references 1 level deep)", linkPath, nestedPath),
				Severity: "suggestion",
				Source:   cue.SourceAgentSkillsIO,
			})
			break // Only report once per referenced file
		}
	}

	return issues
}

// hasNestedReferenceLink checks if a link path indicates a nested reference.
// Only flags actual cross-directory reference chains (links into references/ subdirs),
// not co-located .md files in the same directory.
func hasNestedReferenceLink(nestedPath string) bool {
	return strings.Contains(nestedPath, "references/")
}

// findSubstringLine finds the line number of a substring in content.
func findSubstringLine(content, substr string) int {
	idx := strings.Index(content, substr)
	if idx == -1 {
		return 0
	}
	return strings.Count(content[:idx], "\n") + 1
}
