#!/usr/bin/env bash
# Usage: . testlib.sh
# Simple shell command language test library.
#
# Tests must follow the basic form:
#
#   begin_test "the thing"
#   (
#        set -e
#        echo "hello"
#        false
#   )
#   end_test
#
# When a test fails its stdout and stderr are shown.
#
# Note that tests must `set -e' within the subshell block or failed assertions
# will not cause the test to fail and the result may be misreported.
#
# Copyright (c) 2011-13 by Ryan Tomayko <http://tomayko.com>
# License: MIT

fullfile="$(pwd)/$0"

. "test/testenv.sh"
set -e

# keep track of num tests and failures
tests=0
failures=0

# this runs at process exit
atexit () {
  shutdown

  if [ $failures -gt 0 ]; then
    exit 1
  fi

  exit 0
}

# create the trash dir
trap "atexit" EXIT

SHUTDOWN_LFS=yes
GITSERVER=undefined

# if the file exists, assume another process started it, and will clean it up
# when it's done
if [ -s $LFS_URL_FILE ]; then
  SHUTDOWN_LFS=no
else
  setup || {
    failures=$(( failures + 1 ))
    exit $?
  }
fi

GITSERVER=$(cat "$LFS_URL_FILE")
cd "$TRASHDIR"

# Mark the beginning of a test. A subshell should immediately follow this
# statement.
begin_test () {
    test_status=$?
    [ -n "$test_description" ] && end_test $test_status
    unset test_status

    tests=$(( tests + 1 ))
    test_description="$1"

    exec 3>&1 4>&2
    out="$TRASHDIR/out"
    err="$TRASHDIR/err"
    exec 1>"$out" 2>"$err"

    # reset global git config
    HOME="$TRASHDIR/home"
    rm -rf "$TRASHDIR/home"
    mkdir "$HOME"
    cp "$TESTHOME/.gitconfig" "$HOME/.gitconfig"

    if [ "$OSXKEYFILE" ]; then
      ln -s "$TESTHOME/Library" "$HOME"
    fi

    # allow the subshell to exit non-zero without exiting this process
    set -x +e
}

# Mark the end of a test.
end_test () {
    test_status="${1:-$?}"
    set +x -e
    exec 1>&3 2>&4

    if [ "$test_status" -eq 0 ]; then
        printf "test: %-60s OK\n" "$test_description ..."
    else
        failures=$(( failures + 1 ))
        printf "test: %-60s FAILED\n" "$test_description ..."
        (
            echo "-- stdout --"
            sed 's/^/    /' <"$TRASHDIR/out"
            echo "-- stderr --"
            grep -v -e '^\+ end_test' -e '^+ set +x' <"$TRASHDIR/err" |
                sed 's/^/    /'
        ) 1>&2
    fi
    unset test_description
}
