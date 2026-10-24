package format

import "testing"

func TestDiffExactPreviews(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		original  string
		formatted string
		filename  string
		want      string
	}{
		{
			name:      "identical content is empty",
			original:  "---\nname: test\n---\nContent.\n",
			formatted: "---\nname: test\n---\nContent.\n",
			filename:  "test.md",
			want:      "",
		},
		{
			name:      "inserted line keeps unchanged suffix",
			original:  "line1\nline2\n",
			formatted: "line1\ninserted\nline2\n",
			filename:  "report.md",
			want:      "--- report.md\n+++ report.md (formatted)\n+ inserted\n",
		},
		{
			name:      "deleted line keeps unchanged suffix",
			original:  "line1\nremoved\nline2\n",
			formatted: "line1\nline2\n",
			filename:  "report.md",
			want:      "--- report.md\n+++ report.md (formatted)\n- removed\n",
		},
		{
			name:      "replacement deletes before insert",
			original:  "old\n",
			formatted: "new\n",
			filename:  "swap.md",
			want:      "--- swap.md\n+++ swap.md (formatted)\n- old\n+ new\n",
		},
		{
			name:      "middle change does not cascade over suffix",
			original:  "1\n2\n3\n4\n5\n",
			formatted: "1\n2\nx\n4\n5\n",
			filename:  "chain.md",
			want:      "--- chain.md\n+++ chain.md (formatted)\n- 3\n+ x\n",
		},
		{
			name:      "repeated anchor takes earliest split tie",
			original:  "x\nm\n",
			formatted: "m\nx\n",
			filename:  "anchor.md",
			want:      "--- anchor.md\n+++ anchor.md (formatted)\n- x\n+ x\n",
		},
		{
			name:      "blank-line deletion stays visible",
			original:  "a\n\nb\n",
			formatted: "a\nb\n",
			filename:  "blank.md",
			want:      "--- blank.md\n+++ blank.md (formatted)\n- \n",
		},
		{
			name:      "blank-line insertion stays visible",
			original:  "a\nb\n",
			formatted: "a\n\nb\n",
			filename:  "blank.md",
			want:      "--- blank.md\n+++ blank.md (formatted)\n+ \n",
		},
		{
			name:      "terminal newline stays visible",
			original:  "a",
			formatted: "a\n",
			filename:  "newline.md",
			want:      "--- newline.md\n+++ newline.md (formatted)\n+ \n",
		},
		{
			name:      "carriage return stays in line content",
			original:  "a\r\n",
			formatted: "a\n",
			filename:  "cr.md",
			want:      "--- cr.md\n+++ cr.md (formatted)\n- a\r\n+ a\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := Diff(tt.original, tt.formatted, tt.filename); got != tt.want {
				t.Fatalf("Diff() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDiffAnchorsUnchangedRegions(t *testing.T) {
	t.Parallel()

	original := "keep head\nstale\nkeep middle\nkeep tail\n"
	formatted := "keep head\nkeep middle\nkeep tail\n"
	want := "--- body.md\n+++ body.md (formatted)\n- stale\n"

	if got := Diff(original, formatted, "body.md"); got != want {
		t.Fatalf("Diff() = %q, want %q", got, want)
	}
}
