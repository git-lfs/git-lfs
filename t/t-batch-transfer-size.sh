#!/usr/bin/env bash

. "$(dirname "$0")/testlib.sh"

begin_test "batch storage upload with small batch size"
(
  set -e

  reponame="batch-storage-upload-small-batch"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" batch-storage-repo-upload

  contents1="storage-upload-batch-1"
  contents2="storage-upload-batch-2"
  contents3="storage-upload-batch-3"
  oid1="$(calc_oid "$contents1")"
  oid2="$(calc_oid "$contents2")"
  oid3="$(calc_oid "$contents3")"
  printf "%s" "$contents1" > a.dat
  printf "%s" "$contents2" > b.dat
  printf "%s" "$contents3" > c.dat

  git lfs track "*.dat"
  git add .gitattributes a.dat b.dat c.dat
  git commit -m "initial commit"

  git config --local lfs.transfer.batchSize 1

  GIT_TRACE=1 git push origin main 2>&1 | tee push.log
  if [ "0" -ne "${PIPESTATUS[0]}" ]; then
    echo >&2 "fatal: expected \`git push origin main\` to succeed ..."
    exit 1
  fi

  actual_count="$(grep -c "tq: sending batch of size 1" push.log)"
  [ "3" = "$actual_count" ]

  assert_server_object "$reponame" "$oid1"
  assert_server_object "$reponame" "$oid2"
  assert_server_object "$reponame" "$oid3"
)
end_test

begin_test "batch storage download with small batch size"
(
  set -e

  reponame="batch-storage-download-small-batch"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" batch-storage-repo-download

  contents1="storage-download-batch-1"
  contents2="storage-download-batch-2"
  contents3="storage-download-batch-3"
  oid1="$(calc_oid "$contents1")"
  oid2="$(calc_oid "$contents2")"
  oid3="$(calc_oid "$contents3")"
  printf "%s" "$contents1" > a.dat
  printf "%s" "$contents2" > b.dat
  printf "%s" "$contents3" > c.dat

  git lfs track "*.dat"
  git add .gitattributes a.dat b.dat c.dat
  git commit -m "initial commit"

  git push origin main
  assert_server_object "$reponame" "$oid1"
  assert_server_object "$reponame" "$oid2"
  assert_server_object "$reponame" "$oid3"

  cd ..
  git config --global lfs.transfer.batchSize 1

  GIT_TRACE=1 git clone "$GITSERVER/$reponame" "${reponame}-assert" 2>&1 | tee clone.log
  if [ "0" -ne "${PIPESTATUS[0]}" ]; then
    echo >&2 "fatal: expected \`git clone\` to succeed ..."
    exit 1
  fi

  actual_count="$(grep -c "tq: sending batch of size 1" clone.log)"
  [ "3" = "$actual_count" ]

  cd "${reponame}-assert"
  assert_local_object "$oid1" "${#contents1}"
  assert_local_object "$oid2" "${#contents2}"
  assert_local_object "$oid3" "${#contents3}"
)
end_test
