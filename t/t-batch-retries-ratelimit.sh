#!/usr/bin/env bash

. "$(dirname "$0")/testlib.sh"

begin_test "batch storage upload causes retries"
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

  assert_server_object "$reponame" "$oid"
)
end_test

begin_test "batch storage download causes retries"
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

	assert_local_object "$oid" "${#contents}"
  popd
)
end_test

begin_test "batch clone causes retries"
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
