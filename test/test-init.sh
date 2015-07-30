#!/usr/bin/env bash

. "test/testlib.sh"

begin_test "init again"
(
  set -e

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

  git config --global filter.lfs.smudge "git lfs smudge %f"
  git config --global filter.lfs.clean "git lfs clean %f"

  set +e
  git lfs init 2> init.log
  res=$?
  set -e

  [ "$res" = 2 ]

  grep "clean filter should be" init.log
  [ `grep -c "(MISSING)" init.log` = "0" ]

  [ "git lfs smudge %f" = "$(git config filter.lfs.smudge)" ]
  [ "git lfs clean %f" = "$(git config filter.lfs.clean)" ]

  git lfs init --force
  [ "git-lfs smudge %f" = "$(git config filter.lfs.smudge)" ]
  [ "git-lfs clean %f" = "$(git config filter.lfs.clean)" ]
)
end_test
