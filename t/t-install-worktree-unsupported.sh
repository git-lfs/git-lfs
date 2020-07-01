#!/usr/bin/env bash

. "$(dirname "$0")/testlib.sh"

# These tests rely on behavior found in Git versions less than 2.20.0 to
# perform themselves, specifically:
#   - lack of worktreeConfig extension support
ensure_git_version_isnt $VERSION_HIGHER "2.20.0"

begin_test "install --worktree with unsupported worktreeConfig extension"
(
  set -e

  reponame="$(basename "$0" ".sh")-unsupported"
  mkdir "$reponame"
  cd "$reponame"
  git init

  set +e
  git lfs install --worktree 2>err.log
  res=$?
  set -e

  cat err.log
  grep -i "error" err.log
  grep -- "--worktree" err.log
  [ "0" != "$res" ]
)
end_test
