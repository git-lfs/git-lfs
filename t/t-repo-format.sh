#!/usr/bin/env bash

. "$(dirname "$0")/testlib.sh"

begin_test "repository format version"
(
  set -e

  reponame="lfs-repo-version"
  git init $reponame
  cd $reponame

  [ -z "$(git config --local lfs.repositoryFormatVersion)" ]

  git lfs track '*.dat'

  [ "$(git config --local lfs.repositoryFormatVersion)" = "0" ]

  git config --local lfs.repositoryFormatVersion 1

  git lfs track '*.bin' >output 2>&1 && exit 1
  cat output
  grep "Unknown repository format version: 1" output

  git config --local --unset lfs.repositoryFormatVersion
  # Verify that global settings are ignored.
  git config --global lfs.repositoryFormatVersion 1

  git lfs track '*.bin'
)
end_test
