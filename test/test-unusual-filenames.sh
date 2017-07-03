#!/usr/bin/env bash

. "test/testlib.sh"

reponame="$(basename "$0" ".sh")"

# Leading dashes may be misinterpreted as flags if commands don't use "--"
# before paths.
name1='-dash.dat'
contents1='leading dash'

begin_test "push unusually named files"
(
  set -e

  setup_remote_repo "$reponame"

  clone_repo "$reponame" repo

  git lfs track "*.dat"
  echo "$content1" > "$name1"

  git add -- .gitattributes *.dat
  git commit -m "add files"

  git push origin master | tee push.log
  grep "Git LFS: (1 of 1 files)" push.log
)
end_test

begin_test "pull unusually named files"
(
  set -e

  clone_repo "$reponame" clone

  grep "Downloading $name1" clone.log
)
end_test
