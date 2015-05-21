#!/bin/sh
# Including in script/integration and every test/test-*.sh file.

set -e

# The root directory for the git-lfs repository
ROOTDIR=$(cd $(dirname "$0")/.. && pwd)

# Where Git LFS outputs the compiled binaries
BINPATH="$ROOTDIR/bin"

# Put bin path on PATH
PATH="$BINPATH:$PATH"

# create a temporary work space
TMPDIR=${GIT_LFS_TEST_DIR:-"$(cd $(dirname "$0")/.. && pwd)/tmp"}

# This is unique to every test file, and cleared after every test run.
TRASHDIR="$TMPDIR/$(basename "$0")-$$"

# Points to the git-lfs binary compiled just for the tests
GITLFS="$BINPATH/git-lfs"

# The directory that the test Git server works from.  This cleared at the
# beginning of every test run.
REMOTEDIR="$ROOTDIR/test/remote"

# This is the prefix for Git config files.  See the "Test Suite" section in
# test/README.md
LFS_CONFIG="$REMOTEDIR/config"

# This file contains the URL of the test Git server. See the "Test Suite"
# section in test/README.md
LFS_URL_FILE="$REMOTEDIR/url"

mkdir -p "$TRASHDIR"

. "test/testhelpers.sh"
