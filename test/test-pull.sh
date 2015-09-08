#!/usr/bin/env bash

. "test/testlib.sh"

begin_test "pull"
(
  set -e

  reponame="$(basename "$0" ".sh")"
  setup_remote_repo "$reponame"

  clone_repo "$reponame" clone

  clone_repo "$reponame" repo

  git lfs track "*.dat" 2>&1 | tee track.log
  grep "Tracking \*.dat" track.log

  contents="a"
  contents_oid=$(printf "$contents" | shasum -a 256 | cut -f 1 -d " ")

  printf "$contents" > a.dat
  git add a.dat
  git add .gitattributes
  git commit -m "add a.dat" 2>&1 | tee commit.log
  grep "master (root-commit)" commit.log
  grep "2 files changed" commit.log
  grep "create mode 100644 a.dat" commit.log
  grep "create mode 100644 .gitattributes" commit.log

  [ "a" = "$(cat a.dat)" ]

  assert_pointer "master" "a.dat" "$contents_oid" 1

  refute_server_object "$reponame" "$contents_oid"

  git push origin master 2>&1 | tee push.log
  grep "(1 of 1 files)" push.log
  grep "master -> master" push.log

  assert_server_object "$reponame" "$contents_oid"

  # change to the clone's working directory
  cd ../clone

  git pull 2>&1 | grep "Downloading a.dat (1 B)"

  [ "a" = "$(cat a.dat)" ]

  assert_local_object "$contents_oid" 1


  # Remove the working directory and lfs files
  rm a.dat
  rm -rf .git/lfs/objects
  git lfs pull 2>&1 | grep "(1 of 1 files)"
  [ "a" = "$(cat a.dat)" ]
  assert_local_object "$contents_oid" 1

  # Try with remote arg
  rm a.dat
  rm -rf .git/lfs/objects
  git lfs pull origin 2>&1 | grep "(1 of 1 files)"
  [ "a" = "$(cat a.dat)" ]
  assert_local_object "$contents_oid" 1

  # Remove just the working directory
  rm a.dat
  git lfs pull
  [ "a" = "$(cat a.dat)" ]


  # Test include / exclude filters supplied in gitconfig
  rm -rf .git/lfs/objects
  git config "lfs.fetchinclude" "a*"
  git lfs pull
  assert_local_object "$contents_oid" 1

  rm -rf .git/lfs/objects
  git config --unset "lfs.fetchinclude"
  git config "lfs.fetchexclude" "a*"
  git lfs pull
  refute_local_object "$contents_oid"

  # Test include / exclude filters supplied on the command line
  git config --unset "lfs.fetchexclude"
  rm -rf .git/lfs/objects
  git lfs pull --include="a*"
  assert_local_object "$contents_oid" 1

  rm -rf .git/lfs/objects
  git lfs pull --exclude="a*"
  refute_local_object "$contents_oid"


)
end_test

begin_test "pull: outside git repository"
(
  set +e
  git lfs pull 2>&1 > pull.log
  res=$?

  set -e
  [ "$res" = "128" ]
  grep "Not in a git repository" pull.log
)
end_test
