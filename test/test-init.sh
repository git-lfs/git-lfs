#!/bin/sh

. "test/testlib.sh"

begin_test "init again"
(
  set -e

  tmphome="$(basename "$0" ".sh")"
  mkdir -p $tmphome
  cp $HOME/.gitconfig $tmphome/
  HOME=$PWD/$tmphome
  cd $HOME

  [ "git-lfs smudge %f" = "$(git config filter.lfs.smudge)" ]
  [ "git-lfs clean %f" = "$(git config filter.lfs.clean)" ]

  git lfs init

  [ "git-lfs smudge %f" = "$(git config filter.lfs.smudge)" ]
  [ "git-lfs clean %f" = "$(git config filter.lfs.clean)" ]
)
end_test

begin_test "init with old settings"
(
  set -e

  tmphome="$(basename "$0" ".sh")"
  mkdir -p $tmphome
  HOME=$PWD/$tmphome
  cd $HOME

  git config --global filter.lfs.smudge "git lfs smudge %f"
  git config --global filter.lfs.clean "git lfs clean %f"

  git lfs init 2> init.log

  grep "clean filter should be" init.log

  [ "git lfs smudge %f" = "$(git config filter.lfs.smudge)" ]
  [ "git lfs clean %f" = "$(git config filter.lfs.clean)" ]
)
end_test
