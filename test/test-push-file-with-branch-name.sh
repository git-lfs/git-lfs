#!/usr/bin/env bash

. "test/testlib.sh"

begin_test "push a file with the same name as a branch"
(
  set -e

  reponame="$(basename "$0" ".sh")"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" repo

  git lfs track "master"
  echo "master" > master
  git add .gitattributes master
  git commit -m "add master"

  git lfs push --all origin master 2>&1 | tee push.log
  grep "Uploading LFS objects: 100% (1/1), 7 B" push.log
)
end_test
