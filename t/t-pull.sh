#!/usr/bin/env bash

. "$(dirname "$0")/testlib.sh"

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
  contents3="dir"
  contents3_oid=$(calc_oid "$contents3")

  mkdir dir
  echo "*.log" > .gitignore
  printf "%s" "$contents" > a.dat
  printf "%s" "$contents2" > á.dat
  printf "%s" "$contents3" > dir/dir.dat
  git add .
  git commit -m "add files" 2>&1 | tee commit.log
  grep "master (root-commit)" commit.log
  grep "5 files changed" commit.log
  grep "create mode 100644 a.dat" commit.log
  grep "create mode 100644 .gitattributes" commit.log

  ls -al
  [ "a" = "$(cat a.dat)" ]
  [ "A" = "$(cat "á.dat")" ]
  [ "dir" = "$(cat "dir/dir.dat")" ]

  assert_pointer "master" "a.dat" "$contents_oid" 1
  assert_pointer "master" "á.dat" "$contents2_oid" 1
  assert_pointer "master" "dir/dir.dat" "$contents3_oid" 3

  refute_server_object "$reponame" "$contents_oid"
  refute_server_object "$reponame" "$contents2_oid"
  refute_server_object "$reponame" "$contents33oid"

  echo "initial push"
  git push origin master 2>&1 | tee push.log
  grep "Uploading LFS objects: 100% (3/3), 5 B" push.log
  grep "master -> master" push.log

  assert_server_object "$reponame" "$contents_oid"
  assert_server_object "$reponame" "$contents2_oid"
  assert_server_object "$reponame" "$contents3_oid"

  # change to the clone's working directory
  cd ../clone

  echo "normal pull"
  git pull 2>&1

  [ "a" = "$(cat a.dat)" ]
  [ "A" = "$(cat "á.dat")" ]

  assert_local_object "$contents_oid" 1
  assert_local_object "$contents2_oid" 1
  assert_clean_status

  echo "lfs pull"
  rm -r a.dat á.dat dir # removing files makes the status dirty
  rm -rf .git/lfs/objects
  git lfs pull
  ls -al
  [ "a" = "$(cat a.dat)" ]
  [ "A" = "$(cat "á.dat")" ]
  assert_local_object "$contents_oid" 1
  assert_local_object "$contents2_oid" 1
  git lfs fsck

  echo "lfs pull with remote"
  rm -r a.dat á.dat dir
  rm -rf .git/lfs/objects
  git lfs pull origin
  [ "a" = "$(cat a.dat)" ]
  [ "A" = "$(cat "á.dat")" ]
  assert_local_object "$contents_oid" 1
  assert_local_object "$contents2_oid" 1
  assert_clean_status
  git lfs fsck

  echo "lfs pull with local storage"
  rm a.dat á.dat
  git lfs pull
  [ "a" = "$(cat a.dat)" ]
  [ "A" = "$(cat "á.dat")" ]
  assert_clean_status

  echo "lfs pull with include/exclude filters in gitconfig"
  rm -rf .git/lfs/objects
  git config "lfs.fetchinclude" "a*"
  git lfs pull
  assert_local_object "$contents_oid" 1
  assert_clean_status

  rm -rf .git/lfs/objects
  git config --unset "lfs.fetchinclude"
  git config "lfs.fetchexclude" "a*"
  git lfs pull
  refute_local_object "$contents_oid"
  assert_clean_status

  echo "lfs pull with include/exclude filters in command line"
  git config --unset "lfs.fetchexclude"
  rm -rf .git/lfs/objects
  git lfs pull --include="a*"
  assert_local_object "$contents_oid" 1
  assert_clean_status

  rm -rf .git/lfs/objects
  git lfs pull --exclude="a*"
  refute_local_object "$contents_oid"
  assert_clean_status

  echo "resetting to test status"
  git reset --hard
  assert_clean_status

  echo "lfs pull clean status"
  git lfs pull
  assert_clean_status

  echo "lfs pull with -I"
  git lfs pull -I "*.dat"
  assert_clean_status

  echo "lfs pull in subdir"
  cd dir
  git lfs pull
  assert_clean_status

  echo "lfs pull in subdir with -I"
  git lfs pull -I "*.dat"
  assert_clean_status
)
end_test

