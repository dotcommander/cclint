package lint

import (
	"path/filepath"
	"sync"

	"github.com/dotcommander/cclint/internal/discovery"
)

// DiscoveryCache retains one discovery snapshot (including errors) per cleaned
// absolute lexical root for a single lint invocation. Its zero value is ready
// to use. Callers borrow returned slices read-only; changes on disk are not
// observed until a new cache is used. Symlink aliases remain separate keys.
type DiscoveryCache struct {
	mu      sync.Mutex
	results map[string]discoveryCacheResult
}

type discoveryCacheResult struct {
	files []discovery.File
	err   error
}

// Get discovers a root once, synchronizing discovery and publication for all
// callers. A failure to resolve an absolute root is returned without caching.
func (dc *DiscoveryCache) Get(rootPath string) ([]discovery.File, error) {
	root, err := filepath.Abs(rootPath)
	if err != nil {
		return nil, err
	}
	dc.mu.Lock()
	defer dc.mu.Unlock()
	if result, ok := dc.results[root]; ok {
		return result.files, result.err
	}
	files, err := discovery.NewFileDiscovery(root, false).DiscoverFiles()
	if dc.results == nil {
		dc.results = make(map[string]discoveryCacheResult)
	}
	dc.results[root] = discoveryCacheResult{files: files, err: err}
	return files, err
}
