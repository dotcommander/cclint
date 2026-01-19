package docs_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGitHooksGuide_PreCommitHookExample(t *testing.T) {
	// Given docs/guides/git-hooks.md exists
	path := filepath.Join("..", "guides", "git-hooks.md")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read git-hooks.md: %v", err)
	}

	contentStr := string(content)

	// When read, then pre-commit hook script example is shown
	if !strings.Contains(contentStr, "#!/bin/bash") {
		t.Error("Missing shell script shebang in pre-commit example")
	}

	if !strings.Contains(contentStr, "cclint --staged") {
		t.Error("Missing cclint --staged command in pre-commit example")
	}

	if !strings.Contains(contentStr, ".git/hooks/pre-commit") {
		t.Error("Missing .git/hooks/pre-commit path reference")
	}

	if !strings.Contains(contentStr, "chmod +x") {
		t.Error("Missing chmod +x instruction for making hook executable")
	}
}

func TestGitHooksGuide_StagedAndDiffFlagsExplained(t *testing.T) {
	// Given git-hooks.md
	path := filepath.Join("..", "guides", "git-hooks.md")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read git-hooks.md: %v", err)
	}

	contentStr := string(content)

	// When scanning content, then --staged flag is explained
	stagedHeading := strings.Contains(contentStr, "### `--staged`")
	stagedDescription := strings.Contains(contentStr, "Lint only files that are staged")
	stagedUseCase := strings.Contains(contentStr, "Use case")
	stagedBehavior := strings.Contains(contentStr, "Behavior")

	if !stagedHeading {
		t.Error("Missing --staged flag section heading")
	}
	if !stagedDescription {
		t.Error("Missing --staged flag description")
	}
	if !stagedUseCase {
		t.Error("Missing --staged flag use case")
	}
	if !stagedBehavior {
		t.Error("Missing --staged flag behavior explanation")
	}

	// When scanning content, then --diff flag is explained
	diffHeading := strings.Contains(contentStr, "### `--diff`")
	diffDescription := strings.Contains(contentStr, "Lint all uncommitted changes")
	diffUseCase := strings.Contains(contentStr, "Use case")

	if !diffHeading {
		t.Error("Missing --diff flag section heading")
	}
	if !diffDescription {
		t.Error("Missing --diff flag description")
	}
	if !diffUseCase {
		t.Error("Missing --diff flag use case")
	}
}

func TestGitHooksGuide_HuskyIntegration(t *testing.T) {
	// Given git-hooks.md
	path := filepath.Join("..", "guides", "git-hooks.md")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read git-hooks.md: %v", err)
	}

	contentStr := string(content)

	// When checking sections, then husky integration example is included
	huskyHeading := strings.Contains(contentStr, "## Husky Integration")
	huskyLink := strings.Contains(contentStr, "github.com/typicode/husky")
	huskyInstall := strings.Contains(contentStr, "npm install -D husky")
	huskyInit := strings.Contains(contentStr, "npx husky init")
	huskySetCommand := strings.Contains(contentStr, "npx husky set .husky/pre-commit")

	if !huskyHeading {
		t.Error("Missing Husky Integration section")
	}
	if !huskyLink {
		t.Error("Missing link to Husky repository")
	}
	if !huskyInstall {
		t.Error("Missing Husky installation command")
	}
	if !huskyInit {
		t.Error("Missing Husky init command")
	}
	if !huskySetCommand {
		t.Error("Missing Husky pre-commit hook setup command")
	}

	// Check for lint-staged integration example
	lintStagedSection := strings.Contains(contentStr, "lint-staged")
	if !lintStagedSection {
		t.Error("Missing lint-staged integration example")
	}
}
