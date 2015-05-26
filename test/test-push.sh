#!/bin/sh

. "test/testlib.sh"

begin_test "push"
(
  set -e

  reponame="$(basename "$0" ".sh")"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" repo

  git lfs track "*.dat"
  echo "push a" > a.dat
  git add .gitattributes a.dat
  git commit -m "add a.dat"

  git lfs push origin master 2>&1 |
    tee push.log |
    grep "(1 of 1 files) 7 B / 7 B  100.00 %" || {
      cat push.log
      exit 1
    }

  git checkout -b push-b
  echo "push b" > b.dat
  git add b.dat
  git commit -m "add b.dat"

  git lfs push origin push-b 2>&1 |
    tee push.log |
    grep "(2 of 2 files) 14 B / 14 B  100.00 %" || {
      cat push.log
      exit 1
    }
)
end_test

begin_test "push dry-run"
(
  set -e

  reponame="$(basename "$0" ".sh")-dry-run"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" repo-dry-run

  git lfs track "*.dat"
  echo "push a" > a.dat
  git add .gitattributes a.dat
  git commit -m "add a.dat"

  [ "push a.dat" == "$(git lfs push --dry-run origin master 2>&1)" ]

  git checkout -b push-b
  echo "push b" > b.dat
  git add b.dat
  git commit -m "add b.dat"

  git lfs push --dry-run origin push-b 2>&1 | tee push.log
  grep "push a.dat" push.log
  grep "push b.dat" push.log
)
end_test
