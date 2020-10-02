#!/usr/bin/env bash

. "$(dirname "$0")/testlib.sh"

begin_test "untrack"
(
  set -e

  # no need to setup a remote repo, since this test doesn't need to push or pull

  reponame="untrack"
  git init $reponame
  cd $reponame

  # track *.jpg once
  git lfs track "*.jpg" | grep "Tracking \"\*.jpg\""
  echo "* annex.backend=SHA512E" >> .gitattributes

  git lfs untrack "*.jpg"

  expected="* annex.backend=SHA512E"
  [ "$expected" = "$(cat .gitattributes)" ]
)
end_test

begin_test "untrack outside git repo"
(
  set -e

  reponame="outside"
  mkdir $reponame
  cd $reponame

  git lfs untrack "*.foo" || {
    # this fails if it's run outside of a git repo using GIT_LFS_TEST_DIR

    # git itself returns an exit status of 128
    # $ git show
    # fatal: Not a git repository (or any of the parent directories): .git
    # $ echo "$?"
    # 128

    [ "$?" = "128" ]
    exit 0
  }

  if [ -n "$GIT_LFS_TEST_DIR" ]; then
    echo "GIT_LFS_TEST_DIR should be set outside of any Git repository"
    exit 1
  fi
)
end_test

begin_test "untrack removes escape sequences"
(
  set -e

  reponame="untrack-remove-escape-sequence"
  git init "$reponame"
  cd "$reponame"

  git lfs track " " | grep "Tracking \" \""
  assert_attributes_count "[[:space:]]" "filter=lfs" 1

  git lfs untrack " " | grep "Untracking \" \""
  assert_attributes_count "[[:space:]]" "filter=lfs" 0

  git lfs track "#" | grep "Tracking \"#\""
  assert_attributes_count "\\#" "filter=lfs" 1

  git lfs untrack "#" | grep "Untracking \"#\""
  assert_attributes_count "\\#" "filter=lfs" 0
)
end_test

begin_test "untrack removes prefixed patterns (legacy)"
(
  set -e

  reponame="untrack-removes-prefix-patterns-legacy"
  git init "$reponame"
  cd "$reponame"

  echo "./a.dat filter=lfs diff=lfs merge=lfs" > .gitattributes
  printf "a" > a.dat
  git add .gitattributes a.dat
  git commit -m "initial commit"

  git lfs untrack "./a.dat"

  if [ ! -z "$(cat .gitattributes)" ]; then
    echo >&2 "fatal: expected 'git lfs untrack' to clear .gitattributes"
    exit 1
  fi

  git checkout -- .gitattributes

  git lfs untrack "a.dat"

  if [ ! -z "$(cat .gitattributes)" ]; then
    echo >&2 "fatal: expected 'git lfs untrack' to clear .gitattributes"
    exit 1
  fi
)
end_test

begin_test "untrack removes prefixed patterns (modern)"
(
  set -e

  reponame="untrack-removes-prefix-patterns-modern"
  git init "$reponame"
  cd "$reponame"

  echo "a.dat filter=lfs diff=lfs merge=lfs" > .gitattributes
  printf "a" > a.dat
  git add .gitattributes a.dat
  git commit -m "initial commit"

  git lfs untrack "./a.dat"

  if [ ! -z "$(cat .gitattributes)" ]; then
    echo >&2 "fatal: expected 'git lfs untrack' to clear .gitattributes"
    exit 1
  fi

  git checkout -- .gitattributes

  git lfs untrack "a.dat"

  if [ ! -z "$(cat .gitattributes)" ]; then
    echo >&2 "fatal: expected 'git lfs untrack' to clear .gitattributes"
    exit 1
  fi
)
end_test

begin_test "untrack removes escaped pattern in .gitattributes"
(
  set -e

  reponame="untrack-escaped"
  git init "$reponame"
  cd "$reponame"

  filename="file with spaces.#"

  # emulate multiple instances of the same file in gitattributes
  echo 'file[[:space:]]with[[:space:]]spaces.\# filter=lfs diff=lfs merge=lfs -text' >> .gitattributes
  echo 'file[[:space:]]with[[:space:]]spaces.\# filter=lfs diff=lfs merge=lfs -text' >> .gitattributes
  echo 'file[[:space:]]with[[:space:]]spaces.\# filter=lfs diff=lfs merge=lfs -text' >> .gitattributes

  git lfs untrack "$filename"

  if [ ! -z "$(cat .gitattributes)" ]; then
    echo >&2 "fatal: expected 'git lfs untrack' to clear .gitattributes even if the file name was escaped"
    exit 1
  fi
)
end_test

begin_test "untrack works with GIT_WORK_TREE"
(
  set -e

  reponame="untrack-work-tree"
  export GIT_WORK_TREE="$reponame" GIT_DIR="$reponame-git"
  mkdir "$GIT_WORK_TREE" "$GIT_DIR"
  git init
  git lfs track '*.bin'

  grep -F '*.bin filter=lfs diff=lfs merge=lfs -text' "$reponame/.gitattributes"

  git lfs untrack '*.bin'

  [ ! -s "$reponame/.gitattributes" ]
)
end_test
