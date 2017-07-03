#!/usr/bin/env bash

. "test/testlib.sh"

begin_test "filter-branch (git-lfs/git-lfs#1773)"
(
  set -e

  reponame="filter-branch"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  contents_a="contents (a)"
  printf "$contents_a" > a.dat
  git add a.dat
  git commit -m "add a.dat"

  contents_b="contents (b)"
  printf "$contents_b" > b.dat
  git add b.dat
  git commit -m "add b.dat"

  contents_c="contents (c)"
  printf "$contents_c" > c.dat
  git add c.dat
  git commit -m "add c.dat"

  git filter-branch -f --prune-empty \
    --tree-filter '
      echo >&2 "---"
      git rm --cached -r -q .
      git lfs track "*.dat"
      git add .
    ' --tag-name-filter cat -- --all


  assert_pointer "master" "a.dat" "$(calc_oid "$contents_a")" 12
  assert_pointer "master" "b.dat" "$(calc_oid "$contents_b")" 12
  assert_pointer "master" "c.dat" "$(calc_oid "$contents_c")" 12
)
end_test
