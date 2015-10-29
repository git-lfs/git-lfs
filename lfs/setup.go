package lfs

var (
	// prePushHook invokes `git lfs push` at the pre-push phase.
	prePushHook = &Hook{
		Type:     "pre-push",
		Contents: "#!/bin/sh\ncommand -v git-lfs >/dev/null 2>&1 || { echo >&2 \"\\nThis repository is configured for Git LFS but 'git-lfs' was not found on your path. If you no longer wish to use Git LFS, remove this hook by deleting .git/hooks/pre-push.\\n\"; exit 2; }\ngit lfs pre-push \"$@\"",
		Upgradeables: []string{
			"#!/bin/sh\ngit lfs push --stdin $*",
			"#!/bin/sh\ngit lfs push --stdin \"$@\"",
			"#!/bin/sh\ngit lfs pre-push \"$@\"",
			"#!/bin/sh\ncommand -v git-lfs >/dev/null 2>&1 || { echo >&2 \"\\nThis repository has been set up with Git LFS but Git LFS is not installed.\\n\"; exit 0; }\ngit lfs pre-push \"$@\"",
			"#!/bin/sh\ncommand -v git-lfs >/dev/null 2>&1 || { echo >&2 \"\\nThis repository has been set up with Git LFS but Git LFS is not installed.\\n\"; exit 2; }\ngit lfs pre-push \"$@\"",
		},
	}

	hooks = []*Hook{
		prePushHook,
	}

	filters = &Attribute{
		Section: "filter.lfs",
		Properties: map[string]string{
			"clean":    "git-lfs clean %f",
			"smudge":   "git-lfs smudge %f",
			"required": "true",
		},
	}

	passFilters = &Attribute{
		Section: "filter.lfs",
		Properties: map[string]string{
			"clean":    "git-lfs clean %f",
			"smudge":   "git-lfs smudge --skip %f",
			"required": "true",
		},
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

// UninstallHooks removes all hooks in range of the `hooks` var.
func UninstallHooks() error {
	for _, h := range hooks {
		if err := h.Uninstall(); err != nil {
			return err
		}
	}

	return nil
}

// InstallFilters installs filters necessary for git-lfs to process normal git
// operations. Currently, that list includes:
//   - smudge filter
//   - clean filter
//
// An error will be returned if a filter is unable to be set, or if the required
// filters were not present.
func InstallFilters(opt InstallOptions, passThrough bool) error {
	if passThrough {
		return passFilters.Install(opt)
	}
	return filters.Install(opt)
}

// UninstallFilters proxies into the Uninstall method on the Filters type to
// remove all installed filters.
func UninstallFilters() error {
	filters.Uninstall()
	return nil
}
