// Package queue implements a simple file system queue.  Jobs are stored as
// files in a directory.  Loosely implements something like maildir, without
// any specific code for dealing with email.
package queuedir

import (
	"bytes"
	"fmt"
	"github.com/streadway/simpleuuid"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

type QueueDir struct {
	Path string
}

func New(path string) *QueueDir {
	return &QueueDir{Path: path}
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

type WalkFunc func(id string, body []byte) error

func (q *Queue) Add(reader io.Reader) (string, error) {
	uuid, err := simpleuuid.NewTime(time.Now())
	if err != nil {
		return "", err
	}

	id := uuid.String()
	file, err := os.Create(q.FullPath(id))
	if err == nil {
		defer file.Close()
		_, err = io.Copy(file, reader)
	}
	return id, err
}

func (q *Queue) Walk(cb WalkFunc) error {
	return filepath.Walk(q.Path, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}

		if err != nil {
			return err
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}

		body, err := ioutil.ReadAll(file)
		if err != nil {
			return err
		}

		return cb(filepath.Base(path), body)
	})
}

func (q *Queue) AddString(body string) (string, error) {
	return q.Add(bytes.NewBufferString(body))
}

func (q *Queue) AddBytes(body []byte) (string, error) {
	return q.Add(bytes.NewBuffer(body))
}

func (q *Queue) Move(newqueue *Queue, id string) error {
	return os.Rename(q.FullPath(id), newqueue.FullPath(id))
}

func (q *Queue) Del(id string) error {
	full := q.FullPath(id)
	stat, err := os.Stat(full)
	if err != nil {
		return err
	}

	if stat.IsDir() {
		return fmt.Errorf("%s in %s is a directory", id, q.Path)
	}

	return os.Remove(full)
}

func (q *Queue) FullPath(id string) string {
	return filepath.Join(q.Path, id)
}
