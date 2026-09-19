package thermal

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
)

var projectKeyCache sync.Map // recorded path -> normalized project key

// ProjectDisplayNames turns project keys into the labels a person reads. The
// last path segment is the default; when two keys share that segment, the
// shortest distinguishing tail is added in brackets, which is why two
// directories called "app" do not collapse into one row label.
//
// It lives here rather than in the renderer so the terminal reports and the
// dashboard cannot disagree about what a project is called.
func ProjectDisplayNames(paths []string) map[string]string {
	tails := make([][]string, len(paths))
	for i, p := range paths {
		tails[i] = strings.Split(strings.Trim(filepath.ToSlash(p), "/"), "/")
	}

	out := make(map[string]string, len(paths))
	for i, parts := range tails {
		if len(parts) == 0 {
			continue
		}
		base := parts[len(parts)-1]
		name := base
		for n := 1; n <= len(parts); n++ {
			candidate := strings.Join(parts[len(parts)-n:], "/")
			shared := false
			for j, other := range tails {
				if i == j || n > len(other) {
					continue
				}
				if strings.Join(other[len(other)-n:], "/") == candidate {
					shared = true
					break
				}
			}
			if !shared {
				if n > 1 {
					name = base + " (" + parts[len(parts)-n] + ")"
				}
				break
			}
		}
		out[paths[i]] = name
	}
	return out
}

// ProjectKey normalizes a recorded directory to the nearest enclosing git
// repository root. Subdirectories inside one repository collapse to a single
// project, which is how people think about their work. Symlinks resolve first,
// so the same directory reached through a symlinked parent counts once rather
// than splitting into two projects. When the path has no repository above it,
// or does not exist on disk, the cleaned path is used as-is. Results are
// cached because loaders call this per session.
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
	// Resolve symlinks so /home/user/repo and /mnt/repo/symlink/repo, which can
	// be the same directory, produce one project.
	if resolved, err := filepath.EvalSymlinks(clean); err == nil {
		clean = resolved
	}
	for dir := clean; ; {
		if isRepoRoot(dir) {
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

// isRepoRoot reports whether dir holds a real repository marker. A .git
// directory must contain HEAD, so a stray empty .git directory does not
// swallow every path below it. A .git file marks a worktree or submodule.
func isRepoRoot(dir string) bool {
	info, err := os.Stat(filepath.Join(dir, ".git"))
	if err != nil {
		return false
	}
	if !info.IsDir() {
		return true
	}
	_, err = os.Stat(filepath.Join(dir, ".git", "HEAD"))
	return err == nil
}
