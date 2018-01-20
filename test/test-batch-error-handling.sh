#!/usr/bin/env bash
# This is a sample Git LFS test.  See test/README.md and testhelpers.sh for
# more documentation.

. "test/testlib.sh"

begin_test "batch error handling"
(
  set -e

  # This initializes a new bare git repository in test/remote.
  # These remote repositories are global to every test, so keep the names
  # unique.
  reponame="badbatch" # Server looks for the "badbatch" repo, returns a 203 status
  setup_remote_repo "$reponame"

  # Clone the repository from the test Git server.  This is empty, and will be
  # used to test a "git pull" below. The repo is cloned to $TRASHDIR/clone
  clone_repo "$reponame" clone

  # Clone the repository again to $TRASHDIR/repo. This will be used to commit
  # and push objects.
  clone_repo "$reponame" repo

  # This executes Git LFS from the local repo that was just cloned.
  git lfs track "*.dat" 2>&1 | tee track.log
  grep "Tracking \"\*.dat\"" track.log

  contents="a"
  contents_oid=$(calc_oid "$contents")

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

  # This pushes to the remote repository set up at the top of the test.
  git push origin master 2>&1 | tee push.log
  grep "Unable to parse HTTP response" push.log
)
end_test
