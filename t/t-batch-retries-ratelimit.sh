#!/usr/bin/env bash

. "$(dirname "$0")/testlib.sh"

begin_test "batch upload causes retries"
(
  set -e

  reponame="upload-batch-retry-later"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" batch-repo-upload

  contents="content"
  oid="$(calc_oid "$contents")"
  printf "%s" "$contents" > a.dat

  git lfs track "*.dat"
  git add .gitattributes a.dat
  git commit -m "initial commit"

  GIT_TRACE=1 git push origin main 2>&1 | tee push.log
  if [ "0" -ne "${PIPESTATUS[0]}" ]; then
    echo >&2 "fatal: expected \`git push origin main\` to succeed ..."
    exit 1
  fi

  [ 1 -eq $(grep -c "tq: enqueue retry" push.log) ]

  assert_server_object "$reponame" "$oid"
)
end_test

begin_test "batch upload with multiple files causes retries"
(
  set -e

  reponame="upload-multiple-batch-retry-later"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" batch-repo-upload-multiple

  contents1="content 1"
  oid1="$(calc_oid "$contents1")"
  printf "%s" "$contents1" > a.dat

  contents2="content 2"
  oid2="$(calc_oid "$contents2")"
  printf "%s" "$contents2" > b.dat

  git lfs track "*.dat"
  git add .gitattributes a.dat b.dat
  git commit -m "initial commit"

  GIT_TRACE=1 git push origin main 2>&1 | tee push.log
  if [ "0" -ne "${PIPESTATUS[0]}" ]; then
    echo >&2 "fatal: expected \`git push origin main\` to succeed ..."
    exit 1
  fi

  [ 2 -eq $(grep -c "tq: enqueue retry" push.log) ]
  [ 2 -eq $(grep -c "tq: enqueue retry #1" push.log) ]

  assert_server_object "$reponame" "$oid1"
  assert_server_object "$reponame" "$oid2"
)
end_test

begin_test "batch clone causes retries"
(
  set -e

  reponame="clone-batch-retry-later"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" batch-repo-clone

  contents="content"
  oid="$(calc_oid "$contents")"
  printf "%s" "$contents" > a.dat

  git lfs track "*.dat"
  git add .gitattributes a.dat
  git commit -m "initial commit"

  # Note that the server would apply the default rate limit delay only
  # to this push operation, not the subsequent clone whose handling of
  # 429 responses we want to test.  To avoid this, we first set the
  # available non-rate-limited requests for the repository to the maximum
  # allowed, which allows the push to proceed without delay.  Then we
  # reset the available requests to zero afterwards so the clone will
  # receive a 429 response.
  set_server_rate_limit "batch" "" "$reponame" "" "max"

  git push origin main
  assert_server_object "$reponame" "$oid"

  set_server_rate_limit "batch" "" "$reponame" "" 0

  pushd ..
    GIT_TRACE=1 git lfs clone "$GITSERVER/$reponame" "$reponame-assert" 2>&1 | tee clone.log
    if [ "0" -ne "${PIPESTATUS[0]}" ]; then
      echo >&2 "fatal: expected \`git lfs clone \"$GITSERVER/$reponame\" \"$reponame-assert\"\` to succeed ..."
      exit 1
    fi

    [ 1 -eq $(grep -c "tq: enqueue retry" clone.log) ]

    cd "$reponame-assert"

    assert_local_object "$oid" "${#contents}"
  popd
)
end_test

begin_test "batch clone with multiple files causes retries"
(
  set -e

  reponame="clone-multiple-batch-retry-later"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" batch-repo-clone-multiple

  contents1="content 1"
  oid1="$(calc_oid "$contents1")"
  printf "%s" "$contents1" > a.dat

  contents2="content 2"
  oid2="$(calc_oid "$contents2")"
  printf "%s" "$contents2" > b.dat

  git lfs track "*.dat"
  git add .gitattributes a.dat b.dat
  git commit -m "initial commit"

  # Note that the server would apply the default rate limit delay only
  # to this push operation, not the subsequent clone whose handling of
  # 429 responses we want to test.  To avoid this, we first set the
  # available non-rate-limited requests for the repository to the maximum
  # allowed, which allows the push to proceed without delay.  Then we
  # reset the available requests to zero afterwards so the clone will
  # receive a 429 response.
  set_server_rate_limit "batch" "" "$reponame" "" "max"

  git push origin main
  assert_server_object "$reponame" "$oid1"
  assert_server_object "$reponame" "$oid2"

  set_server_rate_limit "batch" "" "$reponame" "" 0

  pushd ..
    GIT_TRACE=1 git lfs clone "$GITSERVER/$reponame" "$reponame-assert" 2>&1 | tee clone.log
    if [ "0" -ne "${PIPESTATUS[0]}" ]; then
      echo >&2 "fatal: expected \`git lfs clone \"$GITSERVER/$reponame\" \"$reponame-assert\"\` to succeed ..."
      exit 1
    fi

    [ 2 -eq $(grep -c "tq: enqueue retry" clone.log) ]
    [ 2 -eq $(grep -c "tq: enqueue retry #1" clone.log) ]

    cd "$reponame-assert"

    assert_local_object "$oid1" "${#contents1}"
    assert_local_object "$oid2" "${#contents2}"
  popd
)
end_test

begin_test "batch upload causes retries (missing header)"
(
  set -e

  reponame="upload-batch-retry-later-no-header"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" batch-repo-upload-no-header

  contents="content"
  oid="$(calc_oid "$contents")"
  printf "%s" "$contents" > a.dat

  git lfs track "*.dat"
  git add .gitattributes a.dat
  git commit -m "initial commit"

  GIT_TRACE=1 git push origin main 2>&1 | tee push.log
  if [ "0" -ne "${PIPESTATUS[0]}" ]; then
    echo >&2 "fatal: expected \`git push origin main\` to succeed ..."
    exit 1
  fi

  [ 1 -lt $(grep -c "tq: enqueue retry" push.log) ]
  [ 1 -eq $(grep -c "tq: enqueue retry #1" push.log) ]
  [ 1 -eq $(grep -c "tq: enqueue retry #2" push.log) ]

  assert_server_object "$reponame" "$oid"
)
end_test
