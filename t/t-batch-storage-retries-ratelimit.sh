#!/usr/bin/env bash

. "$(dirname "$0")/testlib.sh"

begin_test "batch storage HTTP upload causes delayed retries"
(
  set -e

  reponame="batch-storage-upload-retry-later"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" batch-storage-repo-upload

  contents="storage-upload-retry-later"
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

  [ 1 -eq $(grep -c "tq: retrying object" push.log) ]
  [ 1 -eq $(grep -c "tq: enqueue retry" push.log) ]

  assert_server_object "$reponame" "$oid"
)
end_test

begin_test "batch storage HTTP download causes delayed retries"
(
  set -e

  reponame="batch-storage-download-retry-later"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" batch-storage-repo-download

  contents="storage-download-retry-later"
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

    GIT_TRACE=1 git lfs pull origin main 2>&1 | tee pull.log
    if [ "0" -ne "${PIPESTATUS[0]}" ]; then
      echo >&2 "fatal: expected \`git lfs pull origin main\` to succeed ..."
      exit 1
    fi

    [ 1 -eq $(grep -c "tq: retrying object" pull.log) ]
    [ 1 -eq $(grep -c "tq: enqueue retry" pull.log) ]

    assert_local_object "$oid" "${#contents}"
  popd
)
end_test

begin_test "batch storage HTTP download causes delayed retries during clone"
(
  set -e

  reponame="batch-storage-clone-retry-later"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" batch-storage-repo-clone

  contents="storage-download-retry-later"
  oid="$(calc_oid "$contents")"
  printf "%s" "$contents" > a.dat

  git lfs track "*.dat"
  git add .gitattributes a.dat
  git commit -m "initial commit"

  git push origin main
  assert_server_object "$reponame" "$oid"

  pushd ..
    GIT_TRACE=1 git lfs clone "$GITSERVER/$reponame" "$reponame-assert" 2>&1 | tee clone.log
    if [ "0" -ne "${PIPESTATUS[0]}" ]; then
      echo >&2 "fatal: expected \`git lfs clone \"$GITSERVER/$reponame\" \"$reponame-assert\"\` to succeed ..."
      exit 1
    fi

    [ 1 -eq $(grep -c "tq: retrying object" clone.log) ]
    [ 1 -eq $(grep -c "tq: enqueue retry" clone.log) ]

    cd "$reponame-assert"

    assert_local_object "$oid" "${#contents}"
  popd
)
end_test

begin_test "batch storage HTTP upload causes retries (missing header)"
(
  set -e

  reponame="batch-storage-upload-retry-later-no-header"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" batch-storage-repo-upload-no-header

  contents="storage-upload-retry-later-no-header"
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

  [ 1 -lt $(grep -c "tq: retrying object" push.log) ]
  [ 1 -lt $(grep -c "tq: enqueue retry" push.log) ]
  [ 1 -eq $(grep -c "tq: enqueue retry #1" push.log) ]
  [ 1 -eq $(grep -c "tq: enqueue retry #2" push.log) ]

  assert_server_object "$reponame" "$oid"
)
end_test

begin_test "batch storage HTTP download causes retries (missing header)"
(
  set -e

  reponame="batch-storage-download-retry-later-no-header"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" batch-storage-repo-download-no-header

  contents="storage-download-retry-later-no-header"
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

    GIT_TRACE=1 git lfs pull origin main 2>&1 | tee pull.log
    if [ "0" -ne "${PIPESTATUS[0]}" ]; then
      echo >&2 "fatal: expected \`git lfs pull origin main\` to succeed ..."
      exit 1
    fi

    [ 1 -lt $(grep -c "tq: retrying object" pull.log) ]
    [ 1 -lt $(grep -c "tq: enqueue retry" pull.log) ]
    [ 1 -eq $(grep -c "tq: enqueue retry #1" pull.log) ]
    [ 1 -eq $(grep -c "tq: enqueue retry #2" pull.log) ]

    assert_local_object "$oid" "${#contents}"
  popd
)
end_test

begin_test "batch storage HTTP upload fails when Retry-After exceeds maxRetryTime"
(
  set -e

  reponame="batch-storage-upload-retry-later-exceeds-max"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  git config lfs.transfer.maxRetryTime 5

  # This content announces to the server that it should respond to all
  # object storage PUT requests with a 429 Too Many Requests
  # status code and a Retry-After header until at least 10 seconds
  # have passed after the first object storage PUT request.
  contents="storage-upload-retry-later"
  oid="$(calc_oid "$contents")"
  printf "%s" "$contents" > a.dat

  git lfs track "*.dat"
  git add .gitattributes a.dat
  git commit -m "initial commit"

  GIT_TRACE=1 git push origin main 2>&1 | tee push.log
  [ 0 -ne "${PIPESTATUS[0]}" ]

  grep -E "tq: refusing to retry \"$oid\", retry after [0-9]*s exceeds maximum 5s" push.log

  refute_server_object "$reponame" "$oid"
)
end_test

begin_test "batch storage HTTP download fails when Retry-After exceeds maxRetryTime"
(
  set -e

  reponame="batch-storage-download-retry-later-exceeds-max"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  # This content announces to the server that it should respond to all
  # object storage GET requests with a 429 Too Many Requests
  # status code and a Retry-After header until at least 10 seconds
  # have passed after the first object storage GET request.
  contents="storage-download-retry-later"
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
    git config lfs.transfer.maxRetryTime 5

    GIT_TRACE=1 git lfs pull origin main 2>&1 | tee pull.log
    [ 0 -ne "${PIPESTATUS[0]}" ]

    grep -E "tq: refusing to retry \"$oid\", retry after [0-9]+s exceeds maximum 5s" pull.log

    refute_local_object "$oid"
  popd
)
end_test
