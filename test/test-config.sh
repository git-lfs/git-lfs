#!/usr/bin/env bash

. "test/testlib.sh"

begin_test "default config"
(
  set -e
  reponame="default-config"
  mkdir $reponame
  cd $reponame
  git init
  git remote add origin "$GITSERVER/$reponame"
  git lfs env | tee env.log
  grep "Endpoint=$GITSERVER/$reponame.git/info/lfs (auth=none)" env.log

  git config --file=.gitconfig lfs.url http://gitconfig-file-ignored
  git config --file=.lfsconfig lfs.url http://lfsconfig-file
  git lfs env | tee env.log
  grep "Endpoint=http://lfsconfig-file (auth=none)" env.log

  git config lfs.url http://local-lfsconfig
  git lfs env | tee env.log
  grep "Endpoint=http://local-lfsconfig (auth=none)" env.log
)
end_test

begin_test "extension config"
(
  set -e

  git config --global lfs.extension.env-test.clean "env-test-clean"
  git config --global lfs.extension.env-test.smudge "env-test-smudge"
  git config --global lfs.extension.env-test.priority 0

  reponame="extension-config"
  mkdir $reponame
  cd $reponame
  git init

  expected0="Extension: env-test
    clean = env-test-clean
    smudge = env-test-smudge
    priority = 0"

  [ "$expected0" = "$(git lfs ext)" ]

  # any git config takes precedence over .lfsconfig
  git config --global --unset lfs.extension.env-test.priority

  git config --file=.lfsconfig lfs.extension.env-test.clean "file-env-test-clean"
  git config --file=.lfsconfig lfs.extension.env-test.smudge "file-env-test-smudge"
  git config --file=.lfsconfig lfs.extension.env-test.priority 1
  cat .lfsconfig
  expected1="Extension: env-test
    clean = env-test-clean
    smudge = env-test-smudge
    priority = 1"

  [ "$expected1" = "$(GIT_TRACE=5 git lfs ext)" ]

  git config lfs.extension.env-test.clean "local-env-test-clean"
  git config lfs.extension.env-test.smudge "local-env-test-smudge"
  git config lfs.extension.env-test.priority 2
  expected2="Extension: env-test
    clean = local-env-test-clean
    smudge = local-env-test-smudge
    priority = 2"

  [ "$expected2" = "$(git lfs ext)" ]
)
end_test

begin_test "default config (with gitconfig)"
(
  set -e
  reponame="default-config-with-gitconfig"
  mkdir $reponame
  cd $reponame
  git init
  git remote add origin "$GITSERVER/$reponame"
  git lfs env | tee env.log
  grep "Endpoint=$GITSERVER/$reponame.git/info/lfs (auth=none)" env.log

  git config --file=.gitconfig lfs.url http://gitconfig-file
  git lfs env | tee env.log
  grep "Endpoint=http://gitconfig-file (auth=none)" env.log

  git config lfs.url http://local-gitconfig
  git lfs env | tee env.log
  grep "Endpoint=http://local-gitconfig (auth=none)" env.log
)
end_test

begin_test "extension config (with gitconfig)"
(
  set -e

  git config --global lfs.extension.env-test.clean "env-test-clean"
  git config --global lfs.extension.env-test.smudge "env-test-smudge"
  git config --global lfs.extension.env-test.priority 0

  reponame="extension-config-with-gitconfig"
  mkdir $reponame
  cd $reponame
  git init

  expected0="Extension: env-test
    clean = env-test-clean
    smudge = env-test-smudge
    priority = 0"

  [ "$expected0" = "$(git lfs ext)" ]

  # any git config takes precedence over .gitconfig
  git config --global --unset lfs.extension.env-test.priority

  git config --file=.gitconfig lfs.extension.env-test.clean "file-env-test-clean"
  git config --file=.gitconfig lfs.extension.env-test.smudge "file-env-test-smudge"
  git config --file=.gitconfig lfs.extension.env-test.priority 1
  cat .gitconfig
  expected1="Extension: env-test
    clean = env-test-clean
    smudge = env-test-smudge
    priority = 1"

  [ "$expected1" = "$(GIT_TRACE=5 git lfs ext)" ]

  git config lfs.extension.env-test.clean "local-env-test-clean"
  git config lfs.extension.env-test.smudge "local-env-test-smudge"
  git config lfs.extension.env-test.priority 2
  expected2="Extension: env-test
    clean = local-env-test-clean
    smudge = local-env-test-smudge
    priority = 2"

  [ "$expected2" = "$(git lfs ext)" ]
)
end_test
