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
  grep "Tracking \"\*.dat\"" track.log

  contents="a"
  contents_oid=$(calc_oid "$contents")
  contents2="A"
  contents2_oid=$(calc_oid "$contents2")

  printf "$contents" > a.dat
  printf "$contents2" > á.dat
  git add a.dat á.dat .gitattributes
  git commit -m "add files" 2>&1 | tee commit.log
  grep "master (root-commit)" commit.log
  grep "3 files changed" commit.log
  grep "create mode 100644 a.dat" commit.log
  grep "create mode 100644 .gitattributes" commit.log

  ls -al
  [ "a" = "$(cat a.dat)" ]
  [ "A" = "$(cat "á.dat")" ]

  assert_pointer "master" "a.dat" "$contents_oid" 1
  assert_pointer "master" "á.dat" "$contents2_oid" 1

  refute_server_object "$reponame" "$contents_oid"
  refute_server_object "$reponame" "$contents2_oid"

  echo "initial push"
  git push origin master 2>&1 | tee push.log
  grep "(2 of 2 files)" push.log
  grep "master -> master" push.log

  assert_server_object "$reponame" "$contents_oid"
  assert_server_object "$reponame" "$contents2_oid"

  # change to the clone's working directory
  cd ../clone

  echo "normal pull"
  git pull 2>&1 | tee pull.log
  grep "Downloading a.dat (1 B)" pull.log
  grep "Downloading á.dat (1 B)" pull.log

  [ "a" = "$(cat a.dat)" ]
  [ "A" = "$(cat "á.dat")" ]

  assert_local_object "$contents_oid" 1
  assert_local_object "$contents2_oid" 1

  echo "lfs pull"
  rm a.dat á.dat
  rm -rf .git/lfs/objects
  git lfs pull 2>&1 | grep "(2 of 2 files)"
  ls -al
  [ "a" = "$(cat a.dat)" ]
  [ "A" = "$(cat "á.dat")" ]
  assert_local_object "$contents_oid" 1
  assert_local_object "$contents2_oid" 1

  echo "lfs pull with remote"
  rm a.dat á.dat
  rm -rf .git/lfs/objects
  git lfs pull origin 2>&1 | grep "(2 of 2 files)"
  [ "a" = "$(cat a.dat)" ]
  [ "A" = "$(cat "á.dat")" ]
  assert_local_object "$contents_oid" 1
  assert_local_object "$contents2_oid" 1

  echo "lfs pull with local storage"
  rm a.dat á.dat
  git lfs pull
  [ "a" = "$(cat a.dat)" ]
  [ "A" = "$(cat "á.dat")" ]

  echo "lfs pull with include/exclude filters in gitconfig"
  rm -rf .git/lfs/objects
  git config "lfs.fetchinclude" "a*"
  git lfs pull
  assert_local_object "$contents_oid" 1

  rm -rf .git/lfs/objects
  git config --unset "lfs.fetchinclude"
  git config "lfs.fetchexclude" "a*"
  git lfs pull
  refute_local_object "$contents_oid"

  echo "lfs pull with include/exclude filters in command line"
  git config --unset "lfs.fetchexclude"
  rm -rf .git/lfs/objects
  git lfs pull --include="a*"
  assert_local_object "$contents_oid" 1

  rm -rf .git/lfs/objects
  git lfs pull --exclude="a*"
  refute_local_object "$contents_oid"
)
end_test

begin_test "pull with raw remote url"
(
  set -e
  mkdir raw
  cd raw
  git init
  git lfs install --local --skip-smudge

  git remote add origin $GITSERVER/test-pull
  git pull origin master

  contents="a"
  contents_oid=$(calc_oid "$contents")

  # LFS object not downloaded, pointer in working directory
  refute_local_object "$contents_oid"
  grep "$contents_oid" a.dat

  git lfs pull "$GITSERVER/test-pull"
  echo "pulled!"

  # LFS object downloaded and in working directory
  assert_local_object "$contents_oid" 1
  [ "0" = "$(grep -c "$contents_oid" a.dat)" ]
  [ "a" = "$(cat a.dat)" ]
)
end_test

begin_test "pull: with missing object"
(
  set -e

  # this clone is setup in the first test in this file
  cd clone
  rm -rf .git/lfs/objects

  contents_oid=$(calc_oid "a")
  reponame="$(basename "$0" ".sh")"
  delete_server_object "$reponame" "$contents_oid"
  refute_server_object "$reponame" "$contents_oid"

  # should return non-zero, but should also download all the other valid files too
  git lfs pull 2>&1 | tee pull.log
  pull_exit="${PIPESTATUS[0]}"
  [ "$pull_exit" != "0" ]

  grep "$contents_oid" pull.log

  contents2_oid=$(calc_oid "A")
  assert_local_object "$contents2_oid" 1
  refute_local_object "$contents_oid"
)
end_test

begin_test "pull: outside git repository"
(
  set +e
  git lfs pull 2>&1 > pull.log
  res=$?

  set -e
  if [ "$res" = "0" ]; then
    echo "Passes because $GIT_LFS_TEST_DIR is unset."
    exit 0
  fi
  [ "$res" = "128" ]
  grep "Not in a git repository" pull.log
)
end_test
