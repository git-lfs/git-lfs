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
  grep "Uploading LFS objects: 100% (1/1), 1 B" push.log
)
end_test
