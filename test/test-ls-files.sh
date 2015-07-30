#!/usr/bin/env bash

. "test/testlib.sh"

begin_test "ls-files"
(
  set -e

  mkdir repo
  cd repo
  git init
  git lfs track "*.dat" | grep "Tracking \*.dat"
  echo "some data" > some.dat
  echo "some text" > some.txt
  echo "missing" > missing.dat
  git add missing.dat
  git commit -m "add missing file"
  [ "missing.dat" = "$(git lfs ls-files)" ]

  git rm missing.dat
  git add some.dat some.txt
  git commit -m "added some files, removed missing one"

  git lfs ls-files | tee ls.log
  grep some.dat ls.log
  [ `wc -l < ls.log` = 1 ]
)
end_test
