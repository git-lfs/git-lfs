package gitobj

import (
	"bufio"
	"bytes"
	"fmt"
	"hash"
	"io"
	"sort"
	"strconv"
	"strings"

	"github.com/git-lfs/gitobj/v2/pack"
)

// We define these here instead of using the system ones because not all
// operating systems use the traditional values.  For example, zOS uses
// different values.
const (
	sIFMT      = int32(0170000)
	sIFREG     = int32(0100000)
	sIFDIR     = int32(0040000)
	sIFLNK     = int32(0120000)
	sIFGITLINK = int32(0160000)
)

// Tree encapsulates a Git tree object.
type Tree struct {
	// Entries is the list of entries held by this tree.
	Entries []*TreeEntry
}

// Type implements Object.ObjectType by returning the correct object type for
// Trees, TreeObjectType.
func (t *Tree) Type() ObjectType { return TreeObjectType }

// Decode implements Object.Decode and decodes the uncompressed tree being
// read. It returns the number of uncompressed bytes being consumed off of the
// stream, which should be strictly equal to the size given.
//
// If any error was encountered along the way, that will be returned, along with
// the number of bytes read up to that point.
func (t *Tree) Decode(hash hash.Hash, from io.Reader, size int64) (n int, err error) {
	hashlen := hash.Size()
	buf := bufio.NewReader(from)

	var entries []*TreeEntry
	for {
		modes, err := buf.ReadString(' ')
		if err != nil {
			if err == io.EOF {
				break
			}
			return n, err
		}
		n += len(modes)
		modes = strings.TrimSuffix(modes, " ")

		mode, _ := strconv.ParseInt(modes, 8, 32)

		fname, err := buf.ReadString('\x00')
		if err != nil {
			return n, err
		}
		n += len(fname)
		fname = strings.TrimSuffix(fname, "\x00")

		var sha [pack.MaxHashSize]byte
		if _, err = io.ReadFull(buf, sha[:hashlen]); err != nil {
			return n, err
		}
		n += hashlen

		entries = append(entries, &TreeEntry{
			Name:     fname,
			Oid:      sha[:hashlen],
			Filemode: int32(mode),
		})
	}

	t.Entries = entries

	return n, nil
}

// Encode encodes the tree's contents to the given io.Writer, "w". If there was
// any error copying the tree's contents, that error will be returned.
//
// Otherwise, the number of bytes written will be returned.
func (t *Tree) Encode(to io.Writer) (n int, err error) {
	const entryTmpl = "%s %s\x00%s"

	for _, entry := range t.Entries {
		fmode := strconv.FormatInt(int64(entry.Filemode), 8)

		ne, err := fmt.Fprintf(to, entryTmpl,
			fmode,
			entry.Name,
			entry.Oid)

		if err != nil {
			return n, err
		}

		n = n + ne
	}
	return
}

// Merge performs a merge operation against the given set of `*TreeEntry`'s by
// either replacing existing tree entries of the same name, or appending new
// entries in sub-tree order.
//
// It returns a copy of the tree, and performs the merge in O(n*log(n)) time.
func (t *Tree) Merge(others ...*TreeEntry) *Tree {
	unseen := make(map[string]*TreeEntry)

	// Build a cache of name to *TreeEntry.
	for _, other := range others {
		unseen[other.Name] = other
	}

	// Map the existing entries ("t.Entries") into a new set by either
	// copying an existing entry, or replacing it with a new one.
	entries := make([]*TreeEntry, 0, len(t.Entries))
	for _, entry := range t.Entries {
		if other, ok := unseen[entry.Name]; ok {
			entries = append(entries, other)
			delete(unseen, entry.Name)
		} else {
			oid := make([]byte, len(entry.Oid))
			copy(oid, entry.Oid)

			entries = append(entries, &TreeEntry{
				Filemode: entry.Filemode,
				Name:     entry.Name,
				Oid:      oid,
			})
		}
	}

	// For all the items we haven't replaced into the new set, append them
	// to the entries.
	for _, remaining := range unseen {
		entries = append(entries, remaining)
	}

	// Call sort afterwords, as a tradeoff between speed and spacial
	// complexity. As a future point of optimization, adding new elements
	// (see: above) could be done as a linear pass of the "entries" set.
	//
	// In order to do that, we must have a constant-time lookup of both
	// entries in the existing and new sets. This requires building a
	// map[string]*TreeEntry for the given "others" as well as "t.Entries".
	//
	// Trees can be potentially large, so trade this spacial complexity for
	// an O(n*log(n)) sort.
	sort.Sort(SubtreeOrder(entries))

	return &Tree{Entries: entries}
}

