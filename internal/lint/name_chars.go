package lint

import (
	"unicode"
	"unicode/utf8"
)

// nameCharacterProblem reports the first character class that makes a skill or
// command name unloadable by Claude Code v2.1.299+, or "" when the name is
// loadable. Claude Code stops offering a skill or command whose file/folder
// name or frontmatter name holds hidden, direction-changing, or blank-like
// characters, '<' or '>', or exceeds 256 characters.
func nameCharacterProblem(name string) string {
	for _, r := range name {
		if isDirectionChangingNameRune(r) {
			return "a direction-changing character"
		}
		if isHiddenNameRune(r) {
			return "a hidden character"
		}
		if isBlankLikeNameRune(r) {
			return "a blank-like character"
		}
		if r == '<' || r == '>' {
			return "'<' or '>'"
		}
	}
	if utf8.RuneCountInString(name) > 256 {
		return "more than 256 characters"
	}
	return ""
}

func isHiddenNameRune(r rune) bool {
	if unicode.Is(unicode.Cf, r) || unicode.Is(unicode.Co, r) || unicode.Is(unicode.Cn, r) || unicode.Is(unicode.Cc, r) {
		return true
	}
	if r == 0x2028 || r == 0x2029 {
		return true
	}
	return r >= 0x007F && r <= 0x009F
}

func isDirectionChangingNameRune(r rune) bool {
	return (r >= 0x200E && r <= 0x200F) || (r >= 0x202A && r <= 0x202E) || (r >= 0x2066 && r <= 0x2069)
}

func isBlankLikeNameRune(r rune) bool {
	if r == ' ' {
		return false
	}
	return unicode.Is(unicode.Zs, r) || unicode.Is(unicode.Zl, r) || unicode.Is(unicode.Zp, r)
}
