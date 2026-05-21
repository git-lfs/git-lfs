#!/usr/bin/env bash

. "$(dirname "$0")/testlib.sh"

begin_test "push allow missing object (lfs.allowincompletepush true)"
(
  set -e

  reponame="push-allow-missing-object"
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

  git add present.dat missing.dat
  git commit -m "add objects"

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

begin_test "push allow missing object (lfs.allowincompletepush true) (git-lfs-transfer)"
(
  set -e

  setup_pure_ssh

  reponame="push-ssh-allow-missing-object"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  sshurl=$(ssh_remote "$reponame")
  git config lfs.url "$sshurl"

  git lfs track "*.dat"
  git add .gitattributes
  git commit -m "initial commit"

  present="present"
  present_oid="$(calc_oid "$present")"
  printf "%s" "$present" > present.dat

  missing="missing"
  missing_oid="$(calc_oid "$missing")"
  printf "%s" "$missing" > missing.dat

  git add present.dat missing.dat
  git commit -m "add objects"

  delete_local_object "$missing_oid"

  git config lfs.allowincompletepush true

  GIT_TRACE=1 git push origin main 2>&1 | tee push.log
  if [ "0" -ne "${PIPESTATUS[0]}" ]; then
    echo >&2 "fatal: expected 'git push origin main' to succeed ..."
    exit 1
  fi

  grep "pure SSH connection successful" push.log

  grep "LFS upload missing objects" push.log
  grep "  (missing) missing.dat ($missing_oid)" push.log

  assert_remote_object "$reponame" "$present_oid" "${#present}"
  refute_remote_object "$reponame" "$missing_oid"
)
end_test

begin_test "push reject missing object (lfs.allowincompletepush false)"
(
  set -e

  reponame="push-reject-missing-object"
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

  git add present.dat missing.dat
  git commit -m "add objects"

  delete_local_object "$missing_oid"

  git config lfs.allowincompletepush false

  GIT_TRACE=1 git push origin main 2>&1 | tee push.log
  if [ "1" -ne "${PIPESTATUS[0]}" ]; then
    echo >&2 "fatal: expected 'git push origin main' to fail ..."
    exit 1
  fi

  grep "tq: stopping batched queue, object \"$missing_oid\" missing locally and on remote" push.log
  grep "LFS upload failed:" push.log
  grep "  (missing) missing.dat ($missing_oid)" push.log

  refute_server_object "$reponame" "$present_oid"
  refute_server_object "$reponame" "$missing_oid"
)
end_test

begin_test "push reject missing object (lfs.allowincompletepush false) (git-lfs-transfer)"
(
  set -e

  setup_pure_ssh

  reponame="push-ssh-reject-missing-object"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  sshurl=$(ssh_remote "$reponame")
  git config lfs.url "$sshurl"

  git lfs track "*.dat"
  git add .gitattributes
  git commit -m "initial commit"

  present="present"
  present_oid="$(calc_oid "$present")"
  printf "%s" "$present" > present.dat

  missing="missing"
  missing_oid="$(calc_oid "$missing")"
  printf "%s" "$missing" > missing.dat

  git add present.dat missing.dat
  git commit -m "add objects"

  delete_local_object "$missing_oid"

  git config lfs.allowincompletepush false

  GIT_TRACE=1 git push origin main 2>&1 | tee push.log
  if [ "1" -ne "${PIPESTATUS[0]}" ]; then
    echo >&2 "fatal: expected 'git push origin main' to fail ..."
    exit 1
  fi

  grep "pure SSH connection successful" push.log

  grep "tq: stopping batched queue, object \"$missing_oid\" missing locally and on remote" push.log
  grep "LFS upload failed:" push.log
  grep "  (missing) missing.dat ($missing_oid)" push.log

  refute_remote_object "$reponame" "$present_oid"
  refute_remote_object "$reponame" "$missing_oid"
)
end_test

begin_test "push reject missing object (lfs.allowincompletepush default)"
(
  set -e

  reponame="push-reject-missing-object-default"
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

  git add present.dat missing.dat
  git commit -m "add objects"

  delete_local_object "$missing_oid"

  GIT_TRACE=1 git push origin main 2>&1 | tee push.log
  if [ "1" -ne "${PIPESTATUS[0]}" ]; then
    echo >&2 "fatal: expected 'git push origin main' to fail ..."
    exit 1
  fi

  grep "tq: stopping batched queue, object \"$missing_oid\" missing locally and on remote" push.log
  grep "LFS upload failed:" push.log
  grep "  (missing) missing.dat ($missing_oid)" push.log

  refute_server_object "$reponame" "$present_oid"
  refute_server_object "$reponame" "$missing_oid"
)
end_test

begin_test "push reject missing object (lfs.allowincompletepush default) (git-lfs-transfer)"
(
  set -e

  setup_pure_ssh

  reponame="push-ssh-reject-missing-object-default"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  sshurl=$(ssh_remote "$reponame")
  git config lfs.url "$sshurl"

  git lfs track "*.dat"
  git add .gitattributes
  git commit -m "initial commit"

  present="present"
  present_oid="$(calc_oid "$present")"
  printf "%s" "$present" > present.dat

  missing="missing"
  missing_oid="$(calc_oid "$missing")"
  printf "%s" "$missing" > missing.dat

  git add present.dat missing.dat
  git commit -m "add objects"

  delete_local_object "$missing_oid"

  GIT_TRACE=1 git push origin main 2>&1 | tee push.log
  if [ "1" -ne "${PIPESTATUS[0]}" ]; then
    echo >&2 "fatal: expected 'git push origin main' to fail ..."
    exit 1
  fi

  grep "pure SSH connection successful" push.log

  grep "tq: stopping batched queue, object \"$missing_oid\" missing locally and on remote" push.log
  grep "LFS upload failed:" push.log
  grep "  (missing) missing.dat ($missing_oid)" push.log

  refute_remote_object "$reponame" "$present_oid"
  refute_remote_object "$reponame" "$missing_oid"
)
end_test

begin_test "push reject corrupt object (lfs.allowincompletepush default)"
(
  set -e

  reponame="push-reject-corrupt-object"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  git lfs track "*.dat"
  git add .gitattributes
  git commit -m "initial commit"

  present="present"
  present_oid="$(calc_oid "$present")"
  printf "%s" "$present" > present.dat

  corrupt="corrupt"
  corrupt_oid="$(calc_oid "$corrupt")"
  printf "%s" "$corrupt" > corrupt.dat

  git add present.dat corrupt.dat
  git commit -m "add objects"

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

begin_test "push reject corrupt object (lfs.allowincompletepush default) (git-lfs-transfer)"
(
  set -e

  setup_pure_ssh

  reponame="push-ssh-reject-corrupt-object-default"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  sshurl=$(ssh_remote "$reponame")
  git config lfs.url "$sshurl"

  git lfs track "*.dat"
  git add .gitattributes
  git commit -m "initial commit"

  present="present"
  present_oid="$(calc_oid "$present")"
  printf "%s" "$present" > present.dat

  corrupt="corrupt"
  corrupt_oid="$(calc_oid "$corrupt")"
  printf "%s" "$corrupt" > corrupt.dat

  git add present.dat corrupt.dat
  git commit -m "add objects"

  corrupt_local_object "$corrupt_oid"

  GIT_TRACE=1 git push origin main 2>&1 | tee push.log
  if [ "1" -ne "${PIPESTATUS[0]}" ]; then
    echo >&2 "fatal: expected 'git push origin main' to fail ..."
    exit 1
  fi

  grep "pure SSH connection successful" push.log

  grep "LFS upload failed:" push.log
  grep "  (corrupt) corrupt.dat ($corrupt_oid)" push.log

  assert_remote_object "$reponame" "$present_oid" "${#present}"
  refute_remote_object "$reponame" "$corrupt_oid"
)
end_test
