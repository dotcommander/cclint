package cmd

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCollectFilesToFormatPrecedence(t *testing.T) {
	root := t.TempDir()
	explicit := filepath.Join(root, "explicit.md")
	argument := filepath.Join(root, "argument.md")
	require.NoError(t, os.WriteFile(explicit, []byte("explicit"), 0o600))
	require.NoError(t, os.WriteFile(argument, []byte("argument"), 0o600))

	got, err := collectFilesToFormat(&fmtCommand{
		Files: []string{explicit},
		Args:  []string{argument},
	}, root)
	require.NoError(t, err)
	assert.Equal(t, []string{explicit}, got)

	got, err = collectFilesToFormat(&fmtCommand{Args: []string{"agents", argument}}, root)
	require.NoError(t, err)
	assert.Equal(t, []string{argument}, got, "path arguments must take precedence over component types")
}

func TestCollectFilesToFormatUsesFormatterCapability(t *testing.T) {
	root := t.TempDir()
	agentDir := filepath.Join(root, ".claude", "agents")
	outputStyleDir := filepath.Join(root, ".claude", "output-styles")
	require.NoError(t, os.MkdirAll(agentDir, 0o755))
	require.NoError(t, os.MkdirAll(outputStyleDir, 0o755))
	agent := filepath.Join(agentDir, "agent.md")
	outputStyle := filepath.Join(outputStyleDir, "style.md")
	require.NoError(t, os.WriteFile(agent, []byte("# Agent\n"), 0o600))
	require.NoError(t, os.WriteFile(outputStyle, []byte("# Style\n"), 0o600))

	got, err := collectFilesToFormat(&fmtCommand{Args: []string{"agents"}}, root)
	require.NoError(t, err)
	assert.Equal(t, []string{agent}, got, "component selection must use the formatter capability registry")

	_, err = collectFilesToFormat(&fmtCommand{Args: []string{"output-styles"}}, root)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot access output-styles",
		"output styles must remain filesystem arguments rather than formatter component selectors")
}

func TestCollectFilesToFormatExpandsDirectories(t *testing.T) {
	root := t.TempDir()
	nested := filepath.Join(root, "nested")
	require.NoError(t, os.Mkdir(nested, 0o755))
	first := filepath.Join(root, "first.md")
	second := filepath.Join(nested, "second.MD")
	require.NoError(t, os.WriteFile(first, nil, 0o600))
	require.NoError(t, os.WriteFile(second, nil, 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(root, "ignore.txt"), nil, 0o600))

	got, err := collectFilesToFormat(&fmtCommand{Args: []string{root}}, root)
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{first, second}, got)
}

func TestRunFmtWriteAndCheckModes(t *testing.T) {
	root := t.TempDir()
	agentDir := filepath.Join(root, ".claude", "agents")
	require.NoError(t, os.MkdirAll(agentDir, 0o755))
	path := filepath.Join(agentDir, "test.md")
	unformatted := "---\nname: test\ndescription: Test agent\ntools:\n- Read\n---\n# Test\n"
	require.NoError(t, os.WriteFile(path, []byte(unformatted), 0o600))

	exited := 0
	originalExit := exitFunc
	t.Cleanup(func() { exitFunc = originalExit })
	exitFunc = func(code int) { exited = code }

	quiet := true
	opts := executionOptions{root: &root, quiet: &quiet}
	require.NoError(t, runFmt(opts, &fmtCommand{Check: true, Args: []string{path}}))
	assert.Equal(t, 1, exited)

	exited = 0
	require.NoError(t, runFmt(opts, &fmtCommand{Write: true, Args: []string{path}}))
	assert.Zero(t, exited)
	formatted, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.NotEqual(t, unformatted, string(formatted))

	require.NoError(t, runFmt(opts, &fmtCommand{Check: true, Args: []string{path}}))
	assert.Zero(t, exited)
}

