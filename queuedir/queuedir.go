// Package queue implements a simple file system queue.  Jobs are stored as
// files in a directory.  Loosely implements something like maildir, without
// any specific code for dealing with email.
package queuedir

import (
	"bytes"
	"github.com/streadway/simpleuuid"
	"io"
	"os"
	"path/filepath"
	"time"
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
	uuid, err := simpleuuid.NewTime(time.Now())
	if err != nil {
		return "", err
	}

	id := uuid.String()
	file, err := os.Create(filepath.Join(q.Path, id))
	if err == nil {
		defer file.Close()
		_, err = io.Copy(file, reader)
	}
	return id, err
}

func (q *Queue) AddString(body string) (string, error) {
	return q.Add(bytes.NewBufferString(body))
}

func (q *Queue) AddBytes(body []byte) (string, error) {
	return q.Add(bytes.NewBuffer(body))
}
