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

  # :fire: the "missing" object
  missing_oid_part_1="$(echo "$missing_oid" | cut -b 1-2)"
  missing_oid_part_2="$(echo "$missing_oid" | cut -b 3-4)"
  missing_oid_path=".git/lfs/objects/$missing_oid_part_1/$missing_oid_part_2/$missing_oid"
  rm "$missing_oid_path"

  git config lfs.allowincompletepush true

  git push origin main 2>&1 | tee push.log
  if [ "0" -ne "${PIPESTATUS[0]}" ]; then
    echo >&2 "fatal: expected \`git push origin main\` to succeed ..."
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

  # :fire: the "missing" object
  missing_oid_part_1="$(echo "$missing_oid" | cut -b 1-2)"
  missing_oid_part_2="$(echo "$missing_oid" | cut -b 3-4)"
  missing_oid_path=".git/lfs/objects/$missing_oid_part_1/$missing_oid_part_2/$missing_oid"
  rm "$missing_oid_path"

  git config lfs.allowincompletepush false

  git push origin main 2>&1 | tee push.log
  if [ "1" -ne "${PIPESTATUS[0]}" ]; then
    echo >&2 "fatal: expected \`git push origin main\` to succeed ..."
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
  missing_len="$(printf "%s" "$missing" | wc -c | awk '{ print $1 }')"
  printf "%s" "$missing" > missing.dat
  git add missing.dat
  git commit -m "add missing.dat"

  present="present"
  present_oid="$(calc_oid "$present")"
  present_len="$(printf "%s" "$present" | wc -c | awk '{ print $1 }')"
  printf "%s" "$present" > present.dat
  git add present.dat
  git commit -m "add present.dat"

  assert_local_object "$missing_oid" "$missing_len"
  assert_local_object "$present_oid" "$present_len"

  delete_local_object "$missing_oid"

  refute_local_object "$missing_oid"
  assert_local_object "$present_oid" "$present_len"

  git push origin main 2>&1 | tee push.log

  if [ "0" -eq "${PIPESTATUS[0]}" ]; then
    echo >&2 "fatal: expected 'git push origin main' to exit with non-zero code"
    exit 1
  fi

  grep "LFS upload failed:" push.log
  grep "  (missing) missing.dat ($missing_oid)" push.log

  refute_server_object "$reponame" "$missing_oid"
  assert_server_object "$reponame" "$present_oid"
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
  corrupt_len="$(printf "%s" "$corrupt" | wc -c | awk '{ print $1 }')"
  printf "%s" "$corrupt" > corrupt.dat
  git add corrupt.dat
  git commit -m "add corrupt.dat"

  present="present"
  present_oid="$(calc_oid "$present")"
  present_len="$(printf "%s" "$present" | wc -c | awk '{ print $1 }')"
  printf "%s" "$present" > present.dat
  git add present.dat
  git commit -m "add present.dat"

  assert_local_object "$corrupt_oid" "$corrupt_len"
  assert_local_object "$present_oid" "$present_len"

  corrupt_local_object "$corrupt_oid"

  refute_local_object "$corrupt_oid" "$corrupt_len"
  assert_local_object "$present_oid" "$present_len"

  git push origin main 2>&1 | tee push.log

  if [ "0" -eq "${PIPESTATUS[0]}" ]; then
    echo >&2 "fatal: expected 'git push origin main' to exit with non-zero code"
    exit 1
  fi

  grep "LFS upload failed:" push.log
  grep "  (corrupt) corrupt.dat ($corrupt_oid)" push.log

  refute_server_object "$reponame" "$corrupt_oid"
  assert_server_object "$reponame" "$present_oid"
)
end_test
