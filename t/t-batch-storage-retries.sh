#!/usr/bin/env bash

. "$(dirname "$0")/testlib.sh"

begin_test "batch storage HTTP upload causes retries"
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

begin_test "batch storage HTTP download causes retries"
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

begin_test "batch storage HTTP download retries with Range header"
(
  set -e

  reponame="batch-storage-download-retry-range"
  setup_remote_repo "$reponame"

  clone_repo "$reponame" $reponame

  git lfs track "*.dat" 2>&1 | tee track.log
  grep "Tracking \"\*.dat\"" track.log

 # This content announces to the server that it should interrupt the
 # object download unless a Range header with a positive offset was sent.
  contents="storage-download-retry-range"
  contents_oid=$(calc_oid "$contents")

  printf "%s" "$contents" > a.dat
  git add a.dat
  git add .gitattributes
  git commit -m "add a.dat"
  git push origin main

  assert_server_object "$reponame" "$contents_oid"

  # Test object transfer download with an interrupted initial response,
  # after which the client should fetch the remaining bytes using a request
  # with a Range header.
  rm -rf .git/lfs/objects

  GIT_TRACE=1 GIT_CURL_VERBOSE=1 git lfs fetch 2>&1 | tee fetch.log
  if [ "0" -ne "${PIPESTATUS[0]}" ]; then
    echo >&2 "fatal: expected 'git lfs fetch' to succeed ..."
    exit 1
  fi

  grep "Attempting to resume download of \"$contents_oid\"" fetch.log
  grep "tq: retrying object $contents_oid" fetch.log
  grep "Range: bytes=10-$((${#contents} - 1))" fetch.log

  grep "206 Partial Content" fetch.log
  grep "Content-Range: bytes 10-$((${#contents} - 1))/${#contents}" fetch.log
  grep "xfer: server accepted resume .*$contents_oid" fetch.log

  assert_local_object "$contents_oid" "${#contents}"

)
end_test

begin_test "batch storage HTTP download retries after Range header rejected"
(
  set -e

  reponame="batch-storage-download-retry-range-rejected"
  setup_remote_repo "$reponame"

  clone_repo "$reponame" $reponame

  git lfs track "*.dat" 2>&1 | tee track.log
  grep "Tracking \"\*.dat\"" track.log

  # This content announces to the server that it should interrupt the first
  # object download, then reject the client's Range header request,
  # forcing the client to fall back on re-downloading instead.
  contents="storage-download-retry-range-rejected"
  contents_oid=$(calc_oid "$contents")

  printf "%s" "$contents" > a.dat
  git add a.dat
  git add .gitattributes
  git commit -m "add a.dat"
  git push origin main

  assert_server_object "$reponame" "$contents_oid"

  # Test object transfer download with an interrupted initial response,
  # after which the client should attempt to fetch the remaining bytes using
  # a request with a Range header.  When the server rejects that request,
  # the client should then re-request the entire object.
  rm -rf .git/lfs/objects

  GIT_TRACE=1 GIT_CURL_VERBOSE=1 git lfs fetch 2>&1 | tee fetch.log
  if [ "0" -ne "${PIPESTATUS[0]}" ]; then
    echo >&2 "fatal: expected 'git lfs fetch' to succeed ..."
    exit 1
  fi

  grep "Attempting to resume download of \"$contents_oid\"" fetch.log
  grep "tq: retrying object $contents_oid" fetch.log
  grep "Range: bytes=8-$((${#contents} - 1))" fetch.log

  grep "416 Requested Range Not Satisfiable" fetch.log
  grep "xfer: server rejected resume .*$contents_oid.* re-downloading from start" fetch.log

  assert_local_object "$contents_oid" "${#contents}"
)
end_test

begin_test "batch storage HTTP download retries without invalid Range header"
(
  set -e

  reponame="batch-storage-download-retry-no-invalid-range"
  setup_remote_repo "$reponame"

  clone_repo "$reponame" $reponame

  git lfs track "*.dat" 2>&1 | tee track.log
  grep "Tracking \"\*.dat\"" track.log

  # This string announces to server that we want a test that strictly handles
  # Range headers, rejecting any where the latter part of the range is smaller
  # than the former part.
  contents="storage-download-retry-no-invalid-range"
  contents_oid=$(calc_oid "$contents")

  printf "%s" "$contents" > a.dat
  git add a.dat
  git add .gitattributes
  git commit -m "add a.dat" 2>&1 | tee commit.log
  git push origin main

  assert_server_object "$reponame" "$contents_oid"

  # Delete local copy then fetch it back.
  rm -rf .git/lfs/objects
  refute_local_object "$contents_oid"

  # Create a partial corrupt object.
  mkdir .git/lfs/incomplete
  printf "%s" "${contents/st/aa}" >".git/lfs/incomplete/$contents_oid.tmp"

  # The first download may fail with an error; run a second time to make sure
  # that we detect the corrupt file and retry.
  GIT_TRACE=1 git lfs fetch 2>&1 | tee fetchresume.log
  GIT_TRACE=1 git lfs fetch 2>&1 | tee fetchresume.log
  assert_local_object "$contents_oid" "${#contents}"
)
end_test
