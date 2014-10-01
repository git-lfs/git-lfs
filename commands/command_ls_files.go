package commands

import (
	"github.com/github/git-media/gitmedia"
	"github.com/github/git-media/pointer"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
)

var (
	lsFilesCmd = &cobra.Command{
		Use:   "ls-files",
		Short: "Show information about git media files",
		Run:   lsFilesCommand,
	}
)

func lsFilesCommand(cmd *cobra.Command, args []string) {
	filepath.Walk(gitmedia.LocalLinkDir, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			base, _ := filepath.Rel(gitmedia.LocalLinkDir, path)
			firstTwo := filepath.Dir(base)
			rest := filepath.Base(base)

			link, err := pointer.FindLink(firstTwo + rest)
			if err != nil {
				return nil
			}
			Print(link.Name)
		}
		return nil
	})

}

func init() {
	RootCmd.AddCommand(lsFilesCmd)
}
