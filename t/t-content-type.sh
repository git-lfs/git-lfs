#!/usr/bin/env bash

. "$(dirname "$0")/testlib.sh"

begin_test "content-type: is enabled by default"
(
  set -e

  reponame="content-type-enabled-default"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  git lfs track "*.zip"
  printf "aaaaaaaaaa" > a.txt
  zip -j a.zip a.txt
  rm a.txt

  git add .gitattributes a.zip
  git commit -m "initial commit"
  GIT_CURL_VERBOSE=1 git push origin master 2>&1 | tee push.log

  [ 1 -eq "$(grep -c "Content-Type: application/zip" push.log)" ]
)
end_test

begin_test "content-type: is disabled by configuration"
(
  set -e

  reponame="content-type-disabled-by-configuration"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  git lfs track "*.zip"
  printf "aaaaaaaaaa" > a.txt
  zip -j a.zip a.txt
  rm a.txt

  git add .gitattributes a.zip
  git commit -m "initial commit"
  git config "lfs.$GITSERVER/$reponame.git.contenttype" 0
  GIT_CURL_VERBOSE=1 git push origin master 2>&1 | tee push.log

  [ 1 -eq "$(grep -c "Content-Type: application/zip" push.log)" ]
)
end_test
