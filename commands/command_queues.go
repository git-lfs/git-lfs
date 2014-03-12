package gitmedia

import (
	".."
	"../queuedir"
	"strings"
)

type QueuesCommand struct {
	*Command
}

func (c *QueuesCommand) Run() {
	err := gitmedia.WalkQueues(func(name string, queue *queuedir.Queue) error {
		gitmedia.Print(name)
		return queue.Walk(func(id string, body []byte) error {
			parts := strings.Split(string(body), ":")
			if len(parts) == 2 {
				gitmedia.Print("  " + parts[1])
			} else {
				gitmedia.Print("  " + parts[0])
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
