package crossfile

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
	batch.referenceSnapshot = &referenceSnapshot{root: root, entries: v.readReferenceEntries(root)}
	return &batch
}

func (v *CrossFileValidator) referenceEntries(root string) []referenceEntry {
	if v.referenceSnapshot != nil && v.referenceSnapshot.root == root {
		return v.referenceSnapshot.entries
	}
	return v.readReferenceEntries(root)
}

func (v *CrossFileValidator) readReferenceEntries(root string) []referenceEntry {
	var entries []referenceEntry
	files, _ := v.referenceDiscoverer(root).DiscoverFilesWithRegistry(referenceRegistry())
	for _, file := range files {
		relPath, contents := file.RelPath, file.Contents
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
