package git

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func getUntrackedFiles(rootPath string) ([]string, error) {
	cmd, cancel := gitCommand(rootPath, "ls-files", "--others", "--exclude-standard", "-z")
	defer cancel()
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, gitTimeoutError("ls-files --others", err, output)
	}
	return parseUntrackedPaths(output, rootPath)
}

// ls-files -z emits bare paths, not the status/path pairs used by git diff.
func parseUntrackedPaths(output []byte, rootPath string) ([]string, error) {
	var files []string
	for relPath := range strings.SplitSeq(string(output), "\x00") {
		if relPath == "" || !isRelevantFile(relPath) {
			continue
		}
		absPath := filepath.Join(rootPath, relPath)
		if _, err := os.Stat(absPath); err == nil {
			files = append(files, absPath)
		} else if !os.IsNotExist(err) {
			return nil, fmt.Errorf("stat untracked file %s: %w", relPath, err)
		}
	}
	return files, nil
}
