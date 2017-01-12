#!/usr/bin/env bash

. "test/testlib.sh"

begin_test "creating a lock"
(
  set -e

  reponame="lock_create_simple"
  setup_remote_repo_with_file "$reponame" "a.dat"

  GITLFSLOCKSENABLED=1 git lfs lock "a.dat" | tee lock.log
  grep "'a.dat' was locked" lock.log

  id=$(grep -oh "\((.*)\)" lock.log | tr -d "()")
  assert_server_lock "$reponame" "$id"
)
end_test

begin_test "creating a lock (--json)"
(
  set -e

  reponame="lock_create_simple_json"
  setup_remote_repo_with_file "$reponame" "a_json.dat"

  GITLFSLOCKSENABLED=1 git lfs lock --json "a_json.dat" | tee lock.log
  grep "\"path\":\"a_json.dat\"" lock.log

  id=$(grep -o "\"id\":\".*\"" lock.log | cut -d \" -f 4)
  assert_server_lock "$reponame" "$id"
)
end_test

begin_test "locking a previously locked file"
(
  set -e

  reponame="lock_create_previously_created"
  setup_remote_repo_with_file "$reponame" "b.dat"

  GITLFSLOCKSENABLED=1 git lfs lock "b.dat" | tee lock.log
  grep "'b.dat' was locked" lock.log

  id=$(grep -oh "\((.*)\)" lock.log | tr -d "()")
  assert_server_lock "$reponame" "$id"

  grep "lock already created" <(GITLFSLOCKSENABLED=1 git lfs lock "b.dat" 2>&1)
)
end_test

begin_test "locking a directory"
(
  set -e

  reponame="locking_directories"
  setup_remote_repo "remote_$reponame"
  clone_repo "remote_$reponame" "clone_$reponame"

  git lfs track "*.dat"
  mkdir dir
  echo "a" > dir/a.dat

  git add dir/a.dat .gitattributes

  git commit -m "add dir/a.dat" | tee commit.log
  grep "master (root-commit)" commit.log
  grep "2 files changed" commit.log
  grep "create mode 100644 dir/a.dat" commit.log
  grep "create mode 100644 .gitattributes" commit.log

  git push origin master 2>&1 | tee push.log
  grep "master -> master" push.log

  GITLFSLOCKSENABLED=1 git lfs lock ./dir/ 2>&1 | tee lock.log
  grep "cannot lock directory" lock.log
)
end_test
