#!/usr/bin/env bash
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

  # Regular Git commands can be used.
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
  grep "(1 of 1 files)" push.log
  grep "master -> master" push.log

  assert_server_object "$reponame" "$contents_oid"

  # change to the clone's working directory
  cd ../clone

  git pull

  [ "a" = "$(cat a.dat)" ]

  assert_pointer "master" "a.dat" "$contents_oid" 1
)
end_test

begin_test "happy path on non-origin remote"
(
  set -e

  reponame="happy-without-origin"
  setup_remote_repo "$reponame"

  clone_repo "$reponame" repo-without-origin
  git lfs track "*.dat"
  git add .gitattributes
  git commit -m "track"
  git push origin master

  clone_repo "$reponame" clone-without-origin
  git remote rename origin happy-path

  cd ../repo-without-origin
  echo "a" > a.dat
  git add a.dat
  git commit -m "boom"
  git push origin master

  cd ../clone-without-origin
  echo "remotes:"
  git remote
  git pull happy-path master
)
end_test

begin_test "clears local temp objects"
(
  set -e

  mkdir repo-temp-objects
  cd repo-temp-objects
  git init

  # abcdefghijklmnopqrstuvwxyz0123456789abcdefghijklmnopqrstuvwxyz01
  mkdir -p .git/lfs/objects/go/od
  mkdir -p .git/lfs/tmp/objects

  touch .git/lfs/objects/go/od/goodabcdefghijklmnopqrstuvwxyz0123456789abcdefghijklmnopqrstuvwx
  touch .git/lfs/tmp/objects/goodabcdefghijklmnopqrstuvwxyz0123456789abcdefghijklmnopqrstuvwx-rand123
  touch .git/lfs/tmp/objects/goodabcdefghijklmnopqrstuvwxyz0123456789abcdefghijklmnopqrstuvwx-rand456
  touch .git/lfs/tmp/objects/badabcdefghijklmnopqrstuvwxyz0123456789abcdefghijklmnopqrstuvwxy-rand123
  touch .git/lfs/tmp/objects/badabcdefghijklmnopqrstuvwxyz0123456789abcdefghijklmnopqrstuvwxy-rand456

  GIT_TRACE=5 git lfs env

  # object file exists
  [ -e ".git/lfs/objects/go/od/goodabcdefghijklmnopqrstuvwxyz0123456789abcdefghijklmnopqrstuvwx" ]

  # newer tmp files exist
  [ -e ".git/lfs/tmp/objects/badabcdefghijklmnopqrstuvwxyz0123456789abcdefghijklmnopqrstuvwxy-rand123" ]
  [ -e ".git/lfs/tmp/objects/badabcdefghijklmnopqrstuvwxyz0123456789abcdefghijklmnopqrstuvwxy-rand456" ]

  # existing tmp files were cleaned up
  [ ! -e ".git/lfs/tmp/objects/goodabcdefghijklmnopqrstuvwxyz0123456789abcdefghijklmnopqrstuvwx-rand123" ]
  [ ! -e ".git/lfs/tmp/objects/goodabcdefghijklmnopqrstuvwxyz0123456789abcdefghijklmnopqrstuvwx-rand456" ]
)
end_test
