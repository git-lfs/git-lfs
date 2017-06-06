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

		var typ ObjectType
		switch mode & syscall.S_IFMT {
		case syscall.S_IFREG:
			typ = BlobObjectType
		case syscall.S_IFDIR:
			typ = TreeObjectType
		case syscall.S_IFLNK:
			typ = BlobObjectType
		default:
			if mode == 0xe000 {
				// Mode 0xe000, or a gitlink, has no formal
				// filesystem (`syscall.S_IF<t>`) equivalent.
				//
				// Safeguard that catch here, or otherwise
				// panic.
				typ = CommitObjectType
			} else {
				panic(fmt.Sprintf("git/odb: unknown object type: %q %q", modes, fname))
			}
		}

		entries = append(entries, &TreeEntry{
			Name:     fname,
			Type:     typ,
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
	// Type is the type of entry (either blob: BlobObjectType, or a
	// sub-tree: TreeObjectType).
	Type ObjectType
	// Oid is the object ID for this tree entry.
	Oid []byte
	// Filemode is the filemode of this tree entry on disk.
	Filemode int32
}
