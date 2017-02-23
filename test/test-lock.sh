#!/usr/bin/env bash

. "test/testlib.sh"

begin_test "creating a lock"
(
  set -e

  reponame="lock_create_simple"
  setup_remote_repo_with_file "$reponame" "a.dat"

  git lfs lock --json "a.dat" | tee lock.json
  id=$(assert_lock lock.json a.dat)
  assert_server_lock "$reponame" "$id"
)
end_test

begin_test "create lock with server using client cert"
(
  set -e
  reponame="lock_create_client_cert"
  setup_remote_repo_with_file "$reponame" "cc.dat"

  git config lfs.url "$CLIENTCERTGITSERVER/$reponame.git/info/lfs"
  git lfs lock --json "cc.dat" | tee lock.json
  id=$(assert_lock lock.json cc.dat)
  assert_server_lock "$reponame" "$id"
)
end_test

begin_test "creating a lock (with output)"
(
  set -e

  reponame="lock_create_simple_output"
  setup_remote_repo_with_file "$reponame" "a_output.dat"

  git lfs lock "a_output.dat" | tee lock.log
  grep "Locked a_output.dat" lock.log
  id=$(grep -oh "\((.*)\)" lock.log | tr -d \(\))
  assert_server_lock "$reponame" "$id"
)
end_test

begin_test "locking a previously locked file"
(
  set -e

  reponame="lock_create_previously_created"
  setup_remote_repo_with_file "$reponame" "b.dat"

  git lfs lock --json "b.dat" | tee lock.json
  id=$(assert_lock lock.json b.dat)
  assert_server_lock "$reponame" "$id"

  grep "lock already created" <(git lfs lock "b.dat" 2>&1)
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

  git lfs lock ./dir/ 2>&1 | tee lock.log
  grep "cannot lock directory" lock.log
)
end_test
