#!/usr/bin/env bash

. "test/testlib.sh"

begin_test "checkout"
(
  set -e

  reponame="$(basename "$0" ".sh")"
  setup_remote_repo "$reponame"

  clone_repo "$reponame" repo

  git lfs track "*.dat" 2>&1 | tee track.log
  grep "Tracking \*.dat" track.log

  contents="something something"
  contentsize=19
  contents_oid=$(printf "$contents" | shasum -a 256 | cut -f 1 -d " ")

  # Same content everywhere is ok, just one object in lfs db
  printf "$contents" > file1.dat
  printf "$contents" > file2.dat
  printf "$contents" > file3.dat
  mkdir folder1 folder2
  printf "$contents" > folder1/nested.dat
  printf "$contents" > folder2/nested.dat
  git add file1.dat file2.dat file3.dat folder1/nested.dat folder2/nested.dat
  git add .gitattributes
  git commit -m "add files"

  [ "$contents" = "$(cat file1.dat)" ]
  [ "$contents" = "$(cat file2.dat)" ]
  [ "$contents" = "$(cat file3.dat)" ]
  [ "$contents" = "$(cat folder1/nested.dat)" ]
  [ "$contents" = "$(cat folder2/nested.dat)" ]

  assert_pointer "master" "file1.dat" "$contents_oid" $contentsize

  # Remove the working directory
  rm -rf file1.dat file2.dat file3.dat folder1/nested.dat folder2/nested.dat

  # checkout should replace all
  git lfs checkout
  [ "$contents" = "$(cat file1.dat)" ]
  [ "$contents" = "$(cat file2.dat)" ]
  [ "$contents" = "$(cat file3.dat)" ]
  [ "$contents" = "$(cat folder1/nested.dat)" ]
  [ "$contents" = "$(cat folder2/nested.dat)" ]

  # Remove again
  rm -rf file1.dat file2.dat file3.dat folder1/nested.dat folder2/nested.dat

  # checkout with filters
  git lfs checkout file2.dat
  [ "$contents" = "$(cat file2.dat)" ]
  [ ! -f file1.dat ]
  [ ! -f file3.dat ]
  [ ! -f folder1/nested.dat ]
  [ ! -f folder2/nested.dat ]

  # quotes to avoid shell globbing
  git lfs checkout "file*.dat"
  [ "$contents" = "$(cat file1.dat)" ]
  [ "$contents" = "$(cat file3.dat)" ]
  [ ! -f folder1/nested.dat ]
  [ ! -f folder2/nested.dat ]

  # test subdir context
  pushd folder1
  git lfs checkout nested.dat
  [ "$contents" = "$(cat nested.dat)" ]
  [ ! -f ../folder2/nested.dat ]
  # test '.' in current dir
  rm nested.dat
  git lfs checkout .
  [ "$contents" = "$(cat nested.dat)" ]  
  popd

  # test folder param
  git lfs checkout folder2
  [ "$contents" = "$(cat folder2/nested.dat)" ]

  # test '.' in current dir
  rm -rf file1.dat file2.dat file3.dat folder1/nested.dat folder2/nested.dat
  git lfs checkout .
  [ "$contents" = "$(cat file1.dat)" ]
  [ "$contents" = "$(cat file2.dat)" ]
  [ "$contents" = "$(cat file3.dat)" ]
  [ "$contents" = "$(cat folder1/nested.dat)" ]
  [ "$contents" = "$(cat folder2/nested.dat)" ]

  # test checkout with missing data doesn't fail
  git push origin master
  rm -rf .git/lfs/objects
  rm file*.dat
  git lfs checkout
  [ "$(pointer $contents_oid $contentsize)" = "$(cat file1.dat)" ]
  [ "$(pointer $contents_oid $contentsize)" = "$(cat file2.dat)" ]
  [ "$(pointer $contents_oid $contentsize)" = "$(cat file3.dat)" ]
  [ "$contents" = "$(cat folder1/nested.dat)" ]
  [ "$contents" = "$(cat folder2/nested.dat)" ]

)
end_test
