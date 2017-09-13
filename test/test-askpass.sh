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
  export LFS_ASKPASS_PASSWORD="pass"
  GIT_ASKPASS="lfs-askpass" GIT_TRACE=1 GIT_CURL_VERBOSE=1 git push 2>&1 | tee push.log

  grep "filling with GIT_ASKPASS: lfs-askpass" push.log
)
end_test

begin_test "askpass: push with core.askPass"
(
  set -e

  if [ ! -z "$TRAVIS" ] ; then
    # This test is known to be broken on Travis, so we skip it if the $TRAVIS
    # environment variable is set.
    #
    # See: https://github.com/git-lfs/git-lfs/pull/2500 for more.
    exit 0
  fi

  reponame="askpass-with-config"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  git lfs track "*.dat"
  echo "hello" > a.dat

  git add .gitattributes a.dat
  git commit -m "initial commit"

  # $password is defined from test/cmd/lfstest-gitserver.go (see: skipIfBadAuth)
  export LFS_ASKPASS_PASSWORD="pass"
  git config "core.askPass" "lfs-askpass"
  cat .git/config
  GIT_TRACE=1 GIT_CURL_VERBOSE=1 git push 2>&1 | tee push.log

  grep "filling with GIT_ASKPASS: lfs-askpass" push.log
)
end_test
