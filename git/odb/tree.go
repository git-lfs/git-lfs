package odb

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
	"syscall"
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
func (t *Tree) Decode(from io.Reader, size int64) (n int, err error) {
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

		var sha [20]byte
		if _, err = io.ReadFull(buf, sha[:]); err != nil {
			return n, err
		}
		n += 20

		entries = append(entries, &TreeEntry{
			Name:     fname,
			Oid:      sha[:],
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

// Type is the type of entry (either blob: BlobObjectType, or a sub-tree:
// TreeObjectType).
func (e *TreeEntry) Type() ObjectType {
	switch e.Filemode & syscall.S_IFMT {
	case syscall.S_IFREG:
		return BlobObjectType
	case syscall.S_IFDIR:
		return TreeObjectType
	case syscall.S_IFLNK:
		return BlobObjectType
	default:
		if e.Filemode == 0xe000 {
			// Mode 0xe000, or a gitlink, has no formal filesystem
			// (`syscall.S_IF<t>`) equivalent.
			//
			// Safeguard that catch here, or otherwise panic.
			return CommitObjectType
		} else {
			panic(fmt.Sprintf("git/odb: unknown object type: %o",
				e.Filemode))
		}
	}
}

// SubtreeOrder is an implementation of sort.Interface that sorts a set of
// `*TreeEntry`'s according to "subtree" order. This ordering is required to
// write trees in a correct, readable format to the Git object database.
//
// The format is as follows: entries are sorted lexicographically in byte-order,
// with subtrees (entries of Type() == git/odb.TreeObjectType) being sorted as
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

	if entry.Type() == TreeObjectType {
		return entry.Name + "/"
	}
	return entry.Name + "\x00"
}
