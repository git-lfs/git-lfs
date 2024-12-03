#!/usr/bin/env bash

. "$(dirname "$0")/testlib.sh"

ensure_git_version_isnt $VERSION_LOWER "2.3.0"

export CREDSDIR="$REMOTEDIR/creds-credentials-protect"
setup_creds

# Copy the default record file for the test credential helper to match the
# hostname used in the Git LFS configurations of the tests.
cp "$CREDSDIR/127.0.0.1" "$CREDSDIR/localhost"

begin_test "credentials rejected with line feed"
(
  set -e

  reponame="protect-linefeed"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  contents="a"
  contents_oid=$(calc_oid "$contents")

  git lfs track "*.dat"
  printf "%s" "$contents" >a.dat
  git add .gitattributes a.dat
  git commit -m "add a.dat"

  # Using localhost instead of 127.0.0.1 in the LFS API URL ensures this URL
  # is used when filling credentials rather than the Git remote URL, which
  # would otherwise be used since it would have the same scheme and hostname.
  gitserver="$(echo "$GITSERVER" | sed 's/127\.0\.0\.1/localhost/')"
  testreponame="test%0a$reponame"
  git config lfs.url "$gitserver/$testreponame.git/info/lfs"

  GIT_TRACE=1 git lfs push origin main 2>&1 | tee push.log
  if [ "0" -eq "${PIPESTATUS[0]}" ]; then
    echo >&2 "fatal: expected 'git lfs push' to fail ..."
    exit 1
  fi
  grep "batch response: Git credentials for $gitserver.* not found" push.log
  grep "credential value for path contains newline" push.log
  refute_server_object "$testreponame" "$contents_oid"

  git config credential.protectProtocol false

  GIT_TRACE=1 git lfs push origin main 2>&1 | tee push.log
  if [ "0" -eq "${PIPESTATUS[0]}" ]; then
    echo >&2 "fatal: expected 'git lfs push' to fail ..."
    exit 1
  fi
  grep "batch response: Git credentials for $gitserver.* not found" push.log
  grep "credential value for path contains newline" push.log
  refute_server_object "$testreponame" "$contents_oid"
)
end_test

begin_test "credentials rejected with carriage return"
(
  set -e

  reponame="protect-return"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  contents="a"
  contents_oid=$(calc_oid "$contents")

  git lfs track "*.dat"
  printf "%s" "$contents" >a.dat
  git add .gitattributes a.dat
  git commit -m "add a.dat"

  # Using localhost instead of 127.0.0.1 in the LFS API URL ensures this URL
  # is used when filling credentials rather than the Git remote URL, which
  # would otherwise be used since it would have the same scheme and hostname.
  gitserver="$(echo "$GITSERVER" | sed 's/127\.0\.0\.1/localhost/')"
  testreponame="test%0d$reponame"
  git config lfs.url "$gitserver/$testreponame.git/info/lfs"

  GIT_TRACE=1 git lfs push origin main 2>&1 | tee push.log
  if [ "0" -eq "${PIPESTATUS[0]}" ]; then
    echo >&2 "fatal: expected 'git lfs push' to fail ..."
    exit 1
  fi
  grep "batch response: Git credentials for $gitserver.* not found" push.log
  grep "credential value for path contains carriage return" push.log
  refute_server_object "$testreponame" "$contents_oid"

  git config credential.protectProtocol false

  git lfs push origin main 2>&1 | tee push.log
  if [ "0" -ne "${PIPESTATUS[0]}" ]; then
    echo >&2 "fatal: expected 'git lfs push' to succeed ..."
    exit 1
  fi
  [ $(grep -c "Uploading LFS objects: 100% (1/1)" push.log) -eq 1 ]
  assert_server_object "$testreponame" "$contents_oid"
)
end_test

begin_test "credentials rejected with null byte"
(
  set -e

  reponame="protect-null"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  contents="a"
  contents_oid=$(calc_oid "$contents")

  git lfs track "*.dat"
  printf "%s" "$contents" >a.dat
  git add .gitattributes a.dat
  git commit -m "add a.dat"

  # Using localhost instead of 127.0.0.1 in the LFS API URL ensures this URL
  # is used when filling credentials rather than the Git remote URL, which
  # would otherwise be used since it would have the same scheme and hostname.
  gitserver="$(echo "$GITSERVER" | sed 's/127\.0\.0\.1/localhost/')"
  testreponame="test%00$reponame"
  git config lfs.url "$gitserver/$testreponame.git/info/lfs"

  GIT_TRACE=1 git lfs push origin main 2>&1 | tee push.log
  if [ "0" -eq "${PIPESTATUS[0]}" ]; then
    echo >&2 "fatal: expected 'git lfs push' to fail ..."
    exit 1
  fi
  grep "batch response: Git credentials for $gitserver.* not found" push.log
  grep "credential value for path contains null byte" push.log
  refute_server_object "$testreponame" "$contents_oid"

  git config credential.protectProtocol false

  GIT_TRACE=1 git lfs push origin main 2>&1 | tee push.log
  if [ "0" -eq "${PIPESTATUS[0]}" ]; then
    echo >&2 "fatal: expected 'git lfs push' to fail ..."
    exit 1
  fi
  grep "batch response: Git credentials for $gitserver.* not found" push.log
  grep "credential value for path contains null byte" push.log
  refute_server_object "$testreponame" "$contents_oid"
)
end_test
