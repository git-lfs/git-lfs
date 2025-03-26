#!/usr/bin/env bash

. "$(dirname "$0")/testlib.sh"

begin_test "push with missing objects (lfs.allowincompletepush true)"
(
  set -e

  reponame="push-with-missing-objects"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  git lfs track "*.dat"
  git add .gitattributes
  git commit -m "initial commit"

  present="present"
  present_oid="$(calc_oid "$present")"
  printf "%s" "$present" > present.dat

  missing="missing"
  missing_oid="$(calc_oid "$missing")"
  printf "%s" "$missing" > missing.dat

  git add missing.dat present.dat
  git commit -m "add objects"

  git rm missing.dat
  git commit -m "remove missing"

  delete_local_object "$missing_oid"

  git config lfs.allowincompletepush true

  git push origin main 2>&1 | tee push.log
  if [ "0" -ne "${PIPESTATUS[0]}" ]; then
    echo >&2 "fatal: expected 'git push origin main' to succeed ..."
    exit 1
  fi

  grep "LFS upload missing objects" push.log
  grep "  (missing) missing.dat ($missing_oid)" push.log

  assert_server_object "$reponame" "$present_oid"
  refute_server_object "$reponame" "$missing_oid"
)
end_test

begin_test "push reject missing objects (lfs.allowincompletepush false)"
(
  set -e

  reponame="push-reject-missing-objects"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  git lfs track "*.dat"
  git add .gitattributes
  git commit -m "initial commit"

  present="present"
  present_oid="$(calc_oid "$present")"
  printf "%s" "$present" > present.dat

  missing="missing"
  missing_oid="$(calc_oid "$missing")"
  printf "%s" "$missing" > missing.dat

  git add missing.dat present.dat
  git commit -m "add objects"

  git rm missing.dat
  git commit -m "remove missing"

  delete_local_object "$missing_oid"

  git config lfs.allowincompletepush false

  git push origin main 2>&1 | tee push.log
  if [ "1" -ne "${PIPESTATUS[0]}" ]; then
    echo >&2 "fatal: expected 'git push origin main' to fail ..."
    exit 1
  fi

  grep 'Unable to find source' push.log

  refute_server_object "$reponame" "$present_oid"
  refute_server_object "$reponame" "$missing_oid"
)
end_test

begin_test "push reject missing objects (lfs.allowincompletepush default)"
(
  set -e

  reponame="push-missing-objects"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  git lfs track "*.dat"
  git add .gitattributes
  git commit -m "initial commit"

  missing="missing"
  missing_oid="$(calc_oid "$missing")"
  printf "%s" "$missing" > missing.dat
  git add missing.dat
  git commit -m "add missing.dat"

  present="present"
  present_oid="$(calc_oid "$present")"
  printf "%s" "$present" > present.dat
  git add present.dat
  git commit -m "add present.dat"

  delete_local_object "$missing_oid"

  git push origin main 2>&1 | tee push.log
  if [ "1" -ne "${PIPESTATUS[0]}" ]; then
    echo >&2 "fatal: expected 'git push origin main' to fail ..."
    exit 1
  fi

  grep "LFS upload failed:" push.log
  grep "  (missing) missing.dat ($missing_oid)" push.log

  assert_server_object "$reponame" "$present_oid"
  refute_server_object "$reponame" "$missing_oid"
)
end_test

begin_test "push reject corrupt objects (lfs.allowincompletepush default)"
(
  set -e

  reponame="push-corrupt-objects"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  git lfs track "*.dat"
  git add .gitattributes
  git commit -m "initial commit"

  corrupt="corrupt"
  corrupt_oid="$(calc_oid "$corrupt")"
  printf "%s" "$corrupt" > corrupt.dat
  git add corrupt.dat
  git commit -m "add corrupt.dat"

  present="present"
  present_oid="$(calc_oid "$present")"
  printf "%s" "$present" > present.dat
  git add present.dat
  git commit -m "add present.dat"

  corrupt_local_object "$corrupt_oid"

  git push origin main 2>&1 | tee push.log
  if [ "1" -ne "${PIPESTATUS[0]}" ]; then
    echo >&2 "fatal: expected 'git push origin main' to fail ..."
    exit 1
  fi

  grep "LFS upload failed:" push.log
  grep "  (corrupt) corrupt.dat ($corrupt_oid)" push.log

  assert_server_object "$reponame" "$present_oid"
  refute_server_object "$reponame" "$corrupt_oid"
)
end_test
