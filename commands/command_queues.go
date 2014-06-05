package commands

import (
	"github.com/github/git-media/gitmedia"
	"github.com/github/git-media/queuedir"
	"os"
	"path/filepath"
	"strings"
)

type QueuesCommand struct {
	*Command
}

func (c *QueuesCommand) Run() {
	err := gitmedia.WalkQueues(func(name string, queue *queuedir.Queue) error {
		wd, _ := os.Getwd()
		Print(name)
		return queue.Walk(func(id string, body []byte) error {
			parts := strings.Split(string(body), ":")
			if len(parts) == 2 {
				absPath := filepath.Join(gitmedia.LocalWorkingDir, parts[1])
				relPath, _ := filepath.Rel(wd, absPath)
				Print("  " + relPath)
			} else {
				Print("  " + parts[0])
			}
			return nil
		})
	})

	if err != nil {
		gitmedia.Panic(err, "Error walking queues")
	}
}

func init() {
	registerCommand("queues", func(c *Command) RunnableCommand {
		return &QueuesCommand{Command: c}
	})
}
