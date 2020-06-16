#!/usr/bin/env bash

. "$(dirname "$0")/testlib.sh"

begin_test "install again"
(
  set -eo pipefail

  smudge="$(git config filter.lfs.smudge)"
  clean="$(git config filter.lfs.clean)"
  filter="$(git config filter.lfs.process)"

  [ "$smudge" = "git-lfs smudge -- %f" ]
  [ "$clean" = "git-lfs clean -- %f" ]
  [ "$filter" = "git-lfs filter-process" ]

  GIT_TRACE=1 git lfs install --skip-repo 2>&1 | tee install.log

  if grep -q "--replace-all" install.log; then
    echo >&2 "fatal: unexpected git config --replace-all via 'git lfs install'"
    exit 1
  fi

  [ "$smudge" = "$(git config filter.lfs.smudge)" ]
  [ "$clean" = "$(git config filter.lfs.clean)" ]
  [ "$filter" = "$(git config filter.lfs.process)" ]
)
end_test

begin_test "install with old (non-upgradeable) settings"
(
  set -e

  git config --global filter.lfs.smudge "git-lfs smudge --something %f"
  git config --global filter.lfs.clean "git-lfs clean --something %f"

  git lfs install | tee install.log
  [ "${PIPESTATUS[0]}" = 2 ]

  grep -E "(clean|smudge)\" attribute should be" install.log
  [ `grep -c "(MISSING)" install.log` = "0" ]

  [ "git-lfs smudge --something %f" = "$(git config --global filter.lfs.smudge)" ]
  [ "git-lfs clean --something %f" = "$(git config --global filter.lfs.clean)" ]

  git lfs install --force

  [ "git-lfs smudge -- %f" = "$(git config --global filter.lfs.smudge)" ]
  [ "git-lfs clean -- %f" = "$(git config --global filter.lfs.clean)" ]
)
end_test

begin_test "install with upgradeable settings"
(
  set -e

  git config --global filter.lfs.smudge "git-lfs smudge %f"
  git config --global filter.lfs.clean "git-lfs clean %f"

  # should not need force, should upgrade this old style
  git lfs install
  [ "git-lfs smudge -- %f" = "$(git config --global filter.lfs.smudge)" ]
  [ "git-lfs clean -- %f" = "$(git config --global filter.lfs.clean)" ]
  [ "git-lfs filter-process" = "$(git config --global filter.lfs.process)" ]
)
end_test

