#!/bin/sh
# this should run from the git-lfs project root.

. "test/testlib.sh"

begin_test "happy path"
(
  set -e

  reponame="$(basename "$0")"
  setup_remote_repo $reponame

  echo "set up 'local' test directory with git clone"
  git clone $GITSERVER/$reponame repo
  cd repo

  echo "start the test"

  out=$($GITLFS track "*.dat" 2>&1)
  echo $out | grep "dat"
)
end_test
