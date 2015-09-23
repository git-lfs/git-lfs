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

begin_test "smudge include/exclude"
(
  set -e

  reponame="$(basename "$0" ".sh")-includeexclude"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" includeexclude

  git lfs track "*.dat"
  echo "smudge a" > a.dat
  git add .gitattributes a.dat
  git commit -m "add a.dat"

  pointer="$(pointer fcf5015df7a9089a7aa7fe74139d4b8f7d62e52d5a34f9a87aeffc8e8c668254 9)"

  # smudge works even though it hasn't been pushed, by reading from .git/lfs/objects
  [ "smudge a" = "$(echo "$pointer" | git lfs smudge)" ]

  git push origin master

  # this WOULD download except we're going to prevent it with include/exclude
  rm -rf .git/lfs/objects
  git config "lfs.fetchexclude" "a*"

  [ "$pointer" = "$(echo "$pointer" | git lfs smudge a.dat)" ]
)
end_test

begin_test "smudge with passthrough"
(
  set -e

  reponame="$(basename "$0" ".sh")-passthrough"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "passthrough"

  git lfs track "*.dat"
  echo "smudge a" > a.dat
  git add .gitattributes a.dat
  git commit -m "add a.dat"

  pointer="$(pointer fcf5015df7a9089a7aa7fe74139d4b8f7d62e52d5a34f9a87aeffc8e8c668254 9)"
  [ "smudge a" = "$(echo "$pointer" | git lfs smudge)" ]
  [ "$pointer" = "$(echo "$pointer" | GIT_LFS_SKIP_SMUDGE=1 git lfs smudge)" ]

  git push origin master

  echo "test clone with env"
  export GIT_LFS_SKIP_SMUDGE=1
  env | grep LFS
  clone_repo "$reponame" "passthrough-clone-env"
  [ "$pointer" = "$(cat a.dat)" ]
  [ "0" = "$(grep -c "Downloading a.dat" clone.log)" ]

  git lfs pull
  [ "smudge a" = "$(cat a.dat)" ]

  echo "test clone without env"
  unset GIT_LFS_SKIP_SMUDGE
  env | grep LFS
  clone_repo "$reponame" "no-passhthrough"
  [ "smudge a" = "$(cat a.dat)" ]
  [ "1" = "$(grep -c "Downloading a.dat" clone.log)" ]

  echo "test clone with init --smudge-passthrough"
  git lfs init --skip-smudge
  clone_repo "$reponame" "passthrough-clone-init"
  [ "$pointer" = "$(cat a.dat)" ]
  [ "0" = "$(grep -c "Downloading a.dat" clone.log)" ]

  git lfs init --force
)
end_test
