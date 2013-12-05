package gitmedia

import (
	".."
	"../queuedir"
)

type QueuesCommand struct {
	*Command
}

func (c *QueuesCommand) Run() {
	err := gitmedia.WalkQueues(func(name string, queue *queuedir.Queue) error {
		gitmedia.Print(name)
		return queue.Walk(func(id string, body []byte) error {
			gitmedia.Print("  " + string(body))
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
