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
  echo "$alternate" > .git/objects/info/alternates

  GIT_TRACE=1 git lfs fetch origin master 2>&1 | tee fetch.log
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
  echo "$alternate" > .git/objects/info/alternates
  echo "$alternate_stale" >> .git/objects/info/alternates

  GIT_TRACE=1 git lfs fetch origin master 2>&1 | tee fetch.log
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

  GIT_TRACE=1 git lfs fetch origin master 2>&1 | tee fetch.log
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

  alternate="$TRASHDIR/${reponame}_alternate/.git/objects"
  echo "\"$alternate\"" > .git/objects/info/alternates

  GIT_TRACE=1 git lfs fetch origin master 2>&1 | tee fetch.log
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

  alternate="$TRASHDIR/${reponame}_alternate/.git/objects"

  GIT_ALTERNATE_OBJECT_DIRECTORIES="$alternate" \
  GIT_TRACE=1 \
    git lfs fetch origin master 2>&1 | tee fetch.log
  [ "0" -eq "$(grep -c "sending batch of size 1" fetch.log)" ]
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

  alternate_stale="$TRASHDIR/${reponame}_alternate_stale/.git/objects"
  alternate="$TRASHDIR/${reponame}_alternate/.git/objects"
  sep="$(native_path_list_separator)"

  GIT_ALTERNATE_OBJECT_DIRECTORIES="$alternate_stale$sep$alternate" \
  GIT_TRACE=1 \
    git lfs fetch origin master 2>&1 | tee fetch.log
  [ "0" -eq "$(grep -c "sending batch of size 1" fetch.log)" ]
)
end_test
