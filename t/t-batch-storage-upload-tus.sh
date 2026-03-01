#!/usr/bin/env bash

. "$(dirname "$0")/testlib.sh"

begin_test "batch storage HTTP upload with tus protocol"
(
  set -e

  # This repository name announces to the server that it should include
  # "tus" in the set of transfer adapters it supports.
  reponame="test-tus-upload"
  setup_remote_repo "$reponame"

  clone_repo "$reponame" $reponame
  git config lfs.tusTransfers true

  git lfs track "*.dat"

  # This content announces to the server that it should return a "verify" URL
  # in its Batch API response, so the client can confirm that an upload
  # was successful.  Since we check for a successful upload using an
  # assertion, this extra step is not strictly necessary.  However, it helps
  # demonstrate that "tus" protocol support works with the full Batch API.
  contents_verify="send-verify-action"
  contents_verify_oid="$(calc_oid "$contents_verify")"

  contents="test"
  contents_oid=$(calc_oid "$contents")

  printf "%s" "$contents" > a.dat
  printf "%s" "$contents_verify" > verify.dat
  git add .gitattributes a.dat verify.dat
  git commit -m "initial commit"

  GIT_TRACE=1 GIT_CURL_VERBOSE=1 git push origin main 2>&1 | tee push.log
  if [ "0" -ne "${PIPESTATUS[0]}" ]; then
    echo >&2 "fatal: expected 'git lfs push' to succeed ..."
    exit 1
  fi

  [ 2 -eq "$(grep -c "204 No Content" push.log)" ]
  [ 4 -eq "$(grep -c "Upload-Offset: 0" push.log)" ]
  [ 1 -eq "$(grep -c "Upload-Offset: ${#contents}$" push.log)" ]
  [ 1 -eq "$(grep -c "Upload-Offset: ${#contents_verify}$" push.log)" ]

  [ 1 -eq "$(grep -c "xfer: tus.io uploading \"$contents_oid\" from start" push.log)" ]
  [ 1 -eq "$(grep -c "xfer: tus.io uploading \"$contents_verify_oid\" from start" push.log)" ]
  [ 1 -eq "$(grep -c "xfer: sending tus.io HEAD request for \"$contents_oid\"" push.log)" ]
  [ 1 -eq "$(grep -c "xfer: sending tus.io HEAD request for \"$contents_verify_oid\"" push.log)" ]
  [ 1 -eq "$(grep -c "xfer: sending tus.io PATCH request for \"$contents_oid\"" push.log)" ]
  [ 1 -eq "$(grep -c "xfer: sending tus.io PATCH request for \"$contents_verify_oid\"" push.log)" ]

  [ 0 -eq "$(grep -c "xfer: tus.io resuming upload" push.log)" ]

  assert_server_object "$reponame" "$contents_oid"
  assert_server_object "$reponame" "$contents_verify_oid"
)
end_test

begin_test "batch storage HTTP upload retries with tus protocol"
(
  set -e

  # This repository name announces to the server that it should include
  # "tus" in the set of transfer adapters it supports, and that it should
  # interrupt object uploads if they start from a zero byte offset, but
  # otherwise allow them to complete.
  reponame="test-tus-upload-interrupt"
  setup_remote_repo "$reponame"

  clone_repo "$reponame" $reponame
  git config lfs.tusTransfers true

  git lfs track "*.dat"

  # This content announces to the server that it should return a "verify" URL
  # in its Batch API response, so the client can confirm that an upload
  # was successful.  Since we check for a successful upload using an
  # assertion, this extra step is not strictly necessary.  However, it helps
  # demonstrate that "tus" protocol support works with the full Batch API.
  contents_verify="send-verify-action"
  contents_verify_oid="$(calc_oid "$contents_verify")"

  contents="test"
  contents_oid=$(calc_oid "$contents")

  printf "%s" "$contents" > a.dat
  printf "%s" "$contents_verify" > verify.dat
  git add .gitattributes a.dat verify.dat
  git commit -m "initial commit"

  GIT_TRACE=1 GIT_CURL_VERBOSE=1 git push origin main 2>&1 | tee push.log
  if [ "0" -ne "${PIPESTATUS[0]}" ]; then
    echo >&2 "fatal: expected 'git lfs push' to succeed ..."
    exit 1
  fi

  # The first attempts should start from a zero byte offset and be
  # interrupted after exactly one third of the data is accepted.
  [ 2 -eq "$(grep -c "500 Internal Server Error" push.log)" ]
  [ 4 -eq "$(grep -c "Upload-Offset: 0" push.log)" ]

  # The second attempts should start from where the first attempts were
  # interrupted, and should be successful.
  [ 2 -eq "$(grep -c "204 No Content" push.log)" ]
  [ 2 -eq "$(grep -c "Upload-Offset: $((${#contents} / 3))$" push.log)" ]
  [ 2 -eq "$(grep -c "Upload-Offset: $((${#contents_verify} / 3))$" push.log)" ]
  [ 1 -eq "$(grep -c "Upload-Offset: ${#contents}$" push.log)" ]
  [ 1 -eq "$(grep -c "Upload-Offset: ${#contents_verify}$" push.log)" ]

  [ 1 -eq "$(grep -c "xfer: tus.io uploading \"$contents_oid\" from start" push.log)" ]
  [ 1 -eq "$(grep -c "xfer: tus.io uploading \"$contents_verify_oid\" from start" push.log)" ]
  [ 1 -eq "$(grep -c "xfer: tus.io resuming upload \"$contents_oid\" from $((${#contents} / 3))$" push.log)" ]
  [ 1 -eq "$(grep -c "xfer: tus.io resuming upload \"$contents_verify_oid\" from $((${#contents_verify} / 3))$" push.log)" ]
  [ 2 -eq "$(grep -c "xfer: sending tus.io HEAD request for \"$contents_oid\"" push.log)" ]
  [ 2 -eq "$(grep -c "xfer: sending tus.io HEAD request for \"$contents_verify_oid\"" push.log)" ]
  [ 2 -eq "$(grep -c "xfer: sending tus.io PATCH request for \"$contents_oid\"" push.log)" ]
  [ 2 -eq "$(grep -c "xfer: sending tus.io PATCH request for \"$contents_verify_oid\"" push.log)" ]

  assert_server_object "$reponame" "$contents_oid"
  assert_server_object "$reponame" "$contents_verify_oid"
)
end_test
