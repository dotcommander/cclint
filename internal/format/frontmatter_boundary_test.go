package format

import (
	"reflect"
	"strings"
	"testing"

	"github.com/dotcommander/cclint/internal/textutil"
	"gopkg.in/yaml.v3"
)

func TestFormatterSharedFrontmatterBoundary(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name, input, want string
		wantErr           bool
	}{
		{"LF", "---\nname: test\n---\nBody\n", "---\nname: test\n---\n\nBody\n", false},
		{"CRLF", "---\r\nname: test\r\n--- \t\r\nBody\r\n", "---\nname: test\n---\n\nBody\r\n", false},
		{"EOF closer", "---\nname: test\n---", "---\nname: test\n---\n", false},
		{"EOF closer whitespace", "---\nname: test\n--- \t", "---\nname: test\n---\n", false},
		{"inline opener", "---title\nBody\n", "---title\nBody\n", false},
		{"only opener", "---", "---", true},
		{"unclosed", "---\nname: test\n", "---\nname: test\n", true},
		{"invalid YAML", "---\nname: [\n---\nBody", "---\nname: [\n---\nBody", true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			formatter := &AgentFormatter{}
			got, err := formatter.Format(tt.input)
			if (err != nil) != tt.wantErr || got != tt.want {
				t.Fatalf("Format = %q, %v; want %q, error %v", got, err, tt.want, tt.wantErr)
			}
			if err != nil {
				return
			}
			again, err := formatter.Format(got)
			if err != nil || again != got {
				t.Errorf("formatter not idempotent: %q, %v", again, err)
			}
		})
	}
}

func TestFormatterDelimiterBoundaries(t *testing.T) {
	t.Parallel()
	for _, input := range []string{
		"---\nname: test\ndescription: foo---bar\n---\nBody\n",
		"---\r\nname: test\r\ndescription: foo---bar\r\n---\r\nBody\n",
		"---\nname: test\ndescription: |\n  first\n  ---\n  last\n---\nBody\n",
	} {
		got, err := (&AgentFormatter{}).Format(input)
		if err != nil {
			t.Fatalf("Format(%q): %v", input, err)
		}
		parsed, err := textutil.ExtractRawFrontmatter(got)
		if err != nil {
			t.Fatal(err)
		}
		var values map[string]any
		if err := yaml.Unmarshal([]byte(parsed.YAML), &values); err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(values["description"].(string), "---") {
			t.Fatalf("lost embedded dashes: %q", got)
		}
		if !strings.Contains(got, "Body\n") {
			t.Fatalf("lost body: %q", got)
		}
	}
	input := "---text\nordinary markdown\n"
	got, err := (&SkillFormatter{}).Format(input)
	if err != nil || got != input {
		t.Fatalf("non-delimiter treated as frontmatter: %q, %v", got, err)
	}
}

func TestFormatterPreservesYAMLNodesAndMeaning(t *testing.T) {
	t.Parallel()
	input := "---\n# shared definition\nzebra: &shared\n  quoted: 'true' # retain scalar comment\n  literal: |\n    first---second\n# alias comment\nname: *shared\ndescription: \"12\" # retain description comment\nalpha: !!str 001\n---\nBody\n"
	formatter := &AgentFormatter{}
	got, err := formatter.Format(input)
	if err != nil {
		t.Fatal(err)
	}
	for _, fragment := range []string{"# shared definition", "# retain scalar comment", "# alias comment", "# retain description comment", "&shared", "*shared", "'true'", "!!str 001"} {
		if !strings.Contains(got, fragment) {
			t.Errorf("lost %q in %s", fragment, got)
		}
	}
	rawInput, err := textutil.ExtractRawFrontmatter(input)
	if err != nil {
		t.Fatal(err)
	}
	rawOutput, err := textutil.ExtractRawFrontmatter(got)
	if err != nil {
		t.Fatal(err)
	}
	var original, formatted map[string]any
	if err := yaml.Unmarshal([]byte(rawInput.YAML), &original); err != nil {
		t.Fatal(err)
	}
	if err := yaml.Unmarshal([]byte(rawOutput.YAML), &formatted); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(original, formatted) {
		t.Fatalf("changed YAML meaning: %#v vs %#v", original, formatted)
	}
	again, err := formatter.Format(got)
	if err != nil || again != got {
		t.Fatalf("format is not stable: %v\n%s\n%s", err, got, again)
	}
}

func TestFormatterPreservesRootMerge(t *testing.T) {
	t.Parallel()
	input := "---\ndefaults: &defaults\n  model: sonnet\n<<: *defaults\nname: test\ndescription: Test\n---\nBody\n"
	got, err := (&AgentFormatter{}).Format(input)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := textutil.ExtractRawFrontmatter(got)
	if err != nil {
		t.Fatal(err)
	}
	var values map[string]any
	if err := yaml.Unmarshal([]byte(raw.YAML), &values); err != nil {
		t.Fatal(err)
	}
	if values["model"] != "sonnet" {
		t.Fatalf("lost merge semantics: %#v", values)
	}
	again, err := (&AgentFormatter{}).Format(got)
	if err != nil || got != again {
		t.Fatalf("unstable merge formatting: %v\n%s\n%s", err, got, again)
	}
}
