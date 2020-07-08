#!/usr/bin/env bash

. "$(dirname "$0")/testlib.sh"

begin_test "cherry-pick two commits without lfs cache"
(
  set -e

  reponame="$(basename "$0" ".sh")-cherry-pick-commits"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" cherrypickcommits

  git lfs track "*.dat"
  git add .gitattributes
  git commit -m "initial commit"

  git branch secondbranch

  echo "smudge a" > a.dat
  git add a.dat
  git commit -m "add a.dat"
  commit1=$(git log -n1 --format="%H")

  echo "smudge b" > b.dat
  git add b.dat
  git commit -m "add a.dat"
  commit2=$(git log -n1 --format="%H")

  git push origin main

  git checkout secondbranch
  rm -rf .git/lfs/objects

  git cherry-pick $commit1 $commit2
)
end_test