// Equal returns whether the receiving and given trees are equal, or in other
// words, whether they are represented by the same SHA-1 when saved to the
// object database.
func (t *Tree) Equal(other *Tree) bool {
	if (t == nil) != (other == nil) {
		return false
	}

	if t != nil {
		if len(t.Entries) != len(other.Entries) {
			return false
		}

		for i := 0; i < len(t.Entries); i++ {
			e1 := t.Entries[i]
			e2 := other.Entries[i]

			if !e1.Equal(e2) {
				return false
			}
		}
	}
	return true
}

// TreeEntry encapsulates information about a single tree entry in a tree
// listing.
type TreeEntry struct {
	// Name is the entry name relative to the tree in which this entry is
	// contained.
	Name string
	// Oid is the object ID for this tree entry.
	Oid []byte
	// Filemode is the filemode of this tree entry on disk.
	Filemode int32
}

// Equal returns whether the receiving and given TreeEntry instances are
// identical in name, filemode, and OID.
func (e *TreeEntry) Equal(other *TreeEntry) bool {
	if (e == nil) != (other == nil) {
		return false
	}

	if e != nil {
		return e.Name == other.Name &&
			bytes.Equal(e.Oid, other.Oid) &&
			e.Filemode == other.Filemode
	}
	return true
}

// Type is the type of entry (either blob: BlobObjectType, or a sub-tree:
// TreeObjectType).
func (e *TreeEntry) Type() ObjectType {
	switch e.Filemode & sIFMT {
	case sIFREG:
		return BlobObjectType
	case sIFDIR:
		return TreeObjectType
	case sIFLNK:
		return BlobObjectType
	case sIFGITLINK:
		return CommitObjectType
	default:
		panic(fmt.Sprintf("gitobj: unknown object type: %o",
			e.Filemode))
	}
}

// IsLink returns true if the given TreeEntry is a blob which represents a
// symbolic link (i.e., with a filemode of 0120000.
func (e *TreeEntry) IsLink() bool {
	return e.Filemode & sIFMT == sIFLNK
}

// SubtreeOrder is an implementation of sort.Interface that sorts a set of
// `*TreeEntry`'s according to "subtree" order. This ordering is required to
// write trees in a correct, readable format to the Git object database.
//
// The format is as follows: entries are sorted lexicographically in byte-order,
// with subtrees (entries of Type() == gitobj.TreeObjectType) being sorted as
// if their `Name` fields ended in a "/".
//
// See: https://github.com/git/git/blob/v2.13.0/fsck.c#L492-L525 for more
// details.
type SubtreeOrder []*TreeEntry

// Len implements sort.Interface.Len() and return the length of the underlying
// slice.
func (s SubtreeOrder) Len() int { return len(s) }

// Swap implements sort.Interface.Swap() and swaps the two elements at i and j.
func (s SubtreeOrder) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

// Less implements sort.Interface.Less() and returns whether the element at "i"
// is compared as "less" than the element at "j". In other words, it returns if
// the element at "i" should be sorted ahead of that at "j".
//
// It performs this comparison in lexicographic byte-order according to the
// rules above (see SubtreeOrder).
func (s SubtreeOrder) Less(i, j int) bool {
	return s.Name(i) < s.Name(j)
}

// Name returns the name for a given entry indexed at "i", which is a C-style
// string ('\0' terminated unless it's a subtree), optionally terminated with
// '/' if it's a subtree.
//
// This is done because '/' sorts ahead of '\0', and is compatible with the
// tree order in upstream Git.
func (s SubtreeOrder) Name(i int) string {
	if i < 0 || i >= len(s) {
		return ""
	}

	entry := s[i]
	if entry == nil {
		return ""
	}

	if entry.Type() == TreeObjectType {
		return entry.Name + "/"
	}
	return entry.Name + "\x00"
}
