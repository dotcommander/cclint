package crossfile

import (
	"slices"

	"github.com/dotcommander/cclint/internal/discovery"
)

// ReferencePolicy carries the invocation's discovery policy into reference scans.
// Its zero value excludes descendant symlinks and permits only root-contained paths.
type ReferencePolicy struct {
	Exclude        []string
	FollowSymlinks bool
}

// WithReferencePolicy returns a validator with independent reference discovery
// options. Changing policy invalidates any snapshot selected under the old policy.
func (v *CrossFileValidator) WithReferencePolicy(policy ReferencePolicy) *CrossFileValidator {
	configured := *v
	policy.Exclude = slices.Clone(policy.Exclude)
	configured.referencePolicy = policy
	configured.referenceSnapshot = nil
	return &configured
}

func (v *CrossFileValidator) referenceDiscoverer(root string) *discovery.FileDiscovery {
	return discovery.NewFileDiscovery(root, v.referencePolicy.FollowSymlinks).
		WithExclude(v.referencePolicy.Exclude)
}

func referenceRegistry() []discovery.FileTypeEntry {
	return []discovery.FileTypeEntry{{Type: discovery.FileTypeSkill, Patterns: referenceFileGlobs}}
}

// discoverReferenceFiles retains the standalone default-policy selection helper.
func discoverReferenceFiles(root string) []string {
	v := &CrossFileValidator{}
	files, _ := v.referenceDiscoverer(root).DiscoverPathsWithRegistry(referenceRegistry())
	var paths []string
	for _, file := range files {
		paths = append(paths, file.RelPath)
	}
	return paths
}
