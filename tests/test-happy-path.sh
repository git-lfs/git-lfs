#!/bin/sh
# this should run from the git-lfs project root.
set -e

# cleanup
rm -rf "tests/remote"
rm -rf "tests/local"
mkdir -p "tests/remote"
mkdir -p "tests/local"
ROOTDIR=`pwd`

echo "compile git-lfs"
script/bootstrap
GITLFS="`pwd`/bin/git-lfs"

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


out=$($GITLFS track "*.dat")
echo "$out"
echo "$out" | grep "dat"

echo "ok"
rm -rf $TESTDIR
