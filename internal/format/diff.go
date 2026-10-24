package format

import (
	"bytes"
	"fmt"
	"strings"
)

// Diff renders a changed-only preview between original and formatted content.
// Identical content produces no output; otherwise the actual filename headers
// are followed by one "- <old>" or "+ <new>" line per edit. Equal lines are
// omitted, unmatched intervals render deletions before insertions, and
// alignment uses Hirschberg's linear-space LCS so single insertions or
// deletions never mislabel the unchanged suffix.
func Diff(original, formatted, filename string) string {
	if original == formatted {
		return ""
	}

	var buf bytes.Buffer
	fmt.Fprintf(&buf, "--- %s\n", filename)
	fmt.Fprintf(&buf, "+++ %s (formatted)\n", filename)
	writeLineEdits(&buf, strings.Split(original, "\n"), strings.Split(formatted, "\n"))
	return buf.String()
}

// writeLineEdits aligns two line slices, keeping the strings.Split newline
// representation: terminal empty fields and CR bytes are ordinary lines.
func writeLineEdits(buf *bytes.Buffer, origLines, fmtLines []string) {
	prefix := commonPrefixLength(origLines, fmtLines)
	origLines, fmtLines = origLines[prefix:], fmtLines[prefix:]
	suffix := commonSuffixLength(origLines, fmtLines)
	origLines, fmtLines = origLines[:len(origLines)-suffix], fmtLines[:len(fmtLines)-suffix]
	alignEdits(buf, origLines, fmtLines)
}

// alignEdits writes the edit operations for an interval whose equal bordering
// anchors were already trimmed, using Hirschberg's divide-and-conquer LCS.
func alignEdits(buf *bytes.Buffer, origLines, fmtLines []string) {
	switch {
	case len(origLines) == 0:
		writeOps(buf, "+", fmtLines)
	case len(origLines) == 1:
		alignSingle(buf, origLines[0], fmtLines)
	default:
		middle := len(origLines) / 2
		split := splitIndex(origLines[:middle], origLines[middle:], fmtLines)
		alignEdits(buf, origLines[:middle], fmtLines[:split])
		alignEdits(buf, origLines[middle:], fmtLines[split:])
	}
}

// alignSingle aligns one deletion candidate. A matching anchor splits the
// insertions around the omitted equal line; without one the whole interval is
// a single unmatched range rendered as deletion then insertion.
func alignSingle(buf *bytes.Buffer, line string, others []string) {
	for index, other := range others {
		if other == line {
			writeOps(buf, "+", others[:index])
			writeOps(buf, "+", others[index+1:])
			return
		}
	}
	writeOps(buf, "-", []string{line})
	writeOps(buf, "+", others)
}

// splitIndex partitions fmtLines where the two halves of origLines match the
// most lines combined, taking the earliest maximum so repeated anchors align
// deterministically. Rows go out of scope on return, keeping auxiliary
// storage linear.
func splitIndex(origPrefix, origSuffix, fmtLines []string) int {
	forward := lcsPrefixRow(origPrefix, fmtLines)
	suffix := lcsSuffixRow(origSuffix, fmtLines)
	split, best := 0, -1
	for j := 0; j <= len(fmtLines); j++ {
		if total := forward[j] + suffix[j]; total > best {
			best, split = total, j
		}
	}
	return split
}

// lcsPrefixRow returns the final row of the LCS length table between lines and
// every prefix of against, rolling one row of len(against)+1 counts.
func lcsPrefixRow(lines, against []string) []int {
	row := make([]int, len(against)+1)
	for i := 1; i <= len(lines); i++ {
		diagonal := 0
		for j := 1; j <= len(against); j++ {
			above := row[j]
			switch {
			case lines[i-1] == against[j-1]:
				row[j] = diagonal + 1
			case row[j-1] > above:
				row[j] = row[j-1]
			}
			diagonal = above
		}
	}
	return row
}

// lcsSuffixRow mirrors lcsPrefixRow for line suffixes: row[j] is the LCS
// length between lines and against[j:], walking both slices from the end so
// no reversed copies are needed.
func lcsSuffixRow(lines, against []string) []int {
	row := make([]int, len(against)+1)
	for i := len(lines) - 1; i >= 0; i-- {
		diagonal := 0
		for j := len(against) - 1; j >= 0; j-- {
			above := row[j]
			switch {
			case lines[i] == against[j]:
				row[j] = diagonal + 1
			case row[j+1] > above:
				row[j] = row[j+1]
			}
			diagonal = above
		}
	}
	return row
}

// commonPrefixLength counts equal leading lines.
func commonPrefixLength(a, b []string) int {
	prefix := 0
	for prefix < len(a) && prefix < len(b) && a[prefix] == b[prefix] {
		prefix++
	}
	return prefix
}

// commonSuffixLength counts equal trailing lines.
func commonSuffixLength(a, b []string) int {
	suffix := 0
	for suffix < len(a) && suffix < len(b) && a[len(a)-1-suffix] == b[len(b)-1-suffix] {
		suffix++
	}
	return suffix
}

// writeOps renders one edit line per entry. Empty lines stay visible as bare
// "- "/" + " markers so blank-line edits do not vanish into the header.
func writeOps(buf *bytes.Buffer, marker string, lines []string) {
	for _, line := range lines {
		buf.WriteString(marker)
		buf.WriteByte(' ')
		buf.WriteString(line)
		buf.WriteByte('\n')
	}
}
