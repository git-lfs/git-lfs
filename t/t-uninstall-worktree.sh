#!/usr/bin/env bash

. "$(dirname "$0")/testlib.sh"

# These tests rely on behavior found in Git versions higher than 2.20.0 to
# perform themselves, specifically:
#   - worktreeConfig extension support
ensure_git_version_isnt $VERSION_LOWER "2.20.0"

begin_test "uninstall --worktree outside repository"
(
  set -e

  # If run inside the git-lfs source dir this will update its .git/config & cause issues
  if [ "$GIT_LFS_TEST_DIR" == "" ]; then
    echo "Skipping uninstall --worktree because GIT_LFS_TEST_DIR is not set"
    exit 0
  fi

  has_test_dir || exit 0

  set +e
  git lfs uninstall --worktree >out.log
  res=$?
  set -e

  [ "Not in a git repository." = "$(cat out.log)" ]
  [ "0" != "$res" ]
)
end_test

begin_test "uninstall --worktree with single working tree"
(
  set -e

  # old values that should be ignored by `uninstall --worktree`
  git config --global filter.lfs.smudge "global smudge"
  git config --global filter.lfs.clean "global clean"
  git config --global filter.lfs.process "global filter"

  reponame="$(basename "$0" ".sh")-single-tree"
  mkdir "$reponame"
  cd "$reponame"
  git init
  git lfs install --worktree

  # local configs are correct
  [ "git-lfs smudge -- %f" = "$(git config filter.lfs.smudge)" ]
  [ "git-lfs smudge -- %f" = "$(git config --local filter.lfs.smudge)" ]
  [ "git-lfs smudge -- %f" = "$(git config --worktree filter.lfs.smudge)" ]
  [ "git-lfs clean -- %f" = "$(git config filter.lfs.clean)" ]
  [ "git-lfs clean -- %f" = "$(git config --local filter.lfs.clean)" ]
  [ "git-lfs clean -- %f" = "$(git config --worktree filter.lfs.clean)" ]
  [ "git-lfs filter-process" = "$(git config filter.lfs.process)" ]
  [ "git-lfs filter-process" = "$(git config --local filter.lfs.process)" ]
  [ "git-lfs filter-process" = "$(git config --worktree filter.lfs.process)" ]

  # global configs
  [ "global smudge" = "$(git config --global filter.lfs.smudge)" ]
  [ "global clean" = "$(git config --global filter.lfs.clean)" ]
  [ "global filter" = "$(git config --global filter.lfs.process)" ]

  git lfs uninstall --worktree 2>&1 | tee uninstall.log
  if [ ${PIPESTATUS[0]} -ne 0 ]; then
    echo >&2 "fatal: expected 'git lfs uninstall --worktree' to succeed"
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
  [ "" = "$(git config --worktree filter.lfs.smudge)" ]
  [ "" = "$(git config --local filter.lfs.clean)" ]
  [ "" = "$(git config --worktree filter.lfs.clean)" ]
  [ "" = "$(git config --local filter.lfs.process)" ]
  [ "" = "$(git config --worktree filter.lfs.process)" ]
)
end_test

begin_test "uninstall --worktree with multiple working trees"
(
  set -e

  reponame="$(basename "$0" ".sh")-multi-tree"
  mkdir "$reponame"
  cd "$reponame"
  git init

  # old values that should be ignored by `uninstall --worktree`
  git config --global filter.lfs.smudge "global smudge"
  git config --global filter.lfs.clean "global clean"
  git config --global filter.lfs.process "global filter"
  git config --local filter.lfs.smudge "local smudge"
  git config --local filter.lfs.clean "local clean"
  git config --local filter.lfs.process "local filter"

  touch a.txt
  git add a.txt
  git commit -m "initial commit"

  git config core.repositoryformatversion 1
  git config extensions.worktreeConfig true

  treename="../$reponame-wt"
  git worktree add "$treename"
  cd "$treename"

  git lfs install --worktree

  # worktree configs are correct
  [ "git-lfs smudge -- %f" = "$(git config filter.lfs.smudge)" ]
  [ "git-lfs smudge -- %f" = "$(git config --worktree filter.lfs.smudge)" ]
  [ "git-lfs clean -- %f" = "$(git config filter.lfs.clean)" ]
  [ "git-lfs clean -- %f" = "$(git config --worktree filter.lfs.clean)" ]
  [ "git-lfs filter-process" = "$(git config filter.lfs.process)" ]
  [ "git-lfs filter-process" = "$(git config --worktree filter.lfs.process)" ]

  # local configs are correct
  [ "local smudge" = "$(git config --local filter.lfs.smudge)" ]
  [ "local clean" = "$(git config --local filter.lfs.clean)" ]
  [ "local filter" = "$(git config --local filter.lfs.process)" ]

  # global configs
  [ "global smudge" = "$(git config --global filter.lfs.smudge)" ]
  [ "global clean" = "$(git config --global filter.lfs.clean)" ]
  [ "global filter" = "$(git config --global filter.lfs.process)" ]

  git lfs uninstall --worktree 2>&1 | tee uninstall.log
  if [ ${PIPESTATUS[0]} -ne 0 ]; then
    echo >&2 "fatal: expected 'git lfs uninstall --worktree' to succeed"
    exit 1
  fi
  grep -v "Global Git LFS configuration has been removed." uninstall.log

  # global configs
  [ "global smudge" = "$(git config --global filter.lfs.smudge)" ]
  [ "global clean" = "$(git config --global filter.lfs.clean)" ]
  [ "global filter" = "$(git config --global filter.lfs.process)" ]

  # local configs
  [ "local smudge" = "$(git config filter.lfs.smudge)" ]
  [ "local smudge" = "$(git config --local filter.lfs.smudge)" ]
  [ "local clean" = "$(git config filter.lfs.clean)" ]
  [ "local clean" = "$(git config --local filter.lfs.clean)" ]
  [ "local filter" = "$(git config filter.lfs.process)" ]
  [ "local filter" = "$(git config --local filter.lfs.process)" ]

  # worktree configs are empty
  [ "" = "$(git config --worktree filter.lfs.smudge)" ]
  [ "" = "$(git config --worktree filter.lfs.clean)" ]
  [ "" = "$(git config --worktree filter.lfs.process)" ]
)
end_test

begin_test "uninstall --worktree without worktreeConfig extension"
(
  set -e

  reponame="$(basename "$0" ".sh")-multi-tree-no-config"
  mkdir "$reponame"
  cd "$reponame"
  git init

  touch a.txt
  git add a.txt
  git commit -m "initial commit"

  treename="../$reponame-wt"
  git worktree add "$treename"
  cd "$treename"

  set +e
  git lfs uninstall --worktree >out.log
  res=$?
  set -e

  cat out.log
  grep -E "error running.*git.*config" out.log
  [ "$res" -eq 0 ]
)
end_test

begin_test "uninstall --worktree with conflicting scope"
(
  set -e

  reponame="$(basename "$0" ".sh")-scope-conflict"
  mkdir "$reponame"
  cd "$reponame"
  git init

  set +e
  git lfs uninstall --local --worktree 2>err.log
  res=$?
  set -e

  [ "Only one of --local and --worktree options can be specified." = "$(cat err.log)" ]
  [ "0" != "$res" ]

  set +e
  git lfs uninstall --worktree --system 2>err.log
  res=$?
  set -e

  [ "Only one of --worktree and --system options can be specified." = "$(cat err.log)" ]
  [ "0" != "$res" ]
)
end_test
