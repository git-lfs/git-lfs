#!/usr/bin/env bash

. "$(dirname "$0")/testlib.sh"

begin_test "lfs.httpDownloadEncoding zstd"
(
  set -e

  reponame="accept-encoding-zstd"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  git lfs track "*.dat"
  printf "contents" > a.dat
  git add .gitattributes a.dat
  git commit -m "initial commit"

  git push origin main

  # Now test download with Accept-Encoding header
  rm -rf .git/lfs/objects

  git config lfs.httpDownloadEncoding "zstd"

  GIT_CURL_VERBOSE=1 git lfs pull 2>&1 | tee curl.log

  grep "> Accept-Encoding: zstd" curl.log
)
end_test

begin_test "lfs.<url>.httpDownloadEncoding zstd"
(
  set -e

  reponame="accept-encoding-url-zstd"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  git lfs track "*.dat"
  printf "contents" > a.dat
  git add .gitattributes a.dat
  git commit -m "initial commit"

  git push origin main

  # Now test download with URL-specific Accept-Encoding header
  rm -rf .git/lfs/objects

  # Get the base URL (storage URLs are at /storage/ path on same host)
  baseurl="$(git config remote.origin.url | sed 's|/[^/]*$||')"
  git config "lfs.$baseurl/.httpDownloadEncoding" "zstd"

  GIT_CURL_VERBOSE=1 git lfs pull 2>&1 | tee curl.log

  grep "> Accept-Encoding: zstd" curl.log
)
end_test

begin_test "lfs.httpDownloadEncoding gzip does not set header"
(
  set -e

  reponame="accept-encoding-gzip"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  git lfs track "*.dat"
  printf "contents" > a.dat
  git add .gitattributes a.dat
  git commit -m "initial commit"

  git push origin main

  # Now test download with gzip (should not set explicit header)
  rm -rf .git/lfs/objects

  git config lfs.httpDownloadEncoding "gzip"

  GIT_CURL_VERBOSE=1 git lfs pull 2>&1 | tee curl.log

  # Should not have explicit Accept-Encoding: zstd header
  if grep "> Accept-Encoding: zstd" curl.log; then
    echo "Accept-Encoding should not be zstd when gzip is configured"
    exit 1
  fi
)
end_test

begin_test "lfs.httpDownloadEncoding invalid value"
(
  set -e

  reponame="accept-encoding-invalid"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  git lfs track "*.dat"
  printf "contents" > a.dat
  git add .gitattributes a.dat
  git commit -m "initial commit"

  git push origin main

  # Now test download with invalid Accept-Encoding value
  rm -rf .git/lfs/objects

  git config lfs.httpDownloadEncoding "br"

  # Should fail with an error
  set +e
  git lfs pull 2>&1 | tee pull.log
  exit_code=$?
  set -e

  grep -i "unsupported.*httpDownloadEncoding" pull.log
)
end_test