begin_test "install updates repo hooks"
(
  set -e

  mkdir install-repo-hooks
  cd install-repo-hooks
  git init

  pre_push_hook="#!/bin/sh
command -v git-lfs >/dev/null 2>&1 || { echo >&2 \"\\nThis repository is configured for Git LFS but 'git-lfs' was not found on your path. If you no longer wish to use Git LFS, remove this hook by deleting .git/hooks/pre-push.\\n\"; exit 2; }
git lfs pre-push \"\$@\""

  post_checkout_hook="#!/bin/sh
command -v git-lfs >/dev/null 2>&1 || { echo >&2 \"\\nThis repository is configured for Git LFS but 'git-lfs' was not found on your path. If you no longer wish to use Git LFS, remove this hook by deleting .git/hooks/post-checkout.\\n\"; exit 2; }
git lfs post-checkout \"\$@\""

  post_commit_hook="#!/bin/sh
command -v git-lfs >/dev/null 2>&1 || { echo >&2 \"\\nThis repository is configured for Git LFS but 'git-lfs' was not found on your path. If you no longer wish to use Git LFS, remove this hook by deleting .git/hooks/post-commit.\\n\"; exit 2; }
git lfs post-commit \"\$@\""

  post_merge_hook="#!/bin/sh
command -v git-lfs >/dev/null 2>&1 || { echo >&2 \"\\nThis repository is configured for Git LFS but 'git-lfs' was not found on your path. If you no longer wish to use Git LFS, remove this hook by deleting .git/hooks/post-merge.\\n\"; exit 2; }
git lfs post-merge \"\$@\""

  [ "Updated git hooks.
Git LFS initialized." = "$(git lfs install)" ]
  [ "$pre_push_hook" = "$(cat .git/hooks/pre-push)" ]
  [ "$post_checkout_hook" = "$(cat .git/hooks/post-checkout)" ]
  [ "$post_commit_hook" = "$(cat .git/hooks/post-commit)" ]
  [ "$post_merge_hook" = "$(cat .git/hooks/post-merge)" ]

  # replace old hook
  # more-comprehensive hook update tests are in test-update.sh
  echo "#!/bin/sh
git lfs push --stdin \$*" > .git/hooks/pre-push
  [ "Updated git hooks.
Git LFS initialized." = "$(git lfs install)" ]
  [ "$pre_push_hook" = "$(cat .git/hooks/pre-push)" ]

  # don't replace unexpected hook
  expected="Hook already exists: pre-push

	test

To resolve this, either:
  1: run \`git lfs update --manual\` for instructions on how to merge hooks.
  2: run \`git lfs update --force\` to overwrite your hook."

  echo "test" > .git/hooks/pre-push
  echo "test" > .git/hooks/post-checkout
  echo "test" > .git/hooks/post-commit
  echo "test" > .git/hooks/post-merge
  [ "test" = "$(cat .git/hooks/pre-push)" ]
  [ "$expected" = "$(git lfs install 2>&1)" ]
  [ "test" = "$(cat .git/hooks/pre-push)" ]
  [ "test" = "$(cat .git/hooks/post-checkout)" ]
  [ "test" = "$(cat .git/hooks/post-commit)" ]
  [ "test" = "$(cat .git/hooks/post-merge)" ]

  # Make sure returns non-zero
  set +e
  git lfs install
  if [ $? -eq 0 ]
  then
    exit 1
  fi
  set -e

  # force replace unexpected hook
  [ "Updated git hooks.
Git LFS initialized." = "$(git lfs install --force)" ]
  [ "$pre_push_hook" = "$(cat .git/hooks/pre-push)" ]
  [ "$post_checkout_hook" = "$(cat .git/hooks/post-checkout)" ]
  [ "$post_commit_hook" = "$(cat .git/hooks/post-commit)" ]
  [ "$post_merge_hook" = "$(cat .git/hooks/post-merge)" ]

  has_test_dir || exit 0

  echo "test with bare repository"
  cd ..
  git clone --mirror install-repo-hooks bare-install-repo-hooks
  cd bare-install-repo-hooks
  git lfs env
  git lfs install
  ls -al hooks
  [ "$pre_push_hook" = "$(cat hooks/pre-push)" ]
)
end_test

begin_test "install outside repository directory"
(
  set -e
  if [ -d "hooks" ]; then
    ls -al
    echo "hooks dir exists"
    exit 1
  fi

  git lfs install > check.log 2>&1

  if [ -d "hooks" ]; then
    ls -al
    echo "hooks dir exists"
    exit 1
  fi

  cat check.log

  # doesn't print this because being in a git repo is not necessary for install
  [ "$(grep -c "Not in a git repository" check.log)" = "0" ]
  [ "$(grep -c "Error" check.log)" = "0" ]
)
end_test

begin_test "install --skip-smudge"
(
  set -e

  mkdir install-skip-smudge-test
  cd install-skip-smudge-test

  git lfs install
  [ "git-lfs clean -- %f" = "$(git config --global filter.lfs.clean)" ]
  [ "git-lfs smudge -- %f" = "$(git config --global filter.lfs.smudge)" ]
  [ "git-lfs filter-process" = "$(git config --global filter.lfs.process)" ]

  git lfs install --skip-smudge
  [ "git-lfs clean -- %f" = "$(git config --global filter.lfs.clean)" ]
  [ "git-lfs smudge --skip -- %f" = "$(git config --global filter.lfs.smudge)" ]
  [ "git-lfs filter-process --skip" = "$(git config --global filter.lfs.process)" ]

  git lfs install
  [ "git-lfs clean -- %f" = "$(git config --global filter.lfs.clean)" ]
  [ "git-lfs smudge -- %f" = "$(git config --global filter.lfs.smudge)" ]
  [ "git-lfs filter-process" = "$(git config --global filter.lfs.process)" ]

  [ ! -e "lfs" ]
)
end_test

begin_test "install --local"
(
  set -e

  # old values that should be ignored by `install --local`
  git config --global filter.lfs.smudge "global smudge"
  git config --global filter.lfs.clean "global clean"
  git config --global filter.lfs.process "global filter"

  mkdir install-local-repo
  cd install-local-repo
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
)
end_test

begin_test "install --local with failed permissions"
(
  set -e

  # Windows lacks POSIX permissions.
  [ "$IS_WINDOWS" -eq 1 ] && exit 0

  # Root is exempt from permissions.
  [ "$(id -u)" -eq 0 ] && exit 0

  mkdir install-local-repo-perms
  cd install-local-repo-perms
  git init

  # Make it impossible to write a new .git/config file so we can't write config
  # options.
  chmod 500 .git

  res=0
  git lfs install --local >out.log || res=$?

  # Cleanup fails without this.
  chmod 700 .git

  cat out.log
  grep -E "error running.*git.*config" out.log
  [ "$res" -eq 2 ]
)
end_test

begin_test "install --local outside repository"
(
  set -e

  # If run inside the git-lfs source dir this will update its .git/config & cause issues
  if [ "$GIT_LFS_TEST_DIR" == "" ]; then
    echo "Skipping install --local because GIT_LFS_TEST_DIR is not set"
    exit 0
  fi

  has_test_dir || exit 0

  set +e
  git lfs install --local >out.log
  res=$?
  set -e

  [ "Not in a git repository." = "$(cat out.log)" ]
  [ "0" != "$res" ]
)
end_test

begin_test "install --local with conflicting scope"
(
  set -e

  reponame="$(basename "$0" ".sh")-scope-conflict"
  mkdir "$reponame"
  cd "$reponame"
  git init

  set +e
  git lfs install --local --system 2>err.log
  res=$?
  set -e

  [ "Only one of --local and --system options can be specified." = "$(cat err.log)" ]
  [ "0" != "$res" ]
)
end_test

begin_test "install in directory without access to .git/lfs"
(
  set -e
  mkdir not-a-repo
  cd not-a-repo
  mkdir .git
  touch .git/lfs
  touch lfs

  git config --global filter.lfs.clean whatevs
  [ "whatevs" = "$(git config filter.lfs.clean)" ]

  git lfs install --force

  [ "git-lfs clean -- %f" = "$(git config filter.lfs.clean)" ]
)
end_test

begin_test "install in repo without changing hooks"
(
  set -e
  git init non-lfs-repo
  cd non-lfs-repo

  git lfs install --skip-repo

  # should not install hooks
  [ ! -f .git/hooks/pre-push ]
  [ ! -f .git/hooks/post-checkout ]
  [ ! -f .git/hooks/post-merge ]
  [ ! -f .git/hooks/post-commit ]

  # filters should still be installed
  [ "git-lfs clean -- %f" = "$(git config filter.lfs.clean)" ]
  [ "git-lfs smudge -- %f" = "$(git config filter.lfs.smudge)" ]
  [ "git-lfs filter-process" = "$(git config filter.lfs.process)" ]
)
end_test

begin_test "can install when multiple global values registered"
(
  set -e

  git config --global filter.lfs.smudge "git-lfs smudge --something %f"
  git config --global --add filter.lfs.smudge "git-lfs smudge --something-else %f"

  git lfs install --force
)
end_test
