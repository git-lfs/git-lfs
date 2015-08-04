#!/usr/bin/env bash

. "test/testlib.sh"

begin_test "git credential"
(
  set -e

  printf "git:server" > "$CREDSDIR/credential-test.com"

  echo "protocol=http
host=credential-test.com" | GIT_TERMINAL_PROMPT=0 git credential fill | tee cred.log

  expected="protocol=http
host=credential-test.com
username=git
password=server"

  [ "$expected" = "$(cat cred.log)" ]
)
end_test

begin_test "credentials"
(
  set -e

  reponame="$(basename "$0" ".sh")"
  setup_remote_repo "$reponame"

  printf "path:wrong" > "$CREDSDIR/127.0.0.1--$reponame"

  git config -l

  clone_repo "$reponame" repo

  git lfs track "*.dat" 2>&1 | tee track.log
  grep "Tracking \*.dat" track.log

  contents="a"
  contents_oid=$(printf "$contents" | shasum -a 256 | cut -f 1 -d " ")

  printf "$contents" > a.dat
  git add a.dat
  git add .gitattributes
  git commit -m "add a.dat"

  git push origin master 2>&1 | tee push.log
  grep "(1 of 1 files)" push.log
)
end_test
