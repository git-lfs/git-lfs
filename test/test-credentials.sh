#!/usr/bin/env bash

. "test/testlib.sh"

ensure_git_version_isnt $VERSION_LOWER "2.3.0"

begin_test "credentails with url-specific helper skips askpass"
(
  set -e

  reponame="url-specific-helper"
  setup_remote_repo "$reponame"

  clone_repo "$reponame" "$reponame"
  git config credential.useHttpPath false
  git config credential.helper ""
  git config credential.$GITSERVER.helper "lfstest"

  git lfs track "*.dat"
  echo "hello" > a.dat

  git add .gitattributes a.dat
  git commit -m "initial commit"

  # askpass is skipped
  GIT_ASKPASS="lfs-bad-cmd" GIT_TRACE=1 git push origin master 2>&1 | tee push.log

  [ "0" -eq "$(grep "filling with GIT_ASKPASS" push.log | wc -l)" ]
)
end_test

begin_test "credentials without useHttpPath, with bad path password"
(
  set -e

  reponame="no-httppath-bad-password"
  setup_remote_repo "$reponame"

  printf "path:wrong" > "$CREDSDIR/127.0.0.1--$reponame"

  clone_repo "$reponame" without-path
  git config credential.useHttpPath false
  git checkout -b without-path

  git lfs track "*.dat" 2>&1 | tee track.log
  grep "Tracking \"\*.dat\"" track.log

  printf "a" > a.dat
  git add a.dat
  git add .gitattributes
  git commit -m "add a.dat"

  GIT_TRACE=1 git push origin without-path 2>&1 | tee push.log
  grep "Uploading LFS objects: 100% (1/1), 1 B" push.log

  echo "approvals:"
  [ "1" -eq "$(cat push.log | grep "creds: git credential approve" | wc -l)" ]
  echo "fills:"
  [ "1" -eq "$(cat push.log | grep "creds: git credential fill" | wc -l)" ]

  echo "credential calls have no path:"
  credcalls="$(grep "creds: git credential" push.log)"
  [ "0" -eq "$(echo "$credcalls" | grep "no-httppath-bad-password" | wc -l)" ]
  expected="$(echo "$credcalls" | wc -l)"
  [ "$expected" -eq "$(printf "$credcalls" | grep '", "")' | wc -l)" ]
)
end_test

begin_test "credentials with url-specific useHttpPath, with bad path password"
(
  set -e

  reponame="url-specific-httppath-bad-password"
  setup_remote_repo "$reponame"

  printf "path:wrong" > "$CREDSDIR/127.0.0.1--$reponame"

  clone_repo "$reponame" with-url-specific-path
  git config credential.$GITSERVER.useHttpPath false
  git checkout -b without-path

  git lfs track "*.dat" 2>&1 | tee track.log
  grep "Tracking \"\*.dat\"" track.log

  printf "a" > a.dat
  git add a.dat
  git add .gitattributes
  git commit -m "add a.dat"

  GIT_TRACE=1 git push origin without-path 2>&1 | tee push.log
  grep "Uploading LFS objects: 100% (1/1), 1 B" push.log

  echo "approvals:"
  [ "1" -eq "$(cat push.log | grep "creds: git credential approve" | wc -l)" ]
  echo "fills:"
  [ "1" -eq "$(cat push.log | grep "creds: git credential fill" | wc -l)" ]
)
end_test

begin_test "credentials with useHttpPath, with wrong password"
(
  set -e

  reponame="httppath-bad-password"
  setup_remote_repo "$reponame"

  printf "path:wrong" > "$CREDSDIR/127.0.0.1--$reponame"

  clone_repo "$reponame" with-path-wrong-pass
  git checkout -b with-path-wrong-pass

  git lfs track "*.dat" 2>&1 | tee track.log
  grep "Tracking \"\*.dat\"" track.log

  contents="a"
  contents_oid=$(calc_oid "$contents")

  printf "$contents" > a.dat
  git add a.dat
  git add .gitattributes
  git commit -m "add a.dat"

  GIT_TRACE=1 git push origin with-path-wrong-pass 2>&1 | tee push.log
  [ "0" = "$(grep -c "Uploading LFS objects: 100% (1/1), 0 B" push.log)" ]
  echo "approvals:"
  [ "0" -eq "$(cat push.log | grep "creds: git credential approve" | wc -l)" ]
  echo "fills:"
  [ "2" -eq "$(cat push.log | grep "creds: git credential fill" | wc -l)" ]
)
end_test

