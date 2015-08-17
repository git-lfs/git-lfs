#!/usr/bin/env bash

. "test/testlib.sh"

begin_test "fetch with branches"
(
  set -e

  reponame="$(basename "$0" ".sh")-with-branches"
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

  assert_local_object "$contents_oid" 1

  refute_server_object "$reponame" "$contents_oid"

  git push origin master 2>&1 | tee push.log
  grep "(1 of 1 files)" push.log
  grep "master -> master" push.log

  assert_server_object "$reponame" "$contents_oid"

  # Add a file in a different branch
  git checkout -b newbranch
  b="b"
  b_oid=$(printf "$b" | shasum -a 256 | cut -f 1 -d " ")
  printf "$b" > b.dat
  git add b.dat
  git commit -m "add b.dat"
  assert_local_object "$b_oid" 1

  git push origin newbranch
  assert_server_object "$reponame" "$b_oid"

  # change to the clone's working directory
  cd ../clone

  git pull 2>&1 | grep "Downloading a.dat (1 B)"

  [ "a" = "$(cat a.dat)" ]

  assert_local_object "$contents_oid" 1


  # Remove the working directory and lfs files
  rm -rf .git/lfs/objects

  git lfs fetch 2>&1 | grep "(1 of 1 files)"

  assert_local_object "$contents_oid" 1

  git checkout newbranch
  git checkout master
  rm -rf .git/lfs/objects

  git lfs fetch master newbranch
  assert_local_object "$contents_oid" 1
  assert_local_object "$b_oid" 1

)
end_test

begin_test "fetch with deleted files"
(
  set -e
  reponame="$(basename "$0" ".sh")-with-deleted-files"
  setup_remote_repo "$reponame"

  clone_repo "$reponame" repo-with-deleted-files

  git lfs track "*.dat" 2>&1 | tee track.log
  grep "Tracking \*.dat" track.log

  echo "abc is a big file" > abc.dat
  echo "def is a big file" > def.dat
  git add *.dat

  git commit -m "add files"
  git rm abc.dat
  echo "ghe is a big file" > ghe.dat
  git add ghe.dat
  git commit -m "remove and add one"

  git push origin master 2>&1 | tee push.log
  grep "(3 of 3 files)" push.log
  grep "master -> master" push.log

  git lfs fetch
)
end_test
