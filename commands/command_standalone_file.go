package commands

import (
	"fmt"
	"os"

	"github.com/git-lfs/git-lfs/lfshttp/standalone"
	"github.com/spf13/cobra"
)

func standaloneFileCommand(cmd *cobra.Command, args []string) {
	err := standalone.ProcessStandaloneData(cfg, os.Stdin, os.Stdout)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(2)
	}
}

func init() {
	RegisterCommand("standalone-file", standaloneFileCommand, nil)
}
