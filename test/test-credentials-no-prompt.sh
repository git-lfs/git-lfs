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
