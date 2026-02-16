#!/usr/bin/env bash

. "$(dirname "$0")/testlib.sh"

begin_test "batch storage HTTP download with gzip encoding"
(
  set -e

  reponame="batch-storage-download-encoding-gzip"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  git lfs track "*.dat"

  # This content announces to the server that it should expect an
  # "Accept-Encoding: gzip" header and send a gzip-compressed response.
  contents="storage-download-encoding-gzip"
  contents_oid=$(calc_oid "$contents")
  printf "%s" "$contents" >a.dat

  git add .gitattributes a.dat
  git commit -m "initial commit"

  git push origin main

  # Test object transfer download with an "Accept-Encoding: gzip" header
  # that should be generated automatically without explicit configuration.
  rm -rf .git/lfs/objects

  GIT_TRACE=1 GIT_CURL_VERBOSE=1 git lfs pull 2>&1 | tee pull.log
  if [ "0" -ne "${PIPESTATUS[0]}" ]; then
    echo >&2 "fatal: expected 'git lfs pull' to succeed ..."
    exit 1
  fi

  # We expect one "Accept-Encoding: gzip" header from the Batch API request,
  # prior to the object transfer download request.
  [ 2 -eq "$(grep -c "Accept-Encoding: gzip" pull.log)" ]

  [ 1 -eq "$(grep -c "decompressed gzipped response" pull.log)" ]

  assert_local_object "$contents_oid" "${#contents}"

  # Test again with an explicit configuration.
  rm -rf .git/lfs/objects

  git config lfs.transfer.httpDownloadEncoding gzip

  GIT_TRACE=1 GIT_CURL_VERBOSE=1 git lfs pull 2>&1 | tee pull.log
  if [ "0" -ne "${PIPESTATUS[0]}" ]; then
    echo >&2 "fatal: expected 'git lfs pull' to succeed ..."
    exit 1
  fi

  # We expect one "Accept-Encoding: gzip" header from the Batch API request,
  # prior to the object transfer download request.
  [ 2 -eq "$(grep -c "Accept-Encoding: gzip" pull.log)" ]

  [ 1 -eq "$(grep -c "decompressed gzipped response" pull.log)" ]

  assert_local_object "$contents_oid" "${#contents}"

  # Test again with a URL-specific configuration.
  rm -rf .git/lfs/objects

  git config --unset lfs.transfer.httpDownloadEncoding
  git config "lfs.transfer.$GITSERVER.httpDownloadEncoding" gzip

  GIT_TRACE=1 GIT_CURL_VERBOSE=1 git lfs pull 2>&1 | tee pull.log
  if [ "0" -ne "${PIPESTATUS[0]}" ]; then
    echo >&2 "fatal: expected 'git lfs pull' to succeed ..."
    exit 1
  fi

  # We expect one "Accept-Encoding: gzip" header from the Batch API request,
  # prior to the object transfer download request.
  [ 2 -eq "$(grep -c "Accept-Encoding: gzip" pull.log)" ]

  [ 1 -eq "$(grep -c "decompressed gzipped response" pull.log)" ]

  assert_local_object "$contents_oid" "${#contents}"
)
end_test

begin_test "batch storage HTTP download with zstd encoding"
(
  set -e

  reponame="batch-storage-download-encoding-zstd"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  git lfs track "*.dat"

  # This content announces to the server that it should expect an
  # "Accept-Encoding: zstd" header and send a zstd-compressed response.
  contents="storage-download-encoding-zstd"
  contents_oid=$(calc_oid "$contents")
  printf "%s" "$contents" >a.dat

  git add .gitattributes a.dat
  git commit -m "initial commit"

  git push origin main

  # Test object transfer download with an "Accept-Encoding: zstd" header.
  rm -rf .git/lfs/objects

  git config lfs.transfer.httpDownloadEncoding zstd

  GIT_TRACE=1 GIT_CURL_VERBOSE=1 git lfs pull 2>&1 | tee pull.log
  if [ "0" -ne "${PIPESTATUS[0]}" ]; then
    echo >&2 "fatal: expected 'git lfs pull' to succeed ..."
    exit 1
  fi

  # We expect one "Accept-Encoding: gzip" header from the Batch API request,
  # prior to the object transfer download request.
  [ 1 -eq "$(grep -c "Accept-Encoding: gzip" pull.log)" ]
  [ 1 -eq "$(grep -c "Accept-Encoding: zstd" pull.log)" ]

  [ 1 -eq "$(grep -c "Content-Encoding: zstd" pull.log)" ]

  assert_local_object "$contents_oid" "${#contents}"

  # Test again with a URL-specific configuration.
  rm -rf .git/lfs/objects

  git config --unset lfs.transfer.httpDownloadEncoding
  git config "lfs.transfer.$GITSERVER.httpDownloadEncoding" zstd

  GIT_TRACE=1 GIT_CURL_VERBOSE=1 git lfs pull 2>&1 | tee pull.log
  if [ "0" -ne "${PIPESTATUS[0]}" ]; then
    echo >&2 "fatal: expected 'git lfs pull' to succeed ..."
    exit 1
  fi

  # We expect one "Accept-Encoding: gzip" header from the Batch API request,
  # prior to the object transfer download request.
  [ 1 -eq "$(grep -c "Accept-Encoding: gzip" pull.log)" ]
  [ 1 -eq "$(grep -c "Accept-Encoding: zstd" pull.log)" ]

  [ 1 -eq "$(grep -c "Content-Encoding: zstd" pull.log)" ]

  assert_local_object "$contents_oid" "${#contents}"
)
end_test

begin_test "batch storage HTTP download with invalid encoding"
(
  set -e

  reponame="batch-storage-download-encoding-invalid"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  git lfs track "*.dat"
  printf "contents" > a.dat
  git add .gitattributes a.dat
  git commit -m "initial commit"

  git push origin main

  # Now test download with invalid Accept-Encoding value
  rm -rf .git/lfs/objects

  git config lfs.transfer.httpDownloadEncoding "br"

  # Should fail with an error
  git lfs pull 2>&1 | tee pull.log
  if [ "0" -eq "${PIPESTATUS[0]}" ]; then
    echo >&2 "fatal: expected 'git lfs pull' to fail ..."
    exit 1
  fi

  grep "unsupported lfs\.transfer\.httpDownloadEncoding" pull.log
)
end_test
