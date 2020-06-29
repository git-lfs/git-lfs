#!/usr/bin/env bash

. "$(dirname "$0")/testlib.sh"

begin_test "redirect upload"
(
  set -e

  reponame="redirect-storage-upload"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" redirect-repo-upload

  contents="redirect-storage-upload"
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

  grep "api: redirect" push.log

  assert_server_object "$reponame" "$oid"
)
end_test
