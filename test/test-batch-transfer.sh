#!/bin/sh
# This is a sample Git LFS test.  See test/README.md and testhelpers.sh for
# more documentation.

. "test/testlib.sh"

begin_test "batch transfer"
(
  set -e

  # This initializes a new bare git repository in test/remote.
  # These remote repositories are global to every test, so keep the names
  # unique.
  reponame="$(basename "$0" ".sh")"
  setup_remote_repo "$reponame"

  # Clone the repository from the test Git server.  This is empty, and will be
  # used to test a "git pull" below. The repo is cloned to $TRASHDIR/clone
  clone_repo "$reponame" clone

  # Clone the repository again to $TRASHDIR/repo. This will be used to commit
  # and push objects.
  clone_repo "$reponame" repo

  # This executes Git LFS from the local repo that was just cloned.
  out=$($GITLFS track "*.dat" 2>&1)
  echo "$out" | grep "Tracking \*.dat"

  contents=$(printf "a")
  contents_oid=$(printf "$contents" | shasum -a 256 | cut -f 1 -d " ")

  printf "$contents" > a.dat
  git add a.dat
  git add .gitattributes
  out=$(git commit -m "add a.dat" 2>&1)
  echo "$out" | grep "master (root-commit)"
  echo "$out" | grep "2 files changed"
  echo "$out" | grep "create mode 100644 a.dat"
  echo "$out" | grep "create mode 100644 .gitattributes"

  out=$(cat a.dat)
  if [ "$out" != "a" ]; then
    exit 1
  fi


  # This is a small shell function that runs several git commands together.
  assert_pointer "master" "a.dat" "$contents_oid" 1

  refute_server_object "$contents_oid"

  # Ensure batch transfer is turned on for this repo
  git config --add --local lfs.batch true

  # This pushes to the remote repository set up at the top of the test.
  out=$(git push origin master 2>&1)
  echo "$out" | grep "(1 of 1 files)"
  echo "$out" | grep "master -> master"

  assert_server_object "$contents_oid" "$contents"

  # change to the clone's working directory
  cd ../clone

  git pull 2>&1 | grep "Downloading a.dat (1 B)"

  out=$(cat a.dat)
  if [ "$out" != "a" ]; then
    exit 1
  fi

  assert_pointer "master" "a.dat" "$contents_oid" 1
)
end_test
