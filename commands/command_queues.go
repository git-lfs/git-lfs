package gitmedia

import (
	".."
	"../queuedir"
	"fmt"
)

type QueuesCommand struct {
	*Command
}

func (c *QueuesCommand) Run() {
	err := gitmedia.WalkQueues(func(name string, queue *queuedir.Queue) error {
		fmt.Println(name)
		return queue.Walk(func(id string, body []byte) error {
			fmt.Println("  " + string(body))
			return nil
		})
	})

	if err != nil {
		fmt.Println("Error walking queues")
		fmt.Println(err)
	}
}

func init() {
	registerCommand("queues", func(c *Command) RunnableCommand {
		return &QueuesCommand{Command: c}
	})
}
