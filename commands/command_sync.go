package gitmedia

import (
	".."
	"../client"
	"strings"
)

type SyncCommand struct {
	*Command
}

func (c *SyncCommand) Run() {
	q := gitmedia.UploadQueue()
	q.Walk(func(id string, body []byte) error {
		fileInfo := string(body)
		parts := strings.Split(fileInfo, ":")

		var sha, filename string
		sha = parts[0]
		if len(parts) > 1 {
			filename = parts[1]
		}

		path := gitmedia.LocalMediaPath(sha)

		err := gitmediaclient.Options(path)
		if err != nil {
			gitmedia.Panic(err, "error uploading file %s", filename)
		}

		err = gitmediaclient.Put(path, filename)
		if err != nil {
			gitmedia.Panic(err, "error uploading file %s", sha)
		}

		if err := q.Del(id); err != nil {
			gitmedia.Panic(err, "error removing %s from queue", sha)
		}
		return nil
	})
}

func init() {
	registerCommand("sync", func(c *Command) RunnableCommand {
		return &SyncCommand{Command: c}
	})
}
