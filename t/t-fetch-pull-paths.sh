#!/usr/bin/env bash

. "$(dirname "$0")/testlib.sh"

reponame="$(basename "$0" ".sh")"
top_contents="top"
top_oid=$(calc_oid "$top_contents")
space_contents="space"
space_oid=$(calc_oid "$space_contents")
other_contents="other"
other_oid=$(calc_oid "$other_contents")
literal_contents="literal"
literal_oid=$(calc_oid "$literal_contents")
wildcard_contents="wildcard"
wildcard_oid=$(calc_oid "$wildcard_contents")

begin_test "setup fetch and pull path tests"
(
  set -e

  setup_remote_repo "$reponame"
  clone_repo "$reponame" repo

  git lfs track "*.dat"
  mkdir dir
  printf "%s" "$top_contents" >top.dat
  printf "%s" "$space_contents" >"dir/file with spaces.dat"
  printf "%s" "$other_contents" >dir/other.dat
  printf "%s" "$literal_contents" >"literal[1].dat"
  printf "%s" "$wildcard_contents" >literal1.dat

  git add .
  git commit -m "add LFS files"
  git push origin main

  GIT_LFS_SKIP_SMUDGE=1 clone_repo "$reponame" clone
)
end_test

begin_test "fetch literal paths after a double dash"
(
  set -e
  cd clone
  rm -rf .git/lfs/objects

  git lfs fetch --dry-run origin main -- "literal[1].dat" 2>&1 | tee fetch.log
  grep "fetch $literal_oid => literal\[1\]\.dat" fetch.log
  grep "$wildcard_oid" fetch.log && exit 1 || true

  git lfs fetch origin main -- "literal[1].dat"
  assert_local_object "$literal_oid" 7
  refute_local_object "$wildcard_oid"
  refute_local_object "$top_oid"
)
end_test

begin_test "fetch paths relative to a subdirectory"
(
  set -e
  cd clone
  rm -rf .git/lfs/objects

  pushd dir
    git lfs fetch -- "../top.dat"
  popd

  assert_local_object "$top_oid" 3
  refute_local_object "$space_oid"
  refute_local_object "$other_oid"
)
end_test

begin_test "fetch an absolute literal path"
(
  set -e
  cd clone
  rm -rf .git/lfs/objects

  git lfs fetch -- "$(pwd)/dir/file with spaces.dat"

  assert_local_object "$space_oid" 5
  refute_local_object "$top_oid"
  refute_local_object "$other_oid"
)
end_test

begin_test "fetch all files below a literal directory path"
(
  set -e
  cd clone
  rm -rf .git/lfs/objects

  git lfs fetch -- dir

  assert_local_object "$space_oid" 5
  assert_local_object "$other_oid" 5
  refute_local_object "$top_oid"
  refute_local_object "$literal_oid"
  refute_local_object "$wildcard_oid"
)
end_test

begin_test "fetch paths also respect include filters"
(
  set -e
  cd clone
  rm -rf .git/lfs/objects

  git lfs fetch --include="top.dat" -- "dir/other.dat"

  refute_local_object "$top_oid"
  refute_local_object "$other_oid"
)
end_test

begin_test "fetch paths with refs read from standard input"
(
  set -e
  cd clone
  rm -rf .git/lfs/objects

  printf "main\n" | git lfs fetch origin --stdin -- top.dat

  assert_local_object "$top_oid" 3
  refute_local_object "$space_oid"
  refute_local_object "$other_oid"
)
end_test

begin_test "fetch --all rejects literal paths"
(
  set -e
  cd clone

  if git lfs fetch --all -- top.dat 2>fetch.log; then
    echo >&2 "fatal: expected fetch --all with a path to fail"
    exit 1
  fi
  grep "Cannot combine --all with paths" fetch.log
)
end_test

begin_test "pull a literal path and leave other pointers unchanged"
(
  set -e
  cd clone
  rm -rf .git/lfs/objects
  rm -rf dir top.dat "literal[1].dat" literal1.dat
  GIT_LFS_SKIP_SMUDGE=1 git checkout -f HEAD -- .

  git lfs pull origin -- "dir/file with spaces.dat"

  [ "$space_contents" = "$(cat "dir/file with spaces.dat")" ]
  grep "version https://git-lfs.github.com/spec/v1" top.dat
  grep "version https://git-lfs.github.com/spec/v1" dir/other.dat
  grep "version https://git-lfs.github.com/spec/v1" "literal[1].dat"
  assert_local_object "$space_oid" 5
  refute_local_object "$top_oid"
  refute_local_object "$other_oid"
)
end_test

begin_test "pull a path relative to a subdirectory"
(
  set -e
  cd clone
  rm -rf .git/lfs/objects
  rm -rf dir top.dat "literal[1].dat" literal1.dat
  GIT_LFS_SKIP_SMUDGE=1 git checkout -f HEAD -- .

  pushd dir
    git lfs pull -- "../literal[1].dat"
  popd

  [ "$literal_contents" = "$(cat "literal[1].dat")" ]
  grep "version https://git-lfs.github.com/spec/v1" literal1.dat
  assert_local_object "$literal_oid" 7
  refute_local_object "$wildcard_oid"
)
end_test

begin_test "pull all files below a literal directory path"
(
  set -e
  cd clone
  rm -rf .git/lfs/objects
  rm -rf dir top.dat "literal[1].dat" literal1.dat
  GIT_LFS_SKIP_SMUDGE=1 git checkout -f HEAD -- .

  git lfs pull -- dir

  [ "$space_contents" = "$(cat "dir/file with spaces.dat")" ]
  [ "$other_contents" = "$(cat dir/other.dat)" ]
  grep "version https://git-lfs.github.com/spec/v1" top.dat
  grep "version https://git-lfs.github.com/spec/v1" "literal[1].dat"
  assert_local_object "$space_oid" 5
  assert_local_object "$other_oid" 5
  refute_local_object "$top_oid"
  refute_local_object "$literal_oid"
)
end_test
