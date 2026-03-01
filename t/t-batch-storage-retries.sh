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

  # This content announces to the server that it should strictly validate
  # Range headers, rejecting any where the ending index of the byte range is
  # smaller than the starting index.
  contents="storage-download-retry-no-invalid-range"
  contents_oid=$(calc_oid "$contents")

  printf "%s" "$contents" > a.dat
  git add a.dat
  git add .gitattributes
  git commit -m "add a.dat"
  git push origin main

  assert_server_object "$reponame" "$contents_oid"

  # Test object transfer download by simulating an interrupted and corrupt
  # initial response, after which the client should fetch the remaining bytes
  # using a request with a Range header, then detect that the assembled
  # object data is invalid and report an error.
  #
  # Next, test object transfer download again by simulating a complete but
  # corrupt initial response, after which the client should detect that
  # the object data is invalid and re-request the entire object.
  #
  # Note that the principal purpose of the first check is simply to verify
  # that the client reads the simulated temporary download files this test
  # creates.  If the client's handling of these files changes in the future,
  # this portion of the test should fail, alerting us to the need to
  # rewrite this test.
  #
  # The purpose of the second check is to verify that the client does not
  # send invalid Range headers when a temporary download file equals or
  # exceeds the size of the corresponding object.  Since this results in
  # single normal object transfer download, the test would trivially pass
  # even if the temporary file was simulated incorrectly, so the preceding
  # check exists to ensure we are simulating these files correctly.
  rm -rf .git/lfs/objects

  # Create a corrupt partial download file smaller than the actual object.
  mkdir -p .git/lfs/incomplete
  printf "%s" "aa" >".git/lfs/incomplete/$contents_oid.part"

  GIT_TRACE=1 GIT_CURL_VERBOSE=1 git lfs fetch 2>&1 | tee fetch.log
  if [ "0" -eq "${PIPESTATUS[0]}" ]; then
    echo >&2 "fatal: expected 'git lfs fetch' to fail ..."
    exit 1
  fi

  grep "Attempting to resume download of \"$contents_oid\"" fetch.log
  grep "Range: bytes=2-$((${#contents} - 1))" fetch.log

  grep "206 Partial Content" fetch.log
  grep "Content-Range: bytes 2-$((${#contents} - 1))/${#contents}" fetch.log
  grep "xfer: server accepted resume .*$contents_oid" fetch.log

  grep "expected OID $contents_oid, got" fetch.log
  grep "error: failed to fetch some objects" fetch.log

  refute_local_object "$contents_oid"

  # Create a corrupt partial download file equal in size to the actual object.
  printf "%s" "${contents/st/aa}" >".git/lfs/incomplete/$contents_oid.part"

  GIT_TRACE=1 GIT_CURL_VERBOSE=1 git lfs fetch 2>&1 | tee fetch.log
  if [ "0" -ne "${PIPESTATUS[0]}" ]; then
    echo >&2 "fatal: expected 'git lfs fetch' to succeed ..."
    exit 1
  fi

  [ 0 -eq "$(grep -c "Attempting to resume download of \"$contents_oid\"" fetch.log)" ]
  [ 0 -eq "$(grep -c "tq: retrying object $contents_oid" fetch.log)" ]
  [ 0 -eq "$(grep -c "Range: bytes=" fetch.log)" ]
  [ 0 -eq "$(grep -c "400 Bad Request" fetch.log)" ]

  assert_local_object "$contents_oid" "${#contents}"
)
end_test
