package format

import (
	"strings"
	"testing"
)

func TestFormatterFrontmatterSpacing(t *testing.T) {
	t.Parallel()

	for _, component := range []string{"agent", "command", "skill"} {
		t.Run(component, func(t *testing.T) {
			t.Parallel()

			for _, separator := range []string{"", "\n", "\n\n\n"} {
				t.Run("separator-"+strings.ReplaceAll(separator, "\n", "LF"), func(t *testing.T) {
					t.Parallel()

					frontmatter := "---\nname: helper\ndescription: Test helper\n---\n"
					body := "# Helper\n\nRead the source.\n"
					want := frontmatter + "\n" + body
					formatter := NewComponentFormatter(component)
					got, err := formatter.Format(frontmatter + separator + body)
					if err != nil {
						t.Fatal(err)
					}
					if got != want {
						t.Fatalf("Format() = %q, want %q", got, want)
					}
					again, err := formatter.Format(got)
					if err != nil || again != got {
						t.Fatalf("second Format() = %q, %v; want %q", again, err, got)
					}
					if diff := Diff(want, got, "helper.md"); diff != "" {
						t.Fatalf("canonical input has a diff: %s", diff)
					}
				})
			}
		})
	}
}

func TestFormatterEmptyFrontmatterBody(t *testing.T) {
	t.Parallel()

	formatter := NewComponentFormatter("agent")
	want := "---\nname: helper\ndescription: Test helper\n---\n"
	for _, suffix := range []string{"", "\n", "\n\n\n"} {
		got, err := formatter.Format(want + suffix)
		if err != nil || got != want {
			t.Fatalf("empty body Format() = %q, %v; want %q", got, err, want)
		}
	}
}
