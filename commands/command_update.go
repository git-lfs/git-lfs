package commands

import (
	"regexp"

	"github.com/github/git-lfs/git"
	"github.com/github/git-lfs/lfs"
	"github.com/github/git-lfs/vendor/_nuts/github.com/spf13/cobra"
)

var (
	updateCmd = &cobra.Command{
		Use:   "update",
		Short: "Update local Git LFS configuration",
		Run:   updateCommand,
	}

	updateForce = false
)

// updateCommand is used for updating parts of Git LFS that reside under
// .git/lfs.
func updateCommand(cmd *cobra.Command, args []string) {
	if err := lfs.InstallHooks(updateForce); err != nil {
		Error(err.Error())
		Print("Run `git lfs update --force` to overwrite this hook.")
	} else {
		Print("Updated pre-push hook.")
	}

	lfsAccessRE := regexp.MustCompile(`\Alfs\.(.*)\.access\z`)
	for key, value := range lfs.Config.AllGitConfig() {
		matches := lfsAccessRE.FindStringSubmatch(key)
		if len(matches) < 2 {
			continue
		}

		switch value {
		case "basic":
		case "private":
			git.Config.SetLocal("", key, "basic")
			Print("Updated %s access from %s to %s.", matches[1], value, "basic")
		default:
			git.Config.UnsetLocalKey("", key)
			Print("Removed invalid %s access of %s.", matches[1], value)
		}
	}
}

func init() {
	updateCmd.Flags().BoolVarP(&updateForce, "force", "f", false, "Overwrite hooks.")
	RootCmd.AddCommand(updateCmd)
}
