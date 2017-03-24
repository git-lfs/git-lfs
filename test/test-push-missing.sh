#!/usr/bin/env bash

. "test/testlib.sh"

begin_test "push missing objects"
(
  set -e

  reponame="push-missing-objects"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  git lfs track "*.dat"
  git add .gitattributes
  git commit -m "track *.dat files with LFS"

  missing="missing"
  missing_oid="$(calc_oid "$missing")"
  printf "$missing" > missing.dat

  present="present"
  present_oid="$(calc_oid "$present")"
  printf "$present" > present.dat

  git add missing.dat present.dat
  git commit -m "add two files"

  rm ".git/lfs/objects/${missing_oid:0:2}/${missing_oid:2:2}/$missing_oid"

  git lfs push origin master 2>&1 | tee push.log

  grep "Missing: missing.dat ($missing_oid)" push.log

  assert_server_object "$reponame" "$present_oid"
  refute_server_object "$reponame" "$missing_oid"
)
end_test

begin_test "push missing objects (--strict-mode)"
(
  set -e

  reponame="push-missing-objects-strict-mode"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  git lfs track "*.dat"
  git add .gitattributes
  git commit -m "track *.dat files with LFS"

  missing="missing"
  missing_oid="$(calc_oid "$missing")"
  printf "$missing" > missing.dat

  present="present"
  present_oid="$(calc_oid "$present")"
  printf "$present" > present.dat

  git add missing.dat present.dat
  git commit -m "add two files"

  rm ".git/lfs/objects/${missing_oid:0:2}/${missing_oid:2:2}/$missing_oid"

  git lfs push --strict-mode origin master 2>&1 | tee push.log
  if [ "0" -eq "${PIPESTATUS[0]}" ]; then
    echo >&2 "fatal: expected \"git lfs push --strict-mode\" to fail, didn't ..."
    exit 1
  fi
)
end_test

begin_test "push objects wrong size"
(
  set -e

  reponame="push-objects-wrong-size"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  git lfs track "*.dat"
  git add .gitattributes
  git commit -m "track *.dat files with LFS"

  missing="missing"
  missing_oid="$(calc_oid "$missing")"
  printf "$missing" > missing.dat

  present="present"
  present_oid="$(calc_oid "$present")"
  printf "$present" > present.dat

  git add missing.dat present.dat
  git commit -m "add two files"

  cp /dev/null ".git/lfs/objects/${missing_oid:0:2}/${missing_oid:2:2}/$missing_oid"

  git lfs push origin master 2>&1 | tee push.log

  grep "Missing: missing.dat ($missing_oid)" push.log

  assert_server_object "$reponame" "$present_oid"
  refute_server_object "$reponame" "$missing_oid"
)
end_test

begin_test "push objects wrong size (--strict-mode)"
(
  set -e

  reponame="push-objects-wrong-size-strict-mode"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  git lfs track "*.dat"
  git add .gitattributes
  git commit -m "track *.dat files with LFS"

  missing="missing"
  missing_oid="$(calc_oid "$missing")"
  printf "$missing" > missing.dat

  present="present"
  present_oid="$(calc_oid "$present")"
  printf "$present" > present.dat

  git add missing.dat present.dat
  git commit -m "add two files"

  cp /dev/null ".git/lfs/objects/${missing_oid:0:2}/${missing_oid:2:2}/$missing_oid"

  git lfs push --strict-mode origin master 2>&1 | tee push.log
  if [ "0" -eq "${PIPESTATUS[0]}" ]; then
    echo >&2 "fatal: expected \"git lfs push --strict-mode\" to fail, didn't ..."
    exit 1
  fi
)
end_test
