#!/usr/bin/env bash

. "test/testlib.sh"

begin_test "post-merge with owned locks (into master)"
(
  set -e

  reponame="post-merge-owned"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  git lfs track "*.dat"
  git add .gitattributes
  git commit -m "initial commit"

  printf "contents" > locked_pmo.dat
  git add locked_pmo.dat
  git commit -m "add locked file"

  git push origin master

  GITLFSLOCKSENABLED=1 git lfs lock "locked_pmo.dat" | tee lock.log
  grep "'locked_pmo.dat' was locked" lock.log

  lock_id=$(grep -oh "\((.*)\)" lock.log | tr -d "()")
  assert_server_lock $lock_id

  git checkout -b my-feature

  printf "(more) contents" >> locked_pmo.dat
  git add locked_pmo.dat
  git commit -m "change locked file"

  git push origin my-feature

  assert_server_lock $lock_id

  git checkout master

  git merge my-feature 2>&1 | tee merge.log
  grep "Unlocking 1 files" merge.log
  grep "locked_pmo.dat" merge.log

  git push origin master

  refute_server_lock "$lock_id"
)
end_test

begin_test "post-merge with owned locks (into non-master)"
(
  set -e

  reponame="post-merge-owned-non-master"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  git lfs track "*.dat"
  git add .gitattributes
  git commit -m "initial commit"

  printf "contents" > locked_pmo1.dat
  git add locked_pmo1.dat
  git commit -m "add locked file"

  git push origin master

  GITLFSLOCKSENABLED=1 git lfs lock "locked_pmo1.dat" | tee lock.log
  grep "'locked_pmo1.dat' was locked" lock.log

  lock_id=$(grep -oh "\((.*)\)" lock.log | tr -d "()")
  assert_server_lock $lock_id

  git checkout -b my-feature
  git checkout -b my-sub-feature

  printf "(more) contents" >> locked_pmo1.dat
  git add locked_pmo1.dat
  git commit -m "change locked file"

  git push origin my-sub-feature

  assert_server_lock $lock_id

  git checkout my-feature

  git merge my-sub-feature 2>&1 | tee merge.log
  [ "0" -eq "$(grep -c "Unlocking" merge.log)" ]

  git push origin my-feature

  assert_server_lock "$lock_id"
)
end_test
