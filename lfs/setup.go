package lfs

var (
	// prePushHook invokes `git lfs push` at the pre-push phase.
	prePushHook = &Hook{
		Type:     PrePushHook,
		Contents: "#!/bin/sh\ncommand -v git-lfs >/dev/null 2>&1 || { echo >&2 \"\\nThis repository is configured for Git LFS but 'git-lfs' was not found on your path. If you no longer wish to use Git LFS, remove this hook by deleting .git/hooks/pre-push.\\n\"; exit 2; }\ngit lfs pre-push \"$@\"",
		Upgradeables: []string{
			"#!/bin/sh\ngit lfs push --stdin $*",
			"#!/bin/sh\ngit lfs push --stdin \"$@\"",
			"#!/bin/sh\ngit lfs pre-push \"$@\"",
			"#!/bin/sh\ncommand -v git-lfs >/dev/null 2>&1 || { echo >&2 \"\\nThis repository has been set up with Git LFS but Git LFS is not installed.\\n\"; exit 0; }\ngit lfs pre-push \"$@\"",
			"#!/bin/sh\ncommand -v git-lfs >/dev/null 2>&1 || { echo >&2 \"\\nThis repository has been set up with Git LFS but Git LFS is not installed.\\n\"; exit 2; }\ngit lfs pre-push \"$@\"",
		},
	}

	// hooks is a collection of all hooks to be installed by Git LFS.
	hooks = []*Hook{
		prePushHook,
	}

	// cleanFilter invokes `git lfs clean` as the clean filter.
	cleanFilter = &Filter{Name: "clean", Value: "git-lfs clean %%f"}

	// smudgeFilter invokes `git lfs smudge` as the smudge filter.
	smudgeFilter = &Filter{Name: "smudge", Value: "git-lfs smudge %%f"}

	// XXX(@ttaylorr) not sure if this makes sense as a filter? Perhaps a
	// Settable or Attribute type may be more appropriate.

	requireFilters = &Filter{Name: "required", Value: "true"}

	// filters is a collection of all filters to be installed by Git LFS.
	filters = Filters{
		cleanFilter,
		smudgeFilter,
		requireFilters,
	}
)

// InstallHooks installs all hooks in the `hooks` var.
func InstallHooks(force bool) error {
	for _, h := range hooks {
		if err := h.Install(force); err != nil {
			return err
		}
	}

	return nil
}

// UninstallHooks resmoves all hooks in range of the `hooks` var.
func UninstallHooks() error {
	for _, h := range hooks {
		if err := h.Uninstall(); err != nil {
			return err
		}
	}

	return nil
}

// SetupFilters installs filters necessary for git-lfs to process normal git
// operations. Currently, that list includes:
//   - smudge filter
//   - clean filter
//
// An error will be returned if a filter is unable to be set, or if the required
// filters were not present.
func SetupFilters(force bool) error {
	filters.Setup(force)
	return nil
}

// TeardownFilters proxies into the Teardown method on the Filters type to
// remove all installed filters.
func TeardownFilters() error {
	filters.Teardown()
	return nil
}
