#!/usr/bin/env bash

. "$(dirname "$0")/testlib.sh"

# these tests rely on GIT_TERMINAL_PROMPT to test properly
ensure_git_version_isnt $VERSION_LOWER "2.3.0"

begin_test "download authenticated object"
(
  set -e
  reponame="$(basename "$0" ".sh")"
  setup_remote_repo "$reponame"
  clone_repo "$reponame" without-creds

  git lfs track "*.dat"
  printf "object-authenticated" > hi.dat
  git add hi.dat
  git add .gitattributes
  git commit -m "initial commit"

  GIT_CURL_VERBOSE=1 GIT_TERMINAL_PROMPT=0 git lfs push origin main
)
end_test
