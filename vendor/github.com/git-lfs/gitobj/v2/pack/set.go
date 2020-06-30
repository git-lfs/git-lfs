package pack

import (
	"fmt"
	"hash"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/git-lfs/gitobj/v2/errors"
)

// Set allows access of objects stored across a set of packfiles.
type Set struct {
	// m maps the leading byte of a SHA-1 object name to a set of packfiles
	// that might contain that object, in order of which packfile is most
	// likely to contain that object.
	m map[byte][]*Packfile

	// closeFn is a function that is run by Close(), designated to free
	// resources held by the *Set, like open packfiles.
	closeFn func() error
}

var (
	// nameRe is a regular expression that matches the basename of a
	// filepath that is a packfile.
	//
	// It includes one matchgroup, which is the SHA-1 name of the pack.
	nameRe = regexp.MustCompile(`^(.*)\.pack$`)
)

// NewSet creates a new *Set of all packfiles found in a given object database's
// root (i.e., "/path/to/repo/.git/objects").
//
// It finds all packfiles in the "pack" subdirectory, and instantiates a *Set
// containing them. If there was an error parsing the packfiles in that
// directory, or the directory was otherwise unable to be observed, NewSet
// returns that error.
func NewSet(db string, algo hash.Hash) (*Set, error) {
	pd := filepath.Join(db, "pack")

	paths, err := filepath.Glob(filepath.Join(escapeGlobPattern(pd), "*.pack"))
	if err != nil {
		return nil, err
	}

	packs := make([]*Packfile, 0, len(paths))

	for _, path := range paths {
		submatch := nameRe.FindStringSubmatch(filepath.Base(path))
		if len(submatch) != 2 {
			continue
		}

		name := submatch[1]

		idxf, err := os.Open(filepath.Join(pd, fmt.Sprintf("%s.idx", name)))
		if err != nil {
			// We have a pack (since it matched the regex), but the
			// index is missing or unusable.  Skip this pack and
			// continue on with the next one, as Git does.
			if idxf != nil {
				// In the unlikely event that we did open a
				// file, close it, but discard any error in
				// doing so.
				idxf.Close()
			}
			continue
		}

		packf, err := os.Open(filepath.Join(pd, fmt.Sprintf("%s.pack", name)))
		if err != nil {
			return nil, err
		}

		pack, err := DecodePackfile(packf, algo)
		if err != nil {
			return nil, err
		}

		idx, err := DecodeIndex(idxf, algo)
		if err != nil {
			return nil, err
		}

		pack.idx = idx

		packs = append(packs, pack)
	}
	return NewSetPacks(packs...), nil
}

// globEscapes uses these escapes because filepath.Glob does not understand
// backslash escapes on Windows.
var globEscapes = map[string]string{
	"*": "[*]",
	"?": "[?]",
	"[": "[[]",
}

func escapeGlobPattern(s string) string {
	for char, escape := range globEscapes {
		s = strings.Replace(s, char, escape, -1)
	}
	return s
}

// NewSetPacks creates a new *Set from the given packfiles.
func NewSetPacks(packs ...*Packfile) *Set {
	m := make(map[byte][]*Packfile)

	for i := 0; i < 256; i++ {
		n := byte(i)

		for j := 0; j < len(packs); j++ {
			pack := packs[j]

			var count uint32
			if n == 0 {
				count = pack.idx.fanout[n]
			} else {
				count = pack.idx.fanout[n] - pack.idx.fanout[n-1]
			}

			if count > 0 {
				m[n] = append(m[n], pack)
			}
		}

		sort.Slice(m[n], func(i, j int) bool {
			ni := m[n][i].idx.fanout[n]
			nj := m[n][j].idx.fanout[n]

			return ni > nj
		})
	}

	return &Set{
		m:    m,
		closeFn: func() error {
			for _, pack := range packs {
				if err := pack.Close(); err != nil {
					return err
				}
			}
			return nil
		},
	}
}

// Close closes all open packfiles, returning an error if one was encountered.
func (s *Set) Close() error {
	if s.closeFn == nil {
		return nil
	}
	return s.closeFn()
}

// Object opens (but does not unpack, or, apply the delta-base chain) a given
// object in the first packfile that matches it.
//
// Object searches packfiles contained in the set in order of how many objects
// they have that begin with the first by of the given SHA-1 "name", in
// descending order.
//
// If the object was unable to be found in any of the packfiles, (nil,
// ErrNotFound) will be returned.
//
// If there was otherwise an error opening the object for reading from any of
// the packfiles, it will be returned, and no other packfiles will be searched.
//
// Otherwise, the object will be returned without error.
func (s *Set) Object(name []byte) (*Object, error) {
	return s.each(name, func(p *Packfile) (*Object, error) {
		return p.Object(name)
	})
}

// iterFn is a function that takes a given packfile and opens an object from it.
type iterFn func(p *Packfile) (o *Object, err error)

// each executes the given iterFn "fn" on each Packfile that has any objects
// beginning with a prefix of the SHA-1 "name", in order of which packfiles have
// the most objects beginning with that prefix.
//
// If any invocation of "fn" returns a non-nil error, it will either be a)
// returned immediately, if the error is not ErrIsNotFound, or b) continued
// immediately, if the error is ErrNotFound.
//
// If no packfiles match the given file, return errors.NoSuchObject, along with
// no object.
func (s *Set) each(name []byte, fn iterFn) (*Object, error) {
	var key byte
	if len(name) > 0 {
		key = name[0]
	}

	for _, pack := range s.m[key] {
		o, err := fn(pack)
		if err != nil {
			if IsNotFound(err) {
				continue
			}
			return nil, err
		}
		return o, nil
	}

	return nil, errors.NoSuchObject(name)
}