begin_test "credentials with useHttpPath, with correct password"
(
  set -e

  reponame="$(basename "$0" ".sh")"
  setup_remote_repo "$reponame"

  printf "path:$reponame" > "$CREDSDIR/127.0.0.1--$reponame"

  clone_repo "$reponame" with-path-correct-pass
  git checkout -b with-path-correct-pass

  git lfs track "*.dat" 2>&1 | tee track.log
  grep "Tracking \"\*.dat\"" track.log

  # creating new branch does not re-send any objects existing on other
  # remote branches anymore, generate new object, different from prev tests
  contents="b"
  contents_oid=$(calc_oid "$contents")

  printf "$contents" > b.dat
  git add b.dat
  git add .gitattributes
  git commit -m "add b.dat"

  GIT_TRACE=1 git push origin with-path-correct-pass 2>&1 | tee push.log
  grep "Uploading LFS objects: 100% (1/1), 1 B" push.log
  echo "approvals:"
  [ "1" -eq "$(cat push.log | grep "creds: git credential approve" | wc -l)" ]
  echo "fills:"
  [ "1" -eq "$(cat push.log | grep "creds: git credential fill" | wc -l)" ]
  echo "credential calls have path:"
  credcalls="$(grep "creds: git credential" push.log)"
  [ "0" -eq "$(echo "$credcalls" | grep '", "")' | wc -l)" ]
  expected="$(echo "$credcalls" | wc -l)"
  [ "$expected" -eq "$(printf "$credcalls" | grep "test-credentials" | wc -l)" ]
)
end_test

begin_test "git credential"
(
  set -e

  printf "git:server" > "$CREDSDIR/credential-test.com"
  printf "git:path" > "$CREDSDIR/credential-test.com--some-path"

  mkdir empty
  cd empty
  git init

  echo "protocol=http
host=credential-test.com
path=some/path" | GIT_TERMINAL_PROMPT=0 git credential fill > cred.log
  cat cred.log

  expected="protocol=http
host=credential-test.com
path=some/path
username=git
password=path"

  [ "$expected" = "$(cat cred.log)" ]

  git config credential.useHttpPath false

  echo "protocol=http
host=credential-test.com" | GIT_TERMINAL_PROMPT=0 git credential fill > cred.log
  cat cred.log

  expected="protocol=http
host=credential-test.com
username=git
password=server"
  [ "$expected" = "$(cat cred.log)" ]

  echo "protocol=http
host=credential-test.com
path=some/path" | GIT_TERMINAL_PROMPT=0 git credential fill > cred.log
  cat cred.log

  expected="protocol=http
host=credential-test.com
username=git
password=server"

  [ "$expected" = "$(cat cred.log)" ]
)
end_test


if [[ $(uname) == *"MINGW"* ]]; then
  NETRCFILE="$HOME/_netrc"
else
  NETRCFILE="$HOME/.netrc"
fi


begin_test "credentials from netrc"
(
  set -e

  printf "machine localhost\nlogin netrcuser\npassword netrcpass\n" >> "$NETRCFILE"
  echo $HOME
  echo "GITSERVER $GITSERVER"
  cat $NETRCFILE

  # prevent prompts on Windows particularly
  export SSH_ASKPASS=

  reponame="netrctest"
  setup_remote_repo "$reponame"

  clone_repo "$reponame" repo

  # Need a remote named "localhost" or 127.0.0.1 in netrc will interfere with the other auth
  git remote add "netrc" "$(echo $GITSERVER | sed s/127.0.0.1/localhost/)/netrctest"
  git lfs env

  git lfs track "*.dat"
  echo "push a" > a.dat
  git add .gitattributes a.dat
  git commit -m "add a.dat"

  GIT_TRACE=1 git lfs push netrc master 2>&1 | tee push.log
  grep "Uploading LFS objects: 100% (1/1), 7 B" push.log
  echo "any git credential calls:"
  [ "0" -eq "$(cat push.log | grep "git credential" | wc -l)" ]
)
end_test

