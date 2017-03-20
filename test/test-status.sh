#!/usr/bin/env bash

. "test/testlib.sh"

begin_test "status"
(
  set -e

  mkdir repo-1
  cd repo-1
  git init
  git lfs track "*.dat"
  echo "some data" > file1.dat
  git add file1.dat
  git commit -m "file1.dat"

  echo "other data" > file1.dat
  echo "file2 data" > file2.dat
  git add file2.dat

  echo "file3 data" > file3.dat
  git add file3.dat

  echo "file3 other data" > file3.dat

  expected="On branch master

Git LFS objects to be committed:

	file2.dat (11 B)
	file3.dat (11 B)

Git LFS objects not staged for commit:

	file1.dat"

  [ "$expected" = "$(git lfs status)" ]
)
end_test

begin_test "status --porcelain"
(
  set -e

  mkdir repo-2
  cd repo-2
  git init
  git lfs track "*.dat"
  echo "some data" > file1.dat
  git add file1.dat
  git commit -m "file1.dat"

  echo "other data" > file1.dat
  echo "file2 data" > file2.dat
  git add file2.dat

  echo "file3 data" > file3.dat
  git add file3.dat

  echo "file3 other data" > file3.dat

  expected=" M file1.dat 10
A  file2.dat 11
A  file3.dat 11"

  [ "$expected" = "$(git lfs status --porcelain)" ]
)
end_test


begin_test "status: outside git repository"
(
  set +e
  git lfs status 2>&1 > status.log
  res=$?

  set -e
  if [ "$res" = "0" ]; then
    echo "Passes because $GIT_LFS_TEST_DIR is unset."
    exit 0
  fi
  [ "$res" = "128" ]
  grep "Not in a git repository" status.log
)
end_test

begin_test "status - before initial commit"
(
  set -e

  git init repo-initial
  cd repo-initial
  git lfs track "*.dat"

  # should not fail when nothing to display (ignore output, will be blank)
  git lfs status

  echo "some data" > file1.dat
  git add file1.dat

  expected="
Git LFS objects to be committed:

	file1.dat (10 B)

Git LFS objects not staged for commit:"

  [ "$expected" = "$(git lfs status)" ]
)
end_test

begin_test "status shows multiple files with identical contents"
(
  set -e

  reponame="uniq-status"
  mkdir "$reponame"
  cd "$reponame"

  git init
  git lfs track "*.dat"

  contents="contents"
  printf "$contents" > a.dat
  printf "$contents" > b.dat

  git add --all .

  git lfs status | tee status.log

  [ "1" -eq "$(grep -c "a.dat" status.log)" ]
  [ "1" -eq "$(grep -c "b.dat" status.log)" ]
)
end_test
