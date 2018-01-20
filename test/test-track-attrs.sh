#!/usr/bin/env bash

. "test/testlib.sh"

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

