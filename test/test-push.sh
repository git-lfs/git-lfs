#!/usr/bin/env bash

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

  git lfs push origin master 2>&1 | tee push.log
  grep "(1 of 1 files)" push.log

  git checkout -b push-b
  echo "push b" > b.dat
  git add b.dat
  git commit -m "add b.dat"

  git lfs push origin push-b 2>&1 | tee push.log
  grep "(2 of 2 files)" push.log
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

  [ "push a.dat" = "$(git lfs push --dry-run origin master 2>&1)" ]

  git checkout -b push-b
  echo "push b" > b.dat
  git add b.dat
  git commit -m "add b.dat"

  git lfs push --dry-run origin push-b 2>&1 | tee push.log
  grep "push a.dat" push.log
  grep "push b.dat" push.log
  [ $(wc -l < push.log) -eq 2 ]
)
end_test

begin_test "push object id(s)"
(
  set -e

  reponame="$(basename "$0" ".sh")"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" repo2

  git lfs track "*.dat"
  echo "push a" > a.dat
  git add .gitattributes a.dat
  git commit -m "add a.dat"

  git lfs push --object-id origin \
    4c48d2a6991c9895bcddcf027e1e4907280bcf21975492b1afbade396d6a3340 \
    2>&1 | tee push.log
  grep "(1 of 1 files)" push.log

  echo "push b" > b.dat
  git add b.dat
  git commit -m "add b.dat"

  git lfs push --object-id origin \
    4c48d2a6991c9895bcddcf027e1e4907280bcf21975492b1afbade396d6a3340 \
    82be50ad35070a4ef3467a0a650c52d5b637035e7ad02c36652e59d01ba282b7 \
    2>&1 | tee push.log
  grep "(2 of 2 files)" push.log
)
end_test
