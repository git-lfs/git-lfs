#!/usr/bin/env bash

. "$(dirname "$0")/testlib.sh"

begin_test "alternates (single)"
(
  set -e

  reponame="alternates-single-alternate"
  setup_remote_repo_with_file "$reponame" "a.txt"

  pushd "$TRASHDIR" > /dev/null
    clone_repo "$reponame" "${reponame}_alternate"
  popd > /dev/null

  rm -rf .git/lfs/objects

  alternate="$TRASHDIR/${reponame}_alternate/.git/objects"
  echo "$(native_path "$alternate")" > .git/objects/info/alternates

  GIT_TRACE=1 git lfs fetch origin main 2>&1 | tee fetch.log
  [ "0" -eq "$(grep -c "sending batch of size 1" fetch.log)" ]
)
end_test

begin_test "alternates (multiple)"
(
  set -e

  reponame="alternates-multiple-alternates"
  setup_remote_repo_with_file "$reponame" "a.txt"

  pushd "$TRASHDIR" > /dev/null
    clone_repo "$reponame" "${reponame}_alternate_stale"
    rm -rf .git/lfs/objects
  popd > /dev/null
  pushd "$TRASHDIR" > /dev/null
    clone_repo "$reponame" "${reponame}_alternate"
  popd > /dev/null

  rm -rf .git/lfs/objects

  alternate_stale="$TRASHDIR/${reponame}_alternate_stale/.git/objects"
  alternate="$TRASHDIR/${reponame}_alternate/.git/objects"
  echo "$(native_path "$alternate")" > .git/objects/info/alternates
  echo "$(native_path "$alternate_stale")" >> .git/objects/info/alternates

  GIT_TRACE=1 git lfs fetch origin main 2>&1 | tee fetch.log
  [ "0" -eq "$(grep -c "sending batch of size 1" fetch.log)" ]
)
end_test

begin_test "alternates (commented)"
(
  set -e

  reponame="alternates-commented-alternate"
  setup_remote_repo_with_file "$reponame" "a.txt"

  pushd "$TRASHDIR" > /dev/null
    clone_repo "$reponame" "${reponame}_alternate"
  popd > /dev/null

  rm -rf .git/lfs/objects

  alternate="$TRASHDIR/${reponame}_alternate/.git/objects"
  echo "# $alternate" > .git/objects/info/alternates

  GIT_TRACE=1 git lfs fetch origin main 2>&1 | tee fetch.log
  [ "1" -eq "$(grep -c "sending batch of size 1" fetch.log)" ]
)
end_test

begin_test "alternates (quoted)"
(
  set -e

  reponame="alternates-quoted-alternate"
  setup_remote_repo_with_file "$reponame" "a.txt"

  pushd "$TRASHDIR" > /dev/null
    clone_repo "$reponame" "${reponame}_alternate"
  popd > /dev/null

  rm -rf .git/lfs/objects

  # Normally, a plain native_path call would be sufficient here, but when we
  # use a quoted alternate, Git interprets backslash escapes, and Windows path
  # names look like backslash escapes. As a consequence, we switch to forward
  # slashes to avoid misinterpretation.
  alternate=$(native_path "$TRASHDIR/${reponame}_alternate/.git/objects" | sed -e 's,\\,/,g')
  echo "\"$alternate\"" > .git/objects/info/alternates

  GIT_TRACE=1 git lfs fetch origin main 2>&1 | tee fetch.log
  [ "0" -eq "$(grep -c "sending batch of size 1" fetch.log)" ]
)
end_test

begin_test "alternates (OS environment, single)"
(
  set -e

  reponame="alternates-environment-single-alternate"
  setup_remote_repo_with_file "$reponame" "a.txt"

  pushd "$TRASHDIR" > /dev/null
    clone_repo "$reponame" "${reponame}_alternate"
  popd > /dev/null

  rm -rf .git/lfs/objects
  rm -rf .git/objects/*
  git init

  alternate="$(native_path "$TRASHDIR/${reponame}_alternate/.git/objects")"

  GIT_ALTERNATE_OBJECT_DIRECTORIES="$alternate" \
  GIT_TRACE=1 \
    git lfs fetch origin main 2>&1 | tee fetch.log
  [ "0" -eq "$(grep -c "sending batch of size 1" fetch.log)" ]
  GIT_ALTERNATE_OBJECT_DIRECTORIES="$alternate" \
    git lfs push "$(git config remote.origin.url)" main
)
end_test

begin_test "alternates (OS environment, multiple)"
(
  set -e

  reponame="alternates-environment-multiple-alternates"
  setup_remote_repo_with_file "$reponame" "a.txt"

  pushd "$TRASHDIR" > /dev/null
    clone_repo "$reponame" "${reponame}_alternate_stale"
    rm -rf .git/lfs/objects
  popd > /dev/null
  pushd "$TRASHDIR" > /dev/null
    clone_repo "$reponame" "${reponame}_alternate"
  popd > /dev/null

  rm -rf .git/lfs/objects
  rm -rf .git/objects/*
  git init

  alternate_stale="$(native_path "$TRASHDIR/${reponame}_alternate_stale/.git/objects")"
  alternate="$(native_path "$TRASHDIR/${reponame}_alternate/.git/objects")"
  sep="$(native_path_list_separator)"

  GIT_ALTERNATE_OBJECT_DIRECTORIES="$alternate_stale$sep$alternate" \
  GIT_TRACE=1 \
    git lfs fetch origin main 2>&1 | tee fetch.log
  [ "0" -eq "$(grep -c "sending batch of size 1" fetch.log)" ]
  GIT_ALTERNATE_OBJECT_DIRECTORIES="$alternate_stale$sep$alternate" \
    git lfs push "$(git config remote.origin.url)" main
)
end_test
