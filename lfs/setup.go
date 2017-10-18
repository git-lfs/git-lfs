package lfs

import (
	"fmt"
	"strings"

	"github.com/git-lfs/git-lfs/tools"
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
)

// Get user-readable manual install steps for hooks
func GetHookInstallSteps() string {
	steps := make([]string, 0, len(hooks))
	for _, h := range hooks {
		steps = append(steps, fmt.Sprintf(
			"Add the following to .git/hooks/%s:\n\n%s",
			h.Type, tools.Indent(h.Contents)))
	}

	return strings.Join(steps, "\n\n")
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
