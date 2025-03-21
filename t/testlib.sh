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

. "$(dirname "$0")/testenv.sh"
set -e

# keep track of num tests and failures
tests=0
failures=0

# this runs at process exit
atexit () {
  tap_show_plan "$tests"
  shutdown

  if [ $failures -gt 0 ]; then
    exit 1
  fi

  exit 0
}

# create the trash dir
trap "atexit" SIGKILL SIGINT SIGTERM EXIT

GITSERVER=undefined

setup

GITSERVER=$(cat "$LFS_URL_FILE")
SSLGITSERVER=$(cat "$LFS_SSL_URL_FILE")
CLIENTCERTGITSERVER=$(cat "$LFS_CLIENT_CERT_URL_FILE")
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
    trace="$TRASHDIR/trace"

    exec 1>"$out" 2>"$err"

    # enabling GIT_TRACE can cause Windows git to stall, esp with fd 5
    # other fd numbers like 8/9 don't stall but still don't work, so disable
    if [ $IS_WINDOWS -eq 0 ]; then
      exec 5>"$trace"
      export GIT_TRACE=5
    fi

    # reset global git config
    HOME="$TRASHDIR/home"
    rm -rf "$TRASHDIR/home"
    mkdir "$HOME"
    cp "$TESTHOME/.gitconfig" "$HOME/.gitconfig"

    # allow the subshell to exit non-zero without exiting this process
    set -x +e
}

# Mark the end of a test.
end_test () {
    test_status="${1:-$?}"
    set +x -e
    exec 1>&3 2>&4
    # close fd 5 (GIT_TRACE)
    exec 5>&-

    local dump_output="$LFS_DUMP_TEST_OUTPUT"
    if [ "$test_status" -eq 0 ]; then
        printf "ok %d - %-60s\n" "$tests" "$test_description ..."
    else
        failures=$(( failures + 1 ))
        printf "not ok %d - %-60s\n" "$tests" "$test_description ..."
        dump_output=1
    fi

    if [ -n "$dump_output" ]
    then
        (
            echo "# -- stdout --"
            sed 's/^/#     /' <"$TRASHDIR/out"
            echo "# -- stderr --"
            grep -v -e '^\+ end_test' -e '^+ set +x' <"$TRASHDIR/err" |
                sed 's/^/#     /'
            if [ $IS_WINDOWS -eq 0 ]; then
                echo "# -- git trace --"
                sed 's/^/#    /' <"$TRASHDIR/trace"
            fi
        ) 1>&2
        echo
    fi
    unset test_description
}
