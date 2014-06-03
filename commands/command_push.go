package commands

import (
	"github.com/github/git-media/gitmedia"
	"github.com/github/git-media/gitmediaclient"
	"strings"
)

type PushCommand struct {
	*Command
}

func (c *PushCommand) Run() {
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
	registerCommand("push", func(c *Command) RunnableCommand {
		return &PushCommand{Command: c}
	})
}
