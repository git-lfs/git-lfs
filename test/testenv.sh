#!/usr/bin/env bash
# Including in script/integration and every test/test-*.sh file.

set -e

# The root directory for the git-lfs repository by default.
if [ -z "$ROOTDIR" ]; then
  ROOTDIR=$(cd $(dirname "$0")/.. && pwd)
fi

# Where Git LFS outputs the compiled binaries
BINPATH="$ROOTDIR/bin"

# Put bin path on PATH
PATH="$BINPATH:$PATH"

# create a temporary work space
TMPDIR=${GIT_LFS_TEST_DIR:-"$ROOTDIR/tmp"}

# This is unique to every test file, and cleared after every test run.
TRASHDIR="$TMPDIR/$(basename "$0")-$$"

# The directory that the test Git server works from.  This cleared at the
# beginning of every test run.
REMOTEDIR="$ROOTDIR/test/remote"

# The directory that stores credentials. Credentials are stored in files with
# the username:password with filenames identifying the host (port numbers are
# ignored).
#
#   # stores the credentials for http://127.0.0.1:*
#   $CREDSDIR/127.0.0.1
#
#   # stores the credentials for http://git-server.com
#   $CREDSDIR/git-server.com
#
CREDSDIR="$REMOTEDIR/creds"

# This is the prefix for Git config files.  See the "Test Suite" section in
# test/README.md
LFS_CONFIG="$REMOTEDIR/config"

# This file contains the URL of the test Git server. See the "Test Suite"
# section in test/README.md
LFS_URL_FILE="$REMOTEDIR/url"

# the fake home dir used for the initial setup
TESTHOME="$REMOTEDIR/home"

GIT_CONFIG_NOSYSTEM=1

if [[ `git config --system credential.helper | grep osxkeychain` == "osxkeychain" ]]
then
  OSXKEYFILE="$TMPDIR/temp.keychain"
fi

mkdir -p "$TMPDIR"
mkdir -p "$TRASHDIR"

. "test/testhelpers.sh"
