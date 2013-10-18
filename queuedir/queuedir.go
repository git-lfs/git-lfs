// Package queue implements a simple file system queue.  Jobs are stored as
// files in a directory.  Loosely implements something like maildir, without
// any specific code for dealing with email.
package queuedir

import (
	"io"
	"os"
	"path/filepath"
)

type QueueDir struct {
	Path      string
	TempQueue *Queue
}

func New(path string) (*QueueDir, error) {
	q := &QueueDir{Path: path}
	tq, err := q.Queue("tmp")
	q.TempQueue = tq
	return q, err
}

func (q *QueueDir) Queue(name string) (*Queue, error) {
	qu := &Queue{name, filepath.Join(q.Path, name), q}
	err := os.MkdirAll(qu.Path, 0777)
	return qu, err
}

type Queue struct {
	Name string
	Path string
	Dir  *QueueDir
}

func (q *Queue) Add(reader io.Reader) (string, error) {
	id := "a" // yea this may change
	file, err := os.Create(filepath.Join(q.Path, id))
	if err == nil {
		defer file.Close()
		_, err = io.Copy(file, reader)
	}
	return id, err
}
