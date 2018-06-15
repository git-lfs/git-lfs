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
  git commit -m "initial commit"

  missing="missing"
  missing_oid="$(calc_oid "$missing")"
  missing_len="$(printf "$missing" | wc -c | awk '{ print $1 }')"
  printf "$missing" > missing.dat
  git add missing.dat
  git commit -m "add missing.dat"

  corrupt="corrupt"
  corrupt_oid="$(calc_oid "$corrupt")"
  corrupt_len="$(printf "$corrupt" | wc -c | awk '{ print $1 }')"
  printf "$corrupt" > corrupt.dat
  git add corrupt.dat
  git commit -m "add corrupt.dat"

  present="present"
  present_oid="$(calc_oid "$present")"
  present_len="$(printf "$present" | wc -c | awk '{ print $1 }')"
  printf "$present" > present.dat
  git add present.dat
  git commit -m "add present.dat"

  assert_local_object "$missing_oid" "$missing_len"
  assert_local_object "$corrupt_oid" "$corrupt_len"
  assert_local_object "$present_oid" "$present_len"

  delete_local_object "$missing_oid"
  corrupt_local_object "$corrupt_oid"

  refute_local_object "$missing_oid"
  refute_local_object "$corrupt_oid" "$corrupt_len"
  assert_local_object "$present_oid" "$present_len"

  git config lfs.allowincompletepush false

  git push origin master 2>&1 | tee push.log

  if [ "0" -eq "${PIPESTATUS[0]}" ]; then
    echo >&2 "fatal: expected 'git push origin master' to exit with non-zero code"
    exit 1
  fi

  grep "LFS upload failed:" push.log
  grep "  (missing) missing.dat ($missing_oid)" push.log
  grep "  (corrupt) corrupt.dat ($corrupt_oid)" push.log

  refute_server_object "$reponame" "$missing_oid"
  refute_server_object "$reponame" "$corrupt_oid"
  assert_server_object "$reponame" "$present_oid"
)
end_test
