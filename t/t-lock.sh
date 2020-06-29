#!/usr/bin/env bash

. "$(dirname "$0")/testlib.sh"

begin_test "lock with good ref"
(
  set -e

  reponame="lock-main-branch-required"
  setup_remote_repo_with_file "$reponame" "a.dat"
  clone_repo "$reponame" "$reponame"

  git lfs lock "a.dat" --json 2>&1 | tee lock.json
  if [ "0" -ne "${PIPESTATUS[0]}" ]; then
    echo >&2 "fatal: expected \'git lfs lock \'a.dat\'\' to succeed"
    exit 1
  fi

  id=$(assert_lock lock.json a.dat)
  assert_server_lock "$reponame" "$id" "refs/heads/main"
)
end_test

begin_test "lock with good tracked ref"
(
  set -e

  reponame="lock-tracked-branch-required"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  git lfs track "*.dat"
  echo "a" > a.dat
  git add .gitattributes a.dat
  git commit -m "add a.dat"

  git config push.default upstream
  git config branch.main.merge refs/heads/tracked
  git config branch.main.remote origin
  git push origin main

  git lfs lock "a.dat" --json 2>&1 | tee lock.json
  if [ "0" -ne "${PIPESTATUS[0]}" ]; then
    echo >&2 "fatal: expected \'git lfs lock \'a.dat\'\' to succeed"
    exit 1
  fi

  id=$(assert_lock lock.json a.dat)
  assert_server_lock "$reponame" "$id" "refs/heads/tracked"
)
end_test

begin_test "lock with bad ref"
(
  set -e

  reponame="lock-other-branch-required"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  git lfs track "*.dat"
  echo "a" > a.dat
  git add .gitattributes a.dat
  git commit -m "add a.dat"
  git push origin main:other

  GIT_CURL_VERBOSE=1 git lfs lock "a.dat" 2>&1 | tee lock.json
  if [ "0" -eq "${PIPESTATUS[0]}" ]; then
    echo >&2 "fatal: expected \'git lfs lock \'a.dat\'\' to fail"
    exit 1
  fi

  grep 'Lock failed: Expected ref "refs/heads/other", got "refs/heads/main"' lock.json
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

begin_test "locking a file that doesn't exist"
(
  set -e

  reponame="lock_create_nonexistent"
  setup_remote_repo_with_file "$reponame" "a_output.dat"

  git lfs lock "b_output.dat" | tee lock.log
  grep "Locked b_output.dat" lock.log
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
  grep "main (root-commit)" commit.log
  grep "2 files changed" commit.log
  grep "create mode 100644 dir/a.dat" commit.log
  grep "create mode 100644 .gitattributes" commit.log

  git push origin main 2>&1 | tee push.log
  grep "main -> main" push.log

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

  printf "%s" "$contents" > foo/bar/baz/a.dat
  git add foo/bar/baz/a.dat
  git commit -m "add a.dat"

  git push origin main

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
    echo >&2 "fatal: expected \'git lfs lock \'a.dat\'\' to succeed"
    exit 1
  fi

  id=$(assert_lock lock.json sub/a.dat)
  assert_server_lock "$reponame" "$id"
)
end_test

begin_test "creating a lock (symlinked working directory)"
(
  set -eo pipefail

  if [[ $(uname) == *"MINGW"* ]]; then
    echo >&2 "info: skipped on Windows ..."
    exit 0
  fi

  reponame="lock-in-symlinked-working-directory"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  git lfs track -l "*.dat"
  mkdir -p folder1 folder2
  printf "hello" > folder2/a.dat
  add_symlink "../folder2" "folder1/folder2"

  git add --all .
  git commit -m "initial commit"
  git push origin main

  pushd "$TRASHDIR" > /dev/null
    ln -s "$reponame" "$reponame-symlink"
    cd "$reponame-symlink"

    git lfs lock --json folder1/folder2/a.dat 2>&1 | tee lock.json

    id="$(assert_lock lock.json folder1/folder2/a.dat)"
    assert_server_lock "$reponame" "$id" main
  popd > /dev/null
)
end_test

begin_test "lock with .gitignore"
(
  set -e

  reponame="lock-with-gitignore"
  setup_remote_repo_with_file "$reponame" "a.txt"
  clone_repo "$reponame" "$reponame"

  echo "*.txt filter=lfs diff=lfs merge=lfs -text lockable" > .gitattributes

  git add .gitattributes
  git commit -m ".gitattributes: mark 'a.txt' as lockable"

  rm -f a.txt && git checkout a.txt
  refute_file_writeable a.txt

  echo "*.txt" > .gitignore
  git add .gitignore
  git commit -m ".gitignore: ignore 'a.txt'"
  rm -f a.txt && git checkout a.txt
  refute_file_writeable a.txt
)
end_test


begin_test "lock with .gitignore and lfs.lockignoredfiles"
(
  set -e

  reponame="lock-with-gitignore-and-ignoredfiles"
  setup_remote_repo_with_file "$reponame" "a.txt"
  clone_repo "$reponame" "$reponame"

  git config lfs.lockignoredfiles true
  echo "*.txt filter=lfs diff=lfs merge=lfs -text lockable" > .gitattributes

  git add .gitattributes
  git commit -m ".gitattributes: mark 'a.txt' as lockable"

  rm -f a.txt && git checkout a.txt
  refute_file_writeable a.txt

  echo "*.txt" > .gitignore
  git add .gitignore
  git commit -m ".gitignore: ignore 'a.txt'"
  rm -f a.txt && git checkout a.txt
  refute_file_writeable a.txt
)
end_test
