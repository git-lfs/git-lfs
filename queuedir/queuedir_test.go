package queuedir

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestAdd(t *testing.T) {
	q := Setup(t)
	defer q.Teardown()

	id, err := q.Queue.AddString("BOOM")
	if err != nil {
		t.Fatalf("Cannot add to queue: %s", err)
	}

	assertExist(t, q.Queue, id)
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

func TestMove(t *testing.T) {
	q := Setup(t)
	defer q.Teardown()

	id, err := q.Queue.AddString("BOOM")
	if err != nil {
		t.Fatalf("Cannot add to queue: %s", err)
	}

	assertExist(t, q.Queue, id)

	queue2, err := q.Dir.Queue("test2")
	if err != nil {
		t.Fatalf("Cannot create %s queue: %s", queue2.Name, err)
	}

	err = q.Queue.Move(queue2, id)
	if err != nil {
		t.Fatalf("Cannot move from queue %s to %s: %s", q.Queue.Name, queue2.Name, err)
	}

	assertNotExist(t, q.Queue, id)
	assertExist(t, queue2, id)
}

func TestDel(t *testing.T) {
	q := Setup(t)
	defer q.Teardown()

	id, err := q.Queue.AddString("BOOM")
	if err != nil {
		t.Fatalf("Cannot add to queue: %s", err)
	}

	assertExist(t, q.Queue, id)

	err = q.Queue.Del(id)
	if err != nil {
		t.Fatalf("Error deleting from queue: %s", err)
	}
	assertNotExist(t, q.Queue, id)
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

func assertExist(t *testing.T, q *Queue, id string) {
	if !fileExists(q, id) {
		t.Fatalf("%s does not exist in queue %s", id, q.Name)
	}
}

func assertNotExist(t *testing.T, q *Queue, id string) {
	if fileExists(q, id) {
		t.Fatalf("%s exists in queue %s", id, q.Name)
	}
}

func fileExists(q *Queue, id string) bool {
	_, err := os.Stat(q.FullPath(id))
	return err == nil
}
