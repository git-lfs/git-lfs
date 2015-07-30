#!/usr/bin/env bash

. "test/testlib.sh"

begin_test "batch transfer unsupported on server"
(
  set -e

  # This initializes a new bare git repository in test/remote.
  # These remote repositories are global to every test, so keep the names
  # unique.
  reponame="$(basename "$0" ".sh")"
  setup_remote_repo "$reponame"

  # Clone the repository again to $TRASHDIR/repo. This will be used to commit
  # and push objects.
  clone_repo "$reponame" repo

  # This executes Git LFS from the local repo that was just cloned.
  git lfs track "*.dat" 2>&1 | tee track.log
  grep "Tracking \*.dat" track.log

  contents="a"
  contents_oid=$(printf "$contents" | shasum -a 256 | cut -f 1 -d " ")

  printf "$contents" > a.dat
  git add a.dat
  git add .gitattributes
  git commit -m "add a.dat" 2>&1 | tee commit.log
  grep "master (root-commit)" commit.log
  grep "2 files changed" commit.log
  grep "create mode 100644 a.dat" commit.log
  grep "create mode 100644 .gitattributes" commit.log

  [ "a" = "$(cat a.dat)" ]

  # This is a small shell function that runs several git commands together.
  assert_pointer "master" "a.dat" "$contents_oid" 1

  refute_server_object "$reponame" "$contents_oid"

  # Ensure batch transfer is turned on for this repo
  git config --add --local lfs.batch true

  # Turn off batch support on the server
  if [ -s "$LFS_URL_FILE" ]; then
    curl "$(cat "$LFS_URL_FILE")/stopbatch"
  fi

  # This pushes to the remote repository set up at the top of the test.
  git push origin master 2>&1 | tee push.log
  grep "(1 of 1 files)" push.log
  grep "master -> master" push.log

  assert_server_object "$reponame" "$contents_oid"

  # Re-enable batch support on the server
  if [ -s "$LFS_URL_FILE" ]; then
    curl "$(cat "$LFS_URL_FILE")/startbatch"
  fi

  # Assert that configuration was disabled
  set +e
  git config --file .gitconfig lfs.batch
  if [ $? -eq 0 ]
  then
    exit 1
  fi
  set -e
)
end_test
