#!/usr/bin/env bash

. "test/testlib.sh"

begin_test "smudge"
(
  set -e

  reponame="$(basename "$0" ".sh")"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" repo

  git lfs track "*.dat"
  echo "smudge a" > a.dat
  git add .gitattributes a.dat
  git commit -m "add a.dat"

  # smudge works even though it hasn't been pushed, by reading from .git/lfs/objects
  output="$(pointer fcf5015df7a9089a7aa7fe74139d4b8f7d62e52d5a34f9a87aeffc8e8c668254 9 | git lfs smudge)"
  [ "smudge a" = "$output" ]

  git push origin master

  # download it from the git lfs server
  rm -rf .git/lfs/objects
  output="$(pointer fcf5015df7a9089a7aa7fe74139d4b8f7d62e52d5a34f9a87aeffc8e8c668254 9 | git lfs smudge)"
  [ "smudge a" = "$output" ]
)
end_test

begin_test "smudge --info"
(
  set -e

  cd repo
  output="$(pointer aaaaa15df7a9089a7aa7fe74139d4b8f7d62e52d5a34f9a87aeffc8e8c668254 123 | git lfs smudge --info)"
  [ "123 --" = "$output" ]
)
end_test

begin_test "smudge with invalid pointer"
(
  set -e

  cd repo
  [ "wat" = "$(echo "wat" | git lfs smudge)" ]
)
end_test
