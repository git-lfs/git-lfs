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
	queueDir    *queuedir.QueueDir
	uploadQueue *queuedir.Queue
)
