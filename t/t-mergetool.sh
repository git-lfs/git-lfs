#!/usr/bin/env bash

. "$(dirname "$0")/testlib.sh"

begin_test "mergetool works with large files"
(
  set -e

  reponame="mergetool-works-with-large-files"
  git init "$reponame"
  cd "$reponame"

  git lfs track "*.dat"
  printf "base" > conflict.dat
  git add .gitattributes conflict.dat
  git commit -m "initial commit"

  git checkout -b conflict
  printf "b" > conflict.dat
  git add conflict.dat
  git commit -m "conflict.dat: b"

  git checkout main

  printf "a" > conflict.dat
  git add conflict.dat
  git commit -m "conflict.dat: a"

  set +e
  git merge conflict
  set -e

  git config mergetool.inspect.cmd '
    for i in BASE LOCAL REMOTE; do
      echo "\$$i=$(eval "cat \"\$$i\"")";
    done;
    exit 1
  '
  git config mergetool.inspect.trustExitCode true

  yes | git mergetool \
      --no-prompt \
      --tool=inspect \
      -- conflict.dat 2>&1 \
    | tee mergetool.log

  grep "\$BASE=base" mergetool.log
  grep "\$LOCAL=a" mergetool.log
  grep "\$REMOTE=b" mergetool.log
)
end_test
