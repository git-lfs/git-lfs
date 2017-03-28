#!/usr/bin/env bash

. "test/testlib.sh"

begin_test "verify with retries"
(
  set -e

  reponame="verify-fail-2-times"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  git lfs track "*.dat"
  git add .gitattributes
  git commit -m "initial commit"

  contents="send-verify-action"
  contents_oid="$(calc_oid "$contents")"
  contents_short_oid="$(echo "$contents_oid" | head -c 7)"
  printf "$contents" > a.dat

  git add a.dat
  git commit -m "add a.dat"

  GIT_TRACE=1 GIT_CURL_VERBOSE=1 git push origin master 2>&1 | tee push.log

  grep "Authorization: Basic * * * * *" push.log

  [ "0" -eq "${PIPESTATUS[0]}" ]
  [ "2" -eq "$(grep -c "verify $contents_short_oid attempt" push.log)" ]
)
end_test

begin_test "verify with retries (success without retry)"
(
  set -e

  reponame="verify-fail-0-times"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  git lfs track "*.dat"
  git add .gitattributes
  git commit -m "initial commit"

  contents="send-verify-action"
  contents_oid="$(calc_oid "$contents")"
  contents_short_oid="$(echo "$contents_oid" | head -c 7)"
  printf "$contents" > a.dat

  git add a.dat
  git commit -m "add a.dat"

  GIT_TRACE=1 GIT_CURL_VERBOSE=1 git push origin master 2>&1 | tee push.log

  grep "Authorization: Basic * * * * *" push.log

  [ "0" -eq "${PIPESTATUS[0]}" ]
  [ "1" -eq "$(grep -c "verify $contents_short_oid attempt" push.log)" ]
)
end_test

begin_test "verify with retries (insufficient retries)"
(
  set -e

  reponame="verify-fail-10-times"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  git lfs track "*.dat"
  git add .gitattributes
  git commit -m "initial commit"

  contents="send-verify-action"
  contents_oid="$(calc_oid "$contents")"
  contents_short_oid="$(echo "$contents_oid" | head -c 7)"
  printf "$contents" > a.dat

  git add a.dat
  git commit -m "add a.dat"

  set +e
  GIT_TRACE=1 git push origin master 2>&1 | tee push.log
  if [ "0" -eq "${PIPESTATUS[0]}" ]; then
    echo >&2 "verify: expected \"git push\" to fail, didn't ..."
    exit 1
  fi
  set -e

  [ "3" -eq "$(grep -c "verify $contents_short_oid attempt" push.log)" ]
)
end_test

begin_test "verify with retries (bad .gitconfig)"
(
  set -e

  reponame="bad-config-verify-fail-2-times"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  # Invalid `lfs.transfer.maxverifies` will default to 3.
  git config "lfs.transfer.maxverifies" "-1"

  git lfs track "*.dat"
  git add .gitattributes
  git commit -m "initial commit"

  contents="send-verify-action"
  contents_oid="$(calc_oid "$contents")"
  contents_short_oid="$(echo "$contents_oid" | head -c 7)"
  printf "$contents" > a.dat

  git add a.dat
  git commit -m "add a.dat"

  GIT_TRACE=1 GIT_CURL_VERBOSE=1 git push origin master 2>&1 | tee push.log

  grep "Authorization: Basic * * * * *" push.log

  [ "0" -eq "${PIPESTATUS[0]}" ]
  [ "2" -eq "$(grep -c "verify $contents_short_oid attempt" push.log)" ]
)
end_test
