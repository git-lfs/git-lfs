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
