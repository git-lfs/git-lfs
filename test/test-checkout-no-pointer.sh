#!/usr/bin/env bash

. "test/testlib.sh"

begin_test "git-lfs/git-lfs#1997"
(
  set -e

  reponame="git-lfs-1997"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  git lfs track "*.dat"
  git add .gitattributes
  git commit -m "initial commit"

  git checkout -b my-feature

  contents="first file"
  contents_oid="$(calc_oid "$contents")"
  contents_len="$(printf "$contents" | wc -c)"
  printf "$contents" > a.dat

  git add a.dat
  git commit -m "add a.dat"

  assert_pointer "my-feature" "a.dat" "$contents_oid" "$contents_len"

  git checkout master 2>&1 | tee checkout.log
  if [ "0" -ne "${PIPESTATUS[0]}" ]; then
    echo >&2 "'git checkout master' should have been successful, wasn't ..."
    exit 1
  fi

  [ ! -f a.dat ]
)
end_test
