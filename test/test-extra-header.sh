#!/usr/bin/env bash

. "test/testlib.sh"

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

  GIT_CURL_VERBOSE=1 GIT_TRACE=1 git push origin master 2>&1 | tee curl.log

  grep "> X-Foo: bar" curl.log
  grep "> X-Foo: baz" curl.log
)
end_test
