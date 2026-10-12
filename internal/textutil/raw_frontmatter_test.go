package textutil

import "testing"

func TestRawFrontmatterDelimiterGrammar(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name, input, yaml, body string
		present, wantErr        bool
	}{
		{name: "empty"},
		{name: "plain", input: "Body\n", body: "Body\n"},
		{name: "inline opening", input: "---title\nBody\n", body: "---title\nBody\n"},
		{name: "indented opening", input: " ---\nname: test\n---\n", body: " ---\nname: test\n---\n"},
		{name: "tabs opening", input: "\t---\nname: test\n---\n", body: "\t---\nname: test\n---\n"},
		{name: "bare CR opening", input: "---\r", body: "---\r"},
		{name: "leading blank", input: "\n---\nname: test\n---\n", body: "\n---\nname: test\n---\n"},
		{name: "LF", input: "---\nname: test\n---\nBody", yaml: "name: test\n", body: "\nBody", present: true},
		{name: "CRLF", input: "---\r\nname: test\r\n---\r\nBody\r\n", yaml: "name: test\r\n", body: "\r\nBody\r\n", present: true},
		{name: "trailing whitespace", input: "--- \t\nname: test\n--- \t\r\nBody", yaml: "name: test\n", body: " \t\r\nBody", present: true},
		{name: "empty YAML", input: "---\n---\nBody", body: "\nBody", present: true},
		{name: "EOF closing", input: "---\nname: test\n---", yaml: "name: test\n", present: true},
		{name: "whitespace EOF closing", input: "---\nname: test\n--- \t", yaml: "name: test\n", body: " \t", present: true},
		{name: "YAML is not decoded", input: "---\nbroken: [\n---\nBody", yaml: "broken: [\n", body: "\nBody", present: true},
		{name: "embedded dashes", input: "---\ndescription: \"before---after\"\nscript: |\n  echo ---\n---\nBody", yaml: "description: \"before---after\"\nscript: |\n  echo ---\n", body: "\nBody", present: true},
		{name: "only opening", input: "---", body: "---", wantErr: true},
		{name: "unclosed", input: "---\nname: test\n", body: "---\nname: test\n", wantErr: true},
		{name: "prefix closing", input: "---\nname: test\n---suffix\n", body: "---\nname: test\n---suffix\n", wantErr: true},
		{name: "indented closing", input: "---\nscript: |\n  ---\n", body: "---\nscript: |\n  ---\n", wantErr: true},
		{name: "bare CR closing", input: "---\nname: test\n---\r", body: "---\nname: test\n---\r", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := ExtractRawFrontmatter(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, wantErr = %v", err, tt.wantErr)
			}
			if got.YAML != tt.yaml || got.Body != tt.body || got.HasFrontmatter != tt.present {
				t.Errorf("raw = %#v, want YAML %q, body %q, present %v", got, tt.yaml, tt.body, tt.present)
			}
			if err != nil && err.Error() != "unclosed frontmatter (missing closing ---)" {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestRawFrontmatterExactBodySuffix(t *testing.T) {
	t.Parallel()
	for _, opening := range []string{"---\n", "--- \t\r\n"} {
		for _, closing := range []string{"---\n", "--- \t\r\n", "---", "--- \t"} {
			for _, body := range []string{"", "Body\n", "---inline\n", "葉\r\n\t"} {
				if closing[len(closing)-1] != '\n' && body != "" {
					continue
				}
				input := opening + "name: test\n" + closing + body
				raw, err := ExtractRawFrontmatter(input)
				if err != nil {
					t.Fatal(err)
				}
				if want := closing[3:] + body; raw.Body != want {
					t.Errorf("input %q: body %q, want exact suffix %q", input, raw.Body, want)
				}
				if raw.YAML != "name: test\n" || !raw.HasFrontmatter {
					t.Errorf("input %q: unexpected extraction %#v", input, raw)
				}
			}
		}
	}
}
