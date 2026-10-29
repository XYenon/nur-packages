package fixture

import (
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

type resource struct {
	path    string
	pending int
	changed *sync.Cond
}

var resources = struct {
	sync.Mutex
	groups map[*resource]bool
}{groups: map[*resource]bool{}}

// A task joins only groups whose registered paths contain its file operations.
// Register before starting work; waiting is per resource, not per package or goroutine.
func Begin(paths ...string) func() {
	resources.Lock()
	var joined []*resource
	for group := range resources.groups {
		for _, path := range paths {
			path = filepath.Clean(path)
			if path == group.path || strings.HasPrefix(path, group.path+string(filepath.Separator)) {
				group.pending++
				joined = append(joined, group)
				break
			}
		}
	}
	resources.Unlock()
	return func() {
		resources.Lock()
		defer resources.Unlock()
		for _, group := range joined {
			group.pending--
			group.changed.Broadcast()
		}
	}
}

func TempDir(t testing.TB) string {
	t.Helper()
	path := t.TempDir()
	group := &resource{path: filepath.Clean(path)}
	resources.Lock()
	group.changed = sync.NewCond(&resources.Mutex)
	resources.groups[group] = true
	resources.Unlock()
	t.Cleanup(func() {
		resources.Lock()
		defer resources.Unlock()
		for group.pending != 0 {
			group.changed.Wait()
		}
		delete(resources.groups, group)
	})
	return path
}
