#!/usr/bin/env bash

. "test/testlib.sh"

begin_test "unlocking a lock by path"
(
  set -e

  reponame="unlock_by_path"
  setup_remote_repo_with_file "unlock_by_path" "c.dat"

  git lfs lock --json "c.dat" | tee lock.log

  id=$(assert_lock lock.log c.dat)
  assert_server_lock "$reponame" "$id"

  git lfs unlock "c.dat" 2>&1 | tee unlock.log
  refute_server_lock "$reponame" "$id"
)
end_test

begin_test "force unlocking lock with missing file"
(
  set -e

  reponame="force-unlock-missing-file"
  setup_remote_repo_with_file "$reponame" "a.dat"

  git lfs lock --json "a.dat" | tee lock.log
  id=$(assert_lock lock.log a.dat)
  assert_server_lock "$reponame" "$id"

  git rm a.dat
  git commit -m "a.dat"
  rm *.log *.json # ensure clean git status
  git status

  git lfs unlock "a.dat" 2>&1 | tee unlock.log
  grep "Unable to determine path" unlock.log
  assert_server_lock "$reponame" "$id"

  rm unlock.log
  git lfs unlock --force "a.dat" 2>&1 | tee unlock.log
  refute_server_lock "$reponame" "$id"
)
end_test

begin_test "unlocking a lock (--json)"
(
  set -e

  reponame="unlock_by_path_json"
  setup_remote_repo_with_file "$reponame" "c_json.dat"

  git lfs lock --json "c_json.dat" | tee lock.log

  id=$(assert_lock lock.log c_json.dat)
  assert_server_lock "$reponame" "$id"

  git lfs unlock --json "c_json.dat" 2>&1 | tee unlock.log
  grep "\"unlocked\":true" unlock.log

  refute_server_lock "$reponame" "$id"
)
end_test

begin_test "unlocking a lock by id"
(
  set -e

  reponame="unlock_by_id"
  setup_remote_repo_with_file "unlock_by_id" "d.dat"

  git lfs lock --json "d.dat" | tee lock.log

  id=$(assert_lock lock.log d.dat)
  assert_server_lock "$reponame" "$id"

  git lfs unlock --id="$id" 2>&1 | tee unlock.log
  refute_server_lock "$reponame" "$id"
)
end_test

begin_test "unlocking a lock without sufficient info"
(
  set -e

  reponame="unlock_ambiguous"
  setup_remote_repo_with_file "$reponame" "e.dat"

  git lfs lock --json "e.dat" | tee lock.log

  id=$(assert_lock lock.log e.dat)
  assert_server_lock "$reponame" "$id"

  git lfs unlock 2>&1 | tee unlock.log
  grep "Usage: git lfs unlock" unlock.log
  assert_server_lock "$reponame" "$id"
)
end_test

begin_test "unlocking a lock while uncommitted"
(
  set -e

  reponame="unlock_modified"
  setup_remote_repo_with_file "$reponame" "mod.dat"

  git lfs lock --json "mod.dat" | tee lock.log

  id=$(assert_lock lock.log mod.dat)
  assert_server_lock "$reponame" "$id"

  echo "\nSomething" >> mod.dat

  git lfs unlock "mod.dat" 2>&1 | tee unlock.log
  [ ${PIPESTATUS[0]} -ne "0" ]

  grep "Cannot unlock file with uncommitted changes" unlock.log

  assert_server_lock "$reponame" "$id"

  # should allow after discard
  git checkout mod.dat
  git lfs unlock "mod.dat" 2>&1 | tee unlock.log
  refute_server_lock "$reponame" "$id"
)
end_test

begin_test "unlocking a lock while uncommitted with --force"
(
  set -e

  reponame="unlock_modified_force"
  setup_remote_repo_with_file "$reponame" "modforce.dat"

  git lfs lock --json "modforce.dat" | tee lock.log

  id=$(assert_lock lock.log modforce.dat)
  assert_server_lock "$reponame" "$id"

  echo "\nSomething" >> modforce.dat

  # should allow with --force
  git lfs unlock --force "modforce.dat" 2>&1 | tee unlock.log
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
  git lfs lock --json "untracked.dat" | tee lock.log

  id=$(assert_lock lock.log untracked.dat)
  assert_server_lock "$reponame" "$id"

  git lfs unlock "untracked.dat" 2>&1 | tee unlock.log
  [ ${PIPESTATUS[0]} -ne "0" ]

  grep "Cannot unlock file with uncommitted changes" unlock.log

  assert_server_lock "$reponame" "$id"

  # should allow after add/commit
  git add untracked.dat
  git commit -m "Added untracked"
  git lfs unlock "untracked.dat" 2>&1 | tee unlock.log
  refute_server_lock "$reponame" "$id"
)
end_test
