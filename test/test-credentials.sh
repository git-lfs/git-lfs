#!/usr/bin/env bash

. "test/testlib.sh"

begin_test "git credential"
(
  set -e

  printf "git:server" > "$CREDSDIR/credential-test.com"
  printf "git:path" > "$CREDSDIR/credential-test.com--some-path"

  mkdir empty
  cd empty
  git init

  echo "protocol=http
host=credential-test.com" | GIT_TERMINAL_PROMPT=0 git credential fill | tee cred.log

  expected="protocol=http
host=credential-test.com
username=git
password=server"
  [ "$expected" = "$(cat cred.log)" ]

  echo "protocol=http
host=credential-test.com
path=some/path" | GIT_TERMINAL_PROMPT=0 git credential fill | tee cred.log

  expected="protocol=http
host=credential-test.com
username=git
password=server"

  [ "$expected" = "$(cat cred.log)" ]

  git config credential.useHttpPath true

  echo "protocol=http
host=credential-test.com
path=some/path" | GIT_TERMINAL_PROMPT=0 git credential fill | tee cred.log

  expected="protocol=http
host=credential-test.com
path=some/path
username=git
password=path"

  [ "$expected" = "$(cat cred.log)" ]
)
end_test

begin_test "credentials without useHttpPath, with wrong password"
(
  set -e

  reponame="$(basename "$0" ".sh")"
  setup_remote_repo "$reponame"

  printf "path:wrong" > "$CREDSDIR/127.0.0.1--$reponame"

  clone_repo "$reponame" without-path
  git checkout -b without-path

  git lfs track "*.dat" 2>&1 | tee track.log
  grep "Tracking \*.dat" track.log

  contents="a"
  contents_oid=$(printf "$contents" | shasum -a 256 | cut -f 1 -d " ")

  printf "$contents" > a.dat
  git add a.dat
  git add .gitattributes
  git commit -m "add a.dat"

  git push origin without-path 2>&1 | tee push.log
  grep "(1 of 1 files)" push.log
)
end_test

begin_test "credentials with useHttpPath, with wrong password"
(
  set -e

  reponame="$(basename "$0" ".sh")"
  setup_remote_repo "$reponame"

  printf "path:wrong" > "$CREDSDIR/127.0.0.1--$reponame"

  clone_repo "$reponame" with-path-wrong-pass
  git config credential.useHttpPath true
  git checkout -b with-path-wrong-pass

  git lfs track "*.dat" 2>&1 | tee track.log
  grep "Tracking \*.dat" track.log

  contents="a"
  contents_oid=$(printf "$contents" | shasum -a 256 | cut -f 1 -d " ")

  printf "$contents" > a.dat
  git add a.dat
  git add .gitattributes
  git commit -m "add a.dat"

  git push origin with-path-wrong-pass 2>&1 | tee push.log
  [ "0" = "$(grep -c "(1 of 1 files)" push.log)" ]
)
end_test

begin_test "credentials with useHttpPath, with correct password"
(
  set -e

  reponame="$(basename "$0" ".sh")"
  setup_remote_repo "$reponame"

  printf "path:$reponame" > "$CREDSDIR/127.0.0.1--$reponame"

  clone_repo "$reponame" with-path-correct-pass
  git config credential.useHttpPath true
  git checkout -b with-path-correct-pass

  git lfs track "*.dat" 2>&1 | tee track.log
  grep "Tracking \*.dat" track.log

  contents="a"
  contents_oid=$(printf "$contents" | shasum -a 256 | cut -f 1 -d " ")

  printf "$contents" > a.dat
  git add a.dat
  git add .gitattributes
  git commit -m "add a.dat"

  git push origin with-path-correct-pass 2>&1 | tee push.log
  grep "(1 of 1 files)" push.log
)
end_test
