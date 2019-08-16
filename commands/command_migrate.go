package commands

import (
	"bufio"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/git-lfs/git-lfs/errors"
	"github.com/git-lfs/git-lfs/git"
	"github.com/git-lfs/git-lfs/git/githistory"
	"github.com/git-lfs/git-lfs/tasklog"
	"github.com/git-lfs/gitobj"
	"github.com/spf13/cobra"
)

var (
	// migrateIncludeRefs is a set of Git references to explicitly include
	// in the migration.
	migrateIncludeRefs []string
	// migrateExcludeRefs is a set of Git references to explicitly exclude
	// in the migration.
	migrateExcludeRefs []string

	// migrateYes indicates that an answer of 'yes' should be presumed
	// whenever 'git lfs migrate' asks for user input.
	migrateYes bool

	// migrateSkipFetch assumes that the client has the latest copy of
	// remote references, and thus should not contact the remote for a set
	// of updated references.
	migrateSkipFetch bool

	// migrateEverything indicates the presence of the --everything flag,
	// and instructs 'git lfs migrate' to migrate all local references.
	migrateEverything bool

	// migrateVerbose enables verbose logging
	migrateVerbose bool

	// objectMapFile is the path to the map of old sha1 to new sha1
	// commits
	objectMapFilePath string

	// migrateNoRewrite is the flag indicating whether or not the
	// command should rewrite git history
	migrateNoRewrite bool
	// migrateCommitMessage is the message to use with the commit generated
	// by the migrate command
	migrateCommitMessage string

	// exportRemote is the remote from which to download objects when
	// performing an export
	exportRemote string

	// migrateFixup is the flag indicating whether or not to infer the
	// included and excluded filepath patterns.
	migrateFixup bool
)

// migrate takes the given command and arguments, *gitobj.ObjectDatabase, as well
// as a BlobRewriteFn to apply, and performs a migration.
func migrate(args []string, r *githistory.Rewriter, l *tasklog.Logger, opts *githistory.RewriteOptions) {
	requireInRepo()

	opts, err := rewriteOptions(args, opts, l)
	if err != nil {
		ExitWithError(err)
	}

	_, err = r.Rewrite(opts)
	if err != nil {
		ExitWithError(err)
	}
}

// getObjectDatabase creates a *git.ObjectDatabase from the filesystem pointed
// at the .git directory of the currently checked-out repository.
func getObjectDatabase() (*gitobj.ObjectDatabase, error) {
	dir, err := git.GitCommonDir()
	if err != nil {
		return nil, errors.Wrap(err, "cannot open root")
	}
	alternates, _ := cfg.OSEnv().Get("GIT_ALTERNATE_OBJECT_DIRECTORIES")
	return gitobj.FromFilesystemWithAlternates(filepath.Join(dir, "objects"), cfg.TempDir(), alternates)
}

