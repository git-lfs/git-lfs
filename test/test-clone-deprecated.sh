#!/usr/bin/env bash

. "test/testlib.sh"

ensure_git_version_isnt $VERSION_LOWER "2.15.0"

begin_test "clone (deprecated on new versions of Git)"
(
  set -e

  reponame="clone-deprecated-recent-versions"
  setup_remote_repo "$reponame"

  mkdir -p "$reponame"
  pushd "$reponame" > /dev/null
    git lfs clone "$GITSERVER/$reponame" 2>&1 | tee clone.log
    grep "WARNING: 'git lfs clone' is deprecated and will not be updated" clone.log
  popd > /dev/null
)
end_test