begin_test "pull without clean filter"
(
  set -e

  GIT_LFS_SKIP_SMUDGE=1 git clone $GITSERVER/t-pull no-clean
  cd no-clean
  git lfs uninstall
  git config --list > config.txt
  grep "filter.lfs.clean" config.txt && {
    echo "clean filter still configured:"
    cat config.txt
    exit 1
  }

  contents="a"
  contents_oid=$(calc_oid "$contents")

  # LFS object not downloaded, pointer in working directory
  grep "$contents_oid" a.dat || {
    echo "a.dat not $contents_oid"
    ls -al
    cat a.dat
    exit 1
  }
  assert_local_object "$contents_oid"

  git lfs pull | tee pull.txt
  if [ "0" -ne "${PIPESTATUS[0]}" ]; then
    echo >&2 "fatal: expected pull to succeed ..."
    exit 1
  fi
  grep "Git LFS is not installed" pull.txt
  echo "pulled!"

  # LFS object downloaded, pointer unchanged
  grep "$contents_oid" a.dat || {
    echo "a.dat not $contents_oid"
    ls -al
    cat a.dat
    exit 1
  }
  assert_local_object "$contents_oid" 1
)
end_test

begin_test "pull with raw remote url"
(
  set -e
  mkdir raw
  cd raw
  git init
  git lfs install --local --skip-smudge

  git remote add origin $GITSERVER/t-pull
  git pull origin master

  contents="a"
  contents_oid=$(calc_oid "$contents")

  # LFS object not downloaded, pointer in working directory
  refute_local_object "$contents_oid"
  grep "$contents_oid" a.dat

  git lfs pull "$GITSERVER/t-pull"
  echo "pulled!"

  # LFS object downloaded and in working directory
  assert_local_object "$contents_oid" 1
  [ "0" = "$(grep -c "$contents_oid" a.dat)" ]
  [ "a" = "$(cat a.dat)" ]
)
end_test

begin_test "pull with multiple remotes"
(
  set -e
  mkdir multiple
  cd multiple
  git init
  git lfs install --local --skip-smudge

  git remote add origin "$GITSERVER/t-pull"
  git remote add bad-remote "invalid-url"
  git pull origin master

  contents="a"
  contents_oid=$(calc_oid "$contents")

  # LFS object not downloaded, pointer in working directory
  refute_local_object "$contents_oid"
  grep "$contents_oid" a.dat

  # pull should default to origin instead of bad-remote
  git lfs pull
  echo "pulled!"

  # LFS object downloaded and in working directory
  assert_local_object "$contents_oid" 1
  [ "0" = "$(grep -c "$contents_oid" a.dat)" ]
  [ "a" = "$(cat a.dat)" ]
)
end_test

begin_test "pull with invalid insteadof"
(
  set -e
  mkdir insteadof
  cd insteadof
  git init
  git lfs install --local --skip-smudge

  git remote add origin "$GITSERVER/t-pull"
  git pull origin master

  # set insteadOf to rewrite the href of downloading LFS object.
  git config url."$GITSERVER/storage/invalid".insteadOf "$GITSERVER/storage/"
  # Enable href rewriting explicitly.
  git config lfs.transfer.enablehrefrewrite true

  set +e
  git lfs pull > pull.log 2>&1
  res=$?

  set -e
  [ "$res" = "2" ]

  # check rewritten href is used to download LFS object.
  grep "LFS: Repository or object not found: $GITSERVER/storage/invalid" pull.log

  # lfs-pull succeed after unsetting enableHrefRewrite config
  git config --unset lfs.transfer.enablehrefrewrite
  git lfs pull
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
