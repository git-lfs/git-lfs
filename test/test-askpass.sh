#!/usr/bin/env bash

. "test/testlib.sh"

begin_test "askpass: push with GIT_ASKPASS"
(
  set -e

  reponame="askpass-with-environ"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  git lfs track "*.dat"
  echo "hello" > a.dat

  git add .gitattributes a.dat
  git commit -m "initial commit"

  # $password is defined from test/cmd/lfstest-gitserver.go (see: skipIfBadAuth)
  password="pass"
  GIT_ASKPASS="echo $password" GIT_TRACE=1 GIT_CURL_VERBOSE=1 git push 2>&1 | tee push.log

  grep "filling with GIT_ASKPASS: echo $password" push.log
)
end_test

begin_test "askpass: push with core.askpass"
(
  set -e

  reponame="askpass-with-config"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  git lfs track "*.dat"
  echo "hello" > a.dat

  git add .gitattributes a.dat
  git commit -m "initial commit"

  # $password is defined from test/cmd/lfstest-gitserver.go (see: skipIfBadAuth)
  password="pass"
  git config "core.askpass" "echo $password"
  cat .git/config
  GIT_TRACE=1 GIT_CURL_VERBOSE=1 git push 2>&1 | tee push.log

  grep "filling with GIT_ASKPASS: echo $password" push.log
)
end_test
