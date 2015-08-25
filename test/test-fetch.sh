#!/usr/bin/env bash

. "test/testlib.sh"

begin_test "fetch"
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

  # test with just remote specified
  rm -rf .git/lfs/objects
  git lfs fetch origin 2>&1 | grep "(1 of 1 files)"
  assert_local_object "$contents_oid" 1

  git checkout newbranch
  git checkout master
  rm -rf .git/lfs/objects

  git lfs fetch origin master newbranch
  assert_local_object "$contents_oid" 1
  assert_local_object "$b_oid" 1

  # Test include / exclude filters supplied in gitconfig
  rm -rf .git/lfs/objects
  git config "lfs.fetchinclude" "a*"
  git lfs fetch origin master newbranch
  assert_local_object "$contents_oid" 1
  refute_local_object "$b_oid"
  
  rm -rf .git/lfs/objects
  git config --unset "lfs.fetchinclude"
  git config "lfs.fetchexclude" "a*"
  git lfs fetch origin master newbranch
  refute_local_object "$contents_oid"
  assert_local_object "$b_oid" 1

  rm -rf .git/lfs/objects
  git config "lfs.fetchinclude" "a*,b*"
  git config "lfs.fetchexclude" "c*,d*"
  git lfs fetch origin master newbranch
  assert_local_object "$contents_oid" 1
  assert_local_object "$b_oid" 1
  
  rm -rf .git/lfs/objects
  git config "lfs.fetchinclude" "c*,d*"
  git config "lfs.fetchexclude" "a*,b*"
  git lfs fetch origin master newbranch
  refute_local_object "$contents_oid"
  refute_local_object "$b_oid"

  # Test include / exclude filters supplied on the command line
  git config --unset "lfs.fetchinclude"
  git config --unset "lfs.fetchexclude"
  rm -rf .git/lfs/objects
  git lfs fetch --include="a*" origin master newbranch
  assert_local_object "$contents_oid" 1
  refute_local_object "$b_oid"
  
  rm -rf .git/lfs/objects
  git lfs fetch --exclude="a*" origin master newbranch
  refute_local_object "$contents_oid"
  assert_local_object "$b_oid" 1

  rm -rf .git/lfs/objects
  git lfs fetch -I "a*,b*" -X "c*,d*" origin master newbranch
  assert_local_object "$contents_oid" 1
  assert_local_object "$b_oid" 1
  
  rm -rf .git/lfs/objects
  git lfs fetch --include="c*,d*" --exclude="a*,b*" origin master newbranch
  refute_local_object "$contents_oid"
  refute_local_object "$b_oid"
)
end_test

begin_test "fetch-recent"
(
  set -e

  reponame="fetch-recent"
  setup_remote_repo "$reponame"

  clone_repo "$reponame" "$reponame"

  # just testing for now
  echo '[{
    "CommitDate":"2015-08-14T17:04:36+01:00",
    "Files":[
      {"Filename":"file2.txt","Size":22}]
    }]' | lfstest-testutils addcommits


)
end_test
