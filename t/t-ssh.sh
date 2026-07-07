#!/usr/bin/env bash

. "$(dirname "$0")/testlib.sh"

begin_test "ssh with proxy command in lfs.url (default variant)"
(
  set -e

  reponame="batch-ssh-proxy-default"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  sshurl="${GITSERVER/http:\/\//ssh://-oProxyCommand=ssh-proxy-test/}/$reponame"
  git config lfs.url "$sshurl"

  contents="test"
  oid="$(calc_oid "$contents")"
  git lfs track "*.dat"
  printf "%s" "$contents" > test.dat
  git add .gitattributes test.dat
  git commit -m "initial commit"

  unset GIT_SSH_VARIANT
  GIT_TRACE=1 git push origin main 2>&1 | tee push.log
  if [ "0" -eq "${PIPESTATUS[0]}" ]; then
    echo >&2 "fatal: push succeeded"
    exit 1
  fi

  grep 'expected.*git@127.0.0.1' push.log
  grep "lfs-ssh-echo -- -oProxyCommand" push.log
)
end_test

begin_test "ssh with proxy command in lfs.url (custom variant)"
(
  set -e

  reponame="batch-ssh-proxy-simple"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  sshurl="${GITSERVER/http:\/\//ssh://-oProxyCommand=ssh-proxy-test/}/$reponame"
  git config lfs.url "$sshurl"

  contents="test"
  oid="$(calc_oid "$contents")"
  git lfs track "*.dat"
  printf "%s" "$contents" > test.dat
  git add .gitattributes test.dat
  git commit -m "initial commit"

  export GIT_SSH_VARIANT=simple
  GIT_TRACE=1 git push origin main 2>&1 | tee push.log
  if [ "0" -eq "${PIPESTATUS[0]}" ]; then
    echo >&2 "fatal: push succeeded"
    exit 1
  fi

  grep 'expected.*git@127.0.0.1' push.log
  grep "lfs-ssh-echo oProxyCommand" push.log
)
end_test

begin_test "SSH Gerrit use-case (fallback to HTTPS)"
(
  set -e

  # This repository name announces to the SSH test utility that it should
  # exit with code 1 and the
  reponame="ssh-gerrit-without-lfs-plugin"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  # We construct an invalid host name (127.0.0.1.invalid) to guarantee that
  # the HTTPS fallback URL, which will default to port 443, cannot contact
  # any local or remote services not controlled by our test suite.
  gitserver_hostport="${GITSERVER#http://}"
  gitserver_host="${gitserver_hostport/%:*/}"
  gitserver_port="${gitserver_hostport/#*:/}"
  invalid_host="$gitserver_host.invalid"
  git config lfs.url "ssh://git@$invalid_host:$gitserver_port/$reponame"

  git lfs track "*.dat"

  contents="test"
  printf "%s" "$contents" >test.dat

  git add .gitattributes test.dat
  git commit -m "initial commit"

  GIT_TRACE=1 git push origin main 2>&1 | tee push.log
  [ 0 -ne "${PIPESTATUS[0]}" ]

  # Requests to both the Locking API and the Batch API should fall back to
  # an HTTPS URL, which should in turn fail because the host name is invalid.
  [ 2 -eq "$(grep -c "exec: lfs-ssh-echo.*git-lfs-authenticate /$reponame upload" push.log)" ]
  [ 2 -eq "$(grep -c -F "ssh: git@$invalid_host does not provide git-lfs-authenticate, falling back to guessed LFS endpoint" push.log)" ]
  [ 2 -eq "$(grep -c -F "HTTP: POST https://$invalid_host/$reponame/" push.log)" ]
)
end_test

begin_test "SSH git-lfs-authenticate unavailable (fallback to HTTPS)"
(
  set -e

  # This repository name announces to the SSH test utility that it should
  # exit as if the "git-lfs-authenticate" command was not found.
  reponame="ssh-unavailable"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  # We construct an invalid host name (127.0.0.1.invalid) to guarantee that
  # the HTTPS fallback URL, which will default to port 443, cannot contact
  # any local or remote services not controlled by our test suite.
  gitserver_hostport="${GITSERVER#http://}"
  gitserver_host="${gitserver_hostport/%:*/}"
  gitserver_port="${gitserver_hostport/#*:/}"
  invalid_host="$gitserver_host.invalid"
  git config lfs.url "ssh://git@$invalid_host:$gitserver_port/$reponame"

  git lfs track "*.dat"

  contents="test"
  printf "%s" "$contents" >test.dat

  git add .gitattributes test.dat
  git commit -m "initial commit"

  GIT_TRACE=1 git push origin main 2>&1 | tee push.log
  [ 0 -ne "${PIPESTATUS[0]}" ]

  # Requests to both the Locking API and the Batch API should fall back to
  # an HTTPS URL, which should in turn fail because the host name is invalid.
  [ 2 -eq "$(grep -c "exec: lfs-ssh-echo.*git-lfs-authenticate /$reponame upload" push.log)" ]
  [ 2 -eq "$(grep -c -F "ssh: git@$invalid_host does not provide git-lfs-authenticate, falling back to guessed LFS endpoint" push.log)" ]
  [ 2 -eq "$(grep -c -F "HTTP: POST https://$invalid_host/$reponame/" push.log)" ]
)
end_test
