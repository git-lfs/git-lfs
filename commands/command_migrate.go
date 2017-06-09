package commands

import (
	"path/filepath"
	"strings"

	"github.com/git-lfs/git-lfs/errors"
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

// includeExcludeRefs returns fully-qualified sets of references to include, and
// exclude, or an error if those could not be determined.
//
// They are determined based on the following rules:
//
//   - Include all local refs/heads/<branch> references for each branch
//     specified as an argument.
//   - Include the currently checked out branch if no branches are given as
//     arguments and the --include-ref= or --exclude-ref= flag(s) aren't given.
//   - Include all references given in --include-ref=<ref>.
//   - Exclude all references given in --exclude-ref=<ref>.
func includeExcludeRefs(args []string) (include, exclude []string, err error) {
	hardcore := len(migrateIncludeRefs) > 0 || len(migrateExcludeRefs) > 0

	if len(args) == 0 && !hardcore {
		// If no branches were given explicitly AND neither
		// --include-ref or --exclude-ref flags were given, then add the
		// currently checked out reference.
		current, err := currentRefToMigrate()
		if err != nil {
			return nil, nil, err
		}
		args = append(args, current.Name)
	}

	for _, name := range args {
		// Then, loop through each branch given, resolve that reference,
		// and include it.
		ref, err := git.ResolveRef(name)
		if err != nil {
			return nil, nil, err
		}

		include = append(include, ref.Name)
	}

	if hardcore {
		// If either --include-ref=<ref> or --exclude-ref=<ref> were
		// given, append those to the include and excluded reference
		// set, respectively.
		include = append(include, migrateIncludeRefs...)
		exclude = append(exclude, migrateExcludeRefs...)
	} else {
		// Otherwise, if neither --include-ref=<ref> or
		// --exclude-ref=<ref> were given, include no additional
		// references, and exclude all remote references that are remote
		// branches or remote tags.
		remoteRefs, err := getRemoteRefs()
		if err != nil {
			return nil, nil, err
		}

		exclude = append(exclude, remoteRefs...)
	}

	return include, exclude, nil
}

// getRemoteRefs returns a fully qualified set of references belonging to all
// remotes known by the currently checked-out repository, or an error if those
// references could not be determined.
func getRemoteRefs() ([]string, error) {
	var refs []string

	remotes, err := git.RemoteList()
	if err != nil {
		return nil, err
	}

	for _, remote := range remotes {
		refsForRemote, err := git.RemoteRefs(remote)
		if err != nil {
			return nil, err
		}

		for _, ref := range refsForRemote {
			refs = append(refs, formatRefName(ref, remote))
		}
	}

	return refs, nil
}

// formatRefName returns the fully-qualified name for the given Git reference
// "ref".
func formatRefName(ref *git.Ref, remote string) string {
	var name []string

	switch ref.Type {
	case git.RefTypeRemoteBranch:
		name = []string{"refs", "remotes", remote, ref.Name}
	case git.RefTypeRemoteTag:
		name = []string{"refs", "tags", ref.Name}
	default:
		return ref.Name
	}
	return strings.Join(name, "/")

}

// currentRefToMigrate returns the fully-qualified name of the currently
// checked-out reference, or an error if the reference's type was not a local
// branch.
func currentRefToMigrate() (*git.Ref, error) {
	current, err := git.CurrentRef()
	if err != nil {
		return nil, err
	}

	if current.Type == git.RefTypeOther ||
		current.Type == git.RefTypeRemoteBranch ||
		current.Type == git.RefTypeRemoteTag {

		return nil, errors.Errorf("fatal: cannot migrate non-local ref: %s", current.Name)
	}
	return current, nil
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
