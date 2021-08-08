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

  git push origin main
  assert_server_object "$reponame" "$oid"

  pushd ..
    git lfs clone "$GITSERVER/$reponame" "$reponame-assert"
    if [ "0" -ne "$?" ]; then
	  echo >&2 "fatal: expected \`git lfs clone \"$GITSERVER/$reponame\" \"$reponame-assert\"\` to su``"
	  exit 1
	fi

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

  git push origin main
  assert_server_object "$reponame" "$oid1"
  assert_server_object "$reponame" "$oid2"

  pushd ..
    git lfs clone "$GITSERVER/$reponame" "$reponame-assert"
    if [ "0" -ne "$?" ]; then
	  echo >&2 "fatal: expected \`git lfs clone \"$GITSERVER/$reponame\" \"$reponame-assert\"\` to su``"
	  exit 1
	fi

    cd "$reponame-assert"

    assert_local_object "$oid1" "${#contents1}"
    assert_local_object "$oid2" "${#contents2}"
  popd
)
end_test
