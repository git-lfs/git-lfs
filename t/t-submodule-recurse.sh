#!/usr/bin/env bash

. "$(dirname "$0")/testlib.sh"
reponame="submodule-recurse-test-repo"
submodname="submodule-recurse-test-submodule"

begin_test "submodule with submodule.recurse = true"
(
  set -e

  setup_remote_repo "$reponame"
  setup_remote_repo "$submodname"

  clone_repo "$submodname" submodule

  git lfs track "*.dat" 2>&1 | tee track.log
  grep "Tracking \"\*.dat\"" track.log

  echo "foo" > file.dat
  git add .gitattributes file.dat
  git commit -a -m "add file"
  git push origin main
  subcommit1=$(git rev-parse HEAD)

  echo "bar" > file.dat
  git add file.dat
  git commit -a -m "update file"
  git push origin main
  subcommit2=$(git rev-parse HEAD)

  clone_repo "$reponame" repo
  git submodule add "$GITSERVER/$submodname" submodule
  git submodule update --init --recursive
  git -C submodule reset --hard "$subcommit1"
  git add .gitmodules submodule
  git commit -m "add submodule"
  git push origin main

  git checkout -b feature
  git -C submodule reset --hard "$subcommit2"
  git add .gitmodules submodule
  git commit -m "update submodule"
  git push origin feature

  clone_repo "$reponame" repo-no-recurse
  git submodule update --init --recursive
  git checkout feature

  if [[ -d "submodule/lfs/logs" ]]
  then
    exit 1
  fi

  clone_repo "$reponame" repo-recurse
  git config submodule.recurse true
  git submodule update --init --recursive
  git checkout feature

  if [[ -d "submodule/lfs/logs" ]]
  then
    exit 1
  fi
)
end_test
