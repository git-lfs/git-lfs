package commands

import (
	"fmt"
	"github.com/github/git-media/scanner"
	"github.com/spf13/cobra"
)

var (
	scanCmd = &cobra.Command{
		Use:   "scan",
		Short: "Scan for git media files",
		Run:   scanCommand,
	}
	createLinks = false
)

func scanCommand(cmd *cobra.Command, args []string) {
	pointers, err := scanner.Scan("")
	if err != nil {
		Panic(err, "Failed to scan")
	}

	// Now we have Pointer objects for all git media files.
	// What can we do with them?
	// Create link files
	// Offer to download
	// ?
	// Profit
	for _, p := range pointers {
		fmt.Println("Git Media OID:", p.Name)
		if createLinks {
			p.CreateLink(p.Name)
		}
	}
}

func init() {
	scanCmd.Flags().BoolVarP(&createLinks, "create-links", "l", false, "Create git media link files")
	RootCmd.AddCommand(scanCmd)
}
