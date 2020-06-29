#!/usr/bin/env bash

. "$(dirname "$0")/testlib.sh"

setup_repo () {
  setup_remote_repo_with_file "$1" "$2"

  git lfs track --lockable "*.dat"
  git add -u
  git commit -m 'Mark files lockable'
}

begin_test "unlocking a lock by path with good ref"
(
  set -e

  reponame="unlock-by-path-main-branch-required"
  setup_repo "$reponame" "c.dat"

  git lfs lock --json "c.dat" | tee lock.log

  id=$(assert_lock lock.log c.dat)
  assert_server_lock "$reponame" "$id" "refs/heads/main"

  git lfs unlock --id="$id"
  refute_server_lock "$reponame" "$id" "refs/heads/main"
)
end_test

begin_test "unlocking a lock by path with tracked ref"
(
  set -e

  reponame="unlock-by-path-tracked-branch-required"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  git lfs track "*.dat"
  echo "c" > c.dat
  git add .gitattributes c.dat
  git commit -m "add c.dat"

  git config push.default upstream
  git config branch.main.merge refs/heads/tracked
  git config branch.main.remote origin
  git push origin main

  git lfs lock --json "c.dat" | tee lock.log

  id=$(assert_lock lock.log c.dat)
  assert_server_lock "$reponame" "$id" "refs/heads/tracked"

  git lfs unlock --id="$id"
  refute_server_lock "$reponame" "$id" "refs/heads/tracked"
)
end_test

begin_test "unlocking a lock by path with bad ref"
(
  set -e

  reponame="unlock-by-path-other-branch-required"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  git lfs track "*.dat"
  echo "c" > c.dat
  git add .gitattributes c.dat
  git commit -m "add c.dat"
  git push origin main:other

  git checkout -b other
  git lfs lock --json "c.dat" | tee lock.log

  id=$(assert_lock lock.log c.dat)
  assert_server_lock "$reponame" "$id" "refs/heads/other"

  git checkout main
  git lfs unlock --id="$id" 2>&1 | tee unlock.log
  if [ "0" -eq "${PIPESTATUS[0]}" ]; then
    echo >&2 "fatal: expected 'git lfs lock \'a.dat\'' to fail"
    exit 1
  fi

  assert_server_lock "$reponame" "$id" "refs/heads/other"
  grep 'Expected ref "refs/heads/other", got "refs/heads/main"' unlock.log
)
end_test

begin_test "unlocking a lock by id with bad ref"
(
  set -e

  reponame="unlock-by-id-other-branch-required"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  git lfs track "*.dat"
  echo "c" > c.dat
  git add .gitattributes c.dat
  git commit -m "add c.dat"
  git push origin main:other

  git checkout -b other
  git lfs lock --json "c.dat" | tee lock.log

  id=$(assert_lock lock.log c.dat)
  assert_server_lock "$reponame" "$id" "refs/heads/other"

  git checkout main
  git lfs unlock --id="$id" 2>&1 | tee unlock.log
  if [ "0" -eq "${PIPESTATUS[0]}" ]; then
    echo >&2 "fatal: expected 'git lfs lock \'a.dat\'' to fail"
    exit 1
  fi

  assert_server_lock "$reponame" "$id" "refs/heads/other"
  grep 'Expected ref "refs/heads/other", got "refs/heads/main"' unlock.log
)
end_test

begin_test "unlocking a file makes it readonly"
(
  set -e

  reponame="unlock_set_readonly"
  setup_repo "$reponame" "c.dat"

  git lfs lock --json "c.dat"
  assert_file_writeable c.dat

  git lfs unlock "c.dat"
  refute_file_writeable c.dat
)
end_test

begin_test "unlocking a file ignores readonly"
(
  set -e

  reponame="unlock_set_readonly_ignore"
  setup_repo "$reponame" "c.dat"

  git lfs lock --json "c.dat"
  assert_file_writeable c.dat

  git -c lfs.setlockablereadonly=false lfs unlock "c.dat"
  assert_file_writeable c.dat
)
end_test

