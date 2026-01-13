#!/usr/bin/env bash

. "$(dirname "$0")/testlib.sh"

begin_test "usage: no command specified"
(
  set -e

  git lfs | grep 'git lfs <command> \[<args>\]'
  git lfs help | grep 'git lfs <command> \[<args>\]'
  # Note that "git lfs --help" is handled by Git, not Git LFS.
  git lfs --foo -h | grep 'git lfs <command> \[<args>\]'
  git lfs --foo --help | grep 'git lfs <command> \[<args>\]'

  git-lfs | grep 'git lfs <command> \[<args>\]'
  git-lfs help | grep 'git lfs <command> \[<args>\]'
  git-lfs --help | grep 'git lfs <command> \[<args>\]'
  git-lfs --foo -h | grep 'git lfs <command> \[<args>\]'
  git-lfs --foo --help | grep 'git lfs <command> \[<args>\]'
)
end_test
