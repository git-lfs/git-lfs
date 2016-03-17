#!/usr/bin/env bash

. "test/testlib.sh"

begin_test "uninstall outside repository"
(
  set -e

  smudge="$(git config filter.lfs.smudge)"
  clean="$(git config filter.lfs.clean)"

  printf "$smudge" | grep "git-lfs smudge"
  printf "$clean" | grep "git-lfs clean"

  # uninstall multiple times to trigger https://github.com/github/git-lfs/issues/529
  git lfs uninstall
  git lfs install
  git lfs uninstall | tee uninstall.log
  grep "configuration has been removed" uninstall.log

  [ "" = "$(git config --global filter.lfs.smudge)" ]
  [ "" = "$(git config --global filter.lfs.clean)" ]

  cat $HOME/.gitconfig
  [ "$(grep 'filter "lfs"' $HOME/.gitconfig -c)" = "0" ]
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

  git lfs uninstall

  [ -f .git/hooks/pre-push ] && {
    echo "expected .git/hooks/pre-push to be deleted"
    exit 1
  }
  [ "" = "$(git config filter.lfs.smudge)" ]
  [ "" = "$(git config filter.lfs.clean)" ]
)
end_test

begin_test "uninstall inside repository without git lfs pre-push hook"
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

  git lfs uninstall

  [ -f .git/hooks/pre-push ]
  [ "" = "$(git config filter.lfs.smudge)" ]
  [ "" = "$(git config filter.lfs.clean)" ]
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

  git lfs uninstall hooks

  [ -f .git/hooks/pre-push ] && {
    echo "expected .git/hooks/pre-push to be deleted"
    exit 1
  }

  [ "git-lfs smudge -- %f" = "$(git config filter.lfs.smudge)" ]
  [ "git-lfs clean -- %f" = "$(git config filter.lfs.clean)" ]
)
end_test
