package lint

import (
	"os"
	"regexp"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

// agentCUESchemaPath is relative to the package directory (internal/lint/).
const agentCUESchemaPath = "../cue/schemas/agent.cue"

// readAgentCUESchema reads the embedded CUE agent schema from disk.
// Go tests run with cwd = the package directory, so the relative path is stable.
func readAgentCUESchema(t *testing.T) []byte {
	t.Helper()
	data, err := os.ReadFile(agentCUESchemaPath)
	if err != nil {
		t.Fatalf("cannot read agent CUE schema at %s: %v", agentCUESchemaPath, err)
	}
	return data
}

// quotedTokenRe extracts individual quoted strings from a CUE union body.
var quotedTokenRe = regexp.MustCompile(`"([^"]+)"`)

// extractCUEEnum reads all quoted alternatives from a named CUE type definition
// (e.g. `#Color: "red" | "blue" | ...`) in the given schema bytes.
func extractCUEEnum(schema []byte, typeName string) map[string]struct{} {
	blockRe := regexp.MustCompile(`(?ms)^#` + regexp.QuoteMeta(typeName) + `:\s*((?:"[^"]+"\s*\|?\s*)+)`)
	match := blockRe.FindSubmatch(schema)
	if match == nil {
		return nil
	}
	out := make(map[string]struct{})
	for _, m := range quotedTokenRe.FindAllSubmatch(match[1], -1) {
		out[string(m[1])] = struct{}{}
	}
	return out
}

// extractCUEFieldEnum extracts the inline enum from a struct field like:
// `permissionMode?: "default" | "acceptEdits" | ...`
func extractCUEFieldEnum(schema []byte, fieldName string) map[string]struct{} {
	fieldRe := regexp.MustCompile(`(?m)^\s*` + regexp.QuoteMeta(fieldName) + `\??\s*:\s*((?:"[^"]+"\s*\|?\s*)+)`)
	match := fieldRe.FindSubmatch(schema)
	if match == nil {
		return nil
	}
	out := make(map[string]struct{})
	for _, m := range quotedTokenRe.FindAllSubmatch(match[1], -1) {
		out[string(m[1])] = struct{}{}
	}
	return out
}

// sortedKeys returns sorted keys from a map[string]bool for readable diffs.
func sortedBoolKeys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func sortedStructKeys(m map[string]struct{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// TestEnumParity asserts that the three Go package-level enum maps exactly match
// the corresponding CUE schema enum unions in agent.cue.
func TestEnumParity(t *testing.T) {
	t.Parallel()
	schema := readAgentCUESchema(t)

	t.Run("validColors matches #Color", func(t *testing.T) {
		t.Parallel()
		cueEnum := extractCUEEnum(schema, "Color")
		assert.NotEmpty(t, cueEnum, "extractCUEEnum(Color) returned empty — regex may be wrong")

		goKeys := sortedBoolKeys(validColors)
		cueKeys := sortedStructKeys(cueEnum)

		assert.Equal(t, cueKeys, goKeys,
			"validColors does not match CUE #Color enum;\nGo:  %v\nCUE: %v", goKeys, cueKeys)
	})

	t.Run("validScopes matches #MemoryScope", func(t *testing.T) {
		t.Parallel()
		cueEnum := extractCUEEnum(schema, "MemoryScope")
		assert.NotEmpty(t, cueEnum, "extractCUEEnum(MemoryScope) returned empty — regex may be wrong")

		goKeys := sortedBoolKeys(validScopes)
		cueKeys := sortedStructKeys(cueEnum)

		assert.Equal(t, cueKeys, goKeys,
			"validScopes does not match CUE #MemoryScope enum;\nGo:  %v\nCUE: %v", goKeys, cueKeys)
	})

	t.Run("validModes matches permissionMode field", func(t *testing.T) {
		t.Parallel()
		cueEnum := extractCUEFieldEnum(schema, "permissionMode")
		assert.NotEmpty(t, cueEnum, "extractCUEFieldEnum(permissionMode) returned empty — regex may be wrong")

		goKeys := sortedBoolKeys(validModes)
		cueKeys := sortedStructKeys(cueEnum)

		assert.Equal(t, cueKeys, goKeys,
			"validModes does not match CUE permissionMode field enum;\nGo:  %v\nCUE: %v", goKeys, cueKeys)
	})
}
