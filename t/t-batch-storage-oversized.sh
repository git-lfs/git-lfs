#!/usr/bin/env bash

. "$(dirname "$0")/testlib.sh"

begin_test "batch storage HTTP download of oversized object"
(
  set -e

  reponame="batch-storage-download-oversized-http"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  git lfs track "*.dat"
  git add .gitattributes
  git commit -m "initial commit"

  # This content announces to the server that it should append data to the
  # contents of the object before returning it.
  contents="storage-download-oversized"
  contents_oid="$(calc_oid "$contents")"
  printf "%s" "$contents" > a.dat

  git add a.dat
  git commit -m "add a.dat"

  git push origin main

  assert_server_object "$reponame" "$contents_oid"

  rm -rf .git/lfs/objects

  # We expect the server to append the string "-oversized" to the object's
  # contents, so the transfer fails but the temporary file stays at the
  # object's expected size instead of growing with the extra data.
  oversized_contents="${contents}-oversized"
  oversized_contents_oid="$(calc_oid "$oversized_contents")"
  oversized_contents_size="${#oversized_contents}"

  GIT_TRACE=1 git lfs pull 2>&1 | tee pull.log
  [ 0 -ne "${PIPESTATUS[0]}" ]

  grep "expected OID $contents_oid, got $oversized_contents_oid after ${oversized_contents_size} bytes written" pull.log
  grep "Failed to fetch some objects" pull.log

  refute_local_object "$contents_oid"
  refute_local_object "$oversized_contents_oid"

  [ "$(wc -c <.git/lfs/incomplete/$contents_oid.part)" -eq "${#contents}" ]
)
end_test

begin_test "batch storage HTTP download with oversized gzip encoding"
(
  set -e

  reponame="batch-storage-download-oversized-encoding-gzip"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  git lfs track "*.dat"

  # This content announces to the server that it should expect an
  # "Accept-Encoding: gzip" header and send a gzip-compressed response, and
  # should append data to the object's contents before compressing them.
  contents="storage-download-oversized-encoding-gzip"
  contents_oid=$(calc_oid "$contents")
  printf "%s" "$contents" > a.dat

  git add .gitattributes a.dat
  git commit -m "initial commit"

  git push origin main

  rm -rf .git/lfs/objects

  oversized_contents="${contents}-oversized"
  oversized_contents_oid="$(calc_oid "$oversized_contents")"
  oversized_contents_size="${#oversized_contents}"

  GIT_TRACE=1 GIT_CURL_VERBOSE=1 git lfs pull 2>&1 | tee pull.log
  [ 0 -ne "${PIPESTATUS[0]}" ]

  # We expect one "Accept-Encoding: gzip" header from the Batch API request,
  # prior to the object transfer download request.
  [ 2 -eq "$(grep -c "Accept-Encoding: gzip" pull.log)" ]

  [ 1 -eq "$(grep -c "decompressed gzipped response" pull.log)" ]

  grep "expected OID $contents_oid, got $oversized_contents_oid after ${oversized_contents_size} bytes written" pull.log
  grep "Failed to fetch some objects" pull.log

  refute_local_object "$contents_oid"
  refute_local_object "$oversized_contents_oid"

  # The limit applies after decompression, so the temporary file holds the
  # object's expected size, not the inflated size.
  [ "$(wc -c <.git/lfs/incomplete/$contents_oid.part)" -eq "${#contents}" ]
)
end_test

begin_test "batch storage HTTP download with oversized zstd encoding"
(
  set -e

  reponame="batch-storage-download-oversized-encoding-zstd"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  git lfs track "*.dat"

  # This content announces to the server that it should expect an
  # "Accept-Encoding: zstd" header and send a zstd-compressed response, and
  # should append data to the object's contents before compressing them.
  contents="storage-download-oversized-encoding-zstd"
  contents_oid=$(calc_oid "$contents")
  printf "%s" "$contents" > a.dat

  git add .gitattributes a.dat
  git commit -m "initial commit"

  git push origin main

  rm -rf .git/lfs/objects

  git config lfs.transfer.httpDownloadEncoding zstd

  oversized_contents="${contents}-oversized"
  oversized_contents_oid="$(calc_oid "$oversized_contents")"
  oversized_contents_size="${#oversized_contents}"

  GIT_TRACE=1 GIT_CURL_VERBOSE=1 git lfs pull 2>&1 | tee pull.log
  [ 0 -ne "${PIPESTATUS[0]}" ]

  # We expect one "Accept-Encoding: gzip" header from the Batch API request,
  # prior to the object transfer download request.
  [ 1 -eq "$(grep -c "Accept-Encoding: gzip" pull.log)" ]
  [ 1 -eq "$(grep -c "Accept-Encoding: zstd" pull.log)" ]

  [ 1 -eq "$(grep -c "Content-Encoding: zstd" pull.log)" ]
  [ 1 -eq "$(grep -c "decompressing zstd-encoded response" pull.log)" ]

  grep "expected OID $contents_oid, got $oversized_contents_oid after ${oversized_contents_size} bytes written" pull.log
  grep "Failed to fetch some objects" pull.log

  refute_local_object "$contents_oid"
  refute_local_object "$oversized_contents_oid"

  # The limit applies after decompression, so the temporary file holds the
  # object's expected size, not the inflated size.
  [ "$(wc -c <.git/lfs/incomplete/$contents_oid.part)" -eq "${#contents}" ]
)
end_test
