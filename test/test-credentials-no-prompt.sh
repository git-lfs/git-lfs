#!/usr/bin/env bash

. "test/testlib.sh"

# these tests rely on GIT_TERMINAL_PROMPT to test properly
ensure_git_version_isnt $VERSION_LOWER "2.3.0"

begin_test "attempt private access without credential helper"
(
  set -e

  reponame="$(basename "$0" ".sh")"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" without-creds

  git lfs track "*.dat"
  echo "hi" > hi.dat
  git add hi.dat
  git add .gitattributes
  git commit -m "initial commit"

  git config --global credential.helper lfsnoop
  git config credential.helper lfsnoop
  git config -l

  GIT_TERMINAL_PROMPT=0 git push origin master 2>&1 | tee push.log
  grep "Authorization error: $GITSERVER/$reponame" push.log ||
    grep "Git credentials for $GITSERVER/$reponame not found" push.log
)
end_test

begin_test "askpass: push with bad askpass"
(
  set -e

  reponame="askpass-with-bad-askpass"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  git lfs track "*.dat"
  echo "hello" > a.dat

  git add .gitattributes a.dat
  git commit -m "initial commit"

  git config "credential.helper" ""
  GIT_TERMINAL_PROMPT=0 GIT_ASKPASS="lfs-askpass-2" SSH_ASKPASS="dont-call-me" GIT_TRACE=1 git push origin master 2>&1 | tee push.log
  grep "filling with GIT_ASKPASS" push.log                     # attempt askpass
  grep 'credential fill error: exec: "lfs-askpass-2"' push.log # askpass fails
  grep "creds: git credential fill" push.log                   # attempt git credential
)
end_test
