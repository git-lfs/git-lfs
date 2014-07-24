package commands

import (
	"github.com/github/git-media/gitmedia"
	"github.com/github/git-media/queuedir"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
	"strings"
)

var (
	queuesCmd = &cobra.Command{
		Use:   "queues",
		Short: "Show the status of the internal Git Media queues",
		Run:   queuesCommand,
	}
)

func queuesCommand(cmd *cobra.Command, args []string) {
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
		Panic(err, "Error walking queues")
	}
}

func init() {
	RootCmd.AddCommand(queuesCmd)
}
