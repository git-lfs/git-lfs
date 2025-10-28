package gitattr

import (
	"io"
	"strings"

	"github.com/git-lfs/git-lfs/v3/errors"
	"github.com/git-lfs/git-lfs/v3/tr"
	"github.com/git-lfs/gitobj/v2"
)

// Tree represents the .gitattributes file at one layer of the tree in a Git
// repository.
type Tree struct {
	MP *MacroProcessor

	SystemAttributes *Tree

	UserAttributes *Tree
	// Lines are the lines of the .gitattributes at this level of the tree.
	Lines []Line
	// Children are the named child directories in the repository.
	Children map[string]*Tree

	RepoAttributes *Tree
}

// New constructs a *Tree starting at the given tree "t" and reading objects
// from the given ObjectDatabase. If a tree was not able to be read, an error
// will be propagated up accordingly.
func New(db *gitobj.ObjectDatabase, t *gitobj.Tree) (*Tree, error) {
	processor := NewMacroProcessor()
	return newFromGitTree(db, t, processor)
}

func newFromGitTree(db *gitobj.ObjectDatabase, t *gitobj.Tree, mp *MacroProcessor) (*Tree, error) {
	children := make(map[string]*Tree)
	tree, err := linesInTree(db, t, mp)

	if err != nil {
		return nil, err
	}

	for _, entry := range t.Entries {
		if entry.Type() != gitobj.TreeObjectType {
			continue
		}

		// For every entry in the current tree, parse its sub-trees to
		// see if they might contain a .gitattributes.
		subGitTree, err := db.Tree(entry.Oid)
		if err != nil {
			return nil, err
		}

		subTree, err := newFromGitTree(db, subGitTree, mp)
		if err != nil {
			return nil, err
		}

		if len(subTree.Children) > 0 || len(subTree.Lines) > 0 {
			// Only include entries that have either (1) a
			// .gitattributes in their tree, or (2) a .gitattributes
			// in a sub-tree.
			children[entry.Name] = subTree
		}
	}
	tree.Children = children

	return tree, nil
}

func NewFromReader(mp *MacroProcessor, rdr io.Reader) (*Tree, error) {
	lines, _, err := ParseLines(rdr)
	if err != nil {
		return nil, err
	}
	return &Tree{
		MP:    mp,
		Lines: lines,
	}, nil
}

// linesInTree parses a given tree's .gitattributes and returns a slice of lines
// in that .gitattributes, or an error. If no .gitattributes blob was found,
// return nil.
func linesInTree(db *gitobj.ObjectDatabase, t *gitobj.Tree, mp *MacroProcessor) (*Tree, error) {
	var at int = -1
	for i, e := range t.Entries {
		if e.Name == ".gitattributes" {
			if e.IsLink() {
				return nil, errors.New(tr.Tr.Get("expected '.gitattributes' to be a file, got a symbolic link"))
			}
			at = i
			break
		}
	}

	if at < 0 {
		return &Tree{MP: mp}, nil
	}

	blob, err := db.Blob(t.Entries[at].Oid)
	if err != nil {
		return nil, err
	}
	defer blob.Close()

	return NewFromReader(mp, blob.Contents)
}

// Applied returns a slice of attributes applied to the given path, relative to
// the receiving tree. It traverse through sub-trees in a topological ordering,
// if there are relevant .gitattributes matching that path.
func (t *Tree) Applied(to string) []*Attr {
	var attrs []*Attr
	if !t.MP.didReadMacros {
		if t.SystemAttributes != nil {
			t.MP.ProcessLines(t.SystemAttributes.Lines, true)
		}
		if t.UserAttributes != nil {
			t.MP.ProcessLines(t.UserAttributes.Lines, true)
		}
		t.MP.ProcessLines(t.Lines, true)

		if t.RepoAttributes != nil {
			t.MP.ProcessLines(t.RepoAttributes.Lines, true)
		}
	}

	if t.SystemAttributes != nil {
		attrs = append(attrs, t.SystemAttributes.applied(to)...)
	}

	if t.UserAttributes != nil {
		attrs = append(attrs, t.UserAttributes.applied(to)...)
	}

	attrs = append(attrs, t.applied(to)...)

	if t.RepoAttributes != nil {
		attrs = append(attrs, t.RepoAttributes.applied(to)...)
	}

	return attrs
}

func (t *Tree) applied(to string) []*Attr {
	var attrs []*Attr

	for _, line := range t.MP.ProcessLines(t.Lines, false) {
		if line.Pattern().Match(to) {
			attrs = append(attrs, line.Attrs()...)
		}
	}

	dirPath, remainingPath, found := strings.Cut(to, "/")
	if found {
		if child, ok := t.Children[dirPath]; ok {
			attrs = append(attrs, child.applied(remainingPath)...)
		}
	}

	return attrs
}
