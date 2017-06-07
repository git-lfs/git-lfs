package commands

import (
	"github.com/spf13/cobra"
)

func migrateCommand(cmd *cobra.Command, args []string) {
}

func init() {
	RegisterCommand("migrate", migrateCommand, nil)
}
