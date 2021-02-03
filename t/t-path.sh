#!/usr/bin/env bash

. "$(dirname "$0")/testlib.sh"

begin_test "does not look in current directory for git"
(
  set -e

  reponame="$(basename "$0" ".sh")"
  git init "$reponame"
  cd "$reponame"
  export PATH="$(echo "$PATH" | sed -e "s/:.:/:/g" -e "s/::/:/g")"

  printf "#!/bin/sh\necho exploit >&2\n" > git
  chmod +x git || true
  printf "echo exploit 1>&2\n" > git.bat

  # This needs to succeed.  If it fails, that could be because our malicious
  # "git" is broken but got invoked anyway.
  git lfs env > output.log 2>&1
  ! grep -q 'exploit' output.log
)
end_test

begin_test "does not look in current directory for git with credential helper"
(
  set -e

  reponame="$(basename "$0" ".sh")-credentials"
  setup_remote_repo "$reponame"

  clone_repo "$reponame" credentials-1
  export PATH="$(echo "$PATH" | sed -e "s/:.:/:/g" -e "s/::/:/g")"

  printf "#!/bin/sh\necho exploit >&2\ntouch exploit\n" > git
  chmod +x git || true
  printf "echo exploit 1>&2\r\necho >exploit" > git.bat

  git lfs track "*.dat"
  printf abc > z.dat
  git add z.dat
  git add .gitattributes
  git add git git.bat
  git commit -m "Add files"

  git push origin HEAD
  cd ..

  unset GIT_ASKPASS SSH_ASKPASS

  # This needs to succeed.  If it fails, that could be because our malicious
  # "git" is broken but got invoked anyway.
  GIT_LFS_SKIP_SMUDGE=1 clone_repo "$reponame" credentials-2

  git lfs pull | tee output.log

  ! grep -q 'exploit' output.log
  [ ! -f ../exploit ]
  [ ! -f exploit ]
)
end_test
