#!/usr/bin/env bash

. "test/testlib.sh"

begin_test "unlocking a lock by path"
(
  set -e

  reponame="unlock_by_path"
  setup_remote_repo_with_file "unlock_by_path" "c.dat"

  GITLFSLOCKSENABLED=1 git lfs lock "c.dat" | tee lock.log

  id=$(grep -oh "\((.*)\)" lock.log | tr -d "()")
  assert_server_lock "$reponame" "$id"

  GITLFSLOCKSENABLED=1 git lfs unlock "c.dat" 2>&1 | tee unlock.log
  refute_server_lock "$reponame" "$id"
)
end_test

begin_test "unlocking a lock (--json)"
(
  set -e

  reponame="unlock_by_path_json"
  setup_remote_repo_with_file "$reponame" "c_json.dat"

  GITLFSLOCKSENABLED=1 git lfs lock "c_json.dat" | tee lock.log

  id=$(grep -oh "\((.*)\)" lock.log | tr -d "()")
  assert_server_lock "$reponame" "$id"

  GITLFSLOCKSENABLED=1 git lfs unlock --json "c_json.dat" 2>&1 | tee unlock.log
  grep "\"unlocked\":true" unlock.log

  refute_server_lock "$reponame" "$id"
)
end_test

begin_test "unlocking a lock by id"
(
  set -e

  reponame="unlock_by_id"
  setup_remote_repo_with_file "unlock_by_id" "d.dat"

  GITLFSLOCKSENABLED=1 git lfs lock "d.dat" | tee lock.log

  id=$(grep -oh "\((.*)\)" lock.log | tr -d "()")
  assert_server_lock "$reponame" "$id"

  GITLFSLOCKSENABLED=1 git lfs unlock --id="$id" 2>&1 | tee unlock.log
  refute_server_lock "$reponame" "$id"
)
end_test

begin_test "unlocking a lock without sufficient info"
(
  set -e

  reponame="unlock_ambiguous"
  setup_remote_repo_with_file "$reponame" "e.dat"

  GITLFSLOCKSENABLED=1 git lfs lock "e.dat" | tee lock.log

  id=$(grep -oh "\((.*)\)" lock.log | tr -d "()")
  assert_server_lock "$reponame" "$id"

  GITLFSLOCKSENABLED=1 git lfs unlock 2>&1 | tee unlock.log
  grep "Usage: git lfs unlock" unlock.log
  assert_server_lock "$reponame" "$id"
)
end_test

begin_test "unlocking a lock while uncommitted"
(
  set -e

  reponame="unlock_modified"
  setup_remote_repo_with_file "$reponame" "f.dat"

  GITLFSLOCKSENABLED=1 git lfs lock "f.dat" | tee lock.log

  id=$(grep -oh "\((.*)\)" lock.log | tr -d "()")
  assert_server_lock "$reponame" "$id"

  echo "\nSomething" >> f.dat

  GITLFSLOCKSENABLED=1 git lfs unlock "f.dat" 2>&1 | tee unlock.log
  [ ${PIPESTATUS[0]} -ne "0" ]

  grep "Cannot unlock file with uncommitted changes" unlock.log

  assert_server_lock "$reponame" "$id"

  # should allow after discard
  git checkout f.dat
  GITLFSLOCKSENABLED=1 git lfs unlock "f.dat" 2>&1 | tee unlock.log
  refute_server_lock "$reponame" "$id"
)
end_test

begin_test "unlocking a lock while uncommitted with --force"
(
  set -e

  reponame="unlock_modified_force"
  setup_remote_repo_with_file "$reponame" "g.dat"

  GITLFSLOCKSENABLED=1 git lfs lock "g.dat" | tee lock.log

  id=$(grep -oh "\((.*)\)" lock.log | tr -d "()")
  assert_server_lock "$reponame" "$id"

  echo "\nSomething" >> g.dat

  # should allow with --force
  GITLFSLOCKSENABLED=1 git lfs unlock --force "g.dat" 2>&1 | tee unlock.log
  grep "Warning: unlocking with uncommitted changes" unlock.log
  refute_server_lock "$reponame" "$id"
)
end_test
