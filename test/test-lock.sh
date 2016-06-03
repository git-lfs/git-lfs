#!/usr/bin/env bash

. "test/testlib.sh"

begin_test "creating a lock"
(
  set -e

  setup_remote_repo_with_file "lock_create_simple" "a.dat"

  git lfs lock "a.dat" | tee lock.log
  grep "'a.dat' was locked" lock.log

  id=$(grep -oh "\((.*)\)" lock.log | tr -d "()")
  assert_server_lock $id
)
end_test

begin_test "locking a previously locked file"
(
  set -e

  setup_remote_repo_with_file "lock_create_previously_created" "b.dat"

  git lfs lock "b.dat" | tee lock.log
  grep "'b.dat' was locked" lock.log

  id=$(grep -oh "\((.*)\)" lock.log | tr -d "()")
  assert_server_lock $id

  grep "lock already created" <(git lfs lock "b.dat" 2>&1)
)
end_test
