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

begin_test "batch storage download causes retries"
(
  set -e

  reponame="batch-storage-download-retry"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" batch-storage-repo-download

  contents="storage-download-retry"
  oid="$(calc_oid "$contents")"
  printf "$contents" > a.dat

  git lfs track "*.dat"
  git add .gitattributes a.dat
  git commit -m "initial commit"

  git push origin master
  assert_server_object "$reponame" "$oid"

  pushd ..
    git \
      -c "filter.lfs.process=" \
      -c "filter.lfs.smudge=cat" \
      -c "filter.lfs.required=false" \
      clone "$GITSERVER/$reponame" "$reponame-assert"

    cd "$reponame-assert"

    git config credential.helper lfstest
    git config --local lfs.transfer.maxretries 3

    git lfs pull origin

    assert_local_object "$oid" "${#contents}"
  popd
)
end_test
