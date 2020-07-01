#!/usr/bin/env bash

. "$(dirname "$0")/testlib.sh"

begin_test "uninstall outside repository"
(
  set -e

  mkdir uninstall-test
  cd uninstall-test

  smudge="$(git config filter.lfs.smudge)"
  clean="$(git config filter.lfs.clean)"
  filter="$(git config filter.lfs.process)"

  printf "%s" "$smudge" | grep "git-lfs smudge"
  printf "%s" "$clean" | grep "git-lfs clean"
  printf "%s" "$filter" | grep "git-lfs filter-process"

  # uninstall multiple times to trigger https://github.com/git-lfs/git-lfs/issues/529
  git lfs uninstall

  [ ! -e "lfs" ]

  for opt in "" "--skip-repo"
  do
    git lfs install
    git lfs uninstall $opt | tee uninstall.log
    grep "configuration has been removed" uninstall.log

    [ "" = "$(git config --global filter.lfs.smudge)" ]
    [ "" = "$(git config --global filter.lfs.clean)" ]
    [ "" = "$(git config --global filter.lfs.process)" ]

    cat $HOME/.gitconfig
    [ "$(grep 'filter "lfs"' $HOME/.gitconfig -c)" = "0" ]
  done
)
end_test

begin_test "uninstall outside repository without access to .git/lfs"
(
  set -e

  mkdir uninstall-no-lfs
  cd uninstall-no-lfs

  mkdir .git
  touch .git/lfs
  touch lfs

  [ "" != "$(git config --global filter.lfs.smudge)" ]
  [ "" != "$(git config --global filter.lfs.clean)" ]
  [ "" != "$(git config --global filter.lfs.process)" ]

  git lfs uninstall

  [ "" = "$(git config --global filter.lfs.smudge)" ]
  [ "" = "$(git config --global filter.lfs.clean)" ]
  [ "" = "$(git config --global filter.lfs.process)" ]
)

begin_test "uninstall inside repository with --skip-repo"
(
  set -e

  reponame="$(basename "$0" ".sh")-skip-repo"
  mkdir "$reponame"
  cd "$reponame"
  git init
  git lfs install

  [ -f .git/hooks/pre-push ]
  grep "git-lfs" .git/hooks/pre-push

  [ "git-lfs smudge -- %f" = "$(git config filter.lfs.smudge)" ]
  [ "git-lfs clean -- %f" = "$(git config filter.lfs.clean)" ]
  [ "git-lfs filter-process" = "$(git config filter.lfs.process)" ]

  git lfs uninstall --skip-repo

  [ -f .git/hooks/pre-push ]
  [ "" = "$(git config filter.lfs.smudge)" ]
  [ "" = "$(git config filter.lfs.clean)" ]
  [ "" = "$(git config filter.lfs.process)" ]
)
end_test

begin_test "uninstall inside repository with default pre-push hook"
(
  set -e

  reponame="$(basename "$0" ".sh")-hook"
  mkdir "$reponame"
  cd "$reponame"
  git init
  git lfs install

  [ -f .git/hooks/pre-push ]
  grep "git-lfs" .git/hooks/pre-push

  [ "git-lfs smudge -- %f" = "$(git config filter.lfs.smudge)" ]
  [ "git-lfs clean -- %f" = "$(git config filter.lfs.clean)" ]
  [ "git-lfs filter-process" = "$(git config filter.lfs.process)" ]

  git lfs uninstall

  [ -f .git/hooks/pre-push ] && {
    echo "expected .git/hooks/pre-push to be deleted"
    exit 1
  }
  [ "" = "$(git config filter.lfs.smudge)" ]
  [ "" = "$(git config filter.lfs.clean)" ]
  [ "" = "$(git config filter.lfs.process)" ]
)
end_test

begin_test "uninstall inside repository without lfs pre-push hook"
(
  set -e

  reponame="$(basename "$0" ".sh")-no-hook"
  mkdir "$reponame"
  cd "$reponame"
  git init
  git lfs install
  echo "something something git-lfs" > .git/hooks/pre-push


  [ -f .git/hooks/pre-push ]
  [ "something something git-lfs" = "$(cat .git/hooks/pre-push)" ]

  [ "git-lfs smudge -- %f" = "$(git config filter.lfs.smudge)" ]
  [ "git-lfs clean -- %f" = "$(git config filter.lfs.clean)" ]
  [ "git-lfs filter-process" = "$(git config filter.lfs.process)" ]

  git lfs uninstall

  [ -f .git/hooks/pre-push ]
  [ "" = "$(git config filter.lfs.smudge)" ]
  [ "" = "$(git config filter.lfs.clean)" ]
  [ "" = "$(git config filter.lfs.process)" ]
)
end_test

