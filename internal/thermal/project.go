package thermal

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
)

var projectKeyCache sync.Map // recorded path -> normalized project key

// ProjectKey normalizes a recorded directory to the nearest enclosing git
// repository root. Subdirectories inside one repository collapse to a single
// project, which is how people think about their work. When the path has no
// repository above it, or does not exist on disk, the cleaned path is used
// as-is. Results are cached because loaders call this per session.
//
// Symlinks are deliberately not resolved: the key should read the same way the
// tool recorded it.
func ProjectKey(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return ""
	}
	if v, ok := projectKeyCache.Load(path); ok {
		return v.(string)
	}
	key := resolveProjectKey(path)
	projectKeyCache.Store(path, key)
	return key
}

func resolveProjectKey(path string) string {
	clean := filepath.Clean(path)
	if clean == "." {
		return ""
	}
	for dir := clean; ; {
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return clean
}
