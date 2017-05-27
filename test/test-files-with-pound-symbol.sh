#!/usr/bin/env bash

. "test/testlib.sh"

begin_test "push files with hash symbol"
(
  set -e

  reponame="git-lfs-2041"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  git lfs track "*.dat"
  git add .gitattributes
  git commit -m "initial commit"

  mkdir -p "#files"

  contents="contents"
  contents_oid="$(calc_oid "$contents")"

  printf "$contents" > "#files/a.dat"
  git add "#files/a.dat"
  git commit -m "add a.dat"

  git push origin master

  assert_server_object "$reponame" "$contents_oid"
)
end_test
