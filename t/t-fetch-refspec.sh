#!/usr/bin/env bash

. "$(dirname "$0")/testlib.sh"

begin_test "fetch with good ref"
(
  set -e

  reponame="fetch-main-branch-required"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  git lfs track "*.dat"
  echo "a" > a.dat
  git add .gitattributes a.dat
  git commit -m "add a.dat"

  git push origin main

  # $ echo "a" | shasum -a 256
  oid="87428fc522803d31065e7bce3cf03fe475096631e5e07bbd7a0fde60c4cf25c7"
  assert_local_object "$oid" 2
  assert_server_object "$reponame" "$oid" "refs/heads/main"

  rm -rf .git/lfs/objects
  git lfs fetch --all
  assert_local_object "$oid" 2
)
end_test

begin_test "fetch with tracked ref"
(
  set -e

  reponame="fetch-tracked-branch-required"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  git lfs track "*.dat"
  echo "a" > a.dat
  git add .gitattributes a.dat
  git commit -m "add a.dat"

  git push origin main:tracked

  # $ echo "a" | shasum -a 256
  oid="87428fc522803d31065e7bce3cf03fe475096631e5e07bbd7a0fde60c4cf25c7"
  assert_local_object "$oid" 2
  assert_server_object "$reponame" "$oid" "refs/heads/tracked"

  rm -rf .git/lfs/objects
  git config push.default upstream
  git config branch.main.merge refs/heads/tracked
  git lfs fetch --all
  assert_local_object "$oid" 2
)
end_test

begin_test "fetch with bad ref"
(
  set -e

  reponame="fetch-other-branch-required"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  git lfs track "*.dat"
  echo "a" > a.dat
  git add .gitattributes a.dat
  git commit -m "add a.dat"

  git push origin main:other

  # $ echo "a" | shasum -a 256
  oid="87428fc522803d31065e7bce3cf03fe475096631e5e07bbd7a0fde60c4cf25c7"
  assert_local_object "$oid" 2
  assert_server_object "$reponame" "$oid" "refs/heads/other"

  rm -rf .git/lfs/objects
  GIT_CURL_VERBOSE=1 git lfs fetch --all 2>&1 | tee fetch.log
  if [ "0" -eq "${PIPESTATUS[0]}" ]; then
    echo >&2 "fatal: expected 'git lfs fetch' to fail"
    exit 1
  fi

  grep 'Expected ref "refs/heads/other", got "refs/heads/main"' fetch.log
)
end_test
