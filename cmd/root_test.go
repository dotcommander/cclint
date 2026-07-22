package cmd

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/alecthomas/kong"
	"github.com/dotcommander/cclint/internal/discovery"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCLIParsingKeepsGlobalAndCommandStateSeparate(t *testing.T) {
	var app cli
	parser, err := kong.New(&app, kong.Name("cclint"))
	require.NoError(t, err)

	ctx, err := parser.Parse([]string{
		"lint", "--root", ".", "--quiet=false", "--verbose", "--scores",
		"--improvements", "--format", "json", "--output", "out.json",
		"--fail-on", "warning", "--no-cycle-check=false", "--baseline",
		"--baseline-create=false", "--baseline-path", "base.json", "--type", "agent",
		"--diff=false", "--staged", "file.md",
	})
	require.NoError(t, err)
	assert.Equal(t, "lint <files-or-dirs>", ctx.Command())

	opts := app.executionOptions()
	require.NotNil(t, opts.root)
	assert.True(t, filepath.IsAbs(*opts.root))
	require.NotNil(t, opts.quiet)
	assert.False(t, *opts.quiet)
	require.NotNil(t, opts.verbose)
	assert.True(t, *opts.verbose)
	require.NotNil(t, opts.noCycleCheck)
	assert.False(t, *opts.noCycleCheck)
	require.NotNil(t, opts.baselineCreate)
	assert.False(t, *opts.baselineCreate)
	assert.Equal(t, "json", stringValue(opts.format))
	assert.Equal(t, "warning", stringValue(opts.failOn))

	require.NotNil(t, app.Lint.Type)
	assert.Equal(t, "agent", *app.Lint.Type)
	require.NotNil(t, app.Lint.Diff)
	assert.False(t, *app.Lint.Diff)
	require.NotNil(t, app.Lint.Staged)
	assert.True(t, *app.Lint.Staged)
	assert.Equal(t, []string{"file.md"}, app.Lint.Paths)
}

func TestCLIAbsentFlagsRemainNilAcrossParses(t *testing.T) {
	var first cli
	firstParser, err := kong.New(&first, kong.Name("cclint"))
	require.NoError(t, err)
	_, err = firstParser.Parse([]string{"--quiet=false", "--format", "json"})
	require.NoError(t, err)
	assert.NotNil(t, first.executionOptions().quiet)

	var second cli
	secondParser, err := kong.New(&second, kong.Name("cclint"))
	require.NoError(t, err)
	_, err = secondParser.Parse(nil)
	require.NoError(t, err)
	assert.Nil(t, second.executionOptions().quiet)
	assert.Nil(t, second.executionOptions().format)
}

func TestCLICommandDispatch(t *testing.T) {
	tests := []struct {
		args        []string
		wantCommand string
	}{
		{args: []string{"lint", "--quiet", "file.md"}, wantCommand: "lint <files-or-dirs>"},
		{args: []string{"fmt", "--quiet", "--diff", "file.md"}, wantCommand: "fmt <files>"},
		{args: []string{"summary", "--quiet"}, wantCommand: "summary"},
	}
	for _, tt := range tests {
		var app cli
		parser, err := kong.New(&app, kong.Name("cclint"))
		require.NoError(t, err)
		ctx, err := parser.Parse(tt.args)
		require.NoError(t, err)
		assert.Equal(t, tt.wantCommand, ctx.Command())
	}
}

func TestClassifyArgs(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		wantTypes []discovery.FileType
		wantPaths []string
		wantErr   string
	}{
		{name: "empty"},
		{name: "types", args: []string{"agents", "commands"}, wantTypes: []discovery.FileType{discovery.FileTypeAgent, discovery.FileTypeCommand}},
		{name: "paths", args: []string{"./agents", "custom.md"}, wantPaths: []string{"./agents", "custom.md"}},
		{name: "mixed", args: []string{"agents", "custom.md"}, wantErr: "cannot mix type filters"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := classifyArgs(tt.args)
			if tt.wantErr != "" {
				require.ErrorContains(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantTypes, got.typeFilters)
			assert.Equal(t, tt.wantPaths, got.filePaths)
		})
	}
}

func TestRunRootCommandGitModeTakesDispatchPrecedence(t *testing.T) {
	root := t.TempDir()
	gitInit := exec.Command("git", "init", root)
	require.NoError(t, gitInit.Run())
	quiet := true
	diff := true

	err := runRootCommand(
		executionOptions{root: &root, quiet: &quiet},
		&lintCommand{Diff: &diff, Paths: []string{"agents", "custom.md"}},
	)
	require.NoError(t, err, "git mode must dispatch before mixed-argument classification")
}