begin_test "uninstall hooks inside repository"
(
  set -e

  reponame="$(basename "$0" ".sh")-only-hook"
  mkdir "$reponame"
  cd "$reponame"
  git init
  git lfs install

  [ -f .git/hooks/pre-push ]
  grep "git-lfs" .git/hooks/pre-push

  [ "git-lfs smudge -- %f" = "$(git config filter.lfs.smudge)" ]
  [ "git-lfs clean -- %f" = "$(git config filter.lfs.clean)" ]
  [ "git-lfs filter-process" = "$(git config filter.lfs.process)" ]

  git lfs uninstall hooks

  [ -f .git/hooks/pre-push ] && {
    echo "expected .git/hooks/pre-push to be deleted"
    exit 1
  }

  [ "git-lfs smudge -- %f" = "$(git config filter.lfs.smudge)" ]
  [ "git-lfs clean -- %f" = "$(git config filter.lfs.clean)" ]
  [ "git-lfs filter-process" = "$(git config filter.lfs.process)" ]
)
end_test

begin_test "uninstall --local outside repository"
(
  set -e

  # If run inside the git-lfs source dir this will update its .git/config & cause issues
  if [ "$GIT_LFS_TEST_DIR" == "" ]; then
    echo "Skipping uninstall --local because GIT_LFS_TEST_DIR is not set"
    exit 0
  fi

  has_test_dir || exit 0

  set +e
  git lfs uninstall --local >out.log
  res=$?
  set -e

  [ "Not in a git repository." = "$(cat out.log)" ]
  [ "0" != "$res" ]
)
end_test

begin_test "uninstall --local with conflicting scope"
(
  set -e

  reponame="$(basename "$0" ".sh")-scope-conflict"
  mkdir "$reponame"
  cd "$reponame"
  git init

  set +e
  git lfs uninstall --local --system 2>err.log
  res=$?
  set -e

  [ "Only one of --local and --system options can be specified." = "$(cat err.log)" ]
  [ "0" != "$res" ]
)
end_test

begin_test "uninstall --local"
(
  set -e

  # old values that should be ignored by `uninstall --local`
  git config --global filter.lfs.smudge "global smudge"
  git config --global filter.lfs.clean "global clean"
  git config --global filter.lfs.process "global filter"

  reponame="$(basename "$0" ".sh")-local"
  mkdir "$reponame"
  cd "$reponame"
  git init
  git lfs install --local

  # local configs are correct
  [ "git-lfs smudge -- %f" = "$(git config filter.lfs.smudge)" ]
  [ "git-lfs smudge -- %f" = "$(git config --local filter.lfs.smudge)" ]
  [ "git-lfs clean -- %f" = "$(git config filter.lfs.clean)" ]
  [ "git-lfs clean -- %f" = "$(git config --local filter.lfs.clean)" ]
  [ "git-lfs filter-process" = "$(git config filter.lfs.process)" ]
  [ "git-lfs filter-process" = "$(git config --local filter.lfs.process)" ]

  # global configs
  [ "global smudge" = "$(git config --global filter.lfs.smudge)" ]
  [ "global clean" = "$(git config --global filter.lfs.clean)" ]
  [ "global filter" = "$(git config --global filter.lfs.process)" ]

  git lfs uninstall --local 2>&1 | tee uninstall.log
  if [ ${PIPESTATUS[0]} -ne 0 ]; then
    echo >&2 "fatal: expected 'git lfs uninstall --local' to succeed"
    exit 1
  fi
  grep -v "Global Git LFS configuration has been removed." uninstall.log

  # global configs
  [ "global smudge" = "$(git config filter.lfs.smudge)" ]
  [ "global smudge" = "$(git config --global filter.lfs.smudge)" ]
  [ "global clean" = "$(git config filter.lfs.clean)" ]
  [ "global clean" = "$(git config --global filter.lfs.clean)" ]
  [ "global filter" = "$(git config filter.lfs.process)" ]
  [ "global filter" = "$(git config --global filter.lfs.process)" ]

  # local configs are empty
  [ "" = "$(git config --local filter.lfs.smudge)" ]
  [ "" = "$(git config --local filter.lfs.clean)" ]
  [ "" = "$(git config --local filter.lfs.process)" ]
)
end_test
