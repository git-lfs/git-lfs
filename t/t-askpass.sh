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

begin_test "askpass: push with core.askPass and wrong password"
(
  set -e

  reponame="askpass-with-config-bad-password"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  git lfs track "*.dat"

  contents="a"
  contents_oid="$(calc_oid "$contents")"
  printf "%s" "$contents" >a.dat

  git add .gitattributes a.dat
  git commit -m "initial commit"

  export LFS_ASKPASS_PASSWORD="wrong"
  git config "credential.helper" ""
  git config "core.askPass" "lfs-askpass"

  SSH_ASKPASS="dont-call-me" GIT_TRACE=1 GIT_CURL_VERBOSE=1 git push origin main 2>&1 | tee push.log
  [ 0 -eq "$(grep -c "Uploading LFS objects: 100% (1/1)" push.log)" ]

  # Requests to both the Locking API and the Batch API should receive 403s.
  [ 1 -eq "$(grep -c "Authorization error: $GITSERVER/$reponame.*/locks/verify" push.log)" ]
  [ 1 -eq "$(grep -c "batch response: Authorization error: $GITSERVER/$reponame" push.log)" ]

  GITSERVER_USER="$(printf $GITSERVER | sed -e 's/http:\/\//http:\/\/user@/')"
  [ 2 -eq "$(grep -c "filling with GIT_ASKPASS: lfs-askpass Username for \"$GITSERVER/$reponame\"" push.log)" ]
  [ 2 -eq "$(grep -c "filling with GIT_ASKPASS: lfs-askpass Password for \"$GITSERVER_USER/$reponame\"" push.log)" ]

  refute_server_object "$reponame" "$contents_oid"
)
end_test

begin_test "askpass: push with core.askPass and wrong password and 401 response"
(
  set -e

  reponame="askpass-with-config-bad-password-401-unauth"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  git lfs track "*.dat"

  contents="a"
  contents_oid="$(calc_oid "$contents")"
  printf "%s" "$contents" >a.dat

  git add .gitattributes a.dat
  git commit -m "initial commit"

  export LFS_ASKPASS_PASSWORD="wrong"
  git config "credential.helper" ""
  git config "core.askPass" "lfs-askpass"

  SSH_ASKPASS="dont-call-me" GIT_TRACE=1 GIT_CURL_VERBOSE=1 git push origin main 2>&1 | tee push.log
  [ 0 -eq "$(grep -c "Uploading LFS objects: 100% (1/1)" push.log)" ]

  # Requests to both the Locking API and the Batch API should receive 401s
  # until the maximum number of authentication attempts is reached for both.
  [ 2 -eq "$(grep -c "api: too many authentication attempts" push.log)" ]
  [ 1 -eq "$(grep -c "batch response: too many authentication attempts" push.log)" ]

  # Note that the first request to the Locking API is made without an
  # Authorization header, so no credentials are retrieved for that request
  # and it is not counted toward the authentication attempt limit.
  GITSERVER_USER="$(printf $GITSERVER | sed -e 's/http:\/\//http:\/\/user@/')"
  [ 6 -eq "$(grep -c "filling with GIT_ASKPASS: lfs-askpass Username for \"$GITSERVER/$reponame\"" push.log)" ]
  [ 6 -eq "$(grep -c "filling with GIT_ASKPASS: lfs-askpass Password for \"$GITSERVER_USER/$reponame\"" push.log)" ]

  refute_server_object "$reponame" "$contents_oid"
)
end_test

begin_test "askpass: push with SSH_ASKPASS"
(
  set -e

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

  grep "filling with GIT_ASKPASS" push.log && exit 1
  grep "main -> main" push.log
)
end_test
