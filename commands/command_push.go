package commands

import (
	"fmt"
	"github.com/github/git-media/gitmedia"
	"github.com/github/git-media/gitmediaclient"
	"strings"
)

type PushCommand struct {
	*Command
}

func (c *PushCommand) Run() {
	q, err := gitmedia.UploadQueue()
	if err != nil {
		Panic(err, "Error setting up the queue")
	}

	q.Walk(func(id string, body []byte) error {
		fileInfo := string(body)
		parts := strings.Split(fileInfo, ":")

		var sha, filename string
		sha = parts[0]
		if len(parts) > 1 {
			filename = parts[1]
		}

		path, err := gitmedia.LocalMediaPath(sha)
		if err == nil {
			err = gitmediaclient.Options(path)
		}
		if err != nil {
			Panic(err, "error uploading file %s/%s", sha, filename)
		}

		err = gitmediaclient.Put(path, filename)
		if err != nil {
			Panic(err, "error uploading file %s/%s", sha, filename)
		}
		fmt.Printf("\n")

		if err := q.Del(id); err != nil {
			Panic(err, "error removing %s from queue", sha)
		}

		return nil
	})
}

func init() {
	registerCommand("push", func(c *Command) RunnableCommand {
		return &PushCommand{Command: c}
	})
}
