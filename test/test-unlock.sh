#!/usr/bin/env bash

. "test/testlib.sh"

begin_test "unlocking a lock by path"
(
  set -e

  setup_remote_repo_with_file "unlock_by_path" "a.dat"

  git lfs lock "a.dat" | tee lock.log

  id=$(grep -oh "\((.*)\)" lock.log | tr -d "()")
  assert_server_lock $id

  git lfs unlock "a.dat" 2>&1 | tee unlock.log
  refute_server_lock $id
)
end_test

begin_test "unlocking a lock by id"
(
  set -e

  setup_remote_repo_with_file "unlock_by_id" "a.dat"

  git lfs lock "a.dat" | tee lock.log

  id=$(grep -oh "\((.*)\)" lock.log | tr -d "()")
  assert_server_lock $id

  git lfs unlock --id="$id" 2>&1 | tee unlock.log
  refute_server_lock $id
)
end_test

begin_test "unlocking a lock without sufficient info"
(
  set -e

  setup_remote_repo_with_file "unlock_ambiguous" "a.dat"

  git lfs lock "a.dat" | tee lock.log

  id=$(grep -oh "\((.*)\)" lock.log | tr -d "()")
  assert_server_lock $id

  git lfs unlock 2>&1 | tee unlock.log
  grep "Usage: git lfs unlock" unlock.log
  assert_server_lock $id
)
end_test
