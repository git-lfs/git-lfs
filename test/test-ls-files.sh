#!/bin/sh

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
  git add some.dat some.txt
  git commit -m "added some files"

  [ "some.dat" == "$(git lfs ls-files)" ]
)
end_test
