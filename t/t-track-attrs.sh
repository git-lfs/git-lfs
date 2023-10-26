#!/usr/bin/env bash

. "$(dirname "$0")/testlib.sh"

ensure_git_version_isnt $VERSION_LOWER "2.1.0"

begin_test "track (--no-modify-attrs)"
(
  set -e

  reponame="track-no-modify-attrs"
  git init "$reponame"
  cd "$reponame"

  echo "contents" > a.dat
  git add a.dat

  # Git assumes that identical results from `stat(1)` between the index and
  # working copy are stat dirty. To prevent this, wait at least one second to
  # yield different `stat(1)` results.
  sleep 1

  git commit -m "add a.dat"

  echo "*.dat filter=lfs diff=lfs merge=lfs -text" > .gitattributes

  git add .gitattributes
  git commit -m "asdf"

  [ -z "$(git status --porcelain)" ]

  git lfs track --no-modify-attrs "*.dat"

  [ " M a.dat" = "$(git status --porcelain)" ]
)
end_test

begin_test "track (--dry-run)"
(
  set -e

  reponame="track-dry-run"
  git init "$reponame"
  cd "$reponame"

  git lfs track --dry-run "*.dat"

  echo "contents" > a.dat
  git add a.dat

  git commit -m "add a.dat"
  refute_pointer "main" "a.dat"
)
end_test
