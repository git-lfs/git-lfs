#!/usr/bin/env bash

. "$(dirname "$0")/testlib.sh"

begin_test "askpass: push with GIT_ASKPASS"
(
  set -e

  reponame="askpass-with-git-environ"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  git lfs track "*.dat"
  echo "hello" > a.dat

  git add .gitattributes a.dat
  git commit -m "initial commit"

  # $password is defined from test/cmd/lfstest-gitserver.go (see: skipIfBadAuth)
  export LFS_ASKPASS_USERNAME="user"
  export LFS_ASKPASS_PASSWORD="pass"
  git config "credential.helper" ""
  GIT_ASKPASS="lfs-askpass" SSH_ASKPASS="dont-call-me" GIT_TRACE=1 GIT_CURL_VERBOSE=1 git push origin main 2>&1 | tee push.log

  GITSERVER_USER="$(printf $GITSERVER | sed -e 's/http:\/\//http:\/\/user@/')"

  grep "filling with GIT_ASKPASS: lfs-askpass Username for \"$GITSERVER/$reponame\"" push.log
  grep "filling with GIT_ASKPASS: lfs-askpass Password for \"$GITSERVER_USER/$reponame\"" push.log
  grep "main -> main" push.log
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
  git config "credential.helper" ""
  git config "core.askPass" "lfs-askpass"
  cat .git/config
  SSH_ASKPASS="dont-call-me" GIT_TRACE=1 GIT_CURL_VERBOSE=1 git push origin main 2>&1 | tee push.log

  GITSERVER_USER="$(printf $GITSERVER | sed -e 's/http:\/\//http:\/\/user@/')"

  grep "filling with GIT_ASKPASS: lfs-askpass Username for \"$GITSERVER/$reponame\"" push.log
  grep "filling with GIT_ASKPASS: lfs-askpass Password for \"$GITSERVER_USER/$reponame\"" push.log
  grep "main -> main" push.log
)
end_test

begin_test "askpass: push with SSH_ASKPASS"
(
  set -e

  if [ ! -z "$TRAVIS" ] ; then
    # This test is known to be broken on Travis, so we skip it if the $TRAVIS
    # environment variable is set.
    #
    # See: https://github.com/git-lfs/git-lfs/pull/2500 for more.
    exit 0
  fi

  reponame="askpass-with-ssh-environ"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  git lfs track "*.dat"
  echo "hello" > a.dat

  git add .gitattributes a.dat
  git commit -m "initial commit"

  # $password is defined from test/cmd/lfstest-gitserver.go (see: skipIfBadAuth)
  export LFS_ASKPASS_USERNAME="user"
  export LFS_ASKPASS_PASSWORD="pass"
  git config "credential.helper" ""
  SSH_ASKPASS="lfs-askpass" GIT_TRACE=1 GIT_CURL_VERBOSE=1 git push origin main 2>&1 | tee push.log

  GITSERVER_USER="$(printf $GITSERVER | sed -e 's/http:\/\//http:\/\/user@/')"

  grep "filling with GIT_ASKPASS: lfs-askpass Username for \"$GITSERVER/$reponame\"" push.log
  grep "filling with GIT_ASKPASS: lfs-askpass Password for \"$GITSERVER_USER/$reponame\"" push.log
  grep "main -> main" push.log
)
end_test

begin_test "askpass: defaults to provided credentials"
(
  set -e

  reponame="askpass-provided-creds"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  git lfs track "*.dat"
  echo "hello" > a.dat

  git add .gitattributes a.dat
  git commit -m "initial commit"

  # $password is defined from test/cmd/lfstest-gitserver.go (see: skipIfBadAuth)
  export LFS_ASKPASS_USERNAME="fakeuser"
  export LFS_ASKPASS_PASSWORD="fakepass"
  git config --local "credential.helper" ""

  url=$(git config --get remote.origin.url)
  newurl=${url/http:\/\//http:\/\/user\:pass@}
  git remote set-url origin "$newurl"

  GIT_ASKPASS="lfs-askpass" GIT_TRACE=1 GIT_CURL_VERBOSE=1 git push origin main 2>&1 | tee push.log

  [ ! $(grep "filling with GIT_ASKPASS" push.log) ]
  grep "main -> main" push.log
)
end_test
