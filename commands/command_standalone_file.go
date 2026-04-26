package commands

import (
	"os"

	"github.com/git-lfs/git-lfs/v3/lfshttp/standalone"
	"github.com/spf13/cobra"
)

func standaloneFileCommand(cmd *cobra.Command, args []string) {
	err := standalone.ProcessStandaloneData(cfg, os.Stdin, os.Stdout)
	if err != nil {
		ExitWithError(err)
	}
}

func init() {
	RegisterCommand("standalone-file", standaloneFileCommand, nil)
}
