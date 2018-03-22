#!/usr/bin/env bash

. "test/testlib.sh"

begin_test "mergetool"
(
  set -e

  reponame="mergetool"
  git init "$reponame"
  cd "$reponame"

  git config --add 'mergetool.env.cmd' 'env'

  echo "*.dat filter=lfs merge=lfs diff=lfs mergetool=env" >> .gitattributes
  # git lfs track "*.dat"
  git add .gitattributes
  git commit -m "initial commit"

  git checkout -b conflict
  printf "b" > a.dat
  git add a.dat
  git commit -m "a.dat: b"

  git checkout master
  printf "a" > a.dat
  git add a.dat
  git commit -m "a.dat: a"

  git merge conflict | grep "CONFLICT"

  git lfs mergetool --no-prompt 2>&1 | tee mergetool.log
  exit 1
)
end_test
