#!/usr/bin/env bash

. "$(dirname "$0")/testlib.sh"

begin_test "http.<url>.extraHeader"
(
  set -e

  reponame="copy-headers"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  url="$(git config remote.origin.url).git/info/lfs"
  git config --add "http.$url.extraHeader" "X-Foo: bar"
  git config --add "http.$url.extraHeader" "X-Foo: baz"

  git lfs track "*.dat"
  printf "contents" > a.dat
  git add .gitattributes a.dat
  git commit -m "initial commit"

  GIT_CURL_VERBOSE=1 GIT_TRACE=1 git push origin main 2>&1 | tee curl.log

  grep "> X-Foo: bar" curl.log
  grep "> X-Foo: baz" curl.log
)
end_test

begin_test "http.<url>.extraHeader with authorization"
(
  set -e

  reponame="requirecreds-extraHeader"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  # See: test/cmd/lfstest-gitserver.go:missingRequiredCreds().
  user="requirecreds"
  pass="pass"
  auth="Basic $(echo -n $user:$pass | base64)"

  git config --add "http.extraHeader" "Authorization: $auth"

  git lfs track "*.dat"
  printf "contents" > a.dat
  git add .gitattributes a.dat
  git commit -m "initial commit"

  git push origin main 2>&1 | tee curl.log
  if [ "0" -ne "${PIPESTATUS[0]}" ]; then
    echo >&2 "expected \`git push origin main\` to succeed, didn't"
    exit 1
  fi

  [ "0" -eq "$(grep -c "creds: filling with GIT_ASKPASS" curl.log)" ]
  [ "0" -eq "$(grep -c "creds: git credential approve" curl.log)" ]
  [ "0" -eq "$(grep -c "creds: git credential cache" curl.log)" ]
  [ "0" -eq "$(grep -c "creds: git credential fill" curl.log)" ]
  [ "0" -eq "$(grep -c "creds: git credential reject" curl.log)" ]
)
end_test

begin_test "http.<url>.extraHeader with authorization (casing)"
(
  set -e

  reponame="requirecreds-extraHeaderCasing"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  # See: test/cmd/lfstest-gitserver.go:missingRequiredCreds().
  user="requirecreds"
  pass="pass"
  auth="Basic $(echo -n $user:$pass | base64)"

  git config --local --add lfs.access basic
  # N.B.: "AUTHORIZATION" is not the correct casing, and is therefore the
  # subject of this test. See lfsapi.Client.extraHeaders() for more.
  git config --local --add "http.extraHeader" "AUTHORIZATION: $auth"

  git lfs track "*.dat"
  printf "contents" > a.dat
  git add .gitattributes a.dat
  git commit -m "initial commit"

  git push origin main 2>&1 | tee curl.log
  if [ "0" -ne "${PIPESTATUS[0]}" ]; then
    echo >&2 "expected \`git push origin main\` to succeed, didn't"
    exit 1
  fi

  [ "0" -eq "$(grep -c "creds: filling with GIT_ASKPASS" curl.log)" ]
  [ "0" -eq "$(grep -c "creds: git credential approve" curl.log)" ]
  [ "0" -eq "$(grep -c "creds: git credential cache" curl.log)" ]
  [ "0" -eq "$(grep -c "creds: git credential fill" curl.log)" ]
  [ "0" -eq "$(grep -c "creds: git credential reject" curl.log)" ]
)
end_test

begin_test "http.<url>.extraHeader with mixed-case URLs"
(
  set -e

  reponame="Mixed-Case-Headers"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  # These config options check for several things.
  #
  # First, we check for mixed-case URLs being read properly and not forced to
  # lowercase. Second, we check that the user can specify a config option for
  # the Git URL and have that apply to the LFS URL, which exercises the
  # URLConfig lookup code. Finally, we also write "ExtraHeader" in mixed-case as
  # well to test that we lower-case the rightmost portion of the config key
  # during lookup.
  url="$(git config remote.origin.url).git"
  git config --add "http.$url.ExtraHeader" "X-Foo: bar"
  git config --add "http.$url.ExtraHeader" "X-Foo: baz"

  git lfs track "*.dat"
  printf "contents" > a.dat
  git add .gitattributes a.dat
  git commit -m "initial commit"

  GIT_CURL_VERBOSE=1 GIT_TRACE=1 git push origin main 2>&1 | tee curl.log

  grep "> X-Foo: bar" curl.log
  grep "> X-Foo: baz" curl.log
)
end_test
