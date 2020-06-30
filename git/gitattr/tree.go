package gitattr

import (
	"strings"

	"github.com/git-lfs/gitobj/v2"
)

// Tree represents the .gitattributes file at one layer of the tree in a Git
// repository.
type Tree struct {
	// Lines are the lines of the .gitattributes at this level of the tree.
	Lines []*Line
	// Children are the named child directories in the repository.
	Children map[string]*Tree
}

// New constructs a *Tree starting at the given tree "t" and reading objects
// from the given ObjectDatabase. If a tree was not able to be read, an error
// will be propagated up accordingly.
func New(db *gitobj.ObjectDatabase, t *gitobj.Tree) (*Tree, error) {
	children := make(map[string]*Tree)
	lines, _, err := linesInTree(db, t)
	if err != nil {
		return nil, err
	}

	for _, entry := range t.Entries {
		if entry.Type() != gitobj.TreeObjectType {
			continue
		}

		// For every entry in the current tree, parse its sub-trees to
		// see if they might contain a .gitattributes.
		t, err := db.Tree(entry.Oid)
		if err != nil {
			return nil, err
		}

		at, err := New(db, t)
		if err != nil {
			return nil, err
		}

		if len(at.Children) > 0 || len(at.Lines) > 0 {
			// Only include entries that have either (1) a
			// .gitattributes in their tree, or (2) a .gitattributes
			// in a sub-tree.
			children[entry.Name] = at
		}
	}

	return &Tree{
		Lines:    lines,
		Children: children,
	}, nil
}

// linesInTree parses a given tree's .gitattributes and returns a slice of lines
// in that .gitattributes, or an error. If no .gitattributes blob was found,
// return nil.
func linesInTree(db *gitobj.ObjectDatabase, t *gitobj.Tree) ([]*Line, string, error) {
	var at int = -1
	for i, e := range t.Entries {
		if e.Name == ".gitattributes" {
			at = i
			break
		}
	}

	if at < 0 {
		return nil, "", nil
	}

	blob, err := db.Blob(t.Entries[at].Oid)
	if err != nil {
		return nil, "", err
	}
	defer blob.Close()

	return ParseLines(blob.Contents)
}

// Applied returns a slice of attributes applied to the given path, relative to
// the receiving tree. It traverse through sub-trees in a topological ordering,
// if there are relevant .gitattributes matching that path.
func (t *Tree) Applied(to string) []*Attr {
	var attrs []*Attr
	for _, line := range t.Lines {
		if line.Pattern.Match(to) {
			attrs = append(attrs, line.Attrs...)
		}
	}

	splits := strings.SplitN(to, "/", 2)
	if len(splits) == 2 {
		car, cdr := splits[0], splits[1]
		if child, ok := t.Children[car]; ok {
			attrs = append(attrs, child.Applied(cdr)...)
		}
	}

	return attrs
}
