#!/usr/bin/env bash

. "$(dirname "$0")/testlib.sh"

begin_test "batch storage HTTP download with gzip encoding"
(
  set -e

  reponame="batch-storage-download-encoding-gzip"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  contents="storage-download-encoding-gzip"
  oid="$(calc_oid "$contents")"
  printf "%s" "$contents" > a.dat

  git lfs track "*.dat"
  git add .gitattributes a.dat
  git commit -m "initial commit"

  GIT_CURL_VERBOSE=1 git push origin main | tee push.log
  assert_server_object "$reponame" "$oid"

  pushd ..
    git \
      -c "filter.lfs.process=" \
      -c "filter.lfs.smudge=cat" \
      -c "filter.lfs.required=false" \
      clone "$GITSERVER/$reponame" "$reponame-assert"

    cd "$reponame-assert"

    git config credential.helper lfstest

    GIT_TRACE=1 git lfs pull origin main 2>&1 | tee pull.log
    if [ "0" -ne "${PIPESTATUS[0]}" ]; then
      echo >&2 "fatal: expected \`git lfs pull origin main\` to succeed ..."
      exit 1
    fi

    grep "decompressed gzipped response" pull.log
    assert_local_object "$oid" "${#contents}"
  popd
)
end_test
