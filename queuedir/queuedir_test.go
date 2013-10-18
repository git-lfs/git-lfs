package queuedir

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestAdd(t *testing.T) {
	q := Setup(t)
	defer q.Teardown()

	id, err := q.Queue.Add(bytes.NewBufferString("BOOM"))
	if err != nil {
		t.Fatalf("Cannot add to queue: %s", err)
	}

	file, err := os.Open(filepath.Join(q.Queue.Path, id))
	if err != nil {
		t.Fatalf("Cannot open file: %s", err)
	}

	by, err := ioutil.ReadAll(file)
	if err != nil {
		t.Fatalf("Cannot read file: %s", err)
	}

	if str := string(by); str != "BOOM" {
		t.Fatalf("Expected BOOM, got %s", str)
	}
}

type QueueTest struct {
	Dir   *QueueDir
	Queue *Queue
}

func Setup(t *testing.T) *QueueTest {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Cannot get current working dir: %s", err)
	}

	qdir, err := New(filepath.Join(wd, "test_queuedir"))
	if err != nil {
		t.Fatalf("Cannot create queuedir: %s", err)
	}

	q, err := qdir.Queue("test")
	if err != nil {
		t.Fatalf("Cannot create test queue: %s", err)
	}

	return &QueueTest{qdir, q}
}

func (t *QueueTest) Teardown() {
	os.RemoveAll(t.Dir.Path)
}
