#!/usr/bin/env bash

. "$(dirname "$0")/testlib.sh"

begin_test "content-type: is enabled by default"
(
  set -e

  reponame="content-type-enabled-default"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  git lfs track "*.tar.gz"
  printf "aaaaaaaaaa" > a.txt
  tar -czf a.tar.gz a.txt
  rm a.txt

  git add .gitattributes a.tar.gz
  git commit -m "initial commit"
  GIT_CURL_VERBOSE=1 git push origin main 2>&1 | tee push.log

  [ 1 -eq "$(grep -c "Content-Type: application/x-gzip" push.log)" ]
  [ 0 -eq "$(grep -c "Content-Type: application/octet-stream" push.log)" ]
)
end_test

begin_test "content-type: is disabled by configuration"
(
  set -e

  reponame="content-type-disabled-by-configuration"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  git lfs track "*.tar.gz"
  printf "aaaaaaaaaa" > a.txt
  tar -czf a.tar.gz a.txt
  rm a.txt

  git add .gitattributes a.tar.gz
  git commit -m "initial commit"
  git config "lfs.$GITSERVER.contenttype" 0
  GIT_CURL_VERBOSE=1 git push origin main 2>&1 | tee push.log

  [ 0 -eq "$(grep -c "Content-Type: application/x-gzip" push.log)" ]
  [ 1 -eq "$(grep -c "Content-Type: application/octet-stream" push.log)" ]
)
end_test

begin_test "content-type: warning message"
(
  set -e

  reponame="content-type-warning-message"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" "$reponame"

  git lfs track "*.txt"
  printf "status-storage-422" > a.txt

  git add .gitattributes a.txt
  git commit -m "initial commit"
  git push origin main 2>&1 | tee push.log

  grep "info: Uploading failed due to unsupported Content-Type header(s)." push.log
  grep "info: Consider disabling Content-Type detection with:" push.log
  grep "info:   $ git config lfs.contenttype false" push.log
)
end_test
