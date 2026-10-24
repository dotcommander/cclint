package discovery

import "github.com/bmatcuk/doublestar/v4"

// IsExcluded applies configured globs to a slash-separated, root-relative path.
// Invalid patterns are ignored, matching the discovery exclusion policy.
func IsExcluded(relPath string, patterns []string) bool {
	for _, pattern := range patterns {
		if matched, err := doublestar.Match(pattern, relPath); err == nil && matched {
			return true
		}
	}
	return false
}
