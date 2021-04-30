#!/usr/bin/env bash

. "$(dirname "$0")/testlib.sh"

begin_test "cleans only temp files and directories older than an hour"
(
  set -e

  reponame="$(basename "$0" ".sh")"
  git init "$reponame"
  cd "$reponame"

  git lfs track '*.bin'
  echo foo > abc.bin
  git add abc.bin
  git commit -m 'Add abc.bin'

  tmpdir=.git/lfs/tmp
  mkdir -p "$tmpdir"

  mkdir "$tmpdir/dir-to-preserve"
  touch "$tmpdir/to-preserve"
  touch "$tmpdir/dir-to-preserve/file"
  # git format-patch datestamp; arbitrary timestamp in the past.
  TZ=UTC touch -t 200109170000.00 "$tmpdir/to-destroy"
  TZ=UTC touch -t 200109170000.00 "$tmpdir/dir-to-preserve/file"

  git lfs ls-files >/dev/null

  [ -f "$tmpdir/to-preserve" ]
  [ -f "$tmpdir/dir-to-preserve/file" ]
  [ ! -f "$tmpdir/to-destroy" ]
)
end_test
