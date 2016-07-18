#!/usr/bin/env bash
# Including in script/integration and every test/test-*.sh file.

set -e

UNAME=$(uname -s)
IS_WINDOWS=0
IS_MAC=0
if [[ $UNAME == MINGW* || $UNAME == CYGWIN* ]]
then
  IS_WINDOWS=1
elif [[ $UNAME == *Darwin* ]]
then
  IS_MAC=1
fi

resolve_symlink() {
  local arg=$1
  if [ $IS_WINDOWS == "1" ]; then
    printf '%s' "$arg"
  elif [ $IS_MAC == "1" ]; then
    # no readlink -f on Mac
    local oldwd=$(pwd)
    local target=$arg

    cd `dirname $target`
    target=`basename $target`
    while [ -L "$target" ]
    do
        target=`readlink $target`
        cd `dirname $target`
        target=`basename $target`
    done

    local resolveddir=`pwd -P`
    cd "$oldwd"
    printf '%s' "$resolveddir/$target"

  else
    readlink -f "$arg"
  fi

}

# The root directory for the git-lfs repository by default.
if [ -z "$ROOTDIR" ]; then
  ROOTDIR=$(cd $(dirname "$0")/.. && pwd -P)
fi

# Where Git LFS outputs the compiled binaries
BINPATH="$ROOTDIR/bin"

# Put bin path on PATH
PATH="$BINPATH:$PATH"

# Always provide a test dir outside our git repo if not specified
TEMPDIR_PREFIX="git-lfs_TEMP.XXXXXX"
if [ -z "$GIT_LFS_TEST_DIR" ]; then
    GIT_LFS_TEST_DIR=$(mktemp -d -t "$TEMPDIR_PREFIX")
    GIT_LFS_TEST_DIR=$(resolve_symlink $GIT_LFS_TEST_DIR)
    # cleanup either after single test or at end of integration (except on fail)
    RM_GIT_LFS_TEST_DIR=yes
fi
# create a temporary work space
TMPDIR=$GIT_LFS_TEST_DIR

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

# This file contains the SSL URL of the test Git server. See the "Test Suite"
# section in test/README.md
LFS_SSL_URL_FILE="$REMOTEDIR/sslurl"

# This file contains the self-signed SSL cert of the TLS endpoint of the test Git server.
LFS_CERT_FILE="$REMOTEDIR/cert"

# the fake home dir used for the initial setup
TESTHOME="$REMOTEDIR/home"

GIT_CONFIG_NOSYSTEM=1
GIT_TERMINAL_PROMPT=0

export CREDSDIR

if [[ `git config --system credential.helper | grep osxkeychain` == "osxkeychain" ]]
then
  OSXKEYFILE="$TMPDIR/temp.keychain"
fi

mkdir -p "$TMPDIR"
mkdir -p "$TRASHDIR"


if [ $IS_WINDOWS == "1" ]; then
  # prevent Windows OpenSSH from opening GUI prompts
  SSH_ASKPASS=""
fi
. "test/testhelpers.sh"
