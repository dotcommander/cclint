package discovery

import "path/filepath"

// AllowsPath applies discovery exclusions and symlink containment to an existing
// root-relative path. It does not read file contents or determine component type.
func (fd *FileDiscovery) AllowsPath(relative string) bool {
	if !filepath.IsLocal(relative) {
		return false
	}
	relative = filepath.ToSlash(filepath.Clean(relative))
	return !IsExcluded(relative, fd.exclude) && fd.allowedPath(filepath.Join(fd.rootPath, relative))
}
