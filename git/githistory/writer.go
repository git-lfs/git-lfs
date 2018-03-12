package githistory

import (
	"bytes"

	"github.com/git-lfs/git-lfs/errors"
	"github.com/git-lfs/git-lfs/git/odb"
	"github.com/git-lfs/git-lfs/tasklog"
)

type Pending struct {
	Author    string
	Committer string
	Tree      *odb.Tree

	Message      string
	ExtraHeaders []*odb.ExtraHeader
}

type WriterFn func(*odb.Commit) ([]*Pending, error)

type Writer struct {
	fn WriterFn

	db *odb.ObjectDatabase
	l  *tasklog.Logger
}

func NewWriter(db *odb.ObjectDatabase, l *tasklog.Logger, fn WriterFn) *Writer {
	return &Writer{
		fn: fn,

		db: db,
		l:  l,
	}
}

func (w *Writer) Write(sha []byte) ([]byte, error) {
	start, err := w.db.Commit(sha)
	if err != nil {
		return nil, errors.Wrap(err, "git/githistory: could not find base")
	}

	fntask := w.l.Waiter("migrate: building new commits")
	pendings, err := w.fn(start)
	fntask.Complete()

	if err != nil {
		return nil, errors.Wrap(err, "git/githistory: could not generate history")
	}

	var base []byte = sha
	for _, pending := range pendings {
		base, err = w.commit(pending, base)
		if err != nil {
			return nil, errors.Wrapf(err, "git/githistory: could not add commit")
		}
	}

	root, _ := w.db.Root()

	refs, err := NewRefFinder(w.db, RefFinderLocalOnly).FindRefs()
	if err != nil {
		return nil, errors.Wrap(err, "git/githistory: could not find references")
	}

	updater := &refUpdater{
		CacheFn: func(old []byte) ([]byte, bool) {
			if bytes.Equal(old, sha) {
				return base, true
			}
			return nil, false
		},
		Refs: refs,
		Root: root,

		Logger: w.l,
		db:     w.db,
	}

	if err := updater.UpdateRefs(); err != nil {
		return nil, err
	}
	return base, nil
}

func (w *Writer) commit(p *Pending, base []byte) ([]byte, error) {
	tree, err := w.db.WriteTree(p.Tree)
	if err != nil {
		return nil, errors.Wrap(err, "git/githistory: could not commit tree")
	}

	return w.db.WriteCommit(&odb.Commit{
		Author:    p.Author,
		Committer: p.Committer,
		TreeID:    tree,

		ParentIDs: [][]byte{base},

		ExtraHeaders: p.ExtraHeaders,
		Message:      p.Message,
	})
}
