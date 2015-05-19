#!/bin/sh
# This is a sample Git LFS test.  See test/README.md and testhelpers.sh for
# more documentation.

. "test/testlib.sh"

begin_test "happy path"
(
  set -e

  # This initializes a new bare git repository in test/remote.
  # These remote repositories are global to every test, so keep the names
  # unique.
  reponame="$(basename "$0" ".sh")"
  setup_remote_repo "$reponame"

  # Clone the repository through the test Git server.  It's cloned to
  # $TRASHDIR/repo.
  clone_repo "$reponame" repo

  # This executes Git LFS from the local repo that was just cloned.
  out=$($GITLFS track "*.dat" 2>&1)
  echo "$out" | grep "Tracking \*.dat"

  contents=$(printf "a")
  contents_oid=$(printf "$contents" | shasum -a 256 | cut -f 1 -d " ")

  # Regular Git commands can be used.
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

  # This pushes to the remote repository set up at the top of the test.
  out=$(git push origin master 2>&1)
  echo "$out" | grep "(1 of 1 files) 1 B / 1 B  100.00 %"
  echo "$out" | grep "master -> master"

  assert_server_object "$contents_oid" "$contents"

  # This clones the repository to another subdirectory of $TRASHDIR
  out=$(clone_repo "$reponame" clone)
  echo "$out" | grep "Cloning into 'clone'"
  echo "$out" | grep "Downloading a.dat (1 B)"

  out=$(cat a.dat)
  if [ "$out" != "a" ]; then
    exit 1
  fi

  assert_pointer "master" "a.dat" "$contents_oid" 1
)
end_test
