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
  setup_remote_repo_with_file "$reponame" "mod.dat"

  GITLFSLOCKSENABLED=1 git lfs lock "mod.dat" | tee lock.log

  id=$(grep -oh "\((.*)\)" lock.log | tr -d "()")
  assert_server_lock "$reponame" "$id"

  echo "\nSomething" >> mod.dat

  GITLFSLOCKSENABLED=1 git lfs unlock "mod.dat" 2>&1 | tee unlock.log
  [ ${PIPESTATUS[0]} -ne "0" ]

  grep "Cannot unlock file with uncommitted changes" unlock.log

  assert_server_lock "$reponame" "$id"

  # should allow after discard
  git checkout mod.dat
  GITLFSLOCKSENABLED=1 git lfs unlock "mod.dat" 2>&1 | tee unlock.log
  refute_server_lock "$reponame" "$id"
)
end_test

begin_test "unlocking a lock while uncommitted with --force"
(
  set -e

  reponame="unlock_modified_force"
  setup_remote_repo_with_file "$reponame" "modforce.dat"

  GITLFSLOCKSENABLED=1 git lfs lock "modforce.dat" | tee lock.log

  id=$(grep -oh "\((.*)\)" lock.log | tr -d "()")
  assert_server_lock "$reponame" "$id"

  echo "\nSomething" >> modforce.dat

  # should allow with --force
  GITLFSLOCKSENABLED=1 git lfs unlock --force "modforce.dat" 2>&1 | tee unlock.log
  grep "Warning: unlocking with uncommitted changes" unlock.log
  refute_server_lock "$reponame" "$id"
)
end_test

begin_test "unlocking a lock while untracked"
(
  set -e

  reponame="unlock_untracked"
  setup_remote_repo_with_file "$reponame" "notrelevant.dat"

  git lfs track "*.dat"
  # Create file but don't add it to git
  # Shouldn't be able to unlock it
  echo "something" > untracked.dat
  GITLFSLOCKSENABLED=1 git lfs lock "untracked.dat" | tee lock.log

  id=$(grep -oh "\((.*)\)" lock.log | tr -d "()")
  assert_server_lock "$reponame" "$id"

  GITLFSLOCKSENABLED=1 git lfs unlock "untracked.dat" 2>&1 | tee unlock.log
  [ ${PIPESTATUS[0]} -ne "0" ]

  grep "Cannot unlock file with uncommitted changes" unlock.log

  assert_server_lock "$reponame" "$id"

  # should allow after add/commit
  git add untracked.dat
  git commit -m "Added untracked"
  GITLFSLOCKSENABLED=1 git lfs unlock "untracked.dat" 2>&1 | tee unlock.log
  refute_server_lock "$reponame" "$id"
)
end_test
