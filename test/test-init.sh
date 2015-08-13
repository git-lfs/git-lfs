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

begin_test "init updates repo hooks"
(
  set -e

  mkdir init-repo-hooks
  cd init-repo-hooks
  git init

  pre_push_hook="#!/bin/sh
command -v git-lfs >/dev/null 2>&1 || { echo >&2 \"\\nThis repository has been set up with Git LFS but Git LFS is not installed.\\nDelete \$0 if this repository no longer uses Git LFS.\\n\"; exit 2; }
git lfs pre-push \"\$@\""

  [ "Updated pre-push hook.
Git LFS initialized." = "$(git lfs init)" ]
  [ "$pre_push_hook" = "$(cat .git/hooks/pre-push)" ]

  # replace old hook
  # more-comprehensive hook update tests are in test-update.sh
  echo "#!/bin/sh
git lfs push --stdin \$*" > .git/hooks/pre-push
  [ "Updated pre-push hook.
Git LFS initialized." = "$(git lfs init)" ]
  [ "$pre_push_hook" = "$(cat .git/hooks/pre-push)" ]

  # don't replace unexpected hook
  expected="Hook already exists: pre-push

test

Run \`git lfs update --force\` to overwrite this hook.
Git LFS initialized."

  echo "test" > .git/hooks/pre-push
  [ "test" = "$(cat .git/hooks/pre-push)" ]
  [ "$expected" = "$(git lfs init 2>&1)" ]
  [ "test" = "$(cat .git/hooks/pre-push)" ]

  # force replace unexpected hook
  [ "Updated pre-push hook.
Git LFS initialized." = "$(git lfs init --force)" ]
  [ "$pre_push_hook" = "$(cat .git/hooks/pre-push)" ]
)
end_test