func TestRunGitLintStagedTakesPrecedenceOverDiff(t *testing.T) {
	root := t.TempDir()
	require.NoError(t, exec.Command("git", "init", root).Run())
	for _, args := range [][]string{{"config", "user.email", "test@example.com"}, {"config", "user.name", "Test"}} {
		command := exec.Command("git", args...)
		command.Dir = root
		require.NoError(t, command.Run())
	}
	require.NoError(t, os.WriteFile(filepath.Join(root, "README.md"), []byte("# Test\n"), 0o600))
	for _, args := range [][]string{{"add", "README.md"}, {"commit", "-m", "initial"}} {
		command := exec.Command("git", args...)
		command.Dir = root
		require.NoError(t, command.Run())
	}
	agentsDir := filepath.Join(root, ".claude", "agents")
	require.NoError(t, os.MkdirAll(agentsDir, 0o755))
	valid := "---\nname: staged\ndescription: A deterministic agent. Use PROACTIVELY when testing.\nmodel: sonnet\n---\n\n## Foundation\n\nTest.\n\n## Workflow\n\n1. Test.\n"
	unstagedPath := filepath.Join(agentsDir, "unstaged.md")
	require.NoError(t, os.WriteFile(unstagedPath, []byte(valid), 0o600))
	for _, args := range [][]string{{"add", "-f", ".claude/agents/unstaged.md"}, {"commit", "-m", "add fixture"}} {
		command := exec.Command("git", args...)
		command.Dir = root
		require.NoError(t, command.Run())
	}
	validPath := filepath.Join(agentsDir, "staged.md")
	require.NoError(t, os.WriteFile(validPath, []byte(valid), 0o600))
	add := exec.Command("git", "add", "-f", ".claude/agents/staged.md")
	add.Dir = root
	require.NoError(t, add.Run())
	require.NoError(t, os.WriteFile(unstagedPath, nil, 0o600))

	originalExit := exitFunc
	t.Cleanup(func() { exitFunc = originalExit })
	exitCode := 0
	exitFunc = func(code int) { exitCode = code }
	quiet, staged, diff := true, true, true
	opts := executionOptions{root: &root, quiet: &quiet}

	require.NoError(t, runGitLint(opts, &lintCommand{Staged: &staged, Diff: &diff}))
	assert.Zero(t, exitCode, "staged mode must ignore the unstaged invalid file when both flags are true")

	exitCode = 0
	staged = false
	require.NoError(t, runGitLint(opts, &lintCommand{Staged: &staged, Diff: &diff}))
	assert.Equal(t, 1, exitCode, "diff mode must include the unstaged invalid file")
}

func TestRunRootCommandRunsMultipleTypesSequentially(t *testing.T) {
	root := t.TempDir()
	quiet := true
	err := runRootCommand(
		executionOptions{root: &root, quiet: &quiet},
		&lintCommand{Paths: []string{"agents", "commands"}},
	)
	require.NoError(t, err)
}

func TestRunSingleFileLintAutoDetectsProjectRootWithoutRootFlag(t *testing.T) {
	projectRoot := t.TempDir()
	file := filepath.Join(projectRoot, ".claude", "agents", "external.md")
	contents := "---\nname: external\ndescription: A deterministic agent outside the working directory. Use PROACTIVELY when testing.\nmodel: sonnet\n---\n\n## Foundation\n\nTest.\n\n## Workflow\n\n1. Test.\n"
	require.NoError(t, os.MkdirAll(filepath.Dir(file), 0o755))
	require.NoError(t, os.WriteFile(file, []byte(contents), 0o600))

	originalExit := exitFunc
	t.Cleanup(func() { exitFunc = originalExit })
	exitCode := 0
	exitFunc = func(code int) { exitCode = code }
	quiet := true

	require.NoError(t, runSingleFileLint(
		executionOptions{quiet: &quiet},
		&lintCommand{},
		[]string{file},
	))
	assert.Zero(t, exitCode, "an absent --root must auto-detect the file's project")
}

func TestExitWithErrorUsesProcessSeam(t *testing.T) {
	originalExit := exitFunc
	t.Cleanup(func() { exitFunc = originalExit })
	code := 0
	exitFunc = func(got int) { code = got }

	stderr := captureFile(t, os.Stderr, func() { exitWithError(assert.AnError) })
	assert.Equal(t, 1, code)
	assert.Contains(t, stderr, "Error: assert.AnError general error for testing")
}

func captureFile(t *testing.T, target *os.File, action func()) string {
	t.Helper()
	reader, writer, err := os.Pipe()
	require.NoError(t, err)

	switch target {
	case os.Stdout:
		old := os.Stdout
		os.Stdout = writer
		defer func() { os.Stdout = old }()
	case os.Stderr:
		old := os.Stderr
		os.Stderr = writer
		defer func() { os.Stderr = old }()
	}

	action()
	require.NoError(t, writer.Close())
	var buf bytes.Buffer
	_, err = buf.ReadFrom(reader)
	require.NoError(t, err)
	require.NoError(t, reader.Close())
	return buf.String()
}
