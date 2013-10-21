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
	err := gitmedia.UploadQueue().Walk(func(id string, sha []byte) error {
		path := gitmedia.LocalMediaPath(string(sha))
		return gitmediaclient.Send(path)
	})

	if err != nil {
		gitmedia.Panic(err, "error uploading file")
	}
}

func init() {
	registerCommand("sync", func(c *Command) RunnableCommand {
		return &SyncCommand{Command: c}
	})
}
