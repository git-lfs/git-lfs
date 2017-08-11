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

begin_test "locking a nested file"
(
  set -e

  reponame="locking-nested-file"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  git lfs track "*.dat" --lockable
  git add .gitattributes
  git commit -m "initial commit"

  mkdir -p foo/bar/baz

  contents="contents"
  contents_oid="$(calc_oid "$contents")"

  printf "$contents" > foo/bar/baz/a.dat
  git add foo/bar/baz/a.dat
  git commit -m "add a.dat"

  git push origin master

  assert_server_object "$reponame" "$contents_oid"

  git lfs lock foo/bar/baz/a.dat 2>&1 | tee lock.log
  grep "Locked foo/bar/baz/a.dat" lock.log

  git lfs locks 2>&1 | tee locks.log
  grep "foo/bar/baz/a.dat" locks.log
)
end_test

begin_test "creating a lock (within subdirectory)"
(
  set -e

  reponame="lock_create_within_subdirectory"
  setup_remote_repo_with_file "$reponame" "sub/a.dat"

  cd sub

  git lfs lock --json "a.dat" | tee lock.json
  if [ "0" -ne "${PIPESTATUS[0]}" ]; then
    echo >&2 "fatal: expected 'git lfs lock \'a.dat\'' to succeed"
    exit 1
  fi

  id=$(assert_lock lock.json a.dat)
  assert_server_lock "$reponame" "$id"
)
end_test
