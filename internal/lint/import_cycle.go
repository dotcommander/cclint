package lint

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/dotcommander/cclint/internal/cue"
)

// importPattern matches @import directives outside of code blocks.
// Supports: @path/to/file, @./relative/path, @~/home/path
var importDirectivePattern = regexp.MustCompile(`(?m)^[^` + "`" + `]*@([~./][^\s]+)`)

// ImportGraph tracks @import dependencies between files for cycle detection.
type ImportGraph struct {
	// edges maps each file (absolute path) to the files it imports.
	edges map[string][]string
}

// NewImportGraph creates an empty import graph.
func NewImportGraph() *ImportGraph {
	return &ImportGraph{
		edges: make(map[string][]string),
	}
}

// AddFile registers a file and its imports in the graph.
// filePath should be absolute. importPaths are resolved relative to filePath's directory.
func (g *ImportGraph) AddFile(filePath string, importPaths []string) {
	absFile, _ := filepath.Abs(filePath)
	dir := filepath.Dir(absFile)

	var resolved []string
	for _, imp := range importPaths {
		absImp := resolveImportPath(imp, dir)
		if absImp != "" {
			resolved = append(resolved, absImp)
		}
	}

	g.edges[absFile] = resolved
}

// DetectCycles finds all circular @import chains using DFS with three-color marking.
//
// Algorithm mirrors CrossFileValidator.DetectCycles:
//   - white (0): unvisited
//   - gray  (1): in current recursion stack
//   - black (2): fully explored
//
// A back edge (gray -> gray) indicates a cycle.
func (g *ImportGraph) DetectCycles() [][]string {
	var cycles [][]string

	state := make(map[string]int) // 0=white, 1=gray, 2=black
	var path []string
	inPath := make(map[string]bool)

	var visit func(node string)
	visit = func(node string) {
		state[node] = 1
		path = append(path, node)
		inPath[node] = true

		for _, neighbor := range g.edges[node] {
			if state[neighbor] == 0 {
				visit(neighbor)
			} else if state[neighbor] == 1 && inPath[neighbor] {
				// Gray and in path: cycle detected
				if cycle := extractImportCycle(path, neighbor); cycle != nil {
					cycles = append(cycles, cycle)
				}
			}
			// Black (2): skip
		}

		state[node] = 2
		path = path[:len(path)-1]
		delete(inPath, node)
	}

	// Visit all nodes (sorted for deterministic output)
	for node := range g.edges {
		if state[node] == 0 {
			visit(node)
		}
	}

	return cycles
}

// ExtractImports parses file contents and returns raw import paths.
// Skips imports inside fenced code blocks.
func ExtractImports(contents string) []string {
	var imports []string
	seen := make(map[string]bool)

	lines := strings.Split(contents, "\n")
	inCodeBlock := false

	for _, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "```") {
			inCodeBlock = !inCodeBlock
			continue
		}
		if inCodeBlock {
			continue
		}

		matches := importDirectivePattern.FindAllStringSubmatch(line, -1)
		for _, match := range matches {
			if len(match) >= 2 {
				imp := match[1]
				if !seen[imp] {
					imports = append(imports, imp)
					seen[imp] = true
				}
			}
		}
	}

	return imports
}

// resolveImportPath converts a raw import path to an absolute path.
// Handles ~/ (home directory) and ./ or ../ (relative to baseDir).
func resolveImportPath(importPath, baseDir string) string {
	if strings.HasPrefix(importPath, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return ""
		}
		return filepath.Clean(filepath.Join(home, importPath[1:]))
	}

	if !filepath.IsAbs(importPath) {
		return filepath.Clean(filepath.Join(baseDir, importPath))
	}

	return filepath.Clean(importPath)
}

// FormatImportCycle formats a cycle of absolute paths into a readable chain
// using basenames for brevity.
func FormatImportCycle(cycle []string) string {
	if len(cycle) == 0 {
		return ""
	}

	var sb strings.Builder
	for i, node := range cycle {
		sb.WriteString(filepath.Base(node))
		if i < len(cycle)-1 {
			sb.WriteString(" -> ")
		}
	}
	return sb.String()
}

// extractImportCycle extracts the cycle path from the DFS stack.
// Returns nil if the neighbor is not found in the path.
func extractImportCycle(path []string, neighbor string) []string {
	cycleStart := -1
	for i, p := range path {
		if p == neighbor {
			cycleStart = i
			break
		}
	}
	if cycleStart < 0 {
		return nil
	}
	cycle := make([]string, len(path)-cycleStart+1)
	copy(cycle, path[cycleStart:])
	cycle[len(cycle)-1] = neighbor
	return cycle
}

// DetectImportCycles builds an import graph from the provided file map
// and returns validation errors for any circular import chains found.
//
// files maps absolute file paths to their contents.
func DetectImportCycles(files map[string]string) []cue.ValidationError {
	graph := NewImportGraph()

	for filePath, contents := range files {
		imports := ExtractImports(contents)
		graph.AddFile(filePath, imports)
	}

	cycles := graph.DetectCycles()
	var errors []cue.ValidationError

	for _, cycle := range cycles {
		// Report the error on the first file in the cycle
		errors = append(errors, cue.ValidationError{
			File:     cycle[0],
			Message:  fmt.Sprintf("Circular @import detected: %s", FormatImportCycle(cycle)),
			Severity: "error",
			Source:   cue.SourceCClintObserve,
		})
	}

	return errors
}
