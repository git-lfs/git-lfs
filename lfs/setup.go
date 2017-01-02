package lfs

import (
	"bytes"
	"fmt"
)

var (
	// prePushHook invokes `git lfs pre-push` at the pre-push phase.
	prePushHook = NewStandardHook("pre-push", []string{
		"#!/bin/sh\ngit lfs push --stdin $*",
		"#!/bin/sh\ngit lfs push --stdin \"$@\"",
		"#!/bin/sh\ngit lfs pre-push \"$@\"",
		"#!/bin/sh\ncommand -v git-lfs >/dev/null 2>&1 || { echo >&2 \"\\nThis repository has been set up with Git LFS but Git LFS is not installed.\\n\"; exit 0; }\ngit lfs pre-push \"$@\"",
		"#!/bin/sh\ncommand -v git-lfs >/dev/null 2>&1 || { echo >&2 \"\\nThis repository has been set up with Git LFS but Git LFS is not installed.\\n\"; exit 2; }\ngit lfs pre-push \"$@\"",
	})
	// postCheckoutHook invokes `git lfs post-checkout`
	postCheckoutHook = NewStandardHook("post-checkout", []string{})
	postCommitHook   = NewStandardHook("post-commit", []string{})
	postMergeHook    = NewStandardHook("post-merge", []string{})

	hooks = []*Hook{
		prePushHook,
		postCheckoutHook,
		postCommitHook,
		postMergeHook,
	}

	filters = &Attribute{
		Section: "filter.lfs",
		Properties: map[string]string{
			"clean":    "git-lfs clean -- %f",
			"smudge":   "git-lfs smudge -- %f",
			"process":  "git-lfs filter-process",
			"required": "true",
		},
		Upgradeables: map[string][]string{
			"clean":   []string{"git-lfs clean %f"},
			"smudge":  []string{"git-lfs smudge %f"},
			"process": []string{"git-lfs filter"},
		},
	}

	passFilters = &Attribute{
		Section: "filter.lfs",
		Properties: map[string]string{
			"clean":    "git-lfs clean -- %f",
			"smudge":   "git-lfs smudge --skip -- %f",
			"process":  "git-lfs filter-process --skip",
			"required": "true",
		},
		Upgradeables: map[string][]string{
			"clean":   []string{"git-lfs clean %f"},
			"smudge":  []string{"git-lfs smudge --skip %f"},
			"process": []string{"git-lfs filter --skip"},
		},
	}
)

// Get user-readable manual install steps for hooks
func GetHookInstallSteps() string {

	var buf bytes.Buffer
	for _, h := range hooks {
		buf.WriteString(fmt.Sprintf("Add the following to .git/hooks/%s :\n\n", h.Type))
		buf.WriteString(h.Contents)
		buf.WriteString("\n")
	}
	return buf.String()
}

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