// rewriteOptions returns *githistory.RewriteOptions able to be passed to a
// *githistory.Rewriter that reflect the current arguments and flags passed to
// an invocation of git-lfs-migrate(1).
//
// It is merged with the given "opts". In other words, an identical "opts" is
// returned, where the Include and Exclude fields have been filled based on the
// following rules:
//
// The included and excluded references are determined based on the output of
// includeExcludeRefs (see below for documentation and detail).
//
// If any of the above could not be determined without error, that error will be
// returned immediately.
func rewriteOptions(args []string, opts *githistory.RewriteOptions, l *tasklog.Logger) (*githistory.RewriteOptions, error) {
	include, exclude, err := includeExcludeRefs(l, args)
	if err != nil {
		return nil, err
	}

	return &githistory.RewriteOptions{
		Include: include,
		Exclude: exclude,

		UpdateRefs:        opts.UpdateRefs,
		Verbose:           opts.Verbose,
		ObjectMapFilePath: opts.ObjectMapFilePath,

		BlobFn:            opts.BlobFn,
		TreePreCallbackFn: opts.TreePreCallbackFn,
		TreeCallbackFn:    opts.TreeCallbackFn,
	}, nil
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
func includeExcludeRefs(l *tasklog.Logger, args []string) (include, exclude []string, err error) {
	hardcore := len(migrateIncludeRefs) > 0 || len(migrateExcludeRefs) > 0

	if len(args) == 0 && !hardcore && !migrateEverything {
		// If no branches were given explicitly AND neither
		// --include-ref or --exclude-ref flags were given, then add the
		// currently checked out reference.
		current, err := currentRefToMigrate()
		if err != nil {
			return nil, nil, err
		}
		args = append(args, current.Name)
	}

	if migrateEverything && len(args) > 0 {
		return nil, nil, errors.New("fatal: cannot use --everything with explicit reference arguments")
	}

	for _, name := range args {
		var excluded bool
		if strings.HasPrefix("^", name) {
			name = name[1:]
			excluded = true
		}

		// Then, loop through each branch given, resolve that reference,
		// and include it.
		ref, err := git.ResolveRef(name)
		if err != nil {
			return nil, nil, err
		}

		if excluded {
			exclude = append(exclude, ref.Refspec())
		} else {
			include = append(include, ref.Refspec())
		}
	}

	if hardcore {
		if migrateEverything {
			return nil, nil, errors.New("fatal: cannot use --everything with --include-ref or --exclude-ref")
		}

		// If either --include-ref=<ref> or --exclude-ref=<ref> were
		// given, append those to the include and excluded reference
		// set, respectively.
		include = append(include, migrateIncludeRefs...)
		exclude = append(exclude, migrateExcludeRefs...)
	} else if migrateEverything {
		refs, err := git.AllRefsIn("")
		if err != nil {
			return nil, nil, err
		}

		for _, ref := range refs {
			switch ref.Type {
			case git.RefTypeLocalBranch, git.RefTypeLocalTag,
				git.RefTypeRemoteBranch:

				include = append(include, ref.Refspec())
			case git.RefTypeOther:
				parts := strings.SplitN(ref.Refspec(), "/", 3)
				if len(parts) < 2 {
					continue
				}

				switch parts[1] {
				// The following are GitLab-, GitHub-, VSTS-,
				// and BitBucket-specific reference naming
				// conventions.
				case "merge-requests", "pull", "pull-requests":
					include = append(include, ref.Refspec())
				}
			}
		}
	} else {
		bare, err := git.IsBare()
		if err != nil {
			return nil, nil, errors.Wrap(err, "fatal: unable to determine bareness")
		}

		if !bare {
			// Otherwise, if neither --include-ref=<ref> or
			// --exclude-ref=<ref> were given, include no additional
			// references, and exclude all remote references that
			// are remote branches or remote tags.
			remoteRefs, err := getRemoteRefs(l)
			if err != nil {
				return nil, nil, err
			}

			for remote, refs := range remoteRefs {
				for _, ref := range refs {
					exclude = append(exclude,
						formatRefName(ref, remote))
				}
			}
		}
	}

	return include, exclude, nil
}

// getRemoteRefs returns a fully qualified set of references belonging to all
// remotes known by the currently checked-out repository, or an error if those
// references could not be determined.
func getRemoteRefs(l *tasklog.Logger) (map[string][]*git.Ref, error) {
	refs := make(map[string][]*git.Ref)

	remotes, err := git.RemoteList()
	if err != nil {
		return nil, err
	}

	if !migrateSkipFetch {
		w := l.Waiter("migrate: Fetching remote refs")
		if err := git.Fetch(remotes...); err != nil {
			return nil, err
		}
		w.Complete()
	}

	for _, remote := range remotes {
		var refsForRemote []*git.Ref
		if migrateSkipFetch {
			refsForRemote, err = git.CachedRemoteRefs(remote)
		} else {
			refsForRemote, err = git.RemoteRefs(remote)
		}

		if err != nil {
			return nil, err
		}

		refs[remote] = refsForRemote
	}

	return refs, nil
}

// formatRefName returns the fully-qualified name for the given Git reference
// "ref".
func formatRefName(ref *git.Ref, remote string) string {
	if ref.Type == git.RefTypeRemoteBranch {
		return strings.Join([]string{
			"refs", "remotes", remote, ref.Name}, "/")
	}
	return ref.Refspec()

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
		current.Type == git.RefTypeRemoteBranch {

		return nil, errors.Errorf("fatal: cannot migrate non-local ref: %s", current.Name)
	}
	return current, nil
}

// getHistoryRewriter returns a history rewriter that includes the filepath
// filter given by the --include and --exclude arguments.
func getHistoryRewriter(cmd *cobra.Command, db *gitobj.ObjectDatabase, l *tasklog.Logger) *githistory.Rewriter {
	include, exclude := getIncludeExcludeArgs(cmd)
	filter := buildFilepathFilter(cfg, include, exclude)

	return githistory.NewRewriter(db,
		githistory.WithFilter(filter), githistory.WithLogger(l))
}

func ensureWorkingCopyClean(in io.Reader, out io.Writer) {
	dirty, err := git.IsWorkingCopyDirty()
	if err != nil {
		ExitWithError(errors.Wrap(err,
			"fatal: could not determine if working copy is dirty"))
	}

	if !dirty {
		return
	}

	var proceed bool
	if migrateYes {
		proceed = true
	} else {
		answer := bufio.NewReader(in)
	L:
		for {
			fmt.Fprintf(out, "migrate: override changes in your working copy? [Y/n] ")
			s, err := answer.ReadString('\n')
			if err != nil {
				if err == io.EOF {
					break L
				}
				ExitWithError(errors.Wrap(err,
					"fatal: could not read answer"))
			}

			switch strings.TrimSpace(s) {
			case "n", "N":
				proceed = false
				break L
			case "y", "Y":
				proceed = true
				break L
			}

			if !strings.HasSuffix(s, "\n") {
				fmt.Fprintf(out, "\n")
			}
		}
	}

	if proceed {
		fmt.Fprintf(out, "migrate: changes in your working copy will be overridden ...\n")
	} else {
		Exit("migrate: working copy must not be dirty")
	}
}

func init() {
	info := NewCommand("info", migrateInfoCommand)
	info.Flags().IntVar(&migrateInfoTopN, "top", 5, "--top=<n>")
	info.Flags().StringVar(&migrateInfoAboveFmt, "above", "", "--above=<n>")
	info.Flags().StringVar(&migrateInfoUnitFmt, "unit", "", "--unit=<unit>")

	importCmd := NewCommand("import", migrateImportCommand)
	importCmd.Flags().BoolVar(&migrateVerbose, "verbose", false, "Verbose logging")
	importCmd.Flags().StringVar(&objectMapFilePath, "object-map", "", "Object map file")
	importCmd.Flags().BoolVar(&migrateNoRewrite, "no-rewrite", false, "Add new history without rewriting previous")
	importCmd.Flags().StringVarP(&migrateCommitMessage, "message", "m", "", "With --no-rewrite, an optional commit message")
	importCmd.Flags().BoolVar(&migrateFixup, "fixup", false, "Infer filepaths based on .gitattributes")

	exportCmd := NewCommand("export", migrateExportCommand)
	exportCmd.Flags().BoolVar(&migrateVerbose, "verbose", false, "Verbose logging")
	exportCmd.Flags().StringVar(&objectMapFilePath, "object-map", "", "Object map file")
	exportCmd.Flags().StringVar(&exportRemote, "remote", "", "Remote from which to download objects")

	RegisterCommand("migrate", nil, func(cmd *cobra.Command) {
		cmd.PersistentFlags().StringVarP(&includeArg, "include", "I", "", "Include a list of paths")
		cmd.PersistentFlags().StringVarP(&excludeArg, "exclude", "X", "", "Exclude a list of paths")

		cmd.PersistentFlags().StringSliceVar(&migrateIncludeRefs, "include-ref", nil, "An explicit list of refs to include")
		cmd.PersistentFlags().StringSliceVar(&migrateExcludeRefs, "exclude-ref", nil, "An explicit list of refs to exclude")
		cmd.PersistentFlags().BoolVar(&migrateEverything, "everything", false, "Migrate all local references")
		cmd.PersistentFlags().BoolVar(&migrateSkipFetch, "skip-fetch", false, "Assume up-to-date remote references.")

		cmd.PersistentFlags().BoolVarP(&migrateYes, "yes", "y", false, "Don't prompt for answers.")

		cmd.AddCommand(exportCmd, importCmd, info)
	})
}