func TestEmitFormattedModePrecedence(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.md")
	require.NoError(t, os.WriteFile(path, []byte("original\n"), 0o600))

	stdout := captureFile(t, os.Stdout, func() {
		require.NoError(t, emitFormatted(executionOptions{}, &fmtCommand{Check: true, Diff: true, Write: true}, path, "test.md", "original\n", "formatted\n"))
	})
	assert.Equal(t, "test.md needs formatting\n", stdout)
	contents, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, "original\n", string(contents), "check must take precedence over diff and write")

	stdout = captureFile(t, os.Stdout, func() {
		require.NoError(t, emitFormatted(executionOptions{}, &fmtCommand{Diff: true, Write: true}, path, "test.md", "original\n", "formatted\n"))
	})
	assert.Contains(t, stdout, "--- test.md")
	contents, err = os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, "original\n", string(contents), "diff must take precedence over write")

	require.NoError(t, emitFormatted(executionOptions{quiet: ptr(true)}, &fmtCommand{Write: true}, path, "test.md", "original\n", "formatted\n"))
	contents, err = os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, "formatted\n", string(contents))
}

func TestRunFmtUsesCLIOnlyQuietAndVerbose(t *testing.T) {
	root := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(root, ".cclintrc.yaml"), []byte("quiet: true\nverbose: true\n"), 0o600))
	nonMarkdown := filepath.Join(root, "notes.txt")
	require.NoError(t, os.WriteFile(nonMarkdown, []byte("notes"), 0o600))

	stderr := captureFile(t, os.Stderr, func() {
		require.NoError(t, runFmt(executionOptions{root: &root}, &fmtCommand{Args: []string{nonMarkdown}}))
	})
	assert.Contains(t, stderr, "Skipping "+nonMarkdown,
		"fmt must use absent CLI quiet as false instead of inheriting quiet=true from config")
}

func TestResolveFileTypeOverride(t *testing.T) {
	cmd := &fmtCommand{Type: "agent"}
	fileType, skip, err := resolveFileType(executionOptions{}, cmd, "irrelevant", "irrelevant", t.TempDir())
	require.NoError(t, err)
	assert.False(t, skip)
	assert.Equal(t, "agent", fileType.String())

	cmd.Type = "invalid"
	_, _, err = resolveFileType(executionOptions{}, cmd, "irrelevant", "irrelevant", t.TempDir())
	require.Error(t, err)
}

func TestAtomicFormatReplacementPreservesModeAndLink(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "target.md")
	link := filepath.Join(root, "link.md")
	require.NoError(t, os.WriteFile(target, []byte("original"), 0o640))
	require.NoError(t, os.Symlink(target, link))
	require.NoError(t, replaceFormattedFile(link, "formatted", os.Rename))
	info, err := os.Stat(target)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o640), info.Mode().Perm())
	linkInfo, err := os.Lstat(link)
	require.NoError(t, err)
	assert.NotZero(t, linkInfo.Mode()&os.ModeSymlink)
	content, err := os.ReadFile(target)
	require.NoError(t, err)
	assert.Equal(t, "formatted", string(content))
	assertNoFormatterTemps(t, root)
}

func TestAtomicFormatFailedRenamePreservesOriginal(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "original.md")
	require.NoError(t, os.WriteFile(path, []byte("original"), 0o640))
	canonicalPath, err := filepath.EvalSymlinks(path)
	require.NoError(t, err)
	failure := errors.New("injected rename failure")
	err = replaceFormattedFile(path, "replacement", func(source, target string) error {
		assert.Equal(t, filepath.Dir(canonicalPath), filepath.Dir(source))
		assert.Equal(t, canonicalPath, target)
		prepared, readErr := os.ReadFile(source)
		require.NoError(t, readErr)
		assert.Equal(t, "replacement", string(prepared))
		return failure
	})
	require.ErrorIs(t, err, failure)
	content, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, "original", string(content))
	info, err := os.Stat(path)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o640), info.Mode().Perm())
	assertNoFormatterTemps(t, root)
}

func assertNoFormatterTemps(t *testing.T, root string) {
	t.Helper()
	entries, err := os.ReadDir(root)
	require.NoError(t, err)
	for _, entry := range entries {
		assert.False(t, strings.HasPrefix(entry.Name(), ".cclint-fmt-"), "left temporary %s", entry.Name())
	}
}
