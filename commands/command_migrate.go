package commands

import (
	"bufio"
	"bytes"
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

	objects, err := git.NewObjectScanner()
	if err != nil {
		ExitWithError(err)
	}

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
			if !objects.Scan(blob.Sha1) {
				ExitWithError(objects.Err())
			}

			trackCommand(nil, []string{blob.Filename})

			pbuf := bytes.NewBuffer(nil)

			if err := clean(pbuf, objects.Contents(), blob.Filename); err != nil {
				ExitWithError(err)
			}

			newOid, err := git.HashObject(pbuf)
			if err != nil {
				ExitWithError(err)
			}

			if err := git.UpdateIndex(".gitattributes"); err != nil {
				ExitWithError(err)
			}

			if err := git.UpdateIndexInfo("0644", newOid, blob.Filename); err != nil {
				ExitWithError(err)
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
