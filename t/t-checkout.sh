#!/usr/bin/env bash

. "$(dirname "$0")/testlib.sh"

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
  printf "%s" "$contents" > file1.dat
  printf "%s" "$contents" > file2.dat
  printf "%s" "$contents" > file3.dat
  mkdir folder1 folder2
  printf "%s" "$contents" > folder1/nested.dat
  printf "%s" "$contents" > folder2/nested.dat
  git add file1.dat file2.dat file3.dat folder1/nested.dat folder2/nested.dat
  git add .gitattributes
  git commit -m "add files"

  [ "$contents" = "$(cat file1.dat)" ]
  [ "$contents" = "$(cat file2.dat)" ]
  [ "$contents" = "$(cat file3.dat)" ]
  [ "$contents" = "$(cat folder1/nested.dat)" ]
  [ "$contents" = "$(cat folder2/nested.dat)" ]

  assert_pointer "main" "file1.dat" "$contents_oid" $contentsize

  # Remove the working directory
  rm -rf file1.dat file2.dat file3.dat folder1/nested.dat folder2/nested.dat

  echo "checkout should replace all"
  GIT_TRACE=1 git lfs checkout 2>&1 | tee checkout.log
  [ "$contents" = "$(cat file1.dat)" ]
  [ "$contents" = "$(cat file2.dat)" ]
  [ "$contents" = "$(cat file3.dat)" ]
  [ "$contents" = "$(cat folder1/nested.dat)" ]
  [ "$contents" = "$(cat folder2/nested.dat)" ]
  grep "Checking out LFS objects: 100% (5/5), 95 B" checkout.log
  grep 'accepting "file1.dat"' checkout.log
  ! grep 'rejecting "file1.dat"' checkout.log

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
  git push origin main
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

begin_test "checkout: write-only file"
(
  set -e

  reponame="checkout-locked"
  filename="a.txt"

  setup_remote_repo_with_file "$reponame" "$filename"

  pushd "$TRASHDIR" > /dev/null
    GIT_LFS_SKIP_SMUDGE=1 clone_repo "$reponame" "${reponame}_checkout"

    chmod -w "$filename"

    refute_file_writeable "$filename"
    assert_pointer "refs/heads/main" "$filename" "$(calc_oid "$filename\n")" 6

    git lfs fetch
    git lfs checkout "$filename"

    refute_file_writeable "$filename"
    [ "$filename" = "$(cat "$filename")" ]
  popd > /dev/null
)
end_test

begin_test "checkout: conflicts"
(
  set -e

  reponame="checkout-conflicts"
  filename="file1.dat"

  setup_remote_repo_with_file "$reponame" "$filename"

  pushd "$TRASHDIR" > /dev/null
    clone_repo "$reponame" "${reponame}_checkout"

    git tag base
    git checkout -b first
    echo "abc123" > file1.dat
    git add -u
    git commit -m "first"

    git lfs checkout --to base.txt --base file1.dat 2>&1 | tee output.txt
    grep 'Could not checkout.*not in the middle of a merge' output.txt

    git checkout -b second main
    echo "def456" > file1.dat
    git add -u
    git commit -m "second"

    # This will cause a conflict.
    ! git merge first

    git lfs checkout --to base.txt --base file1.dat
    git lfs checkout --to ours.txt --ours file1.dat
    git lfs checkout --to theirs.txt --theirs file1.dat

    echo "file1.dat" | cmp - base.txt
    echo "abc123" | cmp - theirs.txt
    echo "def456" | cmp - ours.txt
  popd > /dev/null
)
end_test
