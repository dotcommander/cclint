package crossfile

import "os"

// referenceSnapshot holds only parsed trigger tables, not file contents. It is
// owned by one batch invocation; standalone validators never retain it.
type referenceSnapshot struct {
	root    string
	entries []referenceEntry
}

type referenceEntry struct {
	refs     []TriggerRef
	mappings []TriggerMapping
}

// WithReferenceSnapshot returns an invocation-local validator sharing the
// component indexes but not mutable reference-file state with its parent.
// Callers should discard the returned validator after the batch completes.
func (v *CrossFileValidator) WithReferenceSnapshot(root string) *CrossFileValidator {
	batch := *v
	batch.referenceSnapshot = &referenceSnapshot{root: root, entries: readReferenceEntries(root)}
	return &batch
}

func (v *CrossFileValidator) referenceEntries(root string) []referenceEntry {
	if v.referenceSnapshot != nil && v.referenceSnapshot.root == root {
		return v.referenceSnapshot.entries
	}
	return readReferenceEntries(root)
}

func readReferenceEntries(root string) []referenceEntry {
	var entries []referenceEntry
	for _, relPath := range discoverReferenceFiles(root) {
		data, err := os.ReadFile(root + "/" + relPath) //nolint:gosec // G304: path comes from controlled glob inside root
		if err != nil {
			continue
		}
		contents := string(data)
		if !IsTriggerMap(contents) {
			continue
		}
		entries = append(entries, referenceEntry{
			refs:     ParseTriggerTable(relPath, contents),
			mappings: ParseTriggerMappings(relPath, contents),
		})
	}
	return entries
}
