#!/usr/bin/env bash

. "test/testlib.sh"

begin_test "untrack"
(
  set -e

  # no need to setup a remote repo, since this test doesn't need to push or pull

  reponame="untrack"
  git init $reponame
  cd $reponame

  # track *.jpg once
  git lfs track "*.jpg" | grep "Tracking \*.jpg"
  echo "* annex.backend=SHA512E" >> .gitattributes

  git lfs untrack "*.jpg"

  expected="* annex.backend=SHA512E"
  [ "$expected" = "$(cat .gitattributes)" ]
)
end_test

begin_test "untrack outside git repo"
(
  set -e

  reponame="outside"
  mkdir $reponame
  cd $reponame

  git lfs untrack "*.foo" || {
    # this fails if it's run outside of a git repo using GIT_LFS_TEST_DIR

    # git itself returns an exit status of 128
    # $ git show
    # fatal: Not a git repository (or any of the parent directories): .git
    # $ echo "$?"
    # 128

    [ "$?" = "128" ]
    exit 0
  }

  if [ -n "$GIT_LFS_TEST_DIR" ]; then
    echo "GIT_LFS_TEST_DIR should be set outside of any Git repository"
    exit 1
  fi
)
end_test
