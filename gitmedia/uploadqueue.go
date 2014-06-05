package gitmedia

import (
	"github.com/github/git-media/queuedir"
	"path/filepath"
)

func QueueUpload(sha, filename string) error {
	fileBody := sha

	if filename != "" {
		fileBody += ":" + filename
	}

	q, err := UploadQueue()
	if err != nil {
		return err
	}

	_, err = q.AddString(fileBody)
	return err
}

func WalkQueues(cb func(name string, queue *queuedir.Queue) error) error {
	var err error
	for name, queuefunc := range queues {
		q, err := queuefunc()
		if err == nil {
			err = cb(name, q)
		}
		if err != nil {
			return err
		}
	}
	return err
}

func UploadQueue() (*queuedir.Queue, error) {
	if uploadQueue == nil {
		q, err := queueDir.Queue("upload")
		if err != nil {
			return nil, err
		}
		uploadQueue = q
	}

	return uploadQueue, nil
}

func setupQueueDir() *queuedir.QueueDir {
	return queuedir.New(filepath.Join(LocalMediaDir, "queue"))
}

var (
	queues = map[string]func() (*queuedir.Queue, error){
		"upload": UploadQueue,
	}
	queueDir    *queuedir.QueueDir
	uploadQueue *queuedir.Queue
)
