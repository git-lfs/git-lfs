#!/usr/bin/env bash

. "$(dirname "$0")/testlib.sh"

begin_test "fetch urls text"
(
  set -eo pipefail

  reponame="fetch-urls-text"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  git lfs track "*.dat"
  echo "a" > a.dat
  echo "b" > b.dat
  git add .gitattributes a.dat b.dat
  git commit -m "add data"

  git push origin main

  rm a.dat
  git restore a.dat

  # $ echo "a" | shasum -a 256
  oid_a="87428fc522803d31065e7bce3cf03fe475096631e5e07bbd7a0fde60c4cf25c7"
  # $ echo "b" | shasum -a 256
  oid_b="0263829989b6fd954f72baaf2fc64bc2e2f01d692d4de72986ea808f6e99813f"

  git lfs fetch-urls b.dat a.dat | tee fetch.log
  grep "b.dat http://.*" fetch.log
  grep "a.dat http://.*" fetch.log
)
end_test

begin_test "fetch urls json"
(
  set -eo pipefail

  reponame="fetch-urls-json"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  git lfs track "*.dat"
  echo "a" > a.dat
  echo "b" > b.dat
  git add .gitattributes a.dat b.dat
  git commit -m "add data"

  git push origin main

  rm a.dat
  git restore a.dat

  # $ echo "a" | shasum -a 256
  oid_a="87428fc522803d31065e7bce3cf03fe475096631e5e07bbd7a0fde60c4cf25c7"
  # $ echo "b" | shasum -a 256
  oid_b="0263829989b6fd954f72baaf2fc64bc2e2f01d692d4de72986ea808f6e99813f"

  git lfs fetch-urls --json b.dat a.dat | tee fetch.json
  jq -r ".files[] | select(.name==\"a.dat\") | select(.oid==\"$oid_a\") | .href" < fetch.json | tee ahref.log
  curl $(cat ahref.log) -o redownloaded_a.dat
  cmp a.dat redownloaded_a.dat

  jq -r ".files[] | select(.name==\"b.dat\") | select(.oid==\"$oid_b\") | .href" < fetch.json | tee bhref.log
  curl $(cat bhref.log) -o redownloaded_b.dat
  cmp b.dat redownloaded_b.dat
)
end_test