begin_test "credentials from netrc with unknown keyword"
(
  set -e

  printf "machine localhost\nlogin netrcuser\nnot-a-key something\npassword netrcpass\n" >> "$NETRCFILE"
  echo $HOME
  echo "GITSERVER $GITSERVER"
  cat $NETRCFILE

  # prevent prompts on Windows particularly
  export SSH_ASKPASS=

  reponame="netrctest"
  setup_remote_repo "$reponame"

  clone_repo "$reponame" repo2

  # Need a remote named "localhost" or 127.0.0.1 in netrc will interfere with the other auth
  git remote add "netrc" "$(echo $GITSERVER | sed s/127.0.0.1/localhost/)/netrctest"
  git lfs env

  git lfs track "*.dat"
  echo "push a" > a.dat
  git add .gitattributes a.dat
  git commit -m "add a.dat"

  GIT_TRACE=1 git lfs push netrc master 2>&1 | tee push.log
  grep "Uploading LFS objects: 100% (1/1), 7 B" push.log
  echo "any git credential calls:"
  [ "0" -eq "$(cat push.log | grep "git credential" | wc -l)" ]
)
end_test

begin_test "credentials from netrc with bad password"
(
  set -e

  printf "machine localhost\nlogin netrcuser\npassword badpass\n" >> "$NETRCFILE"
  echo $HOME
  echo "GITSERVER $GITSERVER"
  cat $NETRCFILE

  # prevent prompts on Windows particularly
  export SSH_ASKPASS=

  reponame="netrctest"
  setup_remote_repo "$reponame"

  clone_repo "$reponame" repo3

  # Need a remote named "localhost" or 127.0.0.1 in netrc will interfere with the other auth
  git remote add "netrc" "$(echo $GITSERVER | sed s/127.0.0.1/localhost/)/netrctest"
  git lfs env

  git lfs track "*.dat"
  echo "push a" > a.dat
  git add .gitattributes a.dat
  git commit -m "add a.dat"

  git push netrc master 2>&1 | tee push.log
  [ "0" = "$(grep -c "Uploading LFS objects: 100% (1/1), 7 B" push.log)" ]
)
end_test

begin_test "credentials from lfs.url"
(
  set -e

  reponame="requirecreds-lfsurl"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  git lfs track "*.dat"
  echo "push a" > a.dat
  git add .gitattributes a.dat
  git commit -m "add a.dat"

  echo "bad push"
  git lfs env
  git lfs push origin master 2>&1 | tee push.log
  grep "Uploading LFS objects:   0% (0/1), 0 B" push.log

  echo "good push"
  gitserverhost=$(echo "$GITSERVER" | cut -d'/' -f3)
  git config lfs.url http://requirecreds:pass@$gitserverhost/$reponame.git/info/lfs
  git lfs env
  git lfs push origin master 2>&1 | tee push.log
  grep "Uploading LFS objects:   0% (0/1), 0 B" push.log

  echo "bad fetch"
  rm -rf .git/lfs/objects
  git config lfs.url http://$gitserverhost/$reponame.git/info/lfs
  git lfs env
  git lfs fetch --all 2>&1 | tee fetch.log
  grep "Downloading LFS objects:   0% (0/1), 0 B" fetch.log

  echo "good fetch"
  rm -rf .git/lfs/objects
  git config lfs.url http://requirecreds:pass@$gitserverhost/$reponame.git/info/lfs
  git lfs env
  git lfs fetch --all 2>&1 | tee fetch.log
  grep "Downloading LFS objects: 100% (1/1), 7 B" fetch.log
)
end_test

begin_test "credentials from remote.origin.url"
(
  set -e

  reponame="requirecreds-remoteurl"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  git lfs track "*.dat"
  echo "push b" > b.dat
  git add .gitattributes b.dat
  git commit -m "add b.dat"

  echo "bad push"
  git lfs env
  git lfs push origin master 2>&1 | tee push.log
  grep "Uploading LFS objects:   0% (0/1), 0 B" push.log

  echo "good push"
  gitserverhost=$(echo "$GITSERVER" | cut -d'/' -f3)
  git config remote.origin.url http://requirecreds:pass@$gitserverhost/$reponame.git
  git lfs env
  git lfs push origin master 2>&1 | tee push.log
  grep "Uploading LFS objects: 100% (1/1), 7 B" push.log

  echo "bad fetch"
  rm -rf .git/lfs/objects
  git config remote.origin.url http://$gitserverhost/$reponame.git
  git lfs env
  git lfs fetch --all 2>&1 | tee fetch.log
  # Missing authentication causes `git lfs fetch` to fail before the progress
  # meter is printed to the TTY.

  echo "good fetch"
  rm -rf .git/lfs/objects
  git config remote.origin.url http://requirecreds:pass@$gitserverhost/$reponame.git
  git lfs env
  git lfs fetch --all 2>&1 | tee fetch.log
  grep "Downloading LFS objects: 100% (1/1), 7 B" fetch.log
)
end_test
