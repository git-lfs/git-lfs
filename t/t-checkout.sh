#!/usr/bin/env bash

. "$(dirname "$0")/testlib.sh"

begin_test "checkout"
(
  set -e

  reponame="$(basename "$0" ".sh")"
  setup_remote_repo "$reponame"

  clone_repo "$reponame" repo
  rm -f clone.log

  git lfs track "*.dat" 2>&1 | tee track.log
  grep "Tracking \"\*.dat\"" track.log
  rm -f track.log

  contents="something something"
  contentsize=19
  contents_oid=$(calc_oid "$contents")

  # Same content everywhere is ok, just one object in lfs db
  printf "%s" "$contents" > file1.dat
  printf "%s" "$contents" > file2.dat
  printf "%s" "$contents" > file3.dat
  mkdir -p folder1 folder2/folder3/folder4
  printf "%s" "$contents" > folder1/nested.dat
  printf "%s" "$contents" > folder2/folder3/folder4/nested.dat
  git add file1.dat file2.dat file3.dat folder1 folder2
  git add .gitattributes
  git commit -m "add files"

  [ "$contents" = "$(cat file1.dat)" ]
  [ "$contents" = "$(cat file2.dat)" ]
  [ "$contents" = "$(cat file3.dat)" ]
  [ "$contents" = "$(cat folder1/nested.dat)" ]
  [ "$contents" = "$(cat folder2/folder3/folder4/nested.dat)" ]

  assert_pointer "main" "file1.dat" "$contents_oid" $contentsize

  # Remove the working directory
  rm -rf file1.dat file2.dat file3.dat folder1/nested.dat folder2

  echo "checkout should replace all"
  GIT_TRACE=1 git lfs checkout 2>&1 | tee checkout.log
  [ "$contents" = "$(cat file1.dat)" ]
  [ "$contents" = "$(cat file2.dat)" ]
  [ "$contents" = "$(cat file3.dat)" ]
  [ "$contents" = "$(cat folder1/nested.dat)" ]
  [ "$contents" = "$(cat folder2/folder3/folder4/nested.dat)" ]
  grep "Checking out LFS objects: 100% (5/5), 95 B" checkout.log
  grep 'accepting "file1.dat"' checkout.log
  grep 'rejecting "file1.dat"' checkout.log && exit 1
  rm -f checkout.log
  assert_clean_status

  git rm file1.dat

  echo "checkout should skip replacing files deleted in index"
  git lfs checkout
  [ ! -f file1.dat ]
  assert_clean_worktree_with_exceptions "file1\.dat"

  git reset --hard

  # Remove the working directory
  rm -rf file1.dat file2.dat file3.dat folder1/nested.dat folder2

  echo "checkout with filters"
  git lfs checkout file2.dat
  [ "$contents" = "$(cat file2.dat)" ]
  [ ! -f file1.dat ]
  [ ! -f file3.dat ]
  [ ! -f folder1/nested.dat ]
  [ ! -e folder2 ]
  assert_clean_worktree_with_exceptions "(file[13]|nested)\.dat"

  echo "quotes to avoid shell globbing"
  git lfs checkout "file*.dat"
  [ "$contents" = "$(cat file1.dat)" ]
  [ "$contents" = "$(cat file3.dat)" ]
  [ ! -f folder1/nested.dat ]
  [ ! -e folder2 ]
  assert_clean_worktree_with_exceptions "nested\.dat"

  echo "test subdir context"
  rm file1.dat
  pushd folder1
    git lfs checkout nested.dat
    [ "$contents" = "$(cat nested.dat)" ]
    [ ! -f ../file1.dat ]
    [ ! -e ../folder2 ]
    assert_clean_worktree_with_exceptions "(file1|folder4/nested)\.dat"

    # test '.' in current dir
    rm nested.dat
    git lfs checkout .
    [ "$contents" = "$(cat nested.dat)" ]
    [ ! -f ../file1.dat ]
    [ ! -e ../folder2 ]
    assert_clean_worktree_with_exceptions "(file1|folder4/nested)\.dat"

    # test '..' in current dir
    git lfs checkout ..
    [ "$contents" = "$(cat ../file1.dat)" ]
    [ "$contents" = "$(cat ../folder2/folder3/folder4/nested.dat)" ]
    assert_clean_status

    # test glob match with '..' in current dir
    rm -rf ../folder2
    git lfs checkout '../folder2/**'
    [ "$contents" = "$(cat ../folder2/folder3/folder4/nested.dat)" ]
    assert_clean_status
  popd

  echo "test folder param"
  rm -rf folder2
  git lfs checkout folder2
  [ "$contents" = "$(cat folder2/folder3/folder4/nested.dat)" ]
  assert_clean_status

  echo "test folder param with pre-existing directory"
  rm -rf folder2
  mkdir folder2
  git lfs checkout folder2
  [ "$contents" = "$(cat folder2/folder3/folder4/nested.dat)" ]
  assert_clean_status

  echo "test folder param with glob match"
  rm -rf folder2
  git lfs checkout 'folder2/**'
  [ "$contents" = "$(cat folder2/folder3/folder4/nested.dat)" ]
  assert_clean_status

  echo "test '.' in current dir"
  rm -rf file1.dat file2.dat file3.dat folder1 folder2
  git lfs checkout .
  [ "$contents" = "$(cat file1.dat)" ]
  [ "$contents" = "$(cat file2.dat)" ]
  [ "$contents" = "$(cat file3.dat)" ]
  [ "$contents" = "$(cat folder1/nested.dat)" ]
  [ "$contents" = "$(cat folder2/folder3/folder4/nested.dat)" ]
  assert_clean_status

  echo "test pre-existing directories"
  rm -rf folder1/nested.dat folder2/folder3/folder4
  git lfs checkout
  [ "$contents" = "$(cat folder1/nested.dat)" ]
  [ "$contents" = "$(cat folder2/folder3/folder4/nested.dat)" ]
  assert_clean_status

  echo "test checkout with missing data doesn't fail"
  git push origin main
  rm -rf .git/lfs/objects
  rm file*.dat
  git lfs checkout
  [ "$(pointer $contents_oid $contentsize)" = "$(cat file1.dat)" ]
  [ "$(pointer $contents_oid $contentsize)" = "$(cat file2.dat)" ]
  [ "$(pointer $contents_oid $contentsize)" = "$(cat file3.dat)" ]
  [ "$contents" = "$(cat folder1/nested.dat)" ]
  [ "$contents" = "$(cat folder2/folder3/folder4/nested.dat)" ]
  assert_clean_worktree_with_exceptions "file[123]\.dat"
)
end_test

