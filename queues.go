package gitmedia

import (
	"./queuedir"
	"path/filepath"
)

func QueueUpload(sha string) {
	_, err := getUploadQueue().AddString(sha)
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

func getUploadQueue() *queuedir.Queue {
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
		"upload": getUploadQueue,
	}
	queueDir    *queuedir.QueueDir
	uploadQueue *queuedir.Queue
)
