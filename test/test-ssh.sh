#!/usr/bin/env bash

. "test/testlib.sh"

begin_test "ssh with proxy command in lfs.url"
(
  set -e

  reponame="batch-ssh-proxy"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  sshurl="${GITSERVER/http:\/\//ssh://-oProxyCommand=ssh-proxy-test/}/$reponame"
  echo $sshurl
  git config lfs.url "$sshurl"

  contents="test"
  oid="$(calc_oid "$contents")"
  git lfs track "*.dat"
  printf "$contents" > test.dat
  git add .gitattributes test.dat
  git commit -m "initial commit"

  git push origin master 2>&1 | tee push.log
  if [ "0" -eq "${PIPESTATUS[0]}" ]; then
    echo >&2 "fatal: push succeeded"
    exit 1
  fi

  grep "got 4 args" push.log
  grep "lfs-ssh-echo -- -oProxyCommand" push.log
)
end_test