begin_test "checkout: skip directory file conflicts"
(
  set -e

  reponame="checkout-skip-dir-file-conflicts"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  git lfs track "*.dat"

  contents="a"
  contents_oid="$(calc_oid "$contents")"
  mkdir -p dir1 dir2/dir3/dir4
  printf "%s" "$contents" >dir1/a.dat
  printf "%s" "$contents" >dir2/dir3/dir4/a.dat

  git add .gitattributes dir1 dir2
  git commit -m "initial commit"

  rm -rf dir1 dir2/dir3
  touch dir1 dir2/dir3

  git lfs checkout 2>&1 | tee checkout.log
  if [ "0" -ne "${PIPESTATUS[0]}" ]; then
    echo >&2 "fatal: expected checkout to succeed ..."
    exit 1
  fi
  if [ "$IS_WINDOWS" -eq 1 ]; then
    grep 'could not check out "dir1/a\.dat": could not create working directory file' checkout.log
    grep 'could not check out "dir2/dir3/dir4/a\.dat": could not create working directory file' checkout.log
  else
    grep 'Checkout error: stat dir1/a\.dat' checkout.log
    grep 'Checkout error: stat dir2/dir3/dir4/a\.dat' checkout.log
  fi

  [ -f "dir1" ]
  [ -f "dir2/dir3" ]
  assert_clean_index

  pushd dir2
    git lfs checkout 2>&1 | tee checkout.log
    if [ "0" -ne "${PIPESTATUS[0]}" ]; then
      echo >&2 "fatal: expected checkout to succeed ..."
      exit 1
    fi
    if [ "$IS_WINDOWS" -eq 1 ]; then
      grep 'could not check out "dir1/a\.dat": could not create working directory file' checkout.log
      grep 'could not check out "dir2/dir3/dir4/a\.dat": could not create working directory file' checkout.log
    else
      grep 'Checkout error: stat \.\./dir1/a\.dat' checkout.log
      grep 'Checkout error: stat dir3/dir4/a\.dat' checkout.log
    fi
  popd

  [ -f "dir1" ]
  [ -f "dir2/dir3" ]
  assert_clean_index
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
  if [ "0" -ne "${PIPESTATUS[0]}" ]; then
    echo >&2 "fatal: expected checkout to succeed ..."
    exit 1
  fi
  grep "Git LFS is not installed" checkout.txt

  contentsize=19
  contents_oid=$(calc_oid "something something")
  [ "$(pointer $contents_oid $contentsize)" = "$(cat file1.dat)" ]
  [ "$(pointer $contents_oid $contentsize)" = "$(cat file2.dat)" ]
  [ "$(pointer $contents_oid $contentsize)" = "$(cat file3.dat)" ]
  [ "$(pointer $contents_oid $contentsize)" = "$(cat folder1/nested.dat)" ]
  [ "$(pointer $contents_oid $contentsize)" = "$(cat folder2/folder3/folder4/nested.dat)" ]
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

begin_test "checkout: read-only directory"
(
  set -e

  skip_if_root_or_admin "$test_description"

  reponame="checkout-read-only"
  git init "$reponame"
  cd "$reponame"

  git lfs track "*.bin"

  contents="a"
  contents_oid=$(calc_oid "$contents")
  mkdir dir
  printf "%s" "$contents" > dir/a.bin

  git add .gitattributes dir/a.bin
  git commit -m "add dir/a.bin"

  rm dir/a.bin

  if [ "$IS_WINDOWS" -eq 1 ]; then
    icacls dir /inheritance:r
    icacls dir /grant:r Everyone:R
  else
    chmod a-w dir
  fi
  git lfs checkout 2>&1 | tee checkout.log
  # Note that although the checkout command should log an error, at present
  # we still expect a zero exit code.
  if [ "0" -ne "${PIPESTATUS[0]}" ]; then
    echo >&2 "fatal: expected 'git lfs checkout' to succeed ..."
    exit 1
  fi

  assert_local_object "$contents_oid" 1

  [ ! -e dir/a.bin ]

  grep 'could not check out "dir/a.bin"' checkout.log
  grep 'could not create working directory file' checkout.log
  grep 'permission denied' checkout.log
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

    abs_assert_dir="$(canonical_path "$TRASHDIR/${reponame}-assert")"
    abs_theirs_file="$abs_assert_dir/dir1/dir2/theirs.txt"

    rm -rf "$abs_assert_dir"

    git lfs checkout --to base.txt --base file1.dat
    git lfs checkout --to ../ours.txt --ours file1.dat
    git lfs checkout --to "$abs_theirs_file" --theirs file1.dat

    echo "file1.dat" | cmp - base.txt
    echo "def456" | cmp - ../ours.txt
    echo "abc123" | cmp - "$abs_theirs_file"

    rm -rf base.txt ../ours.txt "$abs_assert_dir"
    mkdir -p dir1/dir2

    pushd dir1/dir2
      git lfs checkout --to base.txt --base ../../file1.dat
      git lfs checkout --to ../../../ours.txt --ours ../../file1.dat
      git lfs checkout --to "$abs_theirs_file" --theirs ../../file1.dat
    popd

    echo "file1.dat" | cmp - dir1/dir2/base.txt
    echo "def456" | cmp - ../ours.txt
    echo "abc123" | cmp - "$abs_theirs_file"

    has_native_symlinks && {
      rm -rf "$abs_assert_dir"
      mkdir -p "$abs_assert_dir/link1"
      ln -s link1 "$abs_assert_dir/dir1"

      git lfs checkout --to "$abs_theirs_file" --theirs file1.dat

      [ -L "$abs_assert_dir/dir1" ]
      echo "abc123" | cmp - "$abs_assert_dir/link1/dir2/theirs.txt"
    }

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

begin_test "checkout: pointer extension"
(
  set -e

  reponame="checkout-pointer-extension"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  git lfs track "*.dat"

  setup_case_inverter_extension

  contents="abc"
  inverted_contents_oid="$(calc_oid "$(invert_case "$contents")")"
  mkdir dir1
  printf "%s" "$contents" >dir1/abc.dat

  git add .gitattributes dir1
  git commit -m "initial commit"

  assert_local_object "$inverted_contents_oid" 3

  rm -rf dir1 "$LFSTEST_EXT_LOG"
  git lfs checkout

  [ "$contents" = "$(cat "dir1/abc.dat")" ]
  grep "smudge: dir1/abc.dat" "$LFSTEST_EXT_LOG"

  rm -rf dir1 "$LFSTEST_EXT_LOG"
  mkdir dir2

  pushd dir2
    git lfs checkout
  popd

  [ "$contents" = "$(cat "dir1/abc.dat")" ]

  # Note that at present we expect "git lfs checkout" to run the extension
  # program in the current working directory rather than the repository root,
  # as would occur if it was run within a smudge filter operation started
  # by Git.
  grep "smudge: ../dir1/abc.dat" "$LFSTEST_EXT_LOG"
)
end_test

begin_test "checkout: pointer extension with conflict"
(
  set -e

  reponame="checkout-pointer-extension-conflict"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  git lfs track "*.dat"

  setup_case_inverter_extension

  contents="abc"
  inverted_contents_oid="$(calc_oid "$(invert_case "$contents")")"
  mkdir dir1
  printf "%s" "$contents" >dir1/abc.dat

  git add .gitattributes dir1
  git commit -m "initial commit"

  assert_local_object "$inverted_contents_oid" 3

  git checkout -b theirs
  contents_theirs="Abc"
  printf "%s" "$contents_theirs" >dir1/abc.dat
  git add dir1
  git commit -m "theirs"

  git checkout main
  contents_ours="aBc"
  printf "%s" "$contents_ours" >dir1/abc.dat
  git add dir1
  git commit -m "ours"

  git merge theirs && exit 1

  rm -f "$LFSTEST_EXT_LOG"

  git lfs checkout --to base.txt --base dir1/abc.dat

  printf "%s" "$contents" | cmp - base.txt

  # Note that at present we expect "git lfs checkout" to pass the argument
  # from its --to option to the extension program instead of the pointer's
  # file path.
  grep "smudge: base.txt" "$LFSTEST_EXT_LOG"

  rm -f "$LFSTEST_EXT_LOG"

  pushd dir1
    git lfs checkout --to ../ours.txt --ours abc.dat
  popd

  printf "%s" "$contents_ours" | cmp - ours.txt

  # Note that at present we expect "git lfs checkout" to pass the argument
  # from its --to option to the extension program instead of the pointer's
  # file path.
  grep "smudge: ../ours.txt" "$LFSTEST_EXT_LOG"

  abs_assert_dir="$TRASHDIR/${reponame}-assert"
  abs_theirs_file="$(canonical_path "$abs_assert_dir/dir1/dir2/theirs.txt")"

  rm -rf "$abs_assert_dir" "$LFSTEST_EXT_LOG"
  mkdir dir2

  pushd dir2
    git lfs checkout --to "$abs_theirs_file" --theirs ../dir1/abc.dat
  popd

  printf "%s" "$contents_theirs" | cmp - "$abs_theirs_file"

  # Note that at present we expect "git lfs checkout" to pass the argument
  # from its --to option to the extension program instead of the pointer's
  # file path.
  grep "smudge: $(escape_path "$abs_theirs_file")" "$LFSTEST_EXT_LOG"
)
end_test
