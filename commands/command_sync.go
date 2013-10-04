package gitmedia

import (
	".."
	"../client"
	"fmt"
	"os"
	"path/filepath"
)

type SyncCommand struct {
	*Command
}

func (c *SyncCommand) Setup() {
}

func (c *SyncCommand) Run() {
	filepath.Walk(gitmedia.LocalMediaDir, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}

		if ext := filepath.Ext(path); len(ext) > 0 {
			return nil
		}

		if err := gitmediaclient.Send(path); err != nil {
			fmt.Println("Error uploading:", path)
			panic(err)
		}

		return nil
	})
}

func init() {
	registerCommand("sync", func(c *Command) RunnableCommand {
		return &SyncCommand{Command: c}
	})
}
