#!/usr/bin/env bash

. "test/testlib.sh"

begin_test "git lfs --version is a synonym of git lfs version"
(
  set -e

  reponame="git-lfs-version-synonymous"
  mkdir "$reponame"
  cd "$reponame"

  git lfs version 2>&1 >version.log
  git lfs --version 2>&1 >flag.log

  if [ "$(cat version.log)" != "$(cat flag.log)" ]; then
    echo >&2 "fatal: expected 'git lfs version' and 'git lfs --version' to"
    echo >&2 "produce identical output ..."

    diff -u {version,flag}.log
  fi
)
end_test
