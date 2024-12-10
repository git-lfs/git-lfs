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
  grep 'rejecting "file1.dat"' checkout.log && exit 1

  git rm file1.dat

  echo "checkout should skip replacing files deleted in index"
  git lfs checkout
  [ ! -f file1.dat ]

  git reset --hard

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
  grep "Not in a Git repository" checkout.log
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
    echo "first" > other.txt
    git add other.txt
    git commit -m "first"

    git lfs checkout --to base.txt 2>&1 | tee output.txt
    grep -- '--to and exactly one of --theirs, --ours, and --base must be used together' output.txt

    git lfs checkout --base 2>&1 | tee output.txt
    grep -- '--to and exactly one of --theirs, --ours, and --base must be used together' output.txt

    git lfs checkout --to base.txt --ours --theirs 2>&1 | tee output.txt
    grep -- 'at most one of --base, --theirs, and --ours is allowed' output.txt

    git lfs checkout --to base.txt --base 2>&1 | tee output.txt
    grep -- '--to requires exactly one Git LFS object file path' output.txt

    git lfs checkout --to base.txt --base 2>&1 abc def | tee output.txt
    grep -- '--to requires exactly one Git LFS object file path' output.txt

    git lfs checkout --to base.txt --base file1.dat 2>&1 | tee output.txt
    grep 'Could not checkout.*not in the middle of a merge' output.txt

    git checkout -b second main
    echo "def456" > file1.dat
    git add -u
    echo "second" > other.txt
    git add other.txt
    git commit -m "second"

    # This will cause a conflict.
    git merge first && exit 1

    git lfs checkout --to base.txt --base file1.dat
    git lfs checkout --to ours.txt --ours file1.dat
    git lfs checkout --to theirs.txt --theirs file1.dat

    echo "file1.dat" | cmp - base.txt
    echo "abc123" | cmp - theirs.txt
    echo "def456" | cmp - ours.txt

    git lfs checkout --to base.txt --ours other.txt 2>&1 | tee output.txt
    grep 'Could not find decoder pointer for object' output.txt
  popd > /dev/null
)
end_test


begin_test "checkout: GIT_WORK_TREE"
(
  set -e

  reponame="checkout-work-tree"
  remotename="$(basename "$0" ".sh")"
  export GIT_WORK_TREE="$reponame" GIT_DIR="$reponame-git"
  mkdir "$GIT_WORK_TREE" "$GIT_DIR"
  git init
  git remote add origin "$GITSERVER/$remotename"

  git lfs uninstall --skip-repo

  git fetch origin
  git checkout -B main origin/main

  git lfs install
  git lfs fetch
  git lfs checkout

  contents="something something"
  [ "$contents" = "$(cat "$reponame/file1.dat")" ]
)
end_test

begin_test "checkout: sparse with partial clone and sparse index"
(
  set -e

  # Only test with Git version 2.42.0 as it introduced support for the
  # "objecttype" format option to the "git ls-files" command, which our
  # code requires.
  ensure_git_version_isnt "$VERSION_LOWER" "2.42.0"

  reponame="checkout-sparse"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  git lfs track "*.dat"

  contents1="a"
  contents1_oid=$(calc_oid "$contents1")
  contents2="b"
  contents2_oid=$(calc_oid "$contents2")
  contents3="c"
  contents3_oid=$(calc_oid "$contents3")

  mkdir in-dir out-dir
  printf "%s" "$contents1" >a.dat
  printf "%s" "$contents2" >in-dir/b.dat
  printf "%s" "$contents3" >out-dir/c.dat
  git add .
  git commit -m "add files"

  git push origin main

  assert_server_object "$reponame" "$contents1_oid"
  assert_server_object "$reponame" "$contents2_oid"
  assert_server_object "$reponame" "$contents3_oid"

  # Create a partial clone with a cone-mode sparse checkout of one directory
  # and a sparse index, which is important because otherwise the "git ls-files"
  # command ignores the --sparse option and lists all Git LFS files.
  cd ..
  git clone --filter=tree:0 --depth=1 --no-checkout \
    "$GITSERVER/$reponame" "${reponame}-partial"

  cd "${reponame}-partial"
  git sparse-checkout init --cone --sparse-index
  git sparse-checkout set "in-dir"
  git checkout main

  [ -d "in-dir" ]
  [ ! -e "out-dir" ]

  assert_local_object "$contents1_oid" 1
  assert_local_object "$contents2_oid" 1
  refute_local_object "$contents3_oid"

  # Git LFS objects associated with files outside of the sparse cone
  # should be ignored entirely, rather than just skipped.
  git lfs checkout 2>&1 | tee checkout.log
  if [ "0" -ne "${PIPESTATUS[0]}" ]; then
    echo >&2 "fatal: expected checkout to succeed ..."
    exit 1
  fi
  grep -q 'Skipped checkout for "out-dir/c.dat"' checkout.log && exit 1

  # Fetch all Git LFS objects, including those outside the sparse cone.
  git lfs fetch origin main

  assert_local_object "$contents3_oid" 1

  # Git LFS objects associated with files outside of the sparse cone
  # should not be checked out.
  git lfs checkout 2>&1 | tee checkout.log
  if [ "0" -ne "${PIPESTATUS[0]}" ]; then
    echo >&2 "fatal: expected checkout to succeed ..."
    exit 1
  fi
  grep -q 'Checking out LFS objects: 100% (3/3), 3 B' checkout.log && exit 1

  [ ! -e "out-dir/c.dat" ]
)
end_test
