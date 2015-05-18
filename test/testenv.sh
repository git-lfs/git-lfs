#!/bin/sh

set -e

# Put bin path on PATH
ROOTDIR=$(cd $(dirname "$0")/.. && pwd)
BINPATH="$ROOTDIR/bin"
PATH="$BINPATH:$PATH"

# create a temporary work space
TMPDIR="$(cd $(dirname "$0")/.. && pwd)"/tmp
TRASHDIR="$TMPDIR/$(basename "$0")-$$"

GITLFS="$BINPATH/git-lfs"
REMOTEDIR="$ROOTDIR/test/remote"
LFS_URL_FILE="$REMOTEDIR/url"
LFS_CONFIG="$REMOTEDIR/config"

mkdir -p "$TRASHDIR"

. "test/testhelpers.sh"
