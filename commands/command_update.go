package commands

import (
	"regexp"

	"github.com/git-lfs/git-lfs/v3/tr"
	"github.com/spf13/cobra"
)

var (
	updateForce  = false
	updateManual = false
)

// updateCommand is used for updating parts of Git LFS that reside under
// .git/lfs.
func updateCommand(cmd *cobra.Command, args []string) {
	requireGitVersion()
	setupRepository()

	lfsAccessRE := regexp.MustCompile(`\Alfs\.(.*)\.access\z`)
	for key, _ := range cfg.Git.All() {
		matches := lfsAccessRE.FindStringSubmatch(key)
		if len(matches) < 2 {
			continue
		}

		value, _ := cfg.Git.Get(key)

		switch value {
		case "basic":
		case "private":
			cfg.SetGitLocalKey(key, "basic")
			Print(tr.Tr.Get("Updated %s access from %s to %s.", matches[1], value, "basic"))
		default:
			cfg.UnsetGitLocalKey(key)
			Print(tr.Tr.Get("Removed invalid %s access of %s.", matches[1], value))
		}
	}

	if updateForce && updateManual {
		Exit(tr.Tr.Get("You cannot use --force and --manual options together"))
	}

	if updateManual {
		Print(getHookInstallSteps())
	} else {
		if err := installHooks(updateForce); err != nil {
			Error(err.Error())
			Exit("%s\n  1: %s\n  2: %s",
				tr.Tr.Get("To resolve this, either:"),
				tr.Tr.Get("run `git lfs update --manual` for instructions on how to merge hooks."),
				tr.Tr.Get("run `git lfs update --force` to overwrite your hook."))
		} else {
			Print(tr.Tr.Get("Updated Git hooks."))
		}
	}

}

func init() {
	RegisterCommand("update", updateCommand, func(cmd *cobra.Command) {
		cmd.Flags().BoolVarP(&updateForce, "force", "f", false, "Overwrite existing hooks.")
		cmd.Flags().BoolVarP(&updateManual, "manual", "m", false, "Print instructions for manual install.")
	})
}
