#!/usr/bin/env bash

. "test/testlib.sh"

begin_test "pre-commit with own locked files"
(
  set -e

  reponame="pre-commit-owned-locks"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  git lfs track "*.dat"
  git add .gitattributes
  git commit -m "initial commit"

  contents="locked contents"
  printf "$contents" > locked_pc_auth.dat
  git add locked_pc_auth.dat
  git commit -m "add locked_pc_auth.dat"

  git push origin master

  GITLFSLOCKSENABLED=1 git lfs lock "locked_pc_auth.dat" | tee lock.log
  grep "'locked_pc_auth.dat' was locked" lock.log

  id=$(grep -oh "\((.*)\)" lock.log | tr -d "()")
  assert_server_lock $id

  printf "authorized changes" >> locked_pc_auth.dat
  git add locked_pc_auth.dat
  git commit -m "locked changes"
)
end_test
