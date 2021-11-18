#!/usr/bin/env bash

. "$(dirname "$0")/testlib.sh"

begin_test "ssh with proxy command in lfs.url (default variant)"
(
  set -e

  reponame="batch-ssh-proxy-default"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  sshurl="${GITSERVER/http:\/\//ssh://-oProxyCommand=ssh-proxy-test/}/$reponame"
  git config lfs.url "$sshurl"

  contents="test"
  oid="$(calc_oid "$contents")"
  git lfs track "*.dat"
  printf "%s" "$contents" > test.dat
  git add .gitattributes test.dat
  git commit -m "initial commit"

  unset GIT_SSH_VARIANT
  GIT_TRACE=1 git push origin main 2>&1 | tee push.log
  if [ "0" -eq "${PIPESTATUS[0]}" ]; then
    echo >&2 "fatal: push succeeded"
    exit 1
  fi

  grep 'expected.*git@127.0.0.1' push.log
  grep "lfs-ssh-echo -- -oProxyCommand" push.log
)
end_test

begin_test "ssh with proxy command in lfs.url (custom variant)"
(
  set -e

  reponame="batch-ssh-proxy-simple"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  sshurl="${GITSERVER/http:\/\//ssh://-oProxyCommand=ssh-proxy-test/}/$reponame"
  git config lfs.url "$sshurl"

  contents="test"
  oid="$(calc_oid "$contents")"
  git lfs track "*.dat"
  printf "%s" "$contents" > test.dat
  git add .gitattributes test.dat
  git commit -m "initial commit"

  export GIT_SSH_VARIANT=simple
  GIT_TRACE=1 git push origin main 2>&1 | tee push.log
  if [ "0" -eq "${PIPESTATUS[0]}" ]; then
    echo >&2 "fatal: push succeeded"
    exit 1
  fi

  grep 'expected.*git@127.0.0.1' push.log
  grep "lfs-ssh-echo oProxyCommand" push.log
)
end_test
