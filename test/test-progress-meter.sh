#!/usr/bin/env bash

. "test/testlib.sh"

begin_test "progress meter displays positive progress"
(
  set -e

  reponame="progress-meter"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  git lfs track "*.dat"
  git add .gitattributes
  git commit -m "initial commit"

  for i in `seq 1 128`; do
    printf "$i" > "$i.dat"
  done

  git add *.dat
  git commit -m "add many objects"

  git push origin master 2>&1 | tee push.log
  [ "0" -eq "${PIPESTATUS[0]}" ]

  grep "Uploading LFS objects: 100% (128/128), 276 B" push.log
)
end_test
