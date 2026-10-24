package cmd

import (
	"testing"

	"github.com/dotcommander/cclint/internal/discovery"
	"github.com/dotcommander/cclint/internal/lint"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunTypesLintUsesRegistry(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	quiet := true

	err := runTypesLint(executionOptions{root: &root, quiet: &quiet}, []discovery.FileType{discovery.FileTypeAgent})
	require.NoError(t, err)
}

func TestRunTypesLintRejectsUnregisteredType(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	quiet := true

	err := runTypesLint(executionOptions{root: &root, quiet: &quiet}, []discovery.FileType{discovery.FileTypeUnknown})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no linter for type")
}

func TestSummaryRegistrySelection(t *testing.T) {
	for _, fileType := range []discovery.FileType{
		discovery.FileTypeAgent,
		discovery.FileTypeCommand,
		discovery.FileTypeSkill,
	} {
		entry, ok := lint.LinterForType(fileType)
		require.True(t, ok, "missing registry entry for %s", fileType)
		assert.NotEmpty(t, entry.Name)
		assert.NotNil(t, entry.New)
	}
}