begin_test "unlocking lock removed file"
(
  set -e

  reponame="unlock-removed-file"
  setup_repo "$reponame" "a.dat"

  git lfs lock --json "a.dat" | tee lock.log
  id=$(assert_lock lock.log a.dat)
  assert_server_lock "$reponame" "$id"

  git rm a.dat
  git commit -m "a.dat"
  rm *.log *.json # ensure clean git status
  git status

  git lfs unlock --force "a.dat" 2>&1 | tee unlock.log
  refute_server_lock "$reponame" "$id"
)
end_test

begin_test "unlocking nonexistent file"
(
  set -e

  reponame="unlock-nonexistent-file"
  setup_repo "$reponame" "a.dat"

  git lfs lock --json "b.dat" | tee lock.log
  id=$(assert_lock lock.log b.dat)
  assert_server_lock "$reponame" "$id"

  git lfs unlock --force "b.dat" 2>&1 | tee unlock.log
  refute_server_lock "$reponame" "$id"
)
end_test

begin_test "unlocking unlockable file"
(
  set -e

  reponame="unlock-unlockable-file"
  # Try with lockable patterns.
  setup_repo "$reponame" "a.dat"

  touch README.md
  git add README.md
  git commit -m 'Add README'

  git lfs lock --json "README.md" | tee lock.log
  id=$(assert_lock lock.log README.md)
  assert_server_lock "$reponame" "$id"
  assert_file_writeable "README.md"

  git lfs unlock --force "README.md" 2>&1 | tee unlock.log
  refute_server_lock "$reponame" "$id"
  assert_file_writeable "README.md"

  cd "$TRASHDIR"
  # Try without any lockable patterns.
  setup_remote_repo_with_file "$reponame-2" "a.dat"

  touch README.md
  git add README.md
  git commit -m 'Add README'

  git lfs lock --json "README.md" | tee lock.log
  id=$(assert_lock lock.log README.md)
  assert_server_lock "$reponame-2" "$id"
  assert_file_writeable "README.md"

  git lfs unlock --force "README.md" 2>&1 | tee unlock.log
  refute_server_lock "$reponame-2" "$id"
  assert_file_writeable "README.md"
)
end_test

begin_test "unlocking a lock (--json)"
(
  set -e

  reponame="unlock_by_path_json"
  setup_repo "$reponame" "c_json.dat"

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
  setup_repo "$reponame" "d.dat"

  git lfs lock --json "d.dat" | tee lock.log
  assert_file_writeable d.dat

  id=$(assert_lock lock.log d.dat)
  assert_server_lock "$reponame" "$id"

  git lfs unlock --id="$id"
  refute_file_writeable d.dat
)
end_test

begin_test "unlocking a lock without sufficient info"
(
  set -e

  reponame="unlock_ambiguous"
  setup_repo "$reponame" "e.dat"

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
  setup_repo "$reponame" "mod.dat"

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

begin_test "unlocking a lock with ambiguious arguments"
(
  set -e

  reponame="unlock_ambiguious_args"
  setup_repo "$reponame" "a.dat"

  git lfs lock --json "a.dat" | tee lock.log

  id=$(assert_lock lock.log a.dat)
  assert_server_lock "$reponame" "$id"

  git lfs unlock --id "$id" a.dat 2>&1 | tee unlock.log
  if [ "0" -eq "${PIPESTATUS[0]}" ]; then
    echo >&2 "expected ambiguous \`git lfs unlock\` command to exit, didn't"
    exit 1
  fi

  grep "Usage:" unlock.log
  assert_server_lock "$reponame" "$id"
)
end_test

begin_test "unlocking a lock while uncommitted with --force"
(
  set -e

  reponame="unlock_modified_force"
  setup_repo "$reponame" "modforce.dat"

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
  setup_repo "$reponame" "notrelevant.dat"

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
