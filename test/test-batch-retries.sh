#!/usr/bin/env bash

. "test/testlib.sh"

begin_test "batch storage upload causes retries"
(
  set -e

  reponame="batch-storage-upload-retry"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" batch-storage-repo-upload

  contents="storage-upload-retry"
  oid="$(calc_oid "$contents")"
  printf "$contents" > a.dat

  git lfs track "*.dat"
  git add .gitattributes a.dat
  git commit -m "initial commit"

  git config --local lfs.transfer.maxretries 3
  git push origin master

  assert_server_object "$reponame" "$oid"
)
end_test
