#!/usr/bin/env bash

. "test/testlib.sh"

begin_test "logs"
(
  set -e

  mkdir logs
  cd logs
  git init

  boomtownExit=""
  set +e
  git lfs logs boomtown
  boomtownExit=$?
  set -e

  [ "$boomtownExit" = "2" ]

  logname=`ls .git/lfs/logs`
  logfile=".git/lfs/logs/$logname"
  cat "$logfile"
  echo "... grep ..."
  grep "$ git-lfs logs boomtown" "$logfile"

  [ "$(cat "$logfile")" = "$(git lfs logs last)" ]
)
end_test
