#!/usr/bin/env bash

. "test/testlib.sh"

begin_test "checkout"
(
  set -e

  reponame="$(basename "$0" ".sh")"
  setup_remote_repo "$reponame"

  clone_repo "$reponame" repo

  git lfs track "*.dat" 2>&1 | tee track.log
  grep "Tracking \"\*.dat\"" track.log

  contents="something something"
  contentsize=19
  contents_oid=$(calc_oid "$contents")

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

  echo "checkout should replace all"
  git lfs checkout 2>&1 | tee checkout.log
  [ "$contents" = "$(cat file1.dat)" ]
  [ "$contents" = "$(cat file2.dat)" ]
  [ "$contents" = "$(cat file3.dat)" ]
  [ "$contents" = "$(cat folder1/nested.dat)" ]
  [ "$contents" = "$(cat folder2/nested.dat)" ]
  grep "Checking out LFS objects: 100% (5/5), 95 B" checkout.log

  # Remove the working directory
  rm -rf file1.dat file2.dat file3.dat folder1/nested.dat folder2/nested.dat

  echo "checkout with filters"
  git lfs checkout file2.dat
  [ "$contents" = "$(cat file2.dat)" ]
  [ ! -f file1.dat ]
  [ ! -f file3.dat ]
  [ ! -f folder1/nested.dat ]
  [ ! -f folder2/nested.dat ]

  echo "quotes to avoid shell globbing"
  git lfs checkout "file*.dat"
  [ "$contents" = "$(cat file1.dat)" ]
  [ "$contents" = "$(cat file3.dat)" ]
  [ ! -f folder1/nested.dat ]
  [ ! -f folder2/nested.dat ]

  echo "test subdir context"
  pushd folder1
  git lfs checkout nested.dat
  [ "$contents" = "$(cat nested.dat)" ]
  [ ! -f ../folder2/nested.dat ]
  # test '.' in current dir
  rm nested.dat
  git lfs checkout . 2>&1 | tee checkout.log
  [ "$contents" = "$(cat nested.dat)" ]
  popd

  echo "test folder param"
  git lfs checkout folder2
  [ "$contents" = "$(cat folder2/nested.dat)" ]

  echo "test '.' in current dir"
  rm -rf file1.dat file2.dat file3.dat folder1/nested.dat folder2/nested.dat
  git lfs checkout .
  [ "$contents" = "$(cat file1.dat)" ]
  [ "$contents" = "$(cat file2.dat)" ]
  [ "$contents" = "$(cat file3.dat)" ]
  [ "$contents" = "$(cat folder1/nested.dat)" ]
  [ "$contents" = "$(cat folder2/nested.dat)" ]

  echo "test checkout with missing data doesn't fail"
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

begin_test "checkout: without clean filter"
(
  set -e

  reponame="$(basename "$0" ".sh")"
  git lfs uninstall

  git clone "$GITSERVER/$reponame" checkout-without-clean
  cd checkout-without-clean

  echo "checkout without clean filter"
  git lfs uninstall
  git config --list > config.txt
  grep "filter.lfs.clean" config.txt && {
    echo "clean filter still configured:"
    cat config.txt
    exit 1
  }
  ls -al

  git lfs checkout | tee checkout.txt
  grep "Git LFS is not installed" checkout.txt
  if [ "0" -ne "${PIPESTATUS[0]}" ]; then
    echo >&2 "fatal: expected checkout to succeed ..."
    exit 1
  fi

  contentsize=19
  contents_oid=$(calc_oid "something something")
  [ "$(pointer $contents_oid $contentsize)" = "$(cat file1.dat)" ]
  [ "$(pointer $contents_oid $contentsize)" = "$(cat file2.dat)" ]
  [ "$(pointer $contents_oid $contentsize)" = "$(cat file3.dat)" ]
  [ "$(pointer $contents_oid $contentsize)" = "$(cat folder1/nested.dat)" ]
  [ "$(pointer $contents_oid $contentsize)" = "$(cat folder2/nested.dat)" ]
)
end_test

begin_test "checkout: outside git repository"
(
  set +e
  git lfs checkout 2>&1 > checkout.log
  res=$?

  set -e
  if [ "$res" = "0" ]; then
    echo "Passes because $GIT_LFS_TEST_DIR is unset."
    exit 0
  fi
  [ "$res" = "128" ]
  grep "Not in a git repository" checkout.log
)
end_test
