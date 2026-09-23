#!/usr/bin/env bash

. "$(dirname "$0")/testlib.sh"

begin_test "batch transfer with duplicate OID"
(
  set -e

  reponame="batch-transfer-duplicate-oid"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  git lfs track "*.bin"

  # Content announces to the test server that Batch API responses should
  # include duplicate elements in the "objects" array for this OID.
  contents="send-duplicate-oid"
  contents_oid="$(calc_oid "$contents")"
  printf "%s" "$contents" >test.bin

  git add .gitattributes test.bin
  git commit -m "initial commit"

  set +e
  git push origin main 2>&1 | tee push.log
  res="${PIPESTATUS[0]}"
  set -e

  if [ "0" -eq "$res" ]; then
    echo "push successful with duplicate OID response?"
    exit 1
  fi

  grep "\[${contents_oid}\] The server returned a duplicate OID." push.log
)
end_test
