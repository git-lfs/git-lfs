package localstorage

import (
	"os"
	"path/filepath"

	"github.com/rubyist/tracerx"
)

// AllObjects returns a slice of the the objects stored in this LocalStorage
// object. This does not necessarily mean referenced by commits, just stored.
// Note: reports final SHA only, extensions are ignored.
func (s *LocalStorage) AllObjects() []Object {
	objects := make([]Object, 0, 100)
	for o := range s.ScanObjectsChan() {
		objects = append(objects, o)
	}
	return objects
}

// ScanObjectsChan returns a channel of all the objects stored in this
// LocalStorage object. This does not necessarily mean referenced by commits,
// just stored. You should not alter the store until this channel is closed.
// Note: reports final SHA only, extensions are ignored.
func (s *LocalStorage) ScanObjectsChan() <-chan Object {
	ch := make(chan Object, chanBufSize)

	go func() {
		defer close(ch)
		scanObjects(s.RootDir, ch)
	}()

	return ch
}

func scanObjects(dir string, ch chan<- Object) {
	dirf, err := os.Open(dir)
	if err != nil {
		return
	}
	defer dirf.Close()

	direntries, err := dirf.Readdir(0)
	if err != nil {
		tracerx.Printf("Problem with Readdir in %q: %s", dir, err)
		return
	}

	for _, dirfi := range direntries {
		if dirfi.IsDir() {
			subpath := filepath.Join(dir, dirfi.Name())
			scanObjects(subpath, ch)
		} else {
			// Make sure it's really an object file & not .DS_Store etc
			if oidRE.MatchString(dirfi.Name()) {
				ch <- Object{dirfi.Name(), dirfi.Size()}
			}
		}
	}
}
