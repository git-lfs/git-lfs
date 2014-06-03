package gitmedia

import (
	"github.com/github/git-media/queuedir"
	"path/filepath"
)

func QueueUpload(sha, filename string) {
	fileBody := sha

	if filename != "" {
		fileBody += ":" + filename
	}

	_, err := UploadQueue().AddString(fileBody)
	if err != nil {
		Panic(err, "Unable to add %s to queue", sha)
	}
}

func WalkQueues(cb func(name string, queue *queuedir.Queue) error) error {
	var err error
	for name, queuefunc := range queues {
		err = cb(name, queuefunc())
		if err != nil {
			return err
		}
	}
	return err
}

func UploadQueue() *queuedir.Queue {
	if uploadQueue == nil {
		q, err := queueDir.Queue("upload")
		if err != nil {
			Panic(err, "Error setting up queue")
		}
		uploadQueue = q
	}

	return uploadQueue
}

func setupQueueDir() *queuedir.QueueDir {
	return queuedir.New(filepath.Join(LocalMediaDir, "queue"))
}

var (
	queues = map[string]func() *queuedir.Queue{
		"upload": UploadQueue,
	}
	queueDir    *queuedir.QueueDir
	uploadQueue *queuedir.Queue
)
