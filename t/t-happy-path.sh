#!/usr/bin/env bash
# This is a sample Git LFS test.  See test/README.md and testhelpers.sh for
# more documentation.

. "$(dirname "$0")/testlib.sh"

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
  printf "%s" "$contents" > a.dat
  git add a.dat
  git add .gitattributes
  git commit -m "add a.dat" 2>&1 | tee commit.log
  grep "main (root-commit)" commit.log
  grep "2 files changed" commit.log
  grep "create mode 100644 a.dat" commit.log
  grep "create mode 100644 .gitattributes" commit.log

  [ "a" = "$(cat a.dat)" ]

  # This is a small shell function that runs several git commands together.
  assert_pointer "main" "a.dat" "$contents_oid" 1

  refute_server_object "$reponame" "$contents_oid"

  # This pushes to the remote repository set up at the top of the test.
  git push origin main 2>&1 | tee push.log
  grep "Uploading LFS objects: 100% (1/1), 1 B" push.log
  grep "main -> main" push.log

  assert_server_object "$reponame" "$contents_oid"

  # change to the clone's working directory
  cd ../clone

  git pull origin main

  [ "a" = "$(cat a.dat)" ]

  assert_pointer "main" "a.dat" "$contents_oid" 1
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
  git push origin main

  clone_repo "$reponame" clone-without-origin
  git remote rename origin happy-path

  cd ../repo-without-origin
  echo "a" > a.dat
  git add a.dat
  git commit -m "boom"
  git push origin main

  cd ../clone-without-origin
  echo "remotes:"
  git remote
  git pull happy-path main
)
end_test

begin_test "happy path on good ref"
(
  set -e

  reponame="happy-path-main-branch-required"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  git lfs track "*.dat"
  echo "a" > a.dat
  git add .gitattributes a.dat
  git commit -m "add a.dat"

  git push origin main

  # $ echo "a" | shasum -a 256
  oid="87428fc522803d31065e7bce3cf03fe475096631e5e07bbd7a0fde60c4cf25c7"
  assert_local_object "$oid" 2
  assert_server_object "$reponame" "$oid" "refs/heads/main"

  clone_repo "$reponame" "$reponame-clone"
  assert_local_object "$oid" 2
)
end_test

begin_test "happy path on tracked ref"
(
  set -e

  reponame="happy-path-tracked-branch-required"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  git lfs track "*.dat"
  echo "a" > a.dat
  git add .gitattributes a.dat
  git commit -m "add a.dat"

  git push origin main:tracked

  # $ echo "a" | shasum -a 256
  oid="87428fc522803d31065e7bce3cf03fe475096631e5e07bbd7a0fde60c4cf25c7"
  assert_local_object "$oid" 2
  assert_server_object "$reponame" "$oid" "refs/heads/tracked"

  git lfs clone "$GITSERVER/$reponame" --exclude "*"

  git config credential.helper lfstest
  git config push.default upstream
  git config branch.main.merge refs/heads/tracked

  git checkout
  assert_local_object "$oid" 2
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
