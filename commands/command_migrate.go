package commands

import (
	"bufio"
	"bytes"
	"io"
	"os"

	"github.com/git-lfs/git-lfs/git"
	"github.com/git-lfs/git-lfs/lfs"
	"github.com/spf13/cobra"
)

var (
	migrateInclude string
	migrateExclude string
)

func migrateCommand(cmd *cobra.Command, args []string) {
	lfs.InstallHooks(false)

	var last string

	for scanner := bufio.NewScanner(os.Stdin); scanner.Scan(); {
		rev := scanner.Text()
		if err := git.ReadTree(rev); err != nil {
			ExitWithError(err)
		}

		include, exclude := getIncludeExcludeArgs(cmd)
		filter := buildFilepathFilter(cfg, include, exclude)

		ch, err := lfs.P_lsTreeBlobs(rev, filter)
		if err != nil {
			ExitWithError(err)
		}

		for blob := range ch.Results {
			if err = git.CheckoutIndex(blob.Filename); err != nil {
				ExitWithError(err)
			}

			trackCommand(nil, []string{blob.Filename})

			var flags int = os.O_RDWR
			f, err := os.OpenFile(blob.Filename, flags, os.ModeAppend)
			if err != nil {
				ExitWithError(err)
			}

			pbuf := bytes.NewBuffer(nil)

			// TODO(@ttaylorr): read real contents via
			// git-cat-file(1) instead of faking it, or requiring
			// git.CheckoutIndex().
			clean(pbuf, f, blob.Filename)

			if _, err = f.Seek(0, io.SeekStart); err != nil {
				ExitWithError(err)
			}
			if _, err = io.Copy(f, pbuf); err != nil {
				ExitWithError(err)
			}

			if err = f.Close(); err != nil {
				ExitWithError(err)
			}

			for _, n := range []string{".gitattributes", f.Name()} {
				if err = git.UpdateIndex(n); err != nil {
					ExitWithError(err)
				}
			}

			newtree, err := git.WriteTree("/")
			if err != nil {
				ExitWithError(err)
			}

			newsha, err := git.CommitTree(newtree, last, "lol")
			if err != nil {
				ExitWithError(err)
			}
			last = newsha

			if err = git.UpdateRef("HEAD", newsha); err != nil {
				ExitWithError(err)
			}
		}

		if err = ch.Wait(); err != nil {
			ExitWithError(err)
		}
	}

	checkoutCommand(nil, nil)
}

func init() {
	RegisterCommand("migrate", migrateCommand, func(cmd *cobra.Command) {
		cmd.Flags().StringVarP(&migrateInclude, "include", "I", "", "Include a list of paths")
		cmd.Flags().StringVarP(&migrateExclude, "exclude", "X", "", "Exclude a list of paths")
	})
}
