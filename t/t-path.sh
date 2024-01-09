#!/usr/bin/env bash

. "$(dirname "$0")/testlib.sh"

begin_test "does not look in current directory for git"
(
  set -e

  reponame="$(basename "$0" ".sh")"
  git init "$reponame"
  cd "$reponame"

  cp "$BINPATH/lfstest-badpathcheck$X" "git$X"

  # This should always succeed, even if git-lfs is incorrectly searching for
  # executables in the current directory first, because the "git-lfs env"
  # command ignores all errors when it runs "git config".  So we should always
  # pass this step and then, if our malicious Git was executed, detect
  # its output below.  If this command does fail, something else is wrong.
  PATH="$BINPATH" PATHEXT="$X" "git-lfs$X" env >output.log 2>&1

  grep "exploit" output.log && false
  [ ! -f exploit ]
)
end_test

begin_test "does not look in current directory for git with credential helper"
(
  set -e

  reponame="$(basename "$0" ".sh")-credentials"
  setup_remote_repo "$reponame"

  clone_repo "$reponame" credentials-1

  git lfs track "*.dat"
  printf abc > z.dat
  git add z.dat
  git add .gitattributes

  GITPATH="$(dirname "$(command -v git)")"
  SHELLPATH="$(dirname "$(command -v sh)")"

  # We add our malicious Git to the index and then remove it from the
  # work tree so it is not found early, before we perform our key test.
  # Specifically, our "git push" below will run git-lfs, which then runs
  # "git credential", so if we are looking for Git in the current directory
  # first when running a credential helper, we will fail at that point
  # because our malicious Git will be found first.
  #
  # We prefer to check for this behavior during our "git-lfs pull" further
  # below when we are populating LFS objects into a clone of this repo
  # (which contains the malicious Git), so for now we remove the malicious
  # Git as soon as possible.
  #
  # As of Go 1.19 we also need to specify the GODEBUG environment variable
  # with a value of "execerrdot=0" in order to avoid occasional failures
  # our "git add" command below.  These failures occur due to a specific
  # set of conditions.  First, if the last-modified time of the .git/index
  # file is within a second of that of z.dat, the "git add" command will
  # refresh the Git index (assuming Git was compiled with USE_NSEC=0, as
  # appears to be the case for Git for Windows), and Git LFS will be invoked
  # to "clean" the z.dat file again.
  #
  # If that occurs, then when Git LFS runs it looks for Git, and until we
  # revise Git LFS to rely on Go's os/exec package to not execute programs
  # found in the current working directory (as described in
  # https://go.dev/blog/path-security), the os/exec package will detect our
  # malicious Git program in the current working directory and report an
  # error.  This occurs when Git LFS first initializes a new exec.Cmd
  # structure, even though Git LFS would then locate the true Git executable
  # from our custom PATH and reset the Path member of the Cmd structure
  # before trying to execute the program.
  #
  # Since we explicitly test Git LFS's avoidance of programs in the current
  # working directory using the "git-lfs pull" command further below, here
  # we just want "git add" to succeed, and so for the time being we disable
  # Go's new security checks for this command only.  We will revisit this
  # when we address the larger issue of re-adopting Go's own logic for
  # locating executable programs.
  cp "$BINPATH/lfstest-badpathcheck$X" "git$X"
  GODEBUG=execerrdot=0 \
    PATH="$BINPATH:$GITPATH:$SHELLPATH" "$GITPATH/git$X" add "git$X"
  rm "git$X"

  git commit -m "Add files"
  git push origin HEAD
  cd ..

  unset GIT_ASKPASS SSH_ASKPASS

  # When we call "git clone" below, it will run git-lfs as a smudge filter
  # during the post-clone checkout phase, and specifically will run git-lfs
  # in the newly cloned repository directory which contains a copy of our
  # malicious Git.  So, if we are looking for Git in the current directory
  # first in most cases (and not just when running a credential helper),
  # then when git-lfs runs "git config" we will fail at that point because
  # our malicious Git will be found first.  This occurs even if we specify
  # GIT_LFS_SKIP_SMUDGE=1 because git-lfs will still run "git config".
  #
  # We could ignore errors from clone_repo() and then search for the output
  # of our malicious Git in the t-path-credentials-2 directory; however,
  # this may be somewhat fragile as clone_repo() performs other steps such
  # as changing the current working directory to the new repo clone and
  # attempting to run "git config" there.
  #
  # Instead, since our key check of "git-lfs pull" below will also detect
  # the general failure case where we are looking for Git in the current
  # directory first when running most commands, we temporarily uninstall
  # Git LFS so no smudge filter will execute when "git clone" checks out the
  # repository.
  #
  # We also remove any "exploit" file potentially created by our malicious
  # Git in case it was run anywhere in clone_repo(), which may happen if
  # PATH contains the "." directory already.  Note that we reset PATH
  # to contain only the necessary directories in our key "git-lfs pull"
  # check below.
  git lfs uninstall
  clone_repo "$reponame" t-path-credentials-2
  rm -f exploit
  pushd ..
    git lfs install
  popd

  # As noted, if we are looking for Git in the current directory first
  # only when running a credential helper, then when this runs
  # "git credential", it will find our malicious Git in the current directory
  # and execute it.
  #
  # If we are looking for Git in the current directory first when running
  # most commands (and not just when running a credential helper), then this
  # will also find our malicious Git.  However, in this case it will find it
  # earlier when we try to run "git config" rather than later when we try
  # to run "git credential".
  #
  # We use a pipeline with "tee" here so as to avoid an early failure in the
  # case that our "git-lfs pull" command executes our malicious Git.
  # Unlike "git-lfs env" in the other tests, "git-lfs pull" will halt when
  # it does not receive the normal output from Git.  This in turn halts
  # our test due to our use of the "set -e" option, unless we terminate a
  # pipeline with successful command like "tee".
  PATH="$BINPATH:$GITPATH" PATHEXT="$X" "git-lfs$X" pull 2>&1 | tee output.log

  grep "exploit" output.log && false
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
