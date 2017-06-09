package commands

import (
	"path/filepath"

	"github.com/git-lfs/git-lfs/git"
	"github.com/git-lfs/git-lfs/git/odb"
	"github.com/spf13/cobra"
)

var (
	// migrateIncludeRefs is a set of Git references to explicitly include
	// in the migration.
	migrateIncludeRefs []string
	// migrateExcludeRefs is a set of Git references to explicitly exclude
	// in the migration.
	migrateExcludeRefs []string
)

// getObjectDatabase creates a *git.ObjectDatabase from the filesystem pointed
// at the .git directory of the currently checked-out repository.
func getObjectDatabase() (*odb.ObjectDatabase, error) {
	dir, err := git.GitDir()
	if err != nil {
		return nil, errors.Wrap(err, "cannot open root")
	}
	return odb.FromFilesystem(filepath.Join(dir, "objects"))
}

func init() {
	RegisterCommand("migrate", nil, func(cmd *cobra.Command) {
		// Adding flags directly to cmd.Flags() doesn't apply those
		// flags to any subcommands of the root. Therefore, loop through
		// each subcommand specifically, and include common arguments to
		// each.
		//
		// Once done, link each orphaned command to the
		// `git-lfs-migrate(1)` command as a subcommand (child).

		for _, subcommand := range []*cobra.Command{} {
			subcommand.Flags().StringVarP(&includeArg, "include", "I", "", "Include a list of paths")
			subcommand.Flags().StringVarP(&excludeArg, "exclude", "X", "", "Exclude a list of paths")

			subcommand.Flags().StringSliceVar(&migrateIncludeRefs, "include-ref", nil, "An explicit list of refs to include")
			subcommand.Flags().StringSliceVar(&migrateExcludeRefs, "exclude-ref", nil, "An explicit list of refs to exclude")

			cmd.AddCommand(subcommand)
		}
	})
}
