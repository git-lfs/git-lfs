#!/usr/bin/env bash

. "$(dirname "$0")/testlib.sh"

begin_test "push a file with the same name as a branch"
(
  set -e

  reponame="$(basename "$0" ".sh")"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" repo

  git lfs track "main"
  echo "main" > main
  git add .gitattributes main
  git commit -m "add main"

  git lfs push --all origin main 2>&1 | tee push.log
  grep "Uploading LFS objects: 100% (1/1), [0-9] B" push.log
)
end_test
