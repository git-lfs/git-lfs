#!/usr/bin/env bash

. "test/testlib.sh"

begin_test "init again"
(
  set -e

  smudge="$(git config filter.lfs.smudge)"
  clean="$(git config filter.lfs.clean)"

  printf "$smudge" | grep "git-lfs smudge"
  printf "$clean" | grep "git-lfs clean"

  git lfs init

  [ "$smudge" = "$(git config filter.lfs.smudge)" ]
  [ "$clean" = "$(git config filter.lfs.clean)" ]
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

  cat init.log
  grep -E "(clean|smudge) attribute should be" init.log
  [ `grep -c "(MISSING)" init.log` = "0" ]

  [ "git lfs smudge %f" = "$(git config --global filter.lfs.smudge)" ]
  [ "git lfs clean %f" = "$(git config --global filter.lfs.clean)" ]

  git lfs init --force
  [ "git-lfs smudge %f" = "$(git config --global filter.lfs.smudge)" ]
  [ "git-lfs clean %f" = "$(git config --global filter.lfs.clean)" ]
)
end_test

begin_test "init updates repo hooks"
(
  set -e

  mkdir init-repo-hooks
  cd init-repo-hooks
  git init

  pre_push_hook="#!/bin/sh
command -v git-lfs >/dev/null 2>&1 || { echo >&2 \"\\nThis repository is configured for Git LFS but 'git-lfs' was not found on your path. If you no longer wish to use Git LFS, remove this hook by deleting .git/hooks/pre-push.\\n\"; exit 2; }
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

  has_test_dir || exit 0

  echo "test with bare repository"
  cd ..
  git clone --mirror init-repo-hooks bare-init-repo-hooks
  cd bare-init-repo-hooks
  git lfs env
  git lfs init
  ls -al hooks
  [ "$pre_push_hook" = "$(cat hooks/pre-push)" ]
)
end_test

begin_test "init outside repository directory"
(
  set -e
  if [ -d "hooks" ]; then
    ls -al
    echo "hooks dir exists"
    exit 1
  fi

  git lfs init 2>&1 > check.log

  if [ -d "hooks" ]; then
    ls -al
    echo "hooks dir exists"
    exit 1
  fi

  cat check.log

  # doesn't print this because being in a git repo is not necessary for init
  [ "$(grep -c "Not in a git repository" check.log)" = "0" ]
)
end_test

begin_test "init --skip-smudge"
(
  set -e

  git lfs init
  [ "git-lfs clean %f" = "$(git config --global filter.lfs.clean)" ]
  [ "git-lfs smudge %f" = "$(git config --global filter.lfs.smudge)" ]

  git lfs init --skip-smudge
  [ "git-lfs clean %f" = "$(git config --global filter.lfs.clean)" ]
  [ "git-lfs smudge --skip %f" = "$(git config --global filter.lfs.smudge)" ]

  git lfs init --force
  [ "git-lfs clean %f" = "$(git config --global filter.lfs.clean)" ]
  [ "git-lfs smudge %f" = "$(git config --global filter.lfs.smudge)" ]
)
end_test

begin_test "init --local"
(
  set -e

  # old values that should be ignored by `init --local`
  git config --global filter.lfs.smudge "git lfs smudge %f"
  git config --global filter.lfs.clean "git lfs clean %f"

  mkdir init-local-repo
  cd init-local-repo
  git init
  git lfs init --local

  [ "git-lfs clean %f" = "$(git config filter.lfs.clean)" ]
  [ "git-lfs clean %f" = "$(git config --local filter.lfs.clean)" ]
  [ "git lfs clean %f" = "$(git config --global filter.lfs.clean)" ]
)
end_test

begin_test "init --local outside repository"
(
  # If run inside the git-lfs source dir this will update its .git/config & cause issues
  if [ "$GIT_LFS_TEST_DIR" == "" ]; then
    echo "Skipping init --local because GIT_LFS_TEST_DIR is not set"
    exit 0
  fi

  set +e

  has_test_dir || exit 0

  git lfs init --local 2> err.log
  res=$?

  [ "Not in a git repository." = "$(cat err.log)" ]
  [ "0" != "$res" ]
)
end_test
