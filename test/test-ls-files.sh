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

begin_test "ls-files: outside git repository"
(
  set +e
  git lfs ls-files 2>&1 > ls-files.log
  res=$?

  set -e
  if [ "$res" = "0" ]; then
    echo "Passes because $GIT_LFS_TEST_DIR is unset."
    exit 0
  fi
  [ "$res" = "128" ]
  grep "Not in a git repository" ls-files.log
)
end_test

begin_test "ls-files: with zero files"
(
  set -e
  mkdir empty
  cd empty
  git init
  git lfs track "*.dat"
  git add .gitattributes

  set +e
  git lfs ls-files 2> ls-files.log
  res=$?
  set -e

  cat ls-files.log
  [ "$res" = "2" ]
  grep "Git can't resolve ref:" ls-files.log

  git commit -m "initial commit"
  [ "$(git lfs ls-files)" = "" ]
)
end_test
