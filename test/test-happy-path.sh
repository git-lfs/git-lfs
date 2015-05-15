#!/bin/sh
# this should run from the git-lfs project root.
set -e

wait_for_file() {
  local filename=$1
  for ((n=30; n>0; n--)); do
    if [ -s $filename ]; then
      return 0
    fi

    sleep 0.5
  done

  return 1
}

# cleanup
rm -rf "test/remote"
rm -rf "test/local"
mkdir -p "test/remote"
mkdir -p "test/local"
ROOTDIR=`pwd`

echo "compile git-lfs"
script/bootstrap
GITLFS="`pwd`/bin/git-lfs"

echo "spin up test server"
LFSTEST_URL="test/remote/url" go run "test/cmd/lfstest-gitserver.go" &
wait_for_file "test/remote/url"
GITSERVER=$(cat "test/remote/url")
echo $GITSERVER

echo "set up 'remote' git repository"
REPONAME="$(basename "$0")"
REPODIR="`pwd`/tests/remote/$REPONAME.git"
mkdir -p $REPODIR
cd $REPODIR
git init --bare
git config http.receivepack true
git config receive.denyCurrentBranch ignore

echo "set up 'local' test directory with git clone"
cd $ROOTDIR
TESTDIR="$(mktemp -d "`pwd`/tests/local/XXXXXX")"
cd $TESTDIR
git clone $GITSERVER/$REPONAME repo
cd repo

echo "start the test"

out=$($GITLFS track "*.dat")
echo "$out"
echo "$out" | grep "dat"

# only run if this test complete
rm -rf $TESTDIR

# run after the entire test run regardless of success or failure
curl $GITSERVER/shutdown
