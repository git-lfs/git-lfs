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

begin_test "does not look in current directory for wrong binary using PATHEXT"
(
  set -e

  # Windows is the only platform where Go searches for executable files
  # by appending file extensions from PATHEXT.
  [ "$IS_WINDOWS" -eq 0 ] && exit 0

  reponame="$(basename "$0" ".sh")-notfound"
  git init "$reponame"
  cd "$reponame"

  # Go on Windows always looks in the current directory first when creating
  # a command handler, so we need a dummy git.exe for it to find there since
  # we will restrict PATH to exclude the real Git when we run "git-lfs env"
  # below.  If our git-lfs incorrectly proceeds to run the command handler
  # despite not finding Git in PATH either, Go may then search for a file
  # named "." with any path extension from PATHEXT and execute that file
  # instead, so we create a malicious file named "..exe" to check this case.
  touch "git$X"
  cp "$BINPATH/lfstest-badpathcheck$X" ".$X"

  # This should always succeed, even if git-lfs is incorrectly searching for
  # executables in the current directory first, because the "git-lfs env"
  # command ignores all errors when it runs "git config".  So we should always
  # pass this step and then, if our malicious program was executed, detect
  # its output below.  If this command does fail, something else is wrong.
  PATH="$BINPATH" PATHEXT="$X" "git-lfs$X" env >output.log 2>&1

  grep "exploit" output.log && false
  [ ! -f exploit ]
)
end_test
