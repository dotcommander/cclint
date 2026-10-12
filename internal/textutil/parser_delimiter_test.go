package textutil

import "testing"

func TestParseFrontmatterCompleteDelimiterLines(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name, input, body, description string
		wantErr                        bool
	}{
		{name: "inline opener", input: "---inline\nname: test\n---\nBody", body: "---inline\nname: test\n---\nBody"},
		{name: "unclosed", input: "---\nname: test\n", wantErr: true},
		{name: "only opener", input: "---", wantErr: true},
		{name: "prefix closer", input: "---\nname: test\n---inline\n", wantErr: true},
		{name: "invalid YAML", input: "---\nname: [\n---\nBody", wantErr: true},
		{name: "embedded dashes", input: "---\ndescription: \"before---after\"\n---\nBody", body: "\nBody", description: "before---after"},
		{name: "block dashes", input: "---\ndescription: |\n  before---after\n  ---\n---\nBody", body: "\nBody", description: "before---after\n---\n"},
		{name: "CRLF", input: "---\r\ndescription: test\r\n--- \t\r\nBody\r\n", body: " \t\r\nBody\r\n", description: "test"},
		{name: "EOF closer", input: "---\ndescription: test\n---", description: "test"},
		{name: "EOF closer whitespace", input: "---\ndescription: test\n--- \t", body: " \t", description: "test"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := ParseYAMLFrontmatter(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, wantErr = %v", err, tt.wantErr)
			}
			if err != nil {
				return
			}
			if got.Body != tt.body {
				t.Errorf("body = %q, want %q", got.Body, tt.body)
			}
			if tt.description != "" && got.Data["description"] != tt.description {
				t.Errorf("description = %#v, want %q", got.Data["description"], tt.description)
			}
			if tt.name == "inline opener" && (got.Data == nil || len(got.Data) != 0) {
				t.Errorf("ordinary Markdown needs a non-nil empty map, got %#v", got.Data)
			}
		})
	}
}
