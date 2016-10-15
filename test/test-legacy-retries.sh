#!/usr/bin/env bash

. "test/testlib.sh"

begin_test "legacy upload check causes retries"
(
  set -e

  reponame="legacy-upload-check-retry"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" repo-upload
  git config --local lfs.batch false

  contents="legacy-upload-check-retry"
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

begin_test "legacy download check causes retries"
(
  set -e

  reponame="legacy-download-check-retry"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" repo-download

  contents="legacy-download-check-retry"
  oid="$(calc_oid "$contents")"
  printf "$contents" > a.dat

  git lfs track "*.dat"
  git add .gitattributes a.dat
  git commit -m "initial commit"

  git push origin master
  assert_server_object "$reponame" "$oid"

  pushd ..
    git \
      -c "filter.lfs.smudge=cat" \
      -c "filter.lfs.required=false" \
      clone "$GITSERVER/$reponame" "$reponame-assert"

    cd "$reponame-assert"

    git config credential.helper lfstest
    git config --local lfs.batch false
    git config --local lfs.transfer.maxretries 2

    git lfs pull origin
  popd
)
end_test

begin_test "legacy storage upload causes retries"
(
  set -e

  reponame="legacy-storage-upload-retry"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" legacy-storage-repo-upload

  git config --local lfs.batch false

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

begin_test "legacy storage download causes retries"
(
  set -e

  reponame="legacy-storage-download-retry"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" legacy-storage-repo-download

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
      -c "filter.lfs.smudge=cat" \
      -c "filter.lfs.required=false" \
      clone "$GITSERVER/$reponame" "$reponame-assert"

    cd "$reponame-assert"

    git config credential.helper lfstest
    git config --local lfs.batch false
    git config --local lfs.transfer.maxretries 3

    git lfs pull origin

    assert_local_object "$oid" "${#contents}"
  popd
)
end_test
