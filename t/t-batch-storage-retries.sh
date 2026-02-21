#!/usr/bin/env bash

. "$(dirname "$0")/testlib.sh"

begin_test "batch storage upload causes retries"
(
  set -e

  reponame="batch-storage-upload-retry"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" batch-storage-repo-upload

  contents="storage-upload-retry"
  oid="$(calc_oid "$contents")"
  printf "%s" "$contents" > a.dat

  git lfs track "*.dat"
  git add .gitattributes a.dat
  git commit -m "initial commit"

  git config --local lfs.transfer.maxretries 3

  GIT_TRACE=1 git push origin main 2>&1 | tee push.log
  if [ "0" -ne "${PIPESTATUS[0]}" ]; then
    echo >&2 "fatal: expected \`git push origin main\` to succeed ..."
    exit 1
  fi

  actual_count="$(grep -c "tq: retrying object $oid: Fatal error: Server error" push.log)"
  [ "2" = "$actual_count" ]

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
  printf "%s" "$contents" > a.dat

  git lfs track "*.dat"
  git add .gitattributes a.dat
  git commit -m "initial commit"

  git push origin main
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

    GIT_TRACE=1 git lfs pull origin main 2>&1 | tee pull.log
    if [ "0" -ne "${PIPESTATUS[0]}" ]; then
      echo >&2 "fatal: expected \`git lfs pull origin main\` to succeed ..."
      exit 1
    fi

    actual_count="$(grep -c "tq: retrying object $oid: Fatal error: Server error" pull.log)"
    [ "2" = "$actual_count" ]

    assert_local_object "$oid" "${#contents}"
  popd
)
end_test
