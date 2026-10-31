package discovery

// DiscoverFilesOfTypes scans only selected types, in canonical registry order.
// Callers needing cross-type context should continue to use DiscoverFiles.
func (fd *FileDiscovery) DiscoverFilesOfTypes(types map[FileType]bool) ([]File, error) {
	registry := make([]FileTypeEntry, 0, len(types))
	for _, entry := range DefaultFileTypes {
		if types[entry.Type] {
			registry = append(registry, entry)
		}
	}
	return fd.DiscoverFilesWithRegistry(registry)
}
