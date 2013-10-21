package gitmedia

import (
	".."
	"../client"
)

type SyncCommand struct {
	*Command
}

func (c *SyncCommand) Setup() {
}

func (c *SyncCommand) Run() {
	q := gitmedia.UploadQueue()
	q.Walk(func(id string, body []byte) error {
		sha := string(body)
		path := gitmedia.LocalMediaPath(sha)
		err := gitmediaclient.Send(path)
		if err != nil {
			gitmedia.Panic(err, "error uploading file %s", sha)
		}

		if err := q.Del(id); err != nil {
			gitmedia.Panic(err, "error removing %s from queue", sha)
		}
	})
}

func init() {
	registerCommand("sync", func(c *Command) RunnableCommand {
		return &SyncCommand{Command: c}
	})
}
